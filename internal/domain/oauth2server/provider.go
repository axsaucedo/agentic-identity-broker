package oauth2server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/ory/fosite"
	fositeOAuth2 "github.com/ory/fosite/handler/oauth2"
	"github.com/ory/fosite/handler/pkce"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// Provider is the central wiring for the OAuth2 server mode.
// It constructs fosite handlers with custom strategies and storage adapters,
// and exposes domain-native methods for use by HTTP handlers.
type Provider struct {
	authCodeHandler *fositeOAuth2.AuthorizeExplicitGrantHandler
	ccHandler       *fositeOAuth2.ClientCredentialsGrantHandler
	pkceHandler     *pkce.Handler

	fositeStorage     *FositeStorage
	clientAuth        *ClientAuthService
	signingKeyService *SigningKeyService
	accessStrategy    *JWXAccessTokenStrategy

	config *fosite.Config
	logger *slog.Logger
}

// NewProvider constructs the OAuth2 server provider with fosite handlers.
func NewProvider(
	codeRepo ports.AuthorizationCodeRepository,
	pkceRepo ports.PKCESessionRepository,
	credRepo ports.BrokerClientCredentialRepository,
	agentRepo ports.AgentRepository,
	signingKeyRepo ports.SigningKeyRepository,
	encryption ports.EncryptionPort,
	issuerURI string,
	tokenTTL time.Duration,
	tokenClaimsExpression string,
	logger *slog.Logger,
) (*Provider, error) {
	// Compile token claims CEL expression at startup (FR-013b: fail if invalid)
	customClaimsEval, err := NewTokenClaimsEvaluator(tokenClaimsExpression)
	if err != nil {
		return nil, fmt.Errorf("invalid token_claims_expression: %w", err)
	}

	// Build services
	signingKeyService := NewSigningKeyService(signingKeyRepo, encryption, logger)
	clientAuth := NewClientAuthService(credRepo, agentRepo, logger)

	// Our strategies
	accessStrategy, err := NewJWXAccessTokenStrategy(signingKeyService, signingKeyRepo, issuerURI, tokenTTL, customClaimsEval, logger)
	if err != nil {
		return nil, fmt.Errorf("invalid access token strategy configuration: %w", err)
	}
	codeStrategy := &RandomCodeStrategy{}

	// Storage adapters
	storage := NewFositeStorage(codeRepo, pkceRepo, agentRepo, credRepo, logger)

	config := &fosite.Config{
		AuthorizeCodeLifespan:          60 * time.Second,
		AccessTokenLifespan:            tokenTTL,
		EnforcePKCE:                    true,
		EnablePKCEPlainChallengeMethod: false,
		// Empty AllowedScopes means unrestricted in our domain model.
		// fosite's default WildcardScopeStrategy treats empty as no scopes allowed.
		ScopeStrategy: func(allowedScopes []string, requestedScope string) bool {
			if len(allowedScopes) == 0 {
				return true
			}
			for _, s := range allowedScopes {
				if s == requestedScope {
					return true
				}
			}
			return false
		},
	}

	helper := &fositeOAuth2.HandleHelper{
		AccessTokenStrategy: accessStrategy,
		AccessTokenStorage:  storage,
		Config:              config,
	}

	return &Provider{
		authCodeHandler: &fositeOAuth2.AuthorizeExplicitGrantHandler{
			AccessTokenStrategy:    accessStrategy,
			AuthorizeCodeStrategy:  codeStrategy,
			CoreStorage:            storage,
			TokenRevocationStorage: storage,
			Config:                 config,
		},
		ccHandler: &fositeOAuth2.ClientCredentialsGrantHandler{
			HandleHelper: helper,
			Config:       config,
		},
		pkceHandler: &pkce.Handler{
			AuthorizeCodeStrategy: codeStrategy,
			Storage:               storage,
			Config:                config,
		},
		fositeStorage:     storage,
		clientAuth:        clientAuth,
		signingKeyService: signingKeyService,
		accessStrategy:    accessStrategy,
		config:            config,
		logger:            logger,
	}, nil
}

// ClientAuth returns the client authentication service.
func (p *Provider) ClientAuth() *ClientAuthService {
	return p.clientAuth
}

// SigningKeyService returns the signing key service.
func (p *Provider) SigningKeyService() *SigningKeyService {
	return p.signingKeyService
}

