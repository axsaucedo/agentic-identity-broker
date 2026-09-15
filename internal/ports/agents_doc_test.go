package ports

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPortsAgentsDocumentsApprovedEncryptionContextSubjects(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "resolve caller path")

	content, err := os.ReadFile(filepath.Join(filepath.Dir(thisFile), "AGENTS.md"))
	require.NoError(t, err)

	agents := string(content)
	assert.Contains(t, agents, "service_id", "ports AGENTS must document service-scoped encryption contexts")
	assert.Contains(t, agents, "kid", "ports AGENTS must document signing-key encryption contexts")
}
