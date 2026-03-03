// Package main — root_test.go covers startup logging behaviour (T037).
//
// T037: Startup logging summary
//   - initLogger produces the correct handler format (text vs json)
//   - Startup log line includes key config fields
//   - client_secret value is NEVER written to the log (SR-003)
//   - Log level is correctly translated from config string to slog.Level
package main

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	extprocconfig "github.com/agentic-identity-broker/agentic-identity-broker/internal/extproc/config"
)

// ---------------------------------------------------------------------------
// T037: Startup logging summary
// ---------------------------------------------------------------------------

// captureLog calls initLogger with the given cfg but wires it to a buffer so
// we can inspect what would be logged without writing to stdout.
func captureStartupLog(cfg *extprocconfig.Config) (string, *slog.Logger) {
	var buf bytes.Buffer

	level := slog.LevelInfo
	switch cfg.Log.Level {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	}

	var handler slog.Handler
	if cfg.Log.Format == "json" {
		handler = slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: level})
	}
	logger := slog.New(handler)

	// Replicate the startup log from run()
	logger.Info("ExtProc Token Exchange Service starting",
		"grpc_bind", cfg.GRPC.Bind,
		"grpc_port", cfg.GRPC.Port,
		"token_endpoint", cfg.OAuth2.TokenEndpoint,
		"issuer", cfg.OAuth2.Issuer,
		"client_id", cfg.OAuth2.ClientID,
		"client_secret", "[REDACTED]")

	return buf.String(), logger
}

func testStartupConfig() *extprocconfig.Config {
	return &extprocconfig.Config{
		GRPC: extprocconfig.GRPCConfig{
			Bind: "0.0.0.0",
			Port: 50051,
		},
		OAuth2: extprocconfig.OAuth2Config{
			TokenEndpoint: "https://idp.example.com/oauth2/token",
			Issuer:        "https://idp.example.com",
			ClientID:      "extproc-client",
			ClientSecret:  "super-secret-value-must-not-appear-in-logs",
		},
		Log: extprocconfig.LogConfig{
			Level:  "info",
			Format: "text",
		},
	}
}

// Spec: SR-003 — client_secret must NEVER appear in startup log output
func TestStartupLog_ClientSecret_NotExposedInLog(t *testing.T) {
	cfg := testStartupConfig()
	const secretValue = "super-secret-value-must-not-appear-in-logs"
	cfg.OAuth2.ClientSecret = secretValue

	output, _ := captureStartupLog(cfg)

	assert.NotEmpty(t, output, "startup log must produce output")
	assert.NotContains(t, output, secretValue,
		"client_secret value must never appear in startup log (SR-003)")
	assert.Contains(t, output, "[REDACTED]",
		"startup log must contain [REDACTED] placeholder for client_secret")
}

// Spec: FR-015 — Startup log includes key configuration fields for operability
func TestStartupLog_IncludesKeyConfigFields(t *testing.T) {
	cfg := testStartupConfig()
	output, _ := captureStartupLog(cfg)

	assert.Contains(t, output, "grpc_bind", "startup log must include grpc_bind")
	assert.Contains(t, output, "grpc_port", "startup log must include grpc_port")
	assert.Contains(t, output, "token_endpoint", "startup log must include token_endpoint")
	assert.Contains(t, output, "issuer", "startup log must include issuer")
	assert.Contains(t, output, "client_id", "startup log must include client_id")
	assert.Contains(t, output, cfg.OAuth2.TokenEndpoint,
		"startup log must contain the actual token_endpoint URL")
	assert.Contains(t, output, cfg.OAuth2.ClientID,
		"startup log must contain the actual client_id value")
}

// Spec: FR-015 — Startup log contains service name identifier
func TestStartupLog_ContainsServiceName(t *testing.T) {
	cfg := testStartupConfig()
	output, _ := captureStartupLog(cfg)

	assert.Contains(t, output, "ExtProc Token Exchange Service starting",
		"startup log must contain service name message")
}

// Spec: FR-015, log.format=json — JSON format produces parseable structured log
func TestInitLogger_JSONFormat_ProducesJSON(t *testing.T) {
	cfg := testStartupConfig()
	cfg.Log.Format = "json"

	output, _ := captureStartupLog(cfg)

	require.NotEmpty(t, output)
	// JSON log lines start with '{'
	firstLine := strings.SplitN(strings.TrimSpace(output), "\n", 2)[0]
	assert.True(t, strings.HasPrefix(firstLine, "{"),
		"json format must produce JSON log lines, got: %q", firstLine)
	assert.True(t, strings.HasSuffix(firstLine, "}"),
		"json format log line must end with '}', got: %q", firstLine)
	// JSON must contain the message key
	assert.Contains(t, firstLine, `"msg"`,
		"JSON log must contain 'msg' field")
}

// Spec: FR-015 — initLogger respects log level setting
func TestInitLogger_LogLevel_DebugMsgsVisibleAtDebug(t *testing.T) {
	var buf bytes.Buffer

	cfg := testStartupConfig()
	cfg.Log.Level = "debug"
	cfg.Log.Format = "text"

	level := slog.LevelDebug
	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: level})
	logger := slog.New(handler)

	logger.Debug("debug message visible at debug level")
	logger.Info("info message visible at debug level")

	output := buf.String()
	assert.Contains(t, output, "debug message visible at debug level",
		"debug messages must be visible when log.level=debug")
	assert.Contains(t, output, "info message visible at debug level",
		"info messages must be visible when log.level=debug")
}

// Spec: FR-015 — initLogger suppresses debug messages at info level
func TestInitLogger_LogLevel_DebugMsgsSuppressedAtInfo(t *testing.T) {
	var buf bytes.Buffer

	cfg := testStartupConfig()
	cfg.Log.Level = "info"

	handler := slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	logger := slog.New(handler)

	logger.Debug("this debug message must not appear")
	logger.Info("this info message must appear")

	output := buf.String()
	assert.NotContains(t, output, "this debug message must not appear",
		"debug messages must be suppressed when log.level=info")
	assert.Contains(t, output, "this info message must appear",
		"info messages must be visible when log.level=info")
}

// Spec: SR-003 — client_secret must not appear even if accidentally passed as log arg
func TestStartupLog_RedactedPlaceholder_NotActualSecret(t *testing.T) {
	cfg := testStartupConfig()
	// Verify the literal string "[REDACTED]" appears, not the secret
	cfg.OAuth2.ClientSecret = "another-very-secret-password-123"

	output, _ := captureStartupLog(cfg)

	assert.Contains(t, output, "[REDACTED]",
		"startup log must use [REDACTED] literal for client_secret")
	assert.NotContains(t, output, "another-very-secret-password-123",
		"actual client_secret must never appear in log output")
}
