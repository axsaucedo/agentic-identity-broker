// Package app provides application-layer wiring and orchestration.
package app

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"

	awsencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/aws"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/admin"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/oauth2_sessions"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/jwks"
	jwtauthadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/jwtauth"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	consentservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	domjwtauth "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwtauth"
	oauth2service "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2session"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	tokenexchange "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// App represents a fully-wired application with all services and handlers initialized.
// This is returned by Builder.Build() after dependency injection.
type App struct {
	// Configuration
	Config *ports.Config

	// Repositories
	Storage          *storage.Adapter
	BranchKeyManager ports.BranchKeyManager

	// Domain services
	ConsentService       *consentservice.Service
	ProviderService      *thirdparty.ThirdpartyOAuth2ProviderService
	OAuth2SessionService *oauth2session.OAuth2SessionService
	OAuth2Service        ports.OAuth2Service
	TokenExchangeService *tokenexchange.TokenExchangeService

	// JWT pre-authentication (optional, nil when not configured)
	JWTAuthenticator domjwtauth.JWTAuthenticator

	// Handler groups for routing
	AdminHandlers   *AdminHandlers
	EnduserHandlers *EnduserHandlers

	// Logger
	Logger *slog.Logger
}

// Builder is a chainable builder for constructing App instances.
// Following the builder pattern for flexible configuration and clear intent.
//
// Example usage:
//
//	app, err := NewBuilder().
//		WithConfig(cfg).
//		WithStorage(storageAdapter).
//		WithLogger(logger).
//		Build()
type Builder struct {
	config                 *ports.Config
	storage                *storage.Adapter
	logger                 *slog.Logger
	staticWebResourcesPath string
}

// NewBuilder creates a new application builder.
func NewBuilder() *Builder {
	return &Builder{
		staticWebResourcesPath: "web/dist",
	}
}

// WithConfig sets the application configuration for the builder.
func (b *Builder) WithConfig(cfg *ports.Config) *Builder {
	b.config = cfg
	return b
}

// WithStorage sets the storage adapter for the builder.
func (b *Builder) WithStorage(storage *storage.Adapter) *Builder {
	b.storage = storage
	return b
}

// WithLogger sets the logger for the builder.
func (b *Builder) WithLogger(logger *slog.Logger) *Builder {
	b.logger = logger
	return b
}

func (b *Builder) WithStaticWebResourcesPath(path string) *Builder {
	b.staticWebResourcesPath = path
	return b
}

