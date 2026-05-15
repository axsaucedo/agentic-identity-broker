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
	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	adaptercmd "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/cimd"
	awsencryption "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/encryption/aws"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/admin"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/consent"
	enduserHandlers "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/handlers/enduser"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http/oauth2_sessions"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/jwks"
	jwtauthadapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/jwtauth"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/telemetry"
	agentsservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/agents"
	consentservice "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/consent"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/id"
	domjwe "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwe"
	domjwtauth "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/jwtauth"
	oauth2service "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2"
	domaincimd "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/cimd"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2session"
	domstorage "github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/thirdparty"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/tokenexchange"
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

	// ShutdownTelemetry must be called on graceful shutdown to flush and close
	// all OTel providers. It is always non-nil — when telemetry is disabled it
	// is a no-op.
	ShutdownTelemetry func(context.Context) error
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
	tracerProvider         *sdktrace.TracerProvider // Optional: custom TracerProvider for testing
	cimdFetcher            ports.CIMDFetcher        // Optional: overrides auto-created CIMD fetcher for testing
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

// WithTracerProvider sets a custom TracerProvider for testing.
// When set, this provider is registered as the global provider instead of
// the one created by NewProvider(). Mirrors the WithEncryption precedent.
func (b *Builder) WithTracerProvider(tp *sdktrace.TracerProvider) *Builder {
	b.tracerProvider = tp
	return b
}

// WithCIMDFetcher injects a custom CIMDFetcher, bypassing the production fetcher
// created from CIMDConfig. Intended for testing — allows injecting an HTTP client
// that trusts test TLS certificates (e.g., from httptest.NewTLSServer).
// Only effective when cimd.enabled is true.
func (b *Builder) WithCIMDFetcher(f ports.CIMDFetcher) *Builder {
	b.cimdFetcher = f
	return b
}

