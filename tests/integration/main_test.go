package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/tests/integration/bootstrap"
)

// sharedLS is the single LocalStack container shared across all tests in this package.
// Initialised in TestMain; nil if Docker is unavailable (tests that need it will skip).
var sharedLS *bootstrap.LocalStackContainer

func TestMain(m *testing.M) {
	ctx := context.Background()

	ls, err := bootstrap.StartLocalStackForSuite(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "LocalStack unavailable, skipping container-dependent tests: %v\n", err)
	} else {
		sharedLS = ls
	}

	code := m.Run()

	if sharedLS != nil {
		_ = sharedLS.Terminate(ctx)
		sharedLS.CleanupLocalStackEnvironment()
	}

	os.Exit(code)
}