// Build constructs the App with all wired dependencies.
// Returns error if required dependencies are missing or initialization fails.
//
// Dependencies are created in this order:
// 1. Validate inputs (Config, Storage, Logger required)
// 2. Create domain services (ConsentService, OAuth2Service, OAuth2SessionService)
// 3. Create handler instances (AdminHandlers, EnduserHandlers)
// 4. Return fully-wired App
func (b *Builder) Build() (*App, error) {
	// Validate required dependencies
	if b.config == nil {
		return nil, fmt.Errorf("configuration is required")
	}
	if b.storage == nil {
		return nil, fmt.Errorf("storage adapter is required")
	}
	if b.logger == nil {
		return nil, fmt.Errorf("logger is required")
	}

	// Validate OAuth2AuthServer config when present. Mirrors internal/config/validator.go
	// validateOAuth2AuthServerConfig — ensures multi_agent_client required fields are
	// enforced regardless of startup path (cmd or test bootstrap).
	if b.config.OAuth2AuthServer.UpstreamAuthorizeEndpoint != "" ||
		b.config.OAuth2AuthServer.UpstreamTokenEndpoint != "" ||
		b.config.OAuth2AuthServer.UpstreamIssuerURI != "" {
		if err := b.config.OAuth2AuthServer.Validate(); err != nil {
			return nil, fmt.Errorf("oauth2_authorization_server configuration invalid: %w", err)
		}
	}

	app := &App{
		Config:  b.config,
		Storage: b.storage,
		Logger:  b.logger,
	}

	// Phase 1: Initialize encryption adapter (must happen before domain services).
	// Encryption is mandatory — no fallback. Config must specify memory or aws_kms backend.
	// Constitution Principle VII: Configuration-Driven Design.
	if b.config.Encryption.AWSKMS == nil && b.config.Encryption.Memory == nil {
		return nil, fmt.Errorf("encryption configuration required: set encryption.memory.raw_key or encryption.aws_kms in configuration (no fallback)")
	}

	encryptor, branchKeyManager, err := awsencryption.NewEncryptionAdapter(&b.config.Encryption)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize encryption adapter: %w", err)
	}

	if b.config.Encryption.AWSKMS != nil {
		b.logger.Info("AWS KMS encryption adapter initialized",
			"dynamodb_table", b.config.Encryption.AWSKMS.DynamoDBTableName,
			"dynamodb_region", b.config.Encryption.AWSKMS.DynamoDBRegion,
			"branch_key_ttl", b.config.Encryption.AWSKMS.BranchKeyTTL,
			"dynamodb_read_timeout", b.config.Encryption.AWSKMS.DynamoDBReadTimeout,
			"dynamodb_write_timeout", b.config.Encryption.AWSKMS.DynamoDBWriteTimeout,
			"branch_key_manager_wired", branchKeyManager != nil)
	} else {
		b.logger.Info("Memory encryption adapter initialized",
			"branch_key_manager_wired", branchKeyManager != nil)
	}

	if branchKeyManager != nil {
		app.BranchKeyManager = branchKeyManager
	}

	// Phase 2: Create domain services
	// Constitution Principle VI: domain depends on ports (repository interfaces), not adapters

	// Create ThirdpartyOAuth2ProviderService (handles encryption, decryption, and branch key provisioning).
	// This consolidated domain service replaces the previous ServiceManager + AuthProvider split.
	// encryptor is guaranteed to be initialized from Phase 1.
	if b.storage.Services() != nil {
		app.ProviderService = thirdparty.NewThirdpartyOAuth2ProviderService(
			b.storage.Services(),
			encryptor,
			branchKeyManager, // May be nil if memory backend (no branch key store)
			b.config.Security.SkipThirdpartyHTTPSValidation,
			b.logger,
		)
	}

	// Create consent service if repositories available.
	// ConsentService depends on ProviderService (not raw repository) so all service access
	// goes through the domain service layer including encryption/decryption.
	if b.storage.Agents() != nil && app.ProviderService != nil && b.storage.UserGrants() != nil {
		app.ConsentService = consentservice.NewService(
			b.storage.Agents(),
			app.ProviderService,
			b.storage.UserGrants(),
		)
	}

	// Create OAuth2 service if configuration available.
	// T038: Use NewServiceWithSessions (enables mandatory requirement validation + multi-agent
	// client config) and pass MultiAgentClientConfig from cfg.OAuth2AuthServer.MultiAgentClient.
	if b.config.OAuth2AuthServer.UpstreamAuthorizeEndpoint != "" {
		app.OAuth2Service = oauth2service.NewServiceWithSessions(
			b.storage.Agents(),
			b.storage.UserGrants(),
			b.storage.UserSessions(),
			&oauth2service.OAuth2Config{
				UpstreamAuthorizeEndpoint: b.config.OAuth2AuthServer.UpstreamAuthorizeEndpoint,
				UpstreamTokenEndpoint:     b.config.OAuth2AuthServer.UpstreamTokenEndpoint,
				PublicURL:                 b.config.Server.EndUser.PublicURL,
				SupportedResponseTypes:    b.config.OAuth2AuthServer.SupportedResponseTypes,
				SupportedGrantTypes:       b.config.OAuth2AuthServer.SupportedGrantTypes,
				MultiAgentClient:          b.config.OAuth2AuthServer.MultiAgentClient,
			},
			b.logger,
		)
	}

	// OAuth2SessionService is always created because JWESigningKey is mandatory.
	// Unlike ConsentService and OAuth2Service (which are conditionally created based on
	// storage availability and config), OAuth2SessionService requires the JWESigningKey
	// which is marked as REQUIRED in config validation (internal/config/validator.go).
	// The application will fail to start if JWESigningKey is not provided, so we can
	// safely create OAuth2SessionService unconditionally here.
	//
	// Note: OAuth2SessionService requires a providerService for decrypting client secrets.
	// If services repository is not available, providerService will be nil and OAuth2SessionService
	// will fail to fetch services. This is acceptable since the application is non-functional
	// without the services repository anyway.

	// Decode JWE signing key
	keyBytes, err := base64.StdEncoding.DecodeString(b.config.ThirdPartyOAuth2.JWESigningKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWE signing key: %w", err)
	}

	// Import key as JWK
	jweKey, err := jwk.Import(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to import JWE signing key: %w", err)
	}

	// Build service configuration from application config
	// Constitution Principle VII: Configuration-Driven Design
	cfg := oauth2session.NewConfigFromPorts(b.config.ThirdPartyOAuth2, b.config.Server.EndUser.PublicURL)

	// Create HTTP client for token endpoint with configured timeout
	// Created early to support both OAuth2SessionService and TokenExchangeService
	upstreamClient := &http.Client{
		Timeout: time.Duration(b.config.OAuth2AuthServer.UpstreamTimeoutSeconds) * time.Second,
	}

	app.OAuth2SessionService = oauth2session.NewOAuth2SessionService(
		app.ProviderService,
		b.storage.UserSessions(),
		b.storage.UserGrants(),
		b.storage.Agents(),
		encryptor,
		upstreamClient,
		jweKey,
		cfg,
		b.logger,
	)

	// Create token exchange service if token exchange configuration is available
	// Per Constitution Principle VII (Configuration-Driven Design): only create if configured
	if b.config.TokenExchange.ClaimExtraction.PrincipalExpression != "" &&
		b.config.TokenExchange.Authorization.CEL.Expression != "" {
		// Validate required dependencies
		if app.ConsentService == nil {
			return nil, fmt.Errorf("token exchange service requires consent service, but storage repositories (Agents, Services, UserGrants) are not available")
		}

		// Create CEL evaluator with configuration
		celConfig := tokenexchange.CELEvaluatorConfig{
			PrincipalExpression:     b.config.TokenExchange.ClaimExtraction.PrincipalExpression,
			AgentClientIDExpression: b.config.TokenExchange.ClaimExtraction.AgentClientIDExpression,
			AuthorizationExpression: b.config.TokenExchange.Authorization.CEL.Expression,
			EvaluationTimeout:       b.config.TokenExchange.Authorization.CEL.EvaluationTimeout,
		}

		// T039: When feature is disabled, register resolveAgentIdByClientId CEL function so
		// agent_client_id_expression can look up an agent by its upstream client_id.
		// When enabled, the expression receives the UUID directly from the token — no lookup needed.
		if !b.config.OAuth2AuthServer.MultiAgentClient.Enabled {
			agentRepo := b.storage.Agents()
			celConfig.ResolveAgentIDByClientID = func(clientID string) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				agent, err := agentRepo.GetByClientID(ctx, id.ClientID(clientID))
				if err != nil {
					return "", fmt.Errorf("resolveAgentIdByClientId: %w", err)
				}
				return agent.ID.String(), nil
			}
		}

		celEvaluator, err := tokenexchange.NewCELEvaluator(celConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to create CEL evaluator for token exchange: %w", err)
		}

		// Create JWKS adapter for JWT validation
		// Per spec FR-039: JWKS fetched from upstream OAuth2 server
		jwksAdapter, err := jwks.NewJWKSAdapter(
			b.config.OAuth2AuthServer.UpstreamIssuerURI+"/.well-known/jwks.json",
			upstreamClient,
			15*time.Minute, // min refresh interval
			1*time.Hour,    // max refresh interval
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWKS adapter for token exchange: %w", err)
		}

		// Create JWT validator
		// Per spec SR-001: Client assertion and subject_token JWTs validated against JWKS
		jwtValidator, err := tokenexchange.NewJWTValidator(
			jwksAdapter,
			b.config.OAuth2AuthServer.UpstreamIssuerURI,
			"token-exchange-broker", // Per spec: broker's own identifier in audience claim
			60,                      // Per spec FR-042: 60 second clock skew tolerance
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWT validator for token exchange: %w", err)
		}

		// Create token exchange service
		// Per Constitution Principle VI: service depends on domain service, not raw repository
		// SessionRepository is no longer needed - token lifecycle is managed through OAuth2SessionService
		tokenExchangeService, err := tokenexchange.NewTokenExchangeService(
			jwtValidator,
			celEvaluator,
			app.ProviderService,
			app.OAuth2SessionService,
			app.ConsentService,
			b.storage.Agents(),
			&b.config.TokenExchange,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create token exchange service: %w", err)
		}

		app.TokenExchangeService = tokenExchangeService
	}

	// Create JWT pre-authentication adapter if configured
	// Per Constitution Principle VII: Configuration-Driven Design — only create when JWT block present
	if b.config.Server.EndUser.Authentication.JWT != nil {
		jwtCfg := b.config.Server.EndUser.Authentication.JWT

		// Validate JWT config mutual exclusivity (defense-in-depth, also checked by config validator)
		if jwtCfg.Verification == "none" && jwtCfg.JWKSURI != "" {
			return nil, fmt.Errorf("authentication.jwt: verification 'none' and jwks_uri are mutually exclusive")
		}

		// Create CEL evaluator for JWT claim extraction (domain layer)
		celConfig := domjwtauth.CELEvaluatorConfig{
			PrincipalExpression:   jwtCfg.ClaimExtraction.PrincipalExpression,
			DisplayNameExpression: jwtCfg.ClaimExtraction.DisplayNameExpression,
			EmailExpression:       jwtCfg.ClaimExtraction.EmailExpression,
			PictureURLExpression:  jwtCfg.ClaimExtraction.PictureURLExpression,
		}
		celEval, err := domjwtauth.NewCELEvaluator(celConfig, b.logger)
		if err != nil {
			return nil, fmt.Errorf("failed to create CEL evaluator for JWT pre-auth: %w", err)
		}

		// Create JWT authenticator adapter (uses lestrrat-go/jwx v3)
		jwtAuthenticator, err := jwtauthadapter.NewJWXAuthenticator(jwtauthadapter.JWXAuthenticatorConfig{
			JWTConfig:    jwtCfg,
			CELEvaluator: celEval,
			HTTPClient:   &http.Client{Timeout: 10 * time.Second},
			Logger:       b.logger,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to create JWT authenticator: %w", err)
		}

		app.JWTAuthenticator = jwtAuthenticator
		b.logger.Info("JWT pre-authentication enabled",
			"header_name", jwtCfg.HeaderName,
			"verification", jwtCfg.Verification,
			"has_audience", jwtCfg.ExpectedAudience != "",
			"has_issuer", jwtCfg.ExpectedIssuer != "",
		)
	}

	// Phase 3: Create handler instances

	// Admin handlers
	app.AdminHandlers = &AdminHandlers{
		Agents:   admin.NewAgentsHandler(b.storage.Agents(), app.ProviderService, b.logger, b.config.OAuth2AuthServer.MultiAgentClient.Enabled),
		Services: admin.NewServicesHandler(app.ProviderService, b.config, b.logger),
	}

	// Create agent detail handler with repository dependencies for service requirements (Phase 6)
	agentDetailHandler := consent.NewAgentDetailHandler(app.ConsentService, b.logger).
		WithAgentRepository(b.storage.Agents()).
		WithSessionRepository(b.storage.UserSessions()).
		WithProviderService(app.ProviderService)

	// T040: Build OAuth2TokenHandler — fail-fast if multi-agent verifier construction fails.
	// Config validation makes this error unreachable in practice, but structural fail-closed
	// guarantees (SR-001) are not conditional on upstream validation alone.
	oauth2TokenHandler := &enduser.OAuth2TokenHandler{
		UpstreamTokenURL: b.config.OAuth2AuthServer.UpstreamTokenEndpoint,
		Client:           upstreamClient,
		TokenExchange:    app.TokenExchangeService,
		Logger:           b.logger,
	}
	if b.config.OAuth2AuthServer.MultiAgentClient.Enabled {
		verifier, err := oauth2service.NewMultiAgentTokenVerifier(b.config.OAuth2AuthServer.MultiAgentClient.AgentIDClaimName)
		if err != nil {
			return nil, fmt.Errorf("failed to create multi-agent token verifier: %w", err)
		}
		oauth2TokenHandler.MultiAgentVerifier = verifier
	}

	// Enduser handlers
	app.EnduserHandlers = &EnduserHandlers{
		UserInfo:       consent.NewUserInfoHandler(b.logger),
		Agents:         consent.NewAgentsHandler(app.ConsentService, b.logger),
		AgentDetail:    agentDetailHandler,
		AgentGrants:    consent.NewAgentGrantsHandler(app.ConsentService, b.logger),
		Grants:         consent.NewGrantsHandler(app.ConsentService, b.logger),
		OAuth2Sessions: oauth2_sessions.NewHandler(app.OAuth2SessionService),
		OAuth2Authorize: &enduser.OAuth2AuthorizeHandler{
			Service: app.OAuth2Service,
		},
		OAuth2Token: oauth2TokenHandler,
		OAuth2Metadata: &enduser.OAuth2MetadataHandler{
			Service: app.OAuth2Service,
		},
		SPA: handlers.NewSPAHandler(b.staticWebResourcesPath, b.logger),
	}

	return app, nil
}
