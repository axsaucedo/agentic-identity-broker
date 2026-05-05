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
	credRepo ports.ClientCredentialRepository,
	clientResolver ports.ClientResolver,
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
	clientAuth := NewClientAuthService(credRepo, clientResolver, logger)

	// Our strategies
	accessStrategy, err := NewJWXAccessTokenStrategy(signingKeyService, signingKeyRepo, issuerURI, tokenTTL, customClaimsEval, logger)
	if err != nil {
		return nil, fmt.Errorf("invalid access token strategy configuration: %w", err)
	}
	codeStrategy := &RandomCodeStrategy{}

	// Storage adapters
	storage := NewFositeStorage(codeRepo, pkceRepo, credRepo, clientResolver, logger)

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
// Scope validation and token generation are fully delegated to fosite's ccHandler.
func (p *Provider) HandleClientCredentials(ctx context.Context, clientID string, secret string, requestedScope string) (resp *ports.TokenResponse, err error) {
	defer func() { err = translateFositeError(err) }()

	fositeClient, err := p.fositeStorage.GetClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			return nil, fosite.ErrInvalidClient.WithHintf("client %s not found", clientID)
		}
		return nil, fosite.ErrServerError.WithDebugf("client lookup failed: %v", err)
	}

	cc, ok := fositeClient.(*confidentialClient)
	if !ok {
		return nil, fosite.ErrInvalidClient.WithHintf("client_credentials grant requires a confidential client")
	}

	authClient, err := p.clientAuth.Authenticate(ctx, cc.agent.ID, secret)
	if err != nil {
		return nil, err
	}

	client := &confidentialClient{clientID: clientID, agent: authClient.Agent, credential: authClient.Credential}
	scopes := splitScope(requestedScope)

	session := &fosite.DefaultSession{
		Subject: authClient.Agent.ID.String(),
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AccessToken: time.Now().Add(p.config.AccessTokenLifespan),
		},
	}

	req := fosite.NewAccessRequest(session)
	req.Client = client
	req.GrantTypes = fosite.Arguments{"client_credentials"}
	req.RequestedScope = scopes

	if err := p.ccHandler.HandleTokenEndpointRequest(ctx, req); err != nil {
		return nil, err
	}

	// fosite v0.49 HandleTokenEndpointRequest validates scopes but does not grant them.
	for _, scope := range scopes {
		req.GrantScope(scope)
	}

	fositeResp := fosite.NewAccessResponse()
	if err := p.ccHandler.PopulateTokenEndpointResponse(ctx, req, fositeResp); err != nil {
		return nil, err
	}

	return &ports.TokenResponse{
		AccessToken: fositeResp.GetAccessToken(),
		TokenType:   "Bearer",
		ExpiresIn:   int64(p.config.AccessTokenLifespan.Seconds()),
		Scope:       strings.Join(req.GetGrantedScopes(), " "),
	}, nil
}

// HandleAuthorize processes an authorization endpoint request using the agent's UUID
// as the client identifier. Resolves the agent's broker credentials internally.
//
// PKCE enforcement (EnforcePKCE:true, S256-only) is handled by pkceHandler.
// Scope validation is handled by the configured ScopeStrategy.
// Returns ErrInvalidRedirectURI when the redirect_uri has not been validated so HTTP
// handlers can send a direct JSON response per RFC 6749 §4.1.2.1.
func (p *Provider) HandleAuthorize(
	ctx context.Context,
	clientID string,
	redirectURI string,
	responseType string,
	scope string,
	state string,
	codeChallenge string,
	codeChallengeMethod string,
	principal id.Principal,
) (code string, err error) {
	defer func() { err = translateFositeError(err) }()

	fositeClient, err := p.fositeStorage.GetClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			return "", fosite.ErrInvalidClient.WithHintf("client %s not found", clientID)
		}
		return "", fosite.ErrServerError.WithDebugf("client lookup failed: %v", err)
	}

	h, ok := fositeClient.(agentHolder)
	if !ok {
		return "", fosite.ErrServerError.WithDebugf("internal: expected broker client, got %T", fositeClient)
	}
	agent := h.getAgent()

	// Validate redirect_uri: must be registered and use HTTPS (or loopback HTTP).
	// GetRedirectURIs() returns the agent's registered URIs for confidential clients,
	// and the CIMD document's redirect_uris for public (CIMD) clients.
	if !contains(fositeClient.GetRedirectURIs(), redirectURI) {
		return "", fmt.Errorf("%w: %w", ErrInvalidRedirectURI,
			fosite.ErrInvalidRequest.WithHintf("redirect_uri %q is not registered for this client", redirectURI))
	}
	if !storage.IsValidRedirectURI(redirectURI) {
		return "", fmt.Errorf("%w: %w", ErrInvalidRedirectURI,
			fosite.ErrInvalidRequest.WithHintf("redirect_uri must use HTTPS for non-loopback hosts"))
	}

	if responseType != "code" {
		return "", fosite.ErrUnsupportedResponseType.WithHintf("only 'code' response_type is supported")
	}

	scopes := splitScope(scope)
	if len(agent.AllowedScopes) > 0 && len(scopes) > 0 {
		for _, s := range scopes {
			if !contains(agent.AllowedScopes, s) {
				return "", fosite.ErrInvalidScope.WithHintf("scope %q is not allowed for this client", s)
			}
		}
	}

	session := &fosite.DefaultSession{
		Subject: principal.String(),
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AuthorizeCode: time.Now().Add(60 * time.Second),
		},
	}

	parsedRedirectURI, _ := url.Parse(redirectURI)

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
		"client_id":             {clientID},
		"scope":                 {scope},
		"state":                 {state},
	}

	resp := fosite.NewAuthorizeResponse()
	if err := p.authCodeHandler.HandleAuthorizeEndpointRequest(ctx, authReq, resp); err != nil {
		return "", err
	}
	if err := p.pkceHandler.HandleAuthorizeEndpointRequest(ctx, authReq, resp); err != nil {
		return "", err
	}

	return resp.GetCode(), nil
}

