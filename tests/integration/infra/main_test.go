//go:build integration
// +build integration

package integration

import (
	"context"
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	code := m.Run()

	if sharedLS != nil {
		_ = sharedLS.Terminate(ctx)
		sharedLS.CleanupLocalStackEnvironment()
	}

	os.Exit(code)
}
