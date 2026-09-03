// Package bootstrap provides test infrastructure for E2E testing.
package bootstrap

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"os"
	"strings"
	"sync"
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

// BufferedLogCapture captures structured log output for E2E assertions.
type BufferedLogCapture struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

// NewBufferedJSONLogger returns a JSON logger and in-memory capture for request-level assertions.
func NewBufferedJSONLogger(level slog.Level) (*slog.Logger, *BufferedLogCapture) {
	capture := &BufferedLogCapture{}
	return slog.New(slog.NewJSONHandler(capture, &slog.HandlerOptions{Level: level})), capture
}

// Write appends log bytes to the in-memory capture.
func (c *BufferedLogCapture) Write(p []byte) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.Write(p)
}

// Reset clears the buffered log output.
func (c *BufferedLogCapture) Reset() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.buf.Reset()
}

// Raw returns the captured log output as a string.
func (c *BufferedLogCapture) Raw() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.buf.String()
}

// Records parses each captured JSON log line into a map.
func (c *BufferedLogCapture) Records() ([]map[string]any, error) {
	raw := strings.TrimSpace(c.Raw())
	if raw == "" {
		return nil, nil
	}

	lines := strings.Split(raw, "\n")
	records := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}

		record := map[string]any{}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return nil, err
		}
		records = append(records, record)
	}

	return records, nil
}