func (p *Provider) HandleAuthorizationCodeExchange(
	ctx context.Context,
	clientID string,
	secret string,
	code string,
	redirectURI string,
	codeVerifier string,
) (tokenResp *ports.TokenResponse, err error) {
	defer func() { err = translateFositeError(err) }()

	fositeClient, err := p.fositeStorage.GetClient(ctx, clientID)
	if err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			return nil, fosite.ErrInvalidClient.WithHintf("client %s not found", clientID)
		}
		return nil, fosite.ErrServerError.WithDebugf("client lookup failed: %v", err)
	}

	// Domain-level credential pre-check: peek at stored code to enforce binding.
	authCode, err := p.fositeStorage.codeRepo.FindByCodeHash(ctx, p.authCodeHandler.AuthorizeCodeStrategy.AuthorizeCodeSignature(ctx, code))
	if err != nil {
		if isStorageNotFound(err) {
			return nil, fosite.ErrInvalidGrant.WithHintf("authorization code not found")
		}
		return nil, fosite.ErrServerError.WithDebugf("failed to look up authorization code: %v", err)
	}

	switch bc := fositeClient.(type) {
	case *confidentialClient:
		authedClient, err := p.clientAuth.Authenticate(ctx, bc.agent.ID, secret)
		if err != nil {
			return nil, err
		}
		if authedClient.Credential.AgentID != authCode.AgentID {
			return nil, fosite.ErrInvalidGrant.WithHintf("authorization code was issued to a different client credential")
		}
		// Rebuild client with authenticated credential
		fositeClient = &confidentialClient{clientID: clientID, agent: authedClient.Agent, credential: authedClient.Credential}
	case *publicClient:
		if bc.agent.ID != authCode.AgentID {
			return nil, fosite.ErrInvalidGrant.WithHintf("authorization code was issued to a different client")
		}
	default:
		return nil, fosite.ErrServerError.WithDebugf("unexpected client type %T", fositeClient)
	}

	session := &fosite.DefaultSession{
		Subject: authCode.Principal.String(),
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AccessToken: time.Now().Add(p.config.AccessTokenLifespan),
		},
	}
	req := fosite.NewAccessRequest(session)
	req.Client = fositeClient
	req.GrantTypes = fosite.Arguments{"authorization_code"}
	req.Form = url.Values{
		"code":          {code},
		"redirect_uri":  {redirectURI},
		"code_verifier": {codeVerifier},
		"grant_type":    {"authorization_code"},
	}

	if err := p.authCodeHandler.HandleTokenEndpointRequest(ctx, req); err != nil {
		return nil, err
	}
	if err := p.pkceHandler.HandleTokenEndpointRequest(ctx, req); err != nil {
		return nil, err
	}

	fositeResp := fosite.NewAccessResponse()
	if err := p.authCodeHandler.PopulateTokenEndpointResponse(ctx, req, fositeResp); err != nil {
		return nil, err
	}
	if err := p.pkceHandler.PopulateTokenEndpointResponse(ctx, req, fositeResp); err != nil {
		return nil, err
	}

	return &ports.TokenResponse{
		AccessToken: fositeResp.GetAccessToken(),
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

// isStorageNotFound returns true when err is a domain storage "not found" error.
func isStorageNotFound(err error) bool {
	var se *storage.StorageError
	return errors.As(err, &se) && se.Kind == storage.ErrorKindNotFound
}
