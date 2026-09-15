// Package main is the entry point for the MCP server mock.
// It implements the MCP 2025-11-05 protocol over HTTP (Streamable HTTP transport)
// for use in the Docker Compose integration environment.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/agentic-identity-broker/mock-mcp-server/internal/config"
	"github.com/agentic-identity-broker/mock-mcp-server/internal/server"
)

func main() {
	// Resolve config directory (same convention as other mocks).
	configDir := resolveConfigDir()

	absPath, err := filepath.Abs(configDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to resolve config path: %v\n", err)
		os.Exit(1)
	}

	cfg, err := config.Load(absPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	slog.Info("MCP mock server config loaded", "port", cfg.Server.Port, "bind", cfg.Server.Bind)

	srv := server.New(cfg)

	// Graceful shutdown on SIGINT / SIGTERM.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := srv.Start(); err != nil {
			fmt.Fprintf(os.Stderr, "server error: %v\n", err)
			os.Exit(1)
		}
	}()

	<-sigCh
	slog.Info("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "graceful shutdown failed: %v\n", err)
		os.Exit(1)
	}

	slog.Info("MCP mock server stopped")
}

// resolveConfigDir follows the same convention as other mocks:
// prefer the command-line argument, then the current directory, then the
// project-root-relative path.
func resolveConfigDir() string {
	if len(os.Args) > 1 {
		return os.Args[1]
	}
	if _, err := os.Stat("config.yaml"); err == nil {
		return "."
	}
	return "mocks/mcp-server"
}
