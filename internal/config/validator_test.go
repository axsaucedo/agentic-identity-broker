package config

import (
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

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
			name: "valid config",
			cfg: &ports.Config{
				Log: ports.LogConfig{
					Level:  ports.LogLevelInfo,
					Format: ports.LogFormatText,
				},
				Storage: ports.StorageConfig{
					Backend: "memory",
					Timeouts: ports.StorageTimeouts{
						Read:  5 * time.Second,
						Write: 10 * time.Second,
					},
				},
			},
			wantErr: false,
		},
		{
			name: "invalid log level",
			cfg: &ports.Config{
				Log: ports.LogConfig{
					Level:  ports.LogLevel("invalid"),
					Format: ports.LogFormatText,
				},
			},
			wantErr: true,
		},
		{
			name: "invalid log format",
			cfg: &ports.Config{
				Log: ports.LogConfig{
					Level:  ports.LogLevelInfo,
					Format: ports.LogFormat("invalid"),
				},
			},
			wantErr: true,
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
