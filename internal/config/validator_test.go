package config

import (
	"encoding/base64"
	"strings"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// validTestConfig returns a valid Config for testing with all required fields set.
func validTestConfig() *ports.Config {
	return &ports.Config{
		Log: ports.LogConfig{
			Level:  ports.LogLevelInfo,
			Format: ports.LogFormatText,
		},
		Server: ports.ServerConfig{
			EndUser: ports.ServerInstanceConfig{
				Port:      8000,
				Bind:      "::",
				PublicURL: "http://localhost:8000",
				Authentication: ports.AuthenticationConfig{
					Preauth: ports.PreauthConfig{
						PrincipalHeaderName: "X-Remote-User",
					},
				},
			},
			Admin: ports.ServerInstanceConfig{
				Port:      14000,
				Bind:      "::",
				PublicURL: "http://localhost:14000",
				Authentication: ports.AuthenticationConfig{
					Preauth: ports.PreauthConfig{
						PrincipalHeaderName: "X-Remote-User",
					},
				},
			},
		},
		Storage: ports.StorageConfig{
			Backend: "memory",
			Timeouts: ports.StorageTimeouts{
				Read:  5 * time.Second,
				Write: 10 * time.Second,
			},
		},
		ThirdPartyOAuth2: ports.ThirdPartyOAuth2Config{
			JWESigningKey: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
		},
		Encryption: ports.EncryptionConfig{
			Memory: &ports.MemoryConfig{
				RawKey: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
			},
		},
	}
}

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name    string
		level   ports.LogLevel
		wantErr bool
	}{
		{"valid debug", ports.LogLevelDebug, false},
		{"valid info", ports.LogLevelInfo, false},
		{"valid warn", ports.LogLevelWarn, false},
		{"valid error", ports.LogLevelError, false},
		{"invalid verbose", ports.LogLevel("verbose"), true},
		{"invalid empty", ports.LogLevel(""), true},
		{"invalid trace", ports.LogLevel("trace"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogLevel(tt.level)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLogLevel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLogFormat(t *testing.T) {
	tests := []struct {
		name    string
		format  ports.LogFormat
		wantErr bool
	}{
		{"valid text", ports.LogFormatText, false},
		{"valid json", ports.LogFormatJSON, false},
		{"invalid xml", ports.LogFormat("xml"), true},
		{"invalid empty", ports.LogFormat(""), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLogFormat(tt.format)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLogFormat() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ports.Config
		wantErr bool
	}{
		{
			name:    "valid config",
			cfg:     validTestConfig(),
			wantErr: false,
		},
		{
			name: "invalid log level",
			cfg: func() *ports.Config {
				cfg := validTestConfig()
				cfg.Log.Level = ports.LogLevel("invalid")
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "invalid log format",
			cfg: func() *ports.Config {
				cfg := validTestConfig()
				cfg.Log.Format = ports.LogFormat("invalid")
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "missing enduser principal header name without JWT fails",
			cfg: func() *ports.Config {
				cfg := validTestConfig()
				cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName = ""
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "missing admin principal header name without JWT fails",
			cfg: func() *ports.Config {
				cfg := validTestConfig()
				cfg.Server.Admin.Authentication.Preauth.PrincipalHeaderName = ""
				return cfg
			}(),
			wantErr: true,
		},
		{
			name: "missing enduser principal header name with JWT configured is valid",
			cfg: func() *ports.Config {
				cfg := validTestConfig()
				cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName = ""
				cfg.Server.EndUser.Authentication.JWT = &ports.JWTConfig{
					Verification: "none",
					ClaimExtraction: ports.JWTClaimExtractionConfig{
						PrincipalExpression: "claims.sub",
					},
				}
				return cfg
			}(),
			wantErr: false,
		},
		{
			name: "custom principal header names pass validation",
			cfg: func() *ports.Config {
				cfg := validTestConfig()
				cfg.Server.EndUser.Authentication.Preauth.PrincipalHeaderName = "X-Authenticated-User"
				cfg.Server.Admin.Authentication.Preauth.PrincipalHeaderName = "X-Admin-User"
				return cfg
			}(),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Validate(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify error is ConfigError type when expected
			if err != nil {
				if _, ok := err.(*config.ConfigError); !ok {
					t.Errorf("Validate() error type = %T, want *config.ConfigError", err)
				}
			}
		})
	}
}

// TestValidateJWTConfig tests JWT pre-authentication configuration validation.
// Covers mutual exclusivity checks, required fields, defaults application,
// and CEL expression validation per FR-003a and SR-004.
func TestValidateJWTConfig(t *testing.T) {
	tests := []struct {
		name     string
		jwt      *ports.JWTConfig
		security *ports.SecurityConfig
		wantErr  bool
		errMsg   string
	}{
		{
			name: "valid signed JWT config with JWKS URI",
			jwt: &ports.JWTConfig{
				HeaderName:   "Authorization",
				Verification: "jwks",
				JWKSURI:      "https://auth.example.com/.well-known/jwks.json",
				ClaimExtraction: ports.JWTClaimExtractionConfig{
					PrincipalExpression: "claims.sub",
				},
			},
			security: nil,
			wantErr:  false,
		},
		{
			name: "valid unsigned JWT config",
			jwt: &ports.JWTConfig{
				HeaderName:   "X-JWT-Claims",
				Verification: "none",
				ClaimExtraction: ports.JWTClaimExtractionConfig{
					PrincipalExpression: "claims.sub",
				},
			},
			security: nil,
			wantErr:  false,
		},
		{
			name: "defaults applied when fields are empty",
			jwt: &ports.JWTConfig{
				JWKSURI: "https://auth.example.com/.well-known/jwks.json",
			},
			security: nil,
			wantErr:  false,
		},
		{
			name: "mutual exclusivity: verification none + jwks_uri",
			jwt: &ports.JWTConfig{
				Verification: "none",
				JWKSURI:      "https://auth.example.com/.well-known/jwks.json",
			},
			security: nil,
			wantErr:  true,
			errMsg:   "mutually exclusive",
		},
		{
			name: "missing jwks_uri when verification is jwks",
			jwt: &ports.JWTConfig{
				Verification: "jwks",
				JWKSURI:      "",
			},
			security: nil,
			wantErr:  true,
			errMsg:   "jwks_uri",
		},
		{
			name: "invalid verification mode",
			jwt: &ports.JWTConfig{
				Verification: "custom",
			},
			security: nil,
			wantErr:  true,
			errMsg:   "'jwks' or 'none'",
		},
		{
			name: "HTTPS required for JWKS URI by default",
			jwt: &ports.JWTConfig{
				Verification: "jwks",
				JWKSURI:      "http://auth.example.com/.well-known/jwks.json",
			},
			security: nil,
			wantErr:  true,
			errMsg:   "HTTPS",
		},
		{
			name: "HTTP allowed when skip_thirdparty_https_validation is true",
			jwt: &ports.JWTConfig{
				Verification: "jwks",
				JWKSURI:      "http://auth.example.com/.well-known/jwks.json",
			},
			security: &ports.SecurityConfig{
				SkipThirdpartyHTTPSValidation: true,
			},
			wantErr: false,
		},
		{
			name: "invalid JWKS URI scheme (ftp)",
			jwt: &ports.JWTConfig{
				Verification: "jwks",
				JWKSURI:      "ftp://auth.example.com/jwks.json",
			},
			security: nil,
			wantErr:  true,
			errMsg:   "HTTP or HTTPS URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateJWTConfig(tt.jwt, "server.enduser.authentication.jwt", tt.security)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateJWTConfig() error = %v, wantErr %v", err, tt.wantErr)
			}

			if err != nil && tt.errMsg != "" {
				errStr := err.Error()
				if !strings.Contains(errStr, tt.errMsg) {
					t.Errorf("validateJWTConfig() error = %v, want error containing %q", err, tt.errMsg)
				}
			}
		})
	}
}

// TestValidateJWTConfig_DefaultsApplied verifies that defaults are correctly
// set when JWT config fields are empty.
func TestValidateJWTConfig_DefaultsApplied(t *testing.T) {
	jwt := &ports.JWTConfig{
		JWKSURI: "https://auth.example.com/.well-known/jwks.json",
	}

	err := validateJWTConfig(jwt, "server.enduser.authentication.jwt", nil)
	if err != nil {
		t.Fatalf("validateJWTConfig() unexpected error: %v", err)
	}

	if jwt.HeaderName != "Authorization" {
		t.Errorf("expected HeaderName default 'Authorization', got %q", jwt.HeaderName)
	}
	if jwt.Verification != "jwks" {
		t.Errorf("expected Verification default 'jwks', got %q", jwt.Verification)
	}
	if jwt.ClaimExtraction.PrincipalExpression != "claims.sub" {
		t.Errorf("expected PrincipalExpression default 'claims.sub', got %q",
			jwt.ClaimExtraction.PrincipalExpression)
	}
}

// TestValidate_WithJWTConfig tests that full config validation includes JWT
// validation when JWT config is present.
func TestValidate_WithJWTConfig(t *testing.T) {
	cfg := validTestConfig()
	cfg.Server.EndUser.Authentication.JWT = &ports.JWTConfig{
		Verification: "none",
		JWKSURI:      "https://should-not-be-set.example.com",
	}

	err := Validate(cfg)
	if err == nil {
		t.Fatal("expected validation error for mutual exclusivity, got nil")
	}
	if !strings.Contains(err.Error(), "mutually exclusive") {
		t.Errorf("expected 'mutually exclusive' in error, got: %v", err)
	}
}

func TestValidateThirdPartyOAuth2Config(t *testing.T) {
	tests := []struct {
		name    string
		cfg     *ports.ThirdPartyOAuth2Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid key",
			cfg: &ports.ThirdPartyOAuth2Config{
				JWESigningKey: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef")),
			},
			wantErr: false,
		},
		{
			name: "empty key",
			cfg: &ports.ThirdPartyOAuth2Config{
				JWESigningKey: "",
			},
			wantErr: true,
			errMsg:  "third_party_oauth2.jwe_signing_key",
		},
		{
			name: "invalid base64",
			cfg: &ports.ThirdPartyOAuth2Config{
				JWESigningKey: "not-valid-base64!@#$%",
			},
			wantErr: true,
			errMsg:  "valid base64-encoded string",
		},
		{
			name: "key too short - 16 bytes",
			cfg: &ports.ThirdPartyOAuth2Config{
				JWESigningKey: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef")),
			},
			wantErr: true,
			errMsg:  "exactly 32 bytes when decoded",
		},
		{
			name: "key too long - 64 bytes",
			cfg: &ports.ThirdPartyOAuth2Config{
				JWESigningKey: base64.StdEncoding.EncodeToString([]byte("0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef")),
			},
			wantErr: true,
			errMsg:  "exactly 32 bytes when decoded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateThirdPartyOAuth2Config(tt.cfg)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateThirdPartyOAuth2Config() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Check error message contains expected text
			if err != nil && tt.errMsg != "" {
				errStr := err.Error()
				if !strings.Contains(errStr, tt.errMsg) {
					t.Errorf("validateThirdPartyOAuth2Config() error = %v, want error containing %q", err, tt.errMsg)
				}
			}

			// Verify error is ConfigError type when expected
			if err != nil {
				if _, ok := err.(*config.ConfigError); !ok {
					t.Errorf("validateThirdPartyOAuth2Config() error type = %T, want *config.ConfigError", err)
				}
			}
		})
	}
}
