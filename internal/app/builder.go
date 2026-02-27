// Package app provides application-layer wiring and orchestration.
package app

import (
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/lestrrat-go/jwx/v3/jwk"

	awsencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/aws"
	encmemory "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/memory"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/admin"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/oauth2_sessions"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/jwks"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/noop"
	consentservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
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
	encryption             ports.EncryptionPort   // Optional: custom encryption implementation
	branchKeyManager       ports.BranchKeyManager // Optional: custom branch key manager
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

// WithEncryption sets a custom encryption implementation for the builder.
// If not set, defaults to no-op encryption for development.
// Use this to inject a production encryption adapter.
func (b *Builder) WithEncryption(encryptor ports.EncryptionPort) *Builder {
	b.encryption = encryptor
	return b
}

// WithBranchKeyManager sets a custom branch key manager for the builder.
// If not set, defaults based on keyring type (AWS manager or in-memory).
// Use this to inject a test or custom branch key manager.
func (b *Builder) WithBranchKeyManager(mgr ports.BranchKeyManager) *Builder {
	b.branchKeyManager = mgr
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

	app := &App{
		Config:  b.config,
		Storage: b.storage,
		Logger:  b.logger,
	}

	// Phase 1: Initialize encryption adapter (must happen before domain services)
	// Constitution Principle VII: Configuration-Driven Design
	var encryptor ports.EncryptionPort
	if b.encryption != nil {
		// Builder override takes precedence (for testing)
		encryptor = b.encryption
	} else if b.config.Encryption.AWSKMS != nil || b.config.Encryption.Memory != nil {
		// Production/Development: Use new backend-explicit configuration factory
		// The factory handles backend detection and validation automatically
		adapter, branchKeyManager, err := awsencryption.NewEncryptionAdapter(&b.config.Encryption)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize encryption adapter: %w", err)
		}
		encryptor = adapter

		// Wire branch key manager if available (for all backends except no-op)
		if branchKeyManager != nil && b.branchKeyManager == nil {
			b.branchKeyManager = branchKeyManager
		}

		// Log initialization with backend information
		if b.config.Encryption.AWSKMS != nil {
			b.logger.Info("AWS KMS encryption adapter initialized",
				"dynamodb_table", b.config.Encryption.AWSKMS.DynamoDBTableName,
				"dynamodb_region", b.config.Encryption.AWSKMS.DynamoDBRegion,
				"branch_key_ttl", b.config.Encryption.AWSKMS.BranchKeyTTL,
				"dynamodb_read_timeout", b.config.Encryption.AWSKMS.DynamoDBReadTimeout,
				"dynamodb_write_timeout", b.config.Encryption.AWSKMS.DynamoDBWriteTimeout,
				"branch_key_manager_wired", branchKeyManager != nil)
		} else if b.config.Encryption.Memory != nil {
			b.logger.Info("Memory encryption adapter initialized",
				"branch_key_manager_wired", branchKeyManager != nil)
		}
	} else {
		encryptor = noop.NewNoOpEncryption()
		b.logger.Info("No-op encryption enabled (development mode)")

		// Wire in-memory BranchKeyManager for development if not already set
		if b.branchKeyManager == nil {
			b.branchKeyManager = encmemory.NewInMemoryBranchKeyRepository()
			b.logger.Info("BranchKeyManager wired from in-memory implementation (development mode)")
		}
	}

	// Assign BranchKeyManager to app if wired
	if b.branchKeyManager != nil {
		app.BranchKeyManager = b.branchKeyManager
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
			b.branchKeyManager, // May be nil if no encryption backend configured
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

	// Create OAuth2 service if configuration available
	if b.config.OAuth2AuthServer.UpstreamAuthorizeEndpoint != "" {
		app.OAuth2Service = oauth2service.NewService(
			b.storage.Agents(),
			b.storage.UserGrants(),
			&oauth2service.OAuth2Config{
				UpstreamAuthorizeEndpoint: b.config.OAuth2AuthServer.UpstreamAuthorizeEndpoint,
				UpstreamTokenEndpoint:     b.config.OAuth2AuthServer.UpstreamTokenEndpoint,
				PublicURL:                 b.config.Server.EndUser.PublicURL,
				SupportedResponseTypes:    b.config.OAuth2AuthServer.SupportedResponseTypes,
				SupportedGrantTypes:       b.config.OAuth2AuthServer.SupportedGrantTypes,
			},
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
			&b.config.TokenExchange,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create token exchange service: %w", err)
		}

		app.TokenExchangeService = tokenExchangeService
	}

	// Phase 3: Create handler instances

	// Admin handlers
	app.AdminHandlers = &AdminHandlers{
		Agents:   admin.NewAgentsHandler(b.storage.Agents(), b.storage.Services(), b.logger),
		Services: admin.NewServicesHandler(app.ProviderService, b.config, b.logger),
	}

	// Create agent detail handler with repository dependencies for service requirements (Phase 6)
	agentDetailHandler := consent.NewAgentDetailHandler(app.ConsentService, b.logger).
		WithAgentRepository(b.storage.Agents()).
		WithSessionRepository(b.storage.UserSessions()).
		WithProviderService(app.ProviderService)

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
		OAuth2Token: &enduser.OAuth2TokenHandler{
			UpstreamTokenURL: b.config.OAuth2AuthServer.UpstreamTokenEndpoint,
			Client:           upstreamClient,
			TokenExchange:    app.TokenExchangeService,
			Logger:           b.logger,
		},
		OAuth2Metadata: &enduser.OAuth2MetadataHandler{
			Service: app.OAuth2Service,
		},
		SPA: handlers.NewSPAHandler(b.staticWebResourcesPath, b.logger),
	}

	return app, nil
}