func modeStrategyFor(mode string) oauth2service.ModeStrategy {
	switch mode {
	case "local":
		return oauth2service.NewLocalModeStrategy()
	case "hybrid":
		return oauth2service.NewHybridModeStrategy()
	default: // "proxy"
		return oauth2service.NewProxyModeStrategy()
	}
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

	if err := b.config.OAuth2AuthServer.Validate(); err != nil {
		return nil, fmt.Errorf("oauth2_authorization_server configuration invalid: %w", err)
	}

	app := &App{
		Config:  b.config,
		Storage: b.storage,
		Logger:  b.logger,
	}

	// T029: Initialize telemetry provider
	// Per ADR-011: OTel provider wired at app layer, no port interface needed.
	// When b.tracerProvider is set (test override), register it globally and wrap
	// it in a shutdown function. Otherwise, initialize the full OTel provider from config.
	if b.tracerProvider != nil {
		// Test override: register the provided TracerProvider globally
		otel.SetTracerProvider(b.tracerProvider)
		// Set default propagators for test environment — must match the production
		// default set (ottrace, b3multi, baggage) to ensure span connectivity.
		otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
			ot.OT{},
			b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)),
			propagation.Baggage{},
		))
		app.ShutdownTelemetry = func(ctx context.Context) error {
			return b.tracerProvider.Shutdown(ctx)
		}
	} else {
		// Production path: initialize full provider from config.
		// Use a bounded context for initialization to avoid blocking indefinitely if the
		// OTLP endpoint is unreachable at startup. The exporter timeout is a reasonable bound;
		// per-export retries are handled by the OTel SDK independently of this context.
		initCtx, initCancel := context.WithTimeout(context.Background(), b.config.Telemetry.Exporter.Timeout)
		defer initCancel()
		shutdownTelemetry, telErr := telemetry.NewProvider(initCtx, b.config.Telemetry, b.logger)
		if telErr != nil {
			return nil, fmt.Errorf("failed to initialize telemetry: %w", telErr)
		}
		app.ShutdownTelemetry = shutdownTelemetry
	}

	// T038: Wire OTel slog bridge when telemetry and log export are both enabled.
	// Wraps the base logger handler with a multi-handler that fans log records to both
	// the original handler and the OTel log bridge (otelslog), enabling log-trace correlation.
	if b.config.Telemetry.Enabled && b.config.Telemetry.Logs.Enabled {
		otelHandler := otelslog.NewHandler(b.config.Telemetry.ServiceName,
			otelslog.WithLoggerProvider(global.GetLoggerProvider()))
		b.logger = slog.New(telemetry.NewMultiHandler(b.logger.Handler(), otelHandler))
		app.Logger = b.logger
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

	// Decode and import JWE signing key — required for both OAuth2SessionService and OAuth2Service
	// (CIMD consent flows). Fail fast here before constructing any domain services.
	keyBytes, err := base64.StdEncoding.DecodeString(b.config.ThirdPartyOAuth2.JWESigningKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decode JWE signing key: %w", err)
	}
	jweKey, err := jwk.Import(keyBytes)
	if err != nil {
		return nil, fmt.Errorf("failed to import JWE signing key: %w", err)
	}
	jweTokenService := domjwe.New(jweKey)

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
			b.storage.UserSessions(),
			b.logger,
		)
	}

	// Create OAuth2 service if configuration available.
	// T038: Use NewServiceWithSessions (enables mandatory requirement validation + multi-agent
	// client config) and pass MultiAgentClientConfig from cfg.OAuth2AuthServer.MultiAgentClient.
	// In local mode, also create the service for consent checks and metadata generation.
	var clientResolver ports.ClientResolver
	if b.config.OAuth2AuthServer.Proxy.UpstreamAuthorizeEndpoint != "" || b.config.OAuth2AuthServer.Mode == "local" {
		oauth2Config := &oauth2service.OAuth2Config{
			UpstreamAuthorizeEndpoint: b.config.OAuth2AuthServer.Proxy.UpstreamAuthorizeEndpoint,
			UpstreamTokenEndpoint:     b.config.OAuth2AuthServer.Proxy.UpstreamTokenEndpoint,
			PublicURL:                 b.config.Server.EndUser.PublicURL,
			SupportedResponseTypes:    b.config.OAuth2AuthServer.SupportedResponseTypes,
			SupportedGrantTypes:       b.config.OAuth2AuthServer.SupportedGrantTypes,
			MultiAgentClient:          b.config.OAuth2AuthServer.MultiAgentClient,
			Mode:                      b.config.OAuth2AuthServer.Mode,
			CIMDEnabled:               b.config.OAuth2AuthServer.CIMD.Enabled,
			ModeStrategy:              modeStrategyFor(b.config.OAuth2AuthServer.Mode),
		}
		// In local mode, set correct defaults for supported types
		if b.config.OAuth2AuthServer.Mode == "local" {
			if len(oauth2Config.SupportedResponseTypes) == 0 {
				oauth2Config.SupportedResponseTypes = []string{"code"}
			}
			if len(oauth2Config.SupportedGrantTypes) == 0 {
				oauth2Config.SupportedGrantTypes = []string{"authorization_code", "client_credentials"}
			}
		}

		cimdCfg := b.config.OAuth2AuthServer.CIMD
		if cimdCfg.Enabled {
			activeFetcher := b.cimdFetcher
			if activeFetcher == nil {
				concreteFetcher, fetchErr := adaptercmd.NewFetcher(
					cimdCfg.FetchTimeout,
					int64(cimdCfg.MaxResponseBytes),
					cimdCfg.SSRF.ExtraBlockedCIDRs,
				)
				if fetchErr != nil {
					return nil, fmt.Errorf("failed to create CIMD fetcher: %w", fetchErr)
				}
				if b.config.Telemetry.Enabled && b.config.Telemetry.Traces.Enabled {
					concreteFetcher.WrapTransport(func(base http.RoundTripper) http.RoundTripper {
						return otelhttp.NewTransport(base)
					})
				}
				activeFetcher = concreteFetcher
			}
			cimdCache, cacheErr := domaincimd.NewCIMDCache(cimdCfg.Cache.MinTTL, cimdCfg.Cache.MaxTTL, cimdCfg.Cache.MaxEntries)
			if cacheErr != nil {
				return nil, fmt.Errorf("failed to create CIMD cache: %w", cacheErr)
			}
			cimdSvc := domaincimd.NewService(activeFetcher, cimdCache, cimdCfg.ClientNameBlocklist, b.logger)
			clientResolver = domaincimd.NewCIMDClientResolver(b.storage.Agents(), cimdSvc, b.logger)
			b.logger.Info("CIMD client resolution enabled",
				"fetch_timeout", cimdCfg.FetchTimeout,
				"max_response_bytes", cimdCfg.MaxResponseBytes,
			)
		} else {
			clientResolver = oauth2service.NewOpaqueClientResolver(b.storage.Agents())
		}

		app.OAuth2Service = oauth2service.NewServiceWithClientResolver(
			b.storage.UserGrants(),
			b.storage.UserSessions(),
			clientResolver,
			oauth2Config,
			b.logger,
		).WithJWETokenService(jweTokenService)
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

	// Build service configuration from application config
	// Constitution Principle VII: Configuration-Driven Design
	cfg := oauth2session.NewConfigFromPorts(b.config.ThirdPartyOAuth2, b.config.Server.EndUser.PublicURL)

	// Create HTTP client for token endpoint with configured timeout
	// Created early to support both OAuth2SessionService and TokenExchangeService
	upstreamClient := &http.Client{
		Timeout: time.Duration(b.config.OAuth2AuthServer.Proxy.UpstreamTimeoutSeconds) * time.Second,
	}

	// Wrap the HTTP transport with OTel instrumentation when tracing is enabled.
	// This is the "last resort" layer: even operations without an explicit custom span will
	// still emit a client span and propagate W3C traceparent/tracestate headers to every
	// outgoing HTTP call (JWKS fetches, upstream token proxy, OAuth2 session token exchange).
	if b.config.Telemetry.Enabled && b.config.Telemetry.Traces.Enabled {
		base := upstreamClient.Transport
		if base == nil {
			base = http.DefaultTransport
		}
		upstreamClient.Transport = otelhttp.NewTransport(base)
	}

	app.OAuth2SessionService = oauth2session.NewOAuth2SessionService(
		app.ProviderService,
		b.storage.UserSessions(),
		b.storage.UserGrants(),
		b.storage.Agents(),
		encryptor,
		upstreamClient,
		jweTokenService,
		cfg,
		b.logger,
	)

	// Create agent domain service (used by admin handlers and CEL resolver)
	agentService := agentsservice.NewService(
		b.storage.Agents(),
		app.ProviderService,
		b.logger,
		b.config.OAuth2AuthServer.MultiAgentClient.Enabled,
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
			AgentIDExpression:       b.config.TokenExchange.ClaimExtraction.AgentIDExpression,
			AuthorizationExpression: b.config.TokenExchange.Authorization.CEL.Expression,
			EvaluationTimeout:       b.config.TokenExchange.Authorization.CEL.EvaluationTimeout,
		}

		// T039: When feature is disabled, register resolveAgentIdByClientId CEL function so
		// agent_id_expression can look up an agent by its upstream client_id.
		// When enabled, the expression receives the UUID directly from the token — no lookup needed.
		if !b.config.OAuth2AuthServer.MultiAgentClient.Enabled {
			celConfig.ResolveAgentIDByClientID = func(clientID string) (string, error) {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				agent, err := agentService.ResolveUniqueByClientID(ctx, id.ClientID(clientID))
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
		// Per spec FR-039: JWKS URI discovered from upstream OAuth2 server metadata (RFC 8414)
		discoveryCtx, discoveryCancel := context.WithTimeout(context.Background(),
			time.Duration(b.config.OAuth2AuthServer.Proxy.UpstreamTimeoutSeconds)*time.Second)

		discovered, err := domstorage.DiscoverOAuth2Endpoints(
			discoveryCtx,
			b.config.OAuth2AuthServer.Proxy.UpstreamIssuerURI,
			nil, // use standard /.well-known/oauth-authorization-server path
			b.config.Security.SkipThirdpartyHTTPSValidation,
		)
		discoveryCancel()
		if err != nil {
			return nil, fmt.Errorf("failed to discover OAuth2 server metadata: %w", err)
		}
		if discovered.JWKsURI == "" {
			return nil, fmt.Errorf("OAuth2 server metadata did not include a jwks_uri")
		}

		jwksAdapter, err := jwks.NewJWKSAdapter(
			discovered.JWKsURI,
			upstreamClient,
			15*time.Minute, // min refresh interval
			1*time.Hour,    // max refresh interval
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWKS adapter for token exchange: %w", err)
		}

		// Create JWT validator
		// Per spec SR-001: Client assertion and subject_token JWTs validated against JWKS
		brokerAudience := b.config.TokenExchange.ExpectedAudience
		if brokerAudience == "" {
			brokerAudience = tokenexchange.DefaultBrokerAudience
		}
		jwtValidator, err := tokenexchange.NewJWTValidator(
			jwksAdapter,
			b.config.OAuth2AuthServer.Proxy.UpstreamIssuerURI,
			brokerAudience,
			tokenexchange.DefaultClockSkewTolerance, // Per spec FR-042: 60 second clock skew tolerance
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
		Agents:   admin.NewAgentsHandler(agentService, app.ProviderService, b.logger),
		Services: admin.NewServicesHandler(app.ProviderService, b.config, b.logger),
	}

	agentDetailHandler := consent.NewAgentDetailHandler(app.ConsentService, b.logger, jweTokenService)

	// T040: Build OAuth2TokenHandler — fail-fast if multi-agent verifier construction fails.
	// Config validation makes this error unreachable in practice, but structural fail-closed
	// guarantees (SR-001) are not conditional on upstream validation alone.
	var multiAgentVerifier ports.MultiAgentVerifier
	if b.config.OAuth2AuthServer.MultiAgentClient.Enabled {
		// Discover JWKS URI for multi-agent token signature verification (defense-in-depth, SR-001).
		// The broker is the relying party and must verify that the upstream token has not been tampered with.
		multiAgentDiscoveryCtx, multiAgentDiscoveryCancel := context.WithTimeout(
			context.Background(),
			time.Duration(b.config.OAuth2AuthServer.Proxy.UpstreamTimeoutSeconds)*time.Second,
		)
		multiAgentDiscovered, err := domstorage.DiscoverOAuth2Endpoints(
			multiAgentDiscoveryCtx,
			b.config.OAuth2AuthServer.Proxy.UpstreamIssuerURI,
			nil, // use standard /.well-known/oauth-authorization-server path
			b.config.Security.SkipThirdpartyHTTPSValidation,
		)
		multiAgentDiscoveryCancel()
		if err != nil {
			return nil, fmt.Errorf("failed to discover OAuth2 server metadata for multi-agent verifier: %w", err)
		}
		if multiAgentDiscovered.JWKsURI == "" {
			return nil, fmt.Errorf("OAuth2 server metadata did not include a jwks_uri (required for multi-agent token verification)")
		}

		multiAgentJWKSAdapter, err := jwks.NewJWKSAdapter(
			multiAgentDiscovered.JWKsURI,
			upstreamClient,
			15*time.Minute,
			1*time.Hour,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create JWKS adapter for multi-agent verifier: %w", err)
		}

		verifier, err := oauth2service.NewMultiAgentTokenVerifier(
			b.config.OAuth2AuthServer.MultiAgentClient.AgentIDClaimName,
			multiAgentJWKSAdapter,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create multi-agent token verifier: %w", err)
		}
		multiAgentVerifier = verifier
	}
	var grantHandler enduser.TokenGrantStrategy
	var proceedHandler enduser.AuthorizationProceedStrategy
	var jwksHandler *enduserHandlers.JWKSHandler

	// buildLocalProvider constructs the local token issuance infrastructure.
	// Used in both "local" and "hybrid" modes.
	buildLocalProvider := func() (*oauth2server.Provider, error) {
		signingKeyService := oauth2server.NewSigningKeyService(b.storage.SigningKeys(), encryptor, b.logger)
		clientAuthService := oauth2server.NewClientAuthService(b.storage.BrokerCredentials(), clientResolver, b.logger)
		app.AdminHandlers.ClientCredentials = admin.NewClientCredentialsHandler(b.storage.BrokerCredentials(), b.storage.Agents(), clientAuthService, b.logger)
		app.AdminHandlers.SigningKeys = admin.NewSigningKeysHandler(b.storage.SigningKeys(), signingKeyService, b.logger)

		provider, err := oauth2server.NewProvider(
			b.storage.AuthorizationCodes(),
			b.storage.PKCESessions(),
			b.storage.BrokerCredentials(),
			clientResolver,
			b.storage.SigningKeys(),
			encryptor,
			b.config.Server.EndUser.PublicURL,
			b.config.OAuth2AuthServer.Local.TokenTTL,
			b.config.OAuth2AuthServer.Local.TokenClaimsExpression,
			b.logger,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create OAuth2 server provider: %w", err)
		}

		jwksHandler = enduserHandlers.NewJWKSHandler(signingKeyService, b.logger)
		if err := signingKeyService.EnsureKeyExists(context.Background()); err != nil {
			return nil, fmt.Errorf("failed to ensure signing key exists: %w", err)
		}
		return provider, nil
	}

	// buildProxyStrategies constructs the proxy path strategies.
	// Used in both "proxy" and "hybrid" modes.
	buildProxyStrategies := func() (enduser.TokenGrantStrategy, enduser.AuthorizationProceedStrategy) {
		grant := enduser.NewProxyTokenGrantStrategy(
			b.config.OAuth2AuthServer.Proxy.UpstreamTokenEndpoint,
			upstreamClient,
			multiAgentVerifier,
			b.logger,
		)
		proceed := enduser.NewProxyProceedStrategy()
		return grant, proceed
	}

	switch b.config.OAuth2AuthServer.Mode {
	case "local":
		provider, err := buildLocalProvider()
		if err != nil {
			return nil, err
		}
		grantHandler = enduser.NewLocalGrantStrategy(newLocalMintingStrategy(provider), b.logger)
		proceedHandler = enduser.NewLocalProceedStrategy(newLocalCodeIssuer(provider), b.logger)
		b.logger.Info("OAuth2 server mode: local — local token minting enabled",
			"issuer_uri", b.config.Server.EndUser.PublicURL,
			"token_ttl", b.config.OAuth2AuthServer.Local.TokenTTL,
		)
	case "hybrid":
		provider, err := buildLocalProvider()
		if err != nil {
			return nil, err
		}
		proxyGrant, proxyProceed := buildProxyStrategies()
		localGrant := enduser.NewLocalGrantStrategy(newLocalMintingStrategy(provider), b.logger)
		localProceed := enduser.NewLocalProceedStrategy(newLocalCodeIssuer(provider), b.logger)

		grantHandler = enduser.NewHybridTokenGrantStrategy(proxyGrant, localGrant)
		proceedHandler = enduser.NewHybridProceedStrategy(proxyProceed, localProceed)
		b.logger.Info("OAuth2 server mode: hybrid — proxy and local token minting enabled",
			"issuer_uri", b.config.Server.EndUser.PublicURL,
			"token_ttl", b.config.OAuth2AuthServer.Local.TokenTTL,
		)
	default: // "proxy"
		grantHandler, proceedHandler = buildProxyStrategies()
	}

	oauth2AuthorizeHandler := &enduser.OAuth2AuthorizeHandler{
		Service:        app.OAuth2Service,
		ProceedHandler: proceedHandler,
	}
	oauth2MetadataHandler := &enduser.OAuth2MetadataHandler{
		Service: app.OAuth2Service,
	}

	app.EnduserHandlers = &EnduserHandlers{
		UserInfo:        consent.NewUserInfoHandler(b.logger),
		Agents:          consent.NewAgentsHandler(app.ConsentService, b.logger),
		AgentDetail:     agentDetailHandler,
		AgentGrants:     consent.NewAgentGrantsHandler(app.ConsentService, b.logger),
		Grants:          consent.NewGrantsHandler(app.ConsentService, b.logger, jweTokenService),
		RevokeGrant:     consent.NewRevokeGrantHandler(app.ConsentService, b.logger),
		OAuth2Sessions:  oauth2_sessions.NewHandler(app.OAuth2SessionService),
		OAuth2Authorize: oauth2AuthorizeHandler,
		OAuth2Token: &enduser.OAuth2TokenHandler{
			TokenExchange: app.TokenExchangeService,
			OAuth2Service: app.OAuth2Service,
			Logger:        b.logger,
			GrantHandler:  grantHandler,
		},
		OAuth2Metadata: oauth2MetadataHandler,
		JWKS:           jwksHandler,
		SPA:            handlers.NewSPAHandler(b.staticWebResourcesPath, b.logger),
	}

	return app, nil
}
