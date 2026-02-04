package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config holds the mock OAuth2 server configuration
type Config struct {
	Server Server         `mapstructure:"server"`
	OAuth2 OAuth2Config   `mapstructure:"oauth2"`
	User   MockUserConfig `mapstructure:"mock_user"`
}

// Server holds server configuration
type Server struct {
	Port int    `mapstructure:"port"`
	Bind string `mapstructure:"bind"`
}

// OAuth2Config holds OAuth2 settings
type OAuth2Config struct {
	ClientID             string        `mapstructure:"client_id"`
	ClientSecret         string        `mapstructure:"client_secret"`
	AccessTokenTTL       time.Duration `mapstructure:"access_token_ttl"`
	RefreshTokenTTL      time.Duration `mapstructure:"refresh_token_ttl"`
	AuthorizationCodeTTL time.Duration `mapstructure:"authorization_code_ttl"`
	Scopes               []ScopeConfig `mapstructure:"scopes"`
}

// ScopeConfig holds scope information
type ScopeConfig struct {
	Name        string `mapstructure:"name"`
	Description string `mapstructure:"description"`
}

// MockUserConfig holds mock user information for /userinfo endpoint
type MockUserConfig struct {
	Sub   string `mapstructure:"sub"`
	Name  string `mapstructure:"name"`
	Email string `mapstructure:"email"`
}

// Load loads the configuration from config.yaml
func Load(configPath string) (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configPath)

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate required fields
	if cfg.Server.Port == 0 {
		return nil, fmt.Errorf("server.port is required")
	}
	if cfg.Server.Bind == "" {
		return nil, fmt.Errorf("server.bind is required")
	}
	if cfg.OAuth2.ClientID == "" {
		return nil, fmt.Errorf("oauth2.client_id is required")
	}
	if cfg.OAuth2.ClientSecret == "" {
		return nil, fmt.Errorf("oauth2.client_secret is required")
	}
	if len(cfg.OAuth2.Scopes) == 0 {
		return nil, fmt.Errorf("at least one scope is required")
	}

	return &cfg, nil
}
