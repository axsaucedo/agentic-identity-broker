//go:build integration
// +build integration

package integration

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/agentic-identity-broker/agentic-identity-broker/tests/integration/bootstrap"
)

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
