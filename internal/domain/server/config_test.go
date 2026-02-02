package server

import (
	"strings"
	"testing"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

// TestServerConfigValidation tests server configuration validation with table-driven tests.
func TestServerConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  ports.ServerConfig
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid default configuration",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "valid custom ports",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 3000,
					Bind: "127.0.0.1",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 3001,
					Bind: "127.0.0.1",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 10 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "valid IPv4 addresses",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "0.0.0.0",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "192.168.1.1",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "valid IPv6 addresses",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::1",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "2001:db8::1",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "invalid port: zero",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 0,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "invalid port: negative",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: -1,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "invalid port: too large",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 65536,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "port must be between 1 and 65535",
		},
		{
			name: "duplicate ports",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "ports must be different",
		},
		{
			name: "invalid bind address: bad hostname",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "invalid_@hostname",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "bind address invalid",
		},
		{
			name: "invalid bind address: bad IPv6",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::gggg",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "bind address invalid",
		},
		{
			name: "invalid timeout: zero",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 0,
				},
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "invalid timeout: negative",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: -1 * time.Second,
				},
			},
			wantErr: true,
			errMsg:  "timeout must be positive",
		},
		{
			name: "invalid timeout: too large",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 10 * time.Minute,
				},
			},
			wantErr: true,
			errMsg:  "timeout must be at most 5 minutes",
		},
		{
			name: "valid timeout: minimum boundary",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 1 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "valid timeout: maximum boundary",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 8000,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 14000,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 5 * time.Minute,
				},
			},
			wantErr: false,
		},
		{
			name: "valid port boundaries: port 1",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 1,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 2,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: false,
		},
		{
			name: "valid port boundaries: port 65535",
			config: ports.ServerConfig{
				EndUser: ports.ServerInstanceConfig{
					Port: 65534,
					Bind: "::",
				},
				Admin: ports.ServerInstanceConfig{
					Port: 65535,
					Bind: "::",
				},
				Shutdown: ports.ShutdownConfig{
					Timeout: 30 * time.Second,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateServerConfig(&tt.config)
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateServerConfig() expected error containing %q, got nil", tt.errMsg)
					return
				}
				if tt.errMsg != "" && !strings.Contains(err.Error(), tt.errMsg) {
					t.Errorf("ValidateServerConfig() error = %q, want error containing %q", err.Error(), tt.errMsg)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateServerConfig() unexpected error = %v", err)
				}
			}
		})
	}
}
