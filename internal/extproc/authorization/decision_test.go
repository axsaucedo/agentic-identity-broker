package authorization_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/authorization"
)

func TestParseDecision_Allow(t *testing.T) {
	result := map[string]any{"action": "allow"}
	d := authorization.ParseDecision(result)
	assert.Equal(t, "allow", d.Action)
	assert.Empty(t, d.Reasons)
}

func TestParseDecision_DenyWithReasons(t *testing.T) {
	result := map[string]any{
		"action":  "deny",
		"reasons": []any{"r1", "r2"},
	}
	d := authorization.ParseDecision(result)
	assert.Equal(t, "deny", d.Action)
	assert.Equal(t, []string{"r1", "r2"}, d.Reasons)
}

func TestParseDecision_ApprovalRequired_MappedToDeny(t *testing.T) {
	result := map[string]any{"action": "approval_required"}
	d := authorization.ParseDecision(result)
	assert.Equal(t, "deny", d.Action)
}

func TestParseDecision_CIBARequired_MappedToDeny(t *testing.T) {
	result := map[string]any{"action": "ciba_required"}
	d := authorization.ParseDecision(result)
	assert.Equal(t, "deny", d.Action)
}

func TestParseDecision_NilResult_Deny(t *testing.T) {
	d := authorization.ParseDecision(nil)
	assert.Equal(t, "deny", d.Action)
}

func TestParseDecision_EmptyResult_Deny(t *testing.T) {
	d := authorization.ParseDecision(map[string]any{})
	assert.Equal(t, "deny", d.Action)
}
