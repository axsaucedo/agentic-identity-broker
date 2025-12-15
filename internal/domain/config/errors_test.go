package config

import (
	"errors"
	"strings"
	"testing"
)

func TestConfigError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *ConfigError
		wantMsg string
	}{
		{
			name: "complete error",
			err: &ConfigError{
				Field:    "log.level",
				Value:    "invalid",
				Expected: "debug, info, warn, or error",
				Source:   "cli flag --log-level",
			},
			wantMsg: "configuration error: invalid value \"invalid\" for field \"log.level\" (expected: debug, info, warn, or error) [source: cli flag --log-level]",
		},
		{
			name: "error with wrapped error",
			err: &ConfigError{
				Field:    "config_file",
				Value:    "/path/to/config.yaml",
				Expected: "valid YAML syntax",
				Err:      errors.New("yaml: line 5: mapping values not allowed"),
			},
			wantMsg: "configuration error: invalid value \"/path/to/config.yaml\" for field \"config_file\" (expected: valid YAML syntax): yaml: line 5: mapping values not allowed",
		},
		{
			name: "minimal error",
			err: &ConfigError{
				Field: "log.level",
				Value: "invalid",
			},
			wantMsg: "configuration error: invalid value \"invalid\" for field \"log.level\"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("ConfigError.Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestConfigError_Unwrap(t *testing.T) {
	innerErr := errors.New("inner error")
	configErr := &ConfigError{
		Field: "test",
		Err:   innerErr,
	}

	unwrapped := configErr.Unwrap()
	if unwrapped != innerErr {
		t.Errorf("ConfigError.Unwrap() = %v, want %v", unwrapped, innerErr)
	}

	// Test error chaining with errors.Is
	if !errors.Is(configErr, innerErr) {
		t.Error("errors.Is() should work with wrapped errors")
	}
}

func TestConfigError_NoField(t *testing.T) {
	err := &ConfigError{
		Expected: "valid configuration",
	}

	got := err.Error()
	if !strings.Contains(got, "configuration error") {
		t.Errorf("ConfigError.Error() = %q, should contain 'configuration error'", got)
	}
}