// HandleClientCredentials processes a client_credentials grant type request.
func (p *Provider) HandleClientCredentials(ctx context.Context, agentID id.AgentID, secret string, requestedScope string) (*ports.TokenResponse, error) {
	// Authenticate client
	authClient, err := p.clientAuth.Authenticate(ctx, agentID, secret)
	if err != nil {
		return nil, err
	}

	// Build fosite client
	client := &brokerClient{agent: authClient.Agent, credential: authClient.Credential}

	// Validate scopes
	scopes := splitScope(requestedScope)
	if len(authClient.Agent.AllowedScopes) > 0 && len(scopes) > 0 {
		for _, s := range scopes {
			if !contains(authClient.Agent.AllowedScopes, s) {
				return nil, fmt.Errorf("%w: scope %q not allowed for this agent", ErrInvalidScope, s)
			}
		}
	}

	// Build fosite request
	session := &fosite.DefaultSession{
		Subject: client.GetID(),
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AccessToken: time.Now().Add(p.config.AccessTokenLifespan),
		},
	}

	req := fosite.NewAccessRequest(session)
	req.Client = client
	req.GrantTypes = fosite.Arguments{"client_credentials"}
	req.RequestedScope = scopes
	req.GrantedScope = scopes

	resp := fosite.NewAccessResponse()

	// Generate token via our strategy
	token, sig, err := p.accessStrategy.GenerateAccessToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	resp.SetAccessToken(token)
	resp.SetTokenType("Bearer")
	resp.SetExtra("expires_in", int64(p.config.AccessTokenLifespan.Seconds()))
	resp.SetExtra("scope", requestedScope)
	_ = sig // Signature used for storage lookup (stateless JWT, not stored)

	return &ports.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(p.config.AccessTokenLifespan.Seconds()),
		Scope:       requestedScope,
	}, nil
}

// HandleAuthorize processes an authorization endpoint request using the agent's UUID
// as the client identifier. Resolves the agent's broker credentials internally.
func (p *Provider) HandleAuthorize(
	ctx context.Context,
	agentID id.AgentID,
	redirectURI string,
	responseType string,
	scope string,
	state string,
	codeChallenge string,
	codeChallengeMethod string,
	principal id.Principal,
) (code string, err error) {
	// Look up client by agent ID (uses GetByAgentID internally)
	fositeClient, err := p.fositeStorage.buildClient(ctx, agentID)
	if err != nil {
		var storageErr *storage.StorageError
		if errors.As(err, &storageErr) && storageErr.Kind == storage.ErrorKindNotFound {
			return "", fmt.Errorf("%w: agent %s not found", ErrUnknownClient, agentID)
		}
		return "", fmt.Errorf("%w: client lookup failed", ErrServerError)
	}

	bc, ok := fositeClient.(*brokerClient)
	if !ok {
		return "", fmt.Errorf("internal error: expected *brokerClient, got %T", fositeClient)
	}

	// Validate redirect_uri: must be registered and use HTTPS (or loopback HTTP).
	if !contains(bc.agent.RedirectURIs, redirectURI) {
		return "", fmt.Errorf("%w: %s not registered for agent", ErrInvalidRedirectURI, redirectURI)
	}
	if !isHTTPSOrLoopback(redirectURI) {
		return "", fmt.Errorf("%w: redirect_uri must use HTTPS for non-local hosts", ErrInvalidRedirectURI)
	}

	// Enforce PKCE
	if codeChallenge == "" {
		return "", fmt.Errorf("%w: code_challenge is required (PKCE mandatory)", ErrInvalidRequest)
	}
	if codeChallengeMethod != "S256" {
		return "", fmt.Errorf("%w: only S256 code_challenge_method is supported", ErrInvalidRequest)
	}

	scopes := splitScope(scope)
	if len(bc.agent.AllowedScopes) > 0 && len(scopes) > 0 {
		for _, s := range scopes {
			if !contains(bc.agent.AllowedScopes, s) {
				return "", fmt.Errorf("%w: scope %q not allowed for this agent", ErrInvalidScope, s)
			}
		}
	}

	// Build fosite authorize request — credential ClientID stored in form so
	// CreateAuthorizeCodeSession can persist it for post-rotation binding checks at exchange time.
	session := &fosite.DefaultSession{
		Subject: principal.String(),
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AuthorizeCode: time.Now().Add(60 * time.Second),
		},
	}

	parsedRedirectURI, _ := url.Parse(redirectURI) // already validated above; won't fail

	authReq := fosite.NewAuthorizeRequest()
	authReq.Client = fositeClient
	authReq.RedirectURI = parsedRedirectURI
	authReq.ResponseTypes = fosite.Arguments{responseType}
	authReq.RequestedScope = scopes
	authReq.GrantedScope = scopes
	authReq.State = state
	authReq.Session = session
	authReq.Form = url.Values{
		"redirect_uri":          {redirectURI},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {codeChallengeMethod},
		"response_type":         {responseType},
		"client_id":             {bc.credential.ClientID.String()},
		"scope":                 {scope},
		"state":                 {state},
	}

	resp := fosite.NewAuthorizeResponse()
	if err := p.authCodeHandler.HandleAuthorizeEndpointRequest(ctx, authReq, resp); err != nil {
		return "", fmt.Errorf("%w: failed to handle authorize request", ErrServerError)
	}
	if err := p.pkceHandler.HandleAuthorizeEndpointRequest(ctx, authReq, resp); err != nil {
		return "", fmt.Errorf("%w: failed to store PKCE session", ErrServerError)
	}

	return resp.GetCode(), nil
}

