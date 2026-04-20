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
	storage := NewFositeStorage(codeRepo, agentRepo, credRepo, logger)

	config := &fosite.Config{
		AuthorizeCodeLifespan:          60 * time.Second,
		AccessTokenLifespan:            tokenTTL,
		EnforcePKCE:                    true,
		EnablePKCEPlainChallengeMethod: false,
	}

	helper := &fositeOAuth2.HandleHelper{
		AccessTokenStrategy: accessStrategy,
		AccessTokenStorage:  storage,
		Config:              config,
	}

	return &Provider{
		authCodeHandler: &fositeOAuth2.AuthorizeExplicitGrantHandler{
			AccessTokenStrategy:   accessStrategy,
			AuthorizeCodeStrategy: codeStrategy,
			CoreStorage:           storage,
			Config:                config,
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

	// Build fosite authorize request — use client_id as the client_id in the code record
	session := &fosite.DefaultSession{
		Subject: principal.String(),
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AuthorizeCode: time.Now().Add(60 * time.Second),
		},
	}

	authReq := fosite.NewAuthorizeRequest()
	authReq.Client = fositeClient
	authReq.ResponseTypes = fosite.Arguments{responseType}
	authReq.RequestedScope = scopes
	authReq.GrantedScope = scopes
	authReq.State = state
	authReq.Session = session
	authReq.Form = map[string][]string{
		"redirect_uri":          {redirectURI},
		"code_challenge":        {codeChallenge},
		"code_challenge_method": {"S256"},
		"response_type":         {responseType},
		"client_id":             {bc.credential.ClientID.String()},
		"scope":                 {scope},
		"state":                 {state},
	}

	// Generate authorization code
	authCode, _, err := (&RandomCodeStrategy{}).GenerateAuthorizeCode(ctx, authReq)
	if err != nil {
		return "", fmt.Errorf("failed to generate authorization code: %w", err)
	}

	// Store the code session
	if err := p.fositeStorage.CreateAuthorizeCodeSession(ctx, authCode, authReq); err != nil {
		return "", fmt.Errorf("failed to store authorization code: %w", err)
	}

	return authCode, nil
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

	// Get the code session
	codeHash := sha256Hex(code)
	authCode, err := p.fositeStorage.codeRepo.FindByCodeHash(ctx, codeHash)
	if err != nil {
		return nil, fmt.Errorf("%w: authorization code not found", ErrInvalidGrant)
	}

	// Verify the authenticated client is the one the code was issued to (both agent and credential).
	// The client_id check catches post-rotation redemption: a new credential should not
	// be able to exchange codes issued to a previous credential for the same agent.
	if authedClient.Agent.ID != authCode.AgentID {
		return nil, fmt.Errorf("%w: code was not issued to this client", ErrInvalidGrant)
	}
	if authedClient.Credential.ClientID != authCode.ClientID {
		return nil, fmt.Errorf("%w: code was issued to a different credential", ErrInvalidGrant)
	}

	// Check code is not used
	if authCode.UsedAt != nil {
		return nil, fmt.Errorf("%w: authorization code already used", ErrInvalidGrant)
	}

	// Check code is not expired
	if time.Now().After(authCode.ExpiresAt) {
		return nil, fmt.Errorf("%w: authorization code expired", ErrInvalidGrant)
	}

	// Validate redirect_uri matches
	if authCode.RedirectURI != redirectURI {
		return nil, fmt.Errorf("%w: redirect_uri mismatch", ErrInvalidGrant)
	}

	// Validate PKCE
	if err := verifyPKCE(authCode.CodeChallenge, codeVerifier); err != nil {
		return nil, fmt.Errorf("%w: PKCE verification failed: %v", ErrInvalidGrant, err)
	}

	// Mark code as used — the UPDATE is conditional on used_at IS NULL, so zero rows
	// affected means a concurrent request won the race and already consumed the code.
	if err := p.fositeStorage.codeRepo.MarkUsed(ctx, authCode.ID); err != nil {
		var storageErr *storage.StorageError
		if errors.As(err, &storageErr) && storageErr.Kind == storage.ErrorKindNotFound {
			return nil, fmt.Errorf("%w: authorization code already used", ErrInvalidGrant)
		}
		return nil, fmt.Errorf("failed to mark code as used: %w", err)
	}

	// Build a fosite request for token generation
	client, err := p.fositeStorage.buildClient(ctx, authCode.AgentID)
	if err != nil {
		return nil, fmt.Errorf("failed to look up client: %w", err)
	}

	session := &fosite.DefaultSession{
		Subject: authCode.Principal.String(),
		ExpiresAt: map[fosite.TokenType]time.Time{
			fosite.AccessToken: time.Now().Add(p.config.AccessTokenLifespan),
		},
	}

	scopes := splitScope(authCode.Scope)
	req := fosite.NewAccessRequest(session)
	req.Client = client
	req.GrantTypes = fosite.Arguments{"authorization_code"}
	req.RequestedScope = scopes
	req.GrantedScope = scopes

	// Generate token
	token, _, err := p.accessStrategy.GenerateAccessToken(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	return &ports.TokenResponse{
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(p.config.AccessTokenLifespan.Seconds()),
		Scope:       authCode.Scope,
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
