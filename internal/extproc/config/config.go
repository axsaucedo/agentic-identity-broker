// Package config defines the configuration schema for the ExtProc Token Exchange Service.
// It uses its own schema entirely separate from the identity broker configuration.
// All configuration is loaded via Viper with the EXTPROC_ env var prefix.
package config

import "time"

// Config is the root configuration for the ExtProc Token Exchange Service.
type Config struct {
	GRPC           GRPCConfig           `mapstructure:"grpc"`
	OAuth2         OAuth2Config         `mapstructure:"oauth2"`
	Cache          CacheConfig          `mapstructure:"cache"`
	Log            LogConfig            `mapstructure:"log"`
	CircuitBreaker CircuitBreakerConfig `mapstructure:"circuit_breaker"`
}

// GRPCConfig holds gRPC server settings.
type GRPCConfig struct {
	Bind                 string `mapstructure:"bind"`
	Port                 int    `mapstructure:"port"`
	MaxConcurrentStreams int    `mapstructure:"max_concurrent_streams"`
}

// OAuth2Config holds OAuth2 and token exchange settings.
type OAuth2Config struct {
	TokenEndpoint             string        `mapstructure:"token_endpoint"`
	Issuer                    string        `mapstructure:"issuer"`
	ClientID                  string        `mapstructure:"client_id"`
	ClientSecret              string        `mapstructure:"client_secret"`
	ClientCredentialsEndpoint string        `mapstructure:"client_credentials_endpoint"`
	ClientCredentialsScopes   []string      `mapstructure:"client_credentials_scopes"`
	ClientAssertionType       string        `mapstructure:"client_assertion_type"`
	ExchangeTimeout           time.Duration `mapstructure:"exchange_timeout"`
	TLS                       TLSConfig     `mapstructure:"tls"`
}

// TLSConfig holds TLS settings for outbound HTTP connections.
type TLSConfig struct {
	InsecureSkipVerify bool   `mapstructure:"insecure_skip_verify"`
	CaBundlePath       string `mapstructure:"ca_bundle_path"`
	AllowHTTP          bool   `mapstructure:"allow_http"`
}

// CacheConfig holds token cache settings.
type CacheConfig struct {
	DefaultTTL time.Duration `mapstructure:"default_ttl"`
	MaxTTL     time.Duration `mapstructure:"max_ttl"`
}

// LogConfig holds logging settings.
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

// CircuitBreakerConfig holds circuit breaker settings for the token exchange
// HTTP calls to the identity broker. The circuit breaker prevents a thundering
// herd when the identity broker recovers after an outage.
type CircuitBreakerConfig struct {
	MaxFailures  int           `mapstructure:"max_failures"`
	ResetTimeout time.Duration `mapstructure:"reset_timeout"`
}