func (p *Provider) HandleAuthorizationCodeExchange(
	ctx context.Context,
	agentID id.AgentID,
	secret string,
	code string,
	redirectURI string,
	codeVerifier string,
) (*ports.TokenResponse, error) {
	// Authenticate client
	authedClient, err := p.clientAuth.Authenticate(ctx, agentID, secret)
	if err != nil {
		return nil, err
	}

	// Domain-level credential pre-check: fosite verifies agent-level client identity but does
	// not know about credential rotation. Peek at the stored code to enforce that codes issued
	// to a previous credential cannot be exchanged by a newly rotated one.
	authCode, err := p.fositeStorage.codeRepo.FindByCodeHash(ctx, p.authCodeHandler.AuthorizeCodeStrategy.AuthorizeCodeSignature(ctx, code))
	if err != nil {
		var storageErr *storage.StorageError
		if errors.As(err, &storageErr) && storageErr.Kind == storage.ErrorKindNotFound {
			return nil, fmt.Errorf("%w: authorization code not found", ErrInvalidGrant)
		}
		return nil, fmt.Errorf("failed to look up authorization code: %w", err)
	}
	if authedClient.Credential.ClientID != authCode.ClientID {
		return nil, fmt.Errorf("%w: code was issued to a different credential", ErrInvalidGrant)
	}

	// Build fosite access request
	session := &fosite.DefaultSession{
		Subject: authCode.Principal.String(),
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AccessToken: time.Now().Add(p.config.AccessTokenLifespan),
		},
	}
	client := &brokerClient{agent: authedClient.Agent, credential: authedClient.Credential}
	req := fosite.NewAccessRequest(session)
	req.Client = client
	req.GrantTypes = fosite.Arguments{"authorization_code"}
	req.Form = url.Values{
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
		"grant_type":    {"authorization_code"},
	}

	// Validate code: expiry, client identity match, redirect_uri
	if err := p.authCodeHandler.HandleTokenEndpointRequest(ctx, req); err != nil {
		return nil, mapFositeGrantError(err)
	}

	// Validate PKCE verifier and consume the PKCE session
	if err := p.pkceHandler.HandleTokenEndpointRequest(ctx, req); err != nil {
		return nil, mapFositeGrantError(err)
	}

	// Generate access token and mark code as used
	resp := fosite.NewAccessResponse()
	if err := p.authCodeHandler.PopulateTokenEndpointResponse(ctx, req, resp); err != nil {
		return nil, mapFositeGrantError(err)
	}
	if err := p.pkceHandler.PopulateTokenEndpointResponse(ctx, req, resp); err != nil {
		return nil, mapFositeGrantError(err)
	}

	return &ports.TokenResponse{
		AccessToken: resp.GetAccessToken(),
		TokenType:   "Bearer",
		ExpiresIn:   int64(p.config.AccessTokenLifespan.Seconds()),
		Scope:       strings.Join(req.GetGrantedScopes(), " "),
	}, nil
}

func splitScope(scope string) fosite.Arguments {
	if scope == "" {
		return fosite.Arguments{}
	}
	return strings.Split(scope, " ")
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

// isHTTPSOrLoopback reports whether uri uses HTTPS, or HTTP for loopback hosts.
// Used to enforce HTTPS at runtime on legacy redirect URIs stored before
// Agent.ValidateForCreate began rejecting non-loopback http:// URIs.
func isHTTPSOrLoopback(uriStr string) bool {
	u, err := url.Parse(uriStr)
	if err != nil || u.Host == "" {
		return false
	}
	if u.Scheme == "https" {
		return true
	}
	if u.Scheme == "http" {
		host := u.Hostname()
		return host == "localhost" || host == "127.0.0.1" || host == "::1"
	}
	return false
}

// mapFositeGrantError translates fosite token endpoint errors to domain errors.
// invalid_grant (expired, used, wrong client, PKCE failure) and not_found all become ErrInvalidGrant.
func mapFositeGrantError(err error) error {
	if errors.Is(err, fosite.ErrInvalidGrant) ||
		errors.Is(err, fosite.ErrInvalidatedAuthorizeCode) ||
		errors.Is(err, fosite.ErrNotFound) {
		return fmt.Errorf("%w: %v", ErrInvalidGrant, err)
	}
	return fmt.Errorf("%w: %v", ErrServerError, err)
}
