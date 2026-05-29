// Package bootstrap provides test infrastructure for E2E testing.
package bootstrap

import (
	"io"
	"log/slog"
	"os"
)

// TestLogger returns a logger that writes to os.Stderr when E2E_VERBOSE=1,
// and discards all output otherwise.
// Use this in every test suite's BeforeSuite to avoid I/O contention under
// parallel Ginkgo procs.
func TestLogger(level slog.Level) *slog.Logger {
	w := io.Discard
	if os.Getenv("E2E_VERBOSE") == "1" {
		w = os.Stderr
	}
	return slog.New(slog.NewTextHandler(w, &slog.HandlerOptions{Level: level}))
}
