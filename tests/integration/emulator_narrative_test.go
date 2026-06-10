package integration

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCurrentEncryptionDocsUseUnifiedAWSemulatorNarrative(t *testing.T) {
	t.Run("vendor-neutral example config exists", func(t *testing.T) {
		_, err := os.Stat(repoPath(t, "examples/config/config.aws-emulator.yaml"))
		require.NoError(t, err, "expected vendor-neutral AWS emulator example config")

		_, err = os.Stat(repoPath(t, "examples/config/config.aws-localstack.yaml"))
		assert.ErrorIs(t, err, os.ErrNotExist, "vendor-branded LocalStack example should be retired")
	})

	cases := []struct {
		path           string
		mustContain    []string
		mustNotContain []string
	}{
		{
			path:           "docs/ENCRYPTION_INTEGRATION_GUIDE.md",
			mustContain:    []string{"LocalStack-compatible AWS emulator"},
			mustNotContain: []string{"with LocalStack for testing"},
		},
		{
			path:           "examples/config/config.aws-emulator.yaml",
			mustContain:    []string{"LocalStack-compatible AWS emulator", "floci/floci", "localstack/localstack"},
			mustNotContain: []string{"Use this configuration for local development and testing with LocalStack"},
		},
		{
			path:           "examples/config/config.aws-production-advanced.yaml",
			mustContain:    []string{"AWS emulator scenarios"},
			mustNotContain: []string{"development/LocalStack scenarios"},
		},
		{
			path:           "internal/ports/config.go",
			mustContain:    []string{"LocalStack-compatible AWS emulator", "Use explicit localhost origins during development"},
			mustNotContain: []string{"testing with LocalStack", "development/testing with LocalStack", "allow all origins"},
		},
		{
			path:           "internal/adapters/encryption/aws/config.go",
			mustContain:    []string{"LocalStack-compatible AWS emulator"},
			mustNotContain: []string{"For LocalStack or custom AWS implementations", "For LocalStack only", "LocalStack testing"},
		},
		{
			path:           "docs/operations/deployment-checklist.md",
			mustContain:    []string{"AWS emulator testing"},
			mustNotContain: []string{"LocalStack testing"},
		},
		{
			path:           "specs/012-aws-encryption-vault/contracts/encryption-port.md",
			mustContain:    []string{"AWS emulator"},
			mustNotContain: []string{"testcontainers with LocalStack"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			content, err := os.ReadFile(repoPath(t, tc.path))
			require.NoError(t, err)

			text := string(content)
			for _, want := range tc.mustContain {
				assert.Contains(t, text, want)
			}
			for _, forbidden := range tc.mustNotContain {
				assert.NotContains(t, text, forbidden)
			}
		})
	}
}

func repoPath(t *testing.T, relativePath string) string {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok, "resolve caller path")

	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	return filepath.Join(repoRoot, relativePath)
}
