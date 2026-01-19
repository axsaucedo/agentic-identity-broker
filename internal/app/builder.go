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
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/admin"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/oauth2_sessions"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage/noop"
	consentservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	oauth2service "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2session"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// App represents a fully-wired application with all services and handlers initialized.
// This is returned by Builder.Build() after dependency injection.
type App struct {
	// Configuration
	Config *ports.Config

	// Repositories
	Storage *storage.Adapter

	// Domain services
	ConsentService       *consentservice.Service
	OAuth2SessionService *oauth2session.OAuth2SessionService
	OAuth2Service        ports.OAuth2Service

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
	config     *ports.Config
	storage    *storage.Adapter
	logger     *slog.Logger
	encryption ports.EncryptionPort // Optional: custom encryption implementation
}

// NewBuilder creates a new application builder.
func NewBuilder() *Builder {
	return &Builder{}
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

	// Phase 1: Create domain services
	// Constitution Principle VI: domain depends on ports (repository interfaces), not adapters

	// Create consent service if repositories available
	if b.storage.Agents() != nil && b.storage.Services() != nil && b.storage.UserGrants() != nil {
		app.ConsentService = consentservice.NewService(
			b.storage.Agents(),
			b.storage.Services(),
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

	// Initialize encryption adapter based on configuration or builder override
	// Constitution Principle VII: Configuration-Driven Design
	var encryptor ports.EncryptionPort
	if b.encryption != nil {
		// Builder override takes precedence (for testing)
		encryptor = b.encryption
	} else if b.config.Encryption.KeyEncryptionKey != "" {
		// Production/Staging: Initialize AWS Encryption SDK adapter with configured KEK
		// Supports both AWS KMS ARN and environment variable injection
		// This follows a 6-phase initialization flow:
		//   1. Parse configuration values
		//   2. Apply configuration defaults
		//   3. Validate configuration
		//   4. Create KeyStore with AWS clients
		//   5. Create BranchKeySupplier
		//   6. Create hierarchical keyring

		// Phase 1-2: Parse and apply defaults for BranchKeyTTL
		branchKeyTTL := time.Duration(0)
		if b.config.Encryption.BranchKeyTTL != "" {
			var err error
			branchKeyTTL, err = time.ParseDuration(b.config.Encryption.BranchKeyTTL)
			if err != nil {
				return nil, fmt.Errorf("failed to parse encryption.branch_key_ttl %q: %w", b.config.Encryption.BranchKeyTTL, err)
			}
		}

		// Apply KeyStoreLogicalName default if not specified
		keyStoreLogicalName := b.config.Encryption.KeyStoreLogicalName
		if keyStoreLogicalName == "" {
			keyStoreLogicalName = "IdentityBrokerEncryptionVault"
		}

		// Phase 3: Validate keyring type if specified
		keyringType := b.config.Encryption.KeyringType
		if keyringType == "" {
			keyringType = "hierarchical" // default
		}
		if keyringType != "hierarchical" && keyringType != "raw" {
			return nil, fmt.Errorf("invalid encryption.keyring_type %q: must be 'hierarchical' or 'raw'", keyringType)
		}

		// Phase 4-6: Create AWS encryption adapter (orchestrates KeyStore, Supplier, Keyring creation)
		adapter, err := awsencryption.NewAWSEncryptionAdapterWithConfig(
			b.config.Encryption.KeyEncryptionKey,
			b.config.Encryption.DynamoDBTableName,
			branchKeyTTL,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize AWS encryption adapter: %w", err)
		}
		encryptor = adapter

		// Log initialization with all configuration details
		b.logger.Info("AWS Encryption SDK adapter initialized",
			"keystore_logical_name", keyStoreLogicalName,
			"keyring_type", keyringType,
			"dynamodb_table", b.config.Encryption.DynamoDBTableName,
			"dynamodb_region", b.config.Encryption.DynamoDBRegion,
			"branch_key_ttl", branchKeyTTL,
			"dynamodb_read_timeout", b.config.Encryption.DynamoDBReadTimeout,
			"dynamodb_write_timeout", b.config.Encryption.DynamoDBWriteTimeout)
	} else {
		// Development: No-op encryption for local development without AWS dependencies
		encryptor = noop.NewNoOpEncryption()
		b.logger.Info("No-op encryption enabled (development mode)")
	}

	app.OAuth2SessionService = oauth2session.NewOAuth2SessionService(
		b.storage.Services(),
		b.storage.UserSessions(),
		b.storage.UserGrants(),
		b.storage.Agents(),
		encryptor,
		jweKey,
		cfg,
		b.logger,
	)

	// Phase 2: Create handler instances

	// Admin handlers
	app.AdminHandlers = &AdminHandlers{
		Agents:   admin.NewAgentsHandler(b.storage.Agents(), b.storage.Services(), b.logger),
		Services: admin.NewServicesHandler(b.storage.Services(), b.config, b.logger),
	}

	// Create HTTP client for token endpoint with configured timeout
	upstreamClient := &http.Client{
		Timeout: time.Duration(b.config.OAuth2AuthServer.UpstreamTimeoutSeconds) * time.Second,
	}

	// Create agent detail handler with repository dependencies for service requirements (Phase 6)
	agentDetailHandler := consent.NewAgentDetailHandler(app.ConsentService, b.logger).
		WithAgentRepository(b.storage.Agents()).
		WithSessionRepository(b.storage.UserSessions()).
		WithServiceRepository(b.storage.Services())

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
		},
		OAuth2Metadata: &enduser.OAuth2MetadataHandler{
			Service: app.OAuth2Service,
		},
		SPA: handlers.NewSPAHandler("web/dist/consent", b.logger),
	}

	return app, nil
}
