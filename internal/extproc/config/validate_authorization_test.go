package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
)

// validConfigWithAuthorization returns a valid Config with authorization enabled
// and a policy file path set (required when enabled=true).
// Uses t.TempDir() to ensure cleanup after the test.
func validConfigWithAuthorization(t *testing.T) *config.Config {
	t.Helper()
	cfg := validConfig()

	// Create a real temp file for the policy path so os.Stat passes.
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "authz.rego")
	if err := os.WriteFile(tmpFile, []byte("package test\n"), 0o600); err != nil {
		t.Fatalf("cannot create temp policy file: %v", err)
	}

	cfg.Authorization = config.AuthorizationConfig{
		Enabled: true,
		Policy: config.PolicyConfig{
			Path:     tmpFile,
			Package:  "aib.extproc.authz",
			Decision: "result",
		},
		DefaultDecision:   "deny",
		EvaluationTimeout: 100 * time.Millisecond,
		MaxBodySize:       1048576,
	}
	return cfg
}

func TestValidateAuthorization(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*config.Config)
		wantErr     bool
		errContains string
	}{
		{
			name: "authorization disabled with no policy — valid",
			mutate: func(c *config.Config) {
				c.Authorization.Enabled = false
				c.Authorization.Policy.Path = ""
				c.Authorization.Policy.ConfigFile = ""
			},
			wantErr: false,
		},
		{
			name: "authorization disabled with policy.path — error",
			mutate: func(c *config.Config) {
				c.Authorization.Enabled = false
			},
			wantErr:     true,
			errContains: "authorization.policy.path or authorization.policy.config_file is set but authorization.enabled is false",
		},
		{
			name: "authorization disabled with policy.config_file — error",
			mutate: func(c *config.Config) {
				c.Authorization.Enabled = false
				c.Authorization.Policy.ConfigFile = c.Authorization.Policy.Path
				c.Authorization.Policy.Path = ""
			},
			wantErr:     true,
			errContains: "authorization.policy.path or authorization.policy.config_file is set but authorization.enabled is false",
		},
		{
			name:    "authorization enabled with policy.path — valid",
			mutate:  func(_ *config.Config) {},
			wantErr: false,
		},
		{
			name: "authorization enabled with policy.config_file — valid",
			mutate: func(c *config.Config) {
				// Use the existing temp file from cfg as config_file
				c.Authorization.Policy.ConfigFile = c.Authorization.Policy.Path
				c.Authorization.Policy.Path = ""
			},
			wantErr: false,
		},
		// Rule A1: enabled but no policy source
		{
			name: "ruleA1: enabled but neither path nor config_file — error",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Path = ""
				c.Authorization.Policy.ConfigFile = ""
			},
			wantErr:     true,
			errContains: "authorization.policy.path or authorization.policy.config_file must be set when authorization is enabled",
		},
		// Rule A2: path and config_file are mutually exclusive
		{
			name: "ruleA2: both path and config_file set — error",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Path = "/etc/extproc/policies/authz.rego"
				c.Authorization.Policy.ConfigFile = "/etc/extproc/opa-config.yaml"
			},
			wantErr:     true,
			errContains: "authorization.policy.path and authorization.policy.config_file are mutually exclusive",
		},
		// Rule A3: path traversal in policy.path
		{
			name: "ruleA3: path traversal in policy.path — error",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Path = "/etc/extproc/../../etc/passwd"
			},
			wantErr:     true,
			errContains: "authorization.policy.path must not contain path traversal",
		},
		// Rule A4: path traversal in policy.config_file
		{
			name: "ruleA4: path traversal in policy.config_file — error",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Path = ""
				c.Authorization.Policy.ConfigFile = "/etc/../etc/shadow"
			},
			wantErr:     true,
			errContains: "authorization.policy.config_file must not contain path traversal",
		},
		// Rule A5: default_decision must remain fail-closed
		{
			name: "ruleA5: invalid default_decision — error",
			mutate: func(c *config.Config) {
				c.Authorization.DefaultDecision = "permit"
			},
			wantErr:     true,
			errContains: "authorization.default_decision must be deny",
		},
		{
			name: "ruleA5: default_decision allow — error",
			mutate: func(c *config.Config) {
				c.Authorization.DefaultDecision = "allow"
			},
			wantErr:     true,
			errContains: "authorization.default_decision must be deny",
		},
		{
			name: "ruleA5: default_decision deny with policy.config_file — valid",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.ConfigFile = c.Authorization.Policy.Path
				c.Authorization.Policy.Path = ""
				c.Authorization.DefaultDecision = "deny"
			},
			wantErr: false,
		},
		// Rule A6: evaluation_timeout must be positive when authorization is enabled
		{
			name: "ruleA6: evaluation_timeout zero — error",
			mutate: func(c *config.Config) {
				c.Authorization.EvaluationTimeout = 0
			},
			wantErr:     true,
			errContains: "authorization.evaluation_timeout must be a positive duration",
		},
		{
			name: "ruleA6: evaluation_timeout negative — error",
			mutate: func(c *config.Config) {
				c.Authorization.EvaluationTimeout = -1 * time.Millisecond
			},
			wantErr:     true,
			errContains: "authorization.evaluation_timeout must be a positive duration",
		},
		// Rule A7: max_body_size must be positive when authorization is enabled
		{
			name: "ruleA7: max_body_size zero — error",
			mutate: func(c *config.Config) {
				c.Authorization.MaxBodySize = 0
			},
			wantErr:     true,
			errContains: "authorization.max_body_size must be a positive value",
		},
		{
			name: "ruleA7: max_body_size negative — error",
			mutate: func(c *config.Config) {
				c.Authorization.MaxBodySize = -1
			},
			wantErr:     true,
			errContains: "authorization.max_body_size must be a positive value",
		},
		// Rule A8: policy.package must not be empty when authorization is enabled
		{
			name: "ruleA8: policy.package empty — error",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Package = ""
			},
			wantErr:     true,
			errContains: "authorization.policy.package must not be empty when authorization is enabled",
		},
		// Rule A9: policy.decision must not be empty when authorization is enabled
		{
			name: "ruleA9: policy.decision empty — error",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Decision = ""
			},
			wantErr:     true,
			errContains: "authorization.policy.decision must not be empty when authorization is enabled",
		},
		// File existence: policy.path points to a non-existent file
		{
			name: "file existence: policy.path does not exist — error",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Path = "/tmp/nonexistent-policy-42.rego"
			},
			wantErr:     true,
			errContains: "path not found or not readable",
		},
		// File existence: policy.config_file points to a non-existent file
		{
			name: "file existence: policy.config_file does not exist — error",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Path = ""
				c.Authorization.Policy.ConfigFile = "/tmp/nonexistent-opa-config-42.yaml"
			},
			wantErr:     true,
			errContains: "path not found or not readable",
		},
		// File existence: policy.path is a directory — valid (policy.path supports directories)
		{
			name: "file existence: policy.path is a directory — ok",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Path = os.TempDir()
			},
			wantErr: false,
		},
		// File existence: policy.config_file is a directory — error (config_file must be a file)
		{
			name: "file existence: policy.config_file is a directory — error",
			mutate: func(c *config.Config) {
				c.Authorization.Policy.Path = ""
				c.Authorization.Policy.ConfigFile = os.TempDir()
			},
			wantErr:     true,
			errContains: "expected a file but got a directory",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfigWithAuthorization(t)
			tt.mutate(cfg)

			err := config.Validate(cfg)
			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.True(t, strings.Contains(err.Error(), tt.errContains),
						"expected error to contain %q, got: %s", tt.errContains, err.Error())
				}
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// TestAuthorizationDefaults verifies that authorization is disabled by default.
func TestAuthorizationDefaults(t *testing.T) {
	cfg := validConfig()
	assert.False(t, cfg.Authorization.Enabled)
	require.NoError(t, config.Validate(cfg))
}
