package fixtures

import (
	"encoding/base64"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// DefaultOAuth2Config returns a minimal valid OAuth2 authorization server configuration.
// Uses safe defaults suitable for E2E testing:
// - In-memory storage backend
// - Upstream issuer: http://localhost:19000
// - Upstream authorize endpoint: http://localhost:19000/authorize
// - Upstream token endpoint: http://localhost:19000/token
// - Timeout: 30 seconds
// - Public URL: http://localhost:8000
// - X-Remote-User authentication header
// - ThirdPartyOAuth2: JWE signing key configured for OAuth2 session management
// - StateTokenTTL: 10 minutes
// - PKCE verifier length: 32 bytes
func DefaultOAuth2Config() *ports.Config {
	return &ports.Config{
		Log: ports.LogConfig{
			Level:  ports.LogLevelInfo,
			Format: ports.LogFormatText,
		},
		Server: ports.ServerConfig{
			EndUser: ports.ServerInstanceConfig{
				Port:      8000,
				Bind:      "127.0.0.1",
				PublicURL: "http://localhost:8000",
				Authentication: ports.AuthenticationConfig{
					Preauth: ports.PreauthConfig{
						PrincipalHeaderName: "X-Remote-User",
					},
				},
			},
			Admin: ports.ServerInstanceConfig{
				Port:      14000,
				Bind:      "127.0.0.1",
				PublicURL: "http://localhost:14000",
				Authentication: ports.AuthenticationConfig{
					Preauth: ports.PreauthConfig{
						PrincipalHeaderName: "X-Remote-User",
					},
				},
			},
			Shutdown: ports.ShutdownConfig{
				Timeout: 5 * time.Second,
			},
		},
		Storage: ports.StorageConfig{
			Backend: "memory",
			Timeouts: ports.StorageTimeouts{
				Read:  5 * time.Second,
				Write: 5 * time.Second,
			},
		},
		OAuth2AuthServer: ports.OAuth2AuthServerConfig{
			UpstreamIssuerURI:         "http://localhost:19000",
			UpstreamAuthorizeEndpoint: "http://localhost:19000/authorize",
			UpstreamTokenEndpoint:     "http://localhost:19000/token",
			SupportedResponseTypes:    []string{"code"},
			SupportedGrantTypes:       []string{"authorization_code", "refresh_token"},
			UpstreamTimeoutSeconds:    30,
			Mode:                      "proxy",
		},
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey:      base64.StdEncoding.EncodeToString([]byte("test-32-byte-key-must-be-exact-")),
			StateTokenTTL:      10 * time.Minute,
			PKCEVerifierLength: 32,
		},
		TokenExchange: ports.TokenExchangeConfig{
			ClaimExtraction: ports.ClaimExtractionConfig{
				PrincipalExpression:     "subject_token.sub",
				AgentClientIDExpression: "subject_token.azp",
			},
			Authorization: ports.AuthorizationConfig{
				Type: "cel",
				CEL: ports.CELAuthorizationConfig{
					Expression:        "true",
					EvaluationTimeout: 100 * time.Millisecond,
				},
			},
		},
	}
}

// OAuth2ConfigWithUpstream returns a config with a custom upstream OAuth2 server URL.
// Useful for pointing to a mock upstream server in tests.
// All other settings match DefaultOAuth2Config().
func OAuth2ConfigWithUpstream(upstreamURL string) *ports.Config {
	config := DefaultOAuth2Config()
	config.OAuth2AuthServer.UpstreamIssuerURI = upstreamURL
	config.OAuth2AuthServer.UpstreamAuthorizeEndpoint = upstreamURL + "/oauth/authorize"
	config.OAuth2AuthServer.UpstreamTokenEndpoint = upstreamURL + "/oauth/token"
	return config
}

// OAuth2ConfigWithTimeout returns a config with a custom upstream timeout.
// Useful for testing timeout behavior.
// TimeoutSeconds: specified by caller
// All other settings match DefaultOAuth2Config().
func OAuth2ConfigWithTimeout(timeoutSeconds int) *ports.Config {
	config := DefaultOAuth2Config()
	config.OAuth2AuthServer.UpstreamTimeoutSeconds = timeoutSeconds
	return config
}

// OAuth2ConfigWithLogLevel returns a config with a custom log level.
// Useful for debugging E2E test failures.
// LogLevel: specified by caller (e.g., "debug", "info", "warn", "error")
// All other settings match DefaultOAuth2Config().
func OAuth2ConfigWithLogLevel(level string) *ports.Config {
	config := DefaultOAuth2Config()
	config.Log.Level = ports.LogLevel(level)
	return config
}

// OAuth2ConfigWithPublicURL returns a config with a custom public URL.
// Useful for testing metadata discovery with different base URLs.
// PublicURL: specified by caller
// All other settings match DefaultOAuth2Config().
func OAuth2ConfigWithPublicURL(publicURL string) *ports.Config {
	config := DefaultOAuth2Config()
	config.Server.EndUser.PublicURL = publicURL
	return config
}

// OAuth2ConfigWithStorage returns a config with custom storage backend.
// Useful for testing with different storage backends (not typically used in E2E, but available).
// Backend: specified by caller (e.g., "memory" or "postgres")
// All other settings match DefaultOAuth2Config().
func OAuth2ConfigWithStorage(backend string) *ports.Config {
	config := DefaultOAuth2Config()
	config.Storage.Backend = backend
	return config
}

// TokenExchangeConfigWithCELExpression returns a config with a custom CEL authorization expression.
// Useful for testing different CEL authorization policies.
// Expression: specified by caller (e.g., "claims.iss == 'https://auth.example.com'")
// All other settings match DefaultOAuth2Config().
func TokenExchangeConfigWithCELExpression(expression string) *ports.Config {
	config := DefaultOAuth2Config()
	config.TokenExchange.Authorization.CEL.Expression = expression
	return config
}

// TokenExchangeConfigWithInvalidCELSyntax returns a config with an invalid CEL expression.
// Used to test that invalid CEL syntax is caught at startup (per FR-017).
// InvalidExpression: A malformed CEL expression that should fail compilation.
// This is used in US4-S4 to test startup validation.
func TokenExchangeConfigWithInvalidCELSyntax(invalidExpression string) *ports.Config {
	config := DefaultOAuth2Config()
	config.TokenExchange.Authorization.CEL.Expression = invalidExpression
	return config
}

// TokenExchangeConfigWithClaimExtraction returns a config with custom claim extraction expressions.
// Useful for testing different claim mapping strategies.
// PrincipalExpr: CEL expression to extract principal (e.g., "subject_token.preferred_username")
// AgentExpr: CEL expression to extract agent ID (e.g., "subject_token.client_id")
// All other settings match DefaultOAuth2Config().
func TokenExchangeConfigWithClaimExtraction(principalExpr, agentExpr string) *ports.Config {
	config := DefaultOAuth2Config()
	config.TokenExchange.ClaimExtraction.PrincipalExpression = principalExpr
	config.TokenExchange.ClaimExtraction.AgentClientIDExpression = agentExpr
	return config
}
