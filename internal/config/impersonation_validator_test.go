package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/oauth2/servermode"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
)

func validImpersonationConfig() *ports.ImpersonationConfig {
	return &ports.ImpersonationConfig{
		AudiencePrefix: "https://broker.example.com/impersonation",
		Rules: []ports.ImpersonationRuleConfig{{
			Name: "internal-gateway",
			Roles: map[string]ports.ImpersonationRoleConfig{
				"client_assertion": {ExpectedAudience: "https://broker.example.com/impersonation", PrincipalExpression: "client_assertion.sub"},
				"actor":            {ExpectedAudience: "https://broker.example.com/impersonation", PrincipalExpression: "actor_token.sub"},
				"subject":          {ExpectedAudience: "https://broker.example.com/impersonation", PrincipalExpression: "subject_token.sub", EmailExpression: "has(subject_token.email) ? subject_token.email : ''"},
			},
			TrustedIssuers: []ports.TrustedTokenIssuerConfig{{
				IssuerURI:         "https://idp.example.com",
				AllowedAlgorithms: []string{"ES256", "RS256"},
				SignsRoles:        []string{"client_assertion", "actor", "subject"},
			}},
			Authorization: ports.AuthorizationConfig{
				Type: "cel",
				CEL:  ports.CELAuthorizationConfig{Expression: `client_assertion.sub == "gw"`},
			},
		}},
	}
}

func TestValidateImpersonationConfig_Valid(t *testing.T) {
	err := validateImpersonationConfig(validImpersonationConfig(), servermode.Local, &ports.SecurityConfig{})
	require.NoError(t, err)
}

// validUnverifiedImpersonationConfig accepts an unverified subject: verification none, no
// expected_audience on the subject, subject absent from signs_roles (FR-003d/CR-008).
func validUnverifiedImpersonationConfig() *ports.ImpersonationConfig {
	cfg := validImpersonationConfig()
	subject := cfg.Rules[0].Roles["subject"]
	subject.Verification = ports.SubjectVerificationNone
	subject.ExpectedAudience = ""
	cfg.Rules[0].Roles["subject"] = subject
	cfg.Rules[0].TrustedIssuers[0].SignsRoles = []string{"client_assertion", "actor"}
	return cfg
}

func TestValidateImpersonationConfig_UnverifiedValid(t *testing.T) {
	err := validateImpersonationConfig(validUnverifiedImpersonationConfig(), servermode.Local, &ports.SecurityConfig{})
	require.NoError(t, err)
}

func TestValidateImpersonationConfig_Rejections(t *testing.T) {
	tests := []struct {
		name      string
		mode      servermode.Mode
		mutate    func(c *ports.ImpersonationConfig)
		wantField string
	}{
		{"proxy mode", servermode.Proxy, func(*ports.ImpersonationConfig) {}, "impersonation"},
		{"hybrid mode", servermode.Hybrid, func(*ports.ImpersonationConfig) {}, "impersonation"},
		{"empty audience prefix", servermode.Local, func(c *ports.ImpersonationConfig) { c.AudiencePrefix = "" }, "audience_prefix"},
		{"no rules", servermode.Local, func(c *ports.ImpersonationConfig) { c.Rules = nil }, "rules"},
		{"duplicate rule name", servermode.Local, func(c *ports.ImpersonationConfig) {
			c.Rules = append(c.Rules, c.Rules[0])
		}, "name"},
		{"missing subject role", servermode.Local, func(c *ports.ImpersonationConfig) {
			delete(c.Rules[0].Roles, "subject")
		}, "roles.subject"},
		{"unknown role key", servermode.Local, func(c *ports.ImpersonationConfig) {
			c.Rules[0].Roles["extra"] = ports.ImpersonationRoleConfig{PrincipalExpression: "x"}
		}, "roles"},
		{"missing principal_expression", servermode.Local, func(c *ports.ImpersonationConfig) {
			r := c.Rules[0].Roles["actor"]
			r.PrincipalExpression = ""
			c.Rules[0].Roles["actor"] = r
		}, "principal_expression"},
		{"email on non-subject role", servermode.Local, func(c *ports.ImpersonationConfig) {
			r := c.Rules[0].Roles["actor"]
			r.EmailExpression = "x"
			c.Rules[0].Roles["actor"] = r
		}, "email_expression"},
		{"missing expected_audience", servermode.Local, func(c *ports.ImpersonationConfig) {
			r := c.Rules[0].Roles["client_assertion"]
			r.ExpectedAudience = ""
			c.Rules[0].Roles["client_assertion"] = r
		}, "expected_audience"},
		{"unverified subject with expected_audience", servermode.Local, func(c *ports.ImpersonationConfig) {
			r := c.Rules[0].Roles["subject"]
			r.Verification = ports.SubjectVerificationNone
			c.Rules[0].Roles["subject"] = r
		}, "expected_audience"},
		{"invalid subject verification", servermode.Local, func(c *ports.ImpersonationConfig) {
			r := c.Rules[0].Roles["subject"]
			r.Verification = "sometimes"
			c.Rules[0].Roles["subject"] = r
		}, "verification"},
		{"unverified subject listed in signs_roles", servermode.Local, func(c *ports.ImpersonationConfig) {
			r := c.Rules[0].Roles["subject"]
			r.Verification = ports.SubjectVerificationNone
			r.ExpectedAudience = ""
			c.Rules[0].Roles["subject"] = r
		}, "signs_roles"},
		{"non-cel authorization", servermode.Local, func(c *ports.ImpersonationConfig) {
			c.Rules[0].Authorization.Type = "opa"
		}, "authorization.type"},
		{"empty authorization expression", servermode.Local, func(c *ports.ImpersonationConfig) {
			c.Rules[0].Authorization.CEL.Expression = ""
		}, "authorization.cel.expression"},
		{"empty allowed_algorithms", servermode.Local, func(c *ports.ImpersonationConfig) {
			c.Rules[0].TrustedIssuers[0].AllowedAlgorithms = nil
		}, "allowed_algorithms"},
		{"symmetric algorithm", servermode.Local, func(c *ports.ImpersonationConfig) {
			c.Rules[0].TrustedIssuers[0].AllowedAlgorithms = []string{"HS256"}
		}, "allowed_algorithms"},
		{"duplicate issuer_uri", servermode.Local, func(c *ports.ImpersonationConfig) {
			c.Rules[0].TrustedIssuers = append(c.Rules[0].TrustedIssuers, c.Rules[0].TrustedIssuers[0])
		}, "issuer_uri"},
		{"uncovered signed role", servermode.Local, func(c *ports.ImpersonationConfig) {
			c.Rules[0].TrustedIssuers[0].SignsRoles = []string{"client_assertion", "actor"}
		}, "trusted_issuers"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			cfg := validImpersonationConfig()
			tc.mutate(cfg)
			err := validateImpersonationConfig(cfg, tc.mode, &ports.SecurityConfig{})
			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantField)
		})
	}
}
