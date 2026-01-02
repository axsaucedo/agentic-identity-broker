// Package main is the entry point for the Agentic Identity Broker application.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	httpAdapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/http"
	storageAdapter "github.com/agentic-identity-broker/agentic-identity-broker/internal/adapters/storage"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/server"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "agentic-identity-broker",
	Short: "Agentic Identity Broker - Secure identity management for AI agents",
	Long: `Agentic Identity Broker provides secure identity management,
authentication, and authorization for AI agents and autonomous systems.`,
	RunE: run,
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

// run is the main execution function for the application.
func run(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create configuration loader
	loader := config.NewLoader()
	loader.SetCommand(cmd)

	// Load configuration
	cfg, err := loader.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Emit audit log
	emitAuditLog(loader)

	// Display startup summary
	displayStartupSummary(loader, cfg)

	// Initialize logger
	logger := initializeLogger(cfg.Log)
	logger.Info("Agentic Identity Broker starting",
		"log_level", cfg.Log.Level,
		"log_format", cfg.Log.Format)

	// Initialize storage adapter
	storage, err := storageAdapter.NewAdapter(&cfg.Storage)
	if err != nil {
		return fmt.Errorf("failed to create storage adapter: %w", err)
	}

	// Create server instances
	enduserServer := httpAdapter.NewServer("enduser", cfg.Server.EndUser, logger)
	adminServer := httpAdapter.NewServer("admin", cfg.Server.Admin, logger)

	// Attach repositories to servers for consent management
	// Both servers need these to handle consent-related requests
	if storage.Agents() != nil {
		enduserServer.SetAgentRepository(storage.Agents())
		adminServer.SetAgentRepository(storage.Agents())
	}
	if storage.Services() != nil {
		enduserServer.SetServiceRepository(storage.Services())
		adminServer.SetServiceRepository(storage.Services())
	}
	if storage.UserGrants() != nil {
		enduserServer.SetGrantRepository(storage.UserGrants())
		adminServer.SetGrantRepository(storage.UserGrants())
	}
	if storage.UserSessions() != nil {
		enduserServer.SetSessionRepository(storage.UserSessions())
	}

	// Set OAuth2 configuration for enduser server (only enduser server needs this)
	// Constitution Principle VII (Configuration-Driven Design) compliance
	// The configuration includes JWESigningKey, StateTokenTTL, and PKCEVerifierLength
	enduserServer.SetThirdPartyOAuth2Config(cfg.ThirdPartyOAuth2)

	// Create server manager
	mgr := server.NewManager(enduserServer, adminServer, cfg.Server.Shutdown.Timeout, logger)

	// Setup signal handling for graceful shutdown
	sigCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start servers (blocking)
	logger.Info("Starting dual-port HTTP servers",
		"enduser_port", cfg.Server.EndUser.Port,
		"admin_port", cfg.Server.Admin.Port)

	// Run servers in a goroutine
	var startErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		startErr = mgr.Start(sigCtx)
	}()

	// Wait for either signal or server completion
	<-sigCtx.Done()

	// Signal received, initiate graceful shutdown
	logger.Info("Shutdown signal received, initiating graceful shutdown")
	if err := mgr.Shutdown(context.Background()); err != nil {
		logger.Error("Shutdown error", "error", err)
		return fmt.Errorf("shutdown error: %w", err)
	}

	// Wait for Start() to finish
	<-done
	if startErr != nil && startErr != context.Canceled {
		logger.Error("Server error", "error", startErr)
		return fmt.Errorf("server error: %w", startErr)
	}

	logger.Info("Servers shut down successfully")
	return nil
}

// initializeLogger creates a structured logger based on configuration.
func initializeLogger(logCfg ports.LogConfig) *slog.Logger {
	var level slog.Level
	switch logCfg.Level {
	case ports.LogLevelDebug:
		level = slog.LevelDebug
	case ports.LogLevelInfo:
		level = slog.LevelInfo
	case ports.LogLevelWarn:
		level = slog.LevelWarn
	case ports.LogLevelError:
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	var handler slog.Handler
	if logCfg.Format == ports.LogFormatJSON {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}

// displayStartupSummary shows configuration details at startup.
// Displays all keys, values (redacted if sensitive), and sources.
func displayStartupSummary(loader *config.Loader, cfg interface{}) {
	fmt.Println("=== Configuration Summary ===")

	// Get all sources
	sources := loader.GetSources()

	// Build a map of keys to their source for display
	keyToSource := make(map[string]string)
	for _, source := range sources {
		for _, key := range source.Keys {
			// Higher precedence wins (last one in the slice)
			keyToSource[key] = formatSource(source)
		}
	}

	// Display configuration values
	if c, ok := cfg.(*ports.Config); ok {
		displayConfigValue("log.level", string(c.Log.Level), keyToSource["log.level"])
		displayConfigValue("log.format", string(c.Log.Format), keyToSource["log.format"])
		displayConfigValue("server.enduser.port", fmt.Sprintf("%d", c.Server.EndUser.Port), keyToSource["server.enduser.port"])
		displayConfigValue("server.enduser.bind", c.Server.EndUser.Bind, keyToSource["server.enduser.bind"])
		displayConfigValue("server.admin.port", fmt.Sprintf("%d", c.Server.Admin.Port), keyToSource["server.admin.port"])
		displayConfigValue("server.admin.bind", c.Server.Admin.Bind, keyToSource["server.admin.bind"])
		displayConfigValue("server.shutdown.timeout", c.Server.Shutdown.Timeout.String(), keyToSource["server.shutdown.timeout"])
	}

	fmt.Println("\n=== Configuration Sources ===")
	for _, source := range sources {
		fmt.Printf("  [%d] %s: %s (loaded at %s)\n",
			source.Precedence,
			source.Type,
			source.Path,
			source.LoadedAt.Format(time.RFC3339))
	}
	fmt.Println()
}

// displayConfigValue displays a single configuration value with redaction and source.
func displayConfigValue(key string, value string, source string) {
	displayValue := config.Redact(key, value)
	if source == "" {
		source = "default"
	}
	fmt.Printf("  %s: %v [source: %s]\n", key, displayValue, source)
}

// formatSource formats a ConfigSource for display.
func formatSource(source ports.ConfigSource) string {
	switch source.Type {
	case ports.SourceTypeDefault:
		return "default"
	case ports.SourceTypeEnvFile:
		return fmt.Sprintf("env (%s)", source.Path)
	case ports.SourceTypeYAML:
		return fmt.Sprintf("yaml (%s)", source.Path)
	case ports.SourceTypeCLI:
		return "cli"
	default:
		return string(source.Type)
	}
}

// emitAuditLog outputs a structured JSON audit log to stdout.
// Includes all configuration sources, keys loaded, and redacted keys.
func emitAuditLog(loader *config.Loader) {
	sources := loader.GetSources()

	// Build audit log structure
	auditLog := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"event":     "configuration_loaded",
		"sources":   make([]map[string]interface{}, 0, len(sources)),
	}

	// Collect all keys and identify redacted ones
	allKeys := make(map[string]bool)
	redactedKeys := make([]string, 0)

	for _, source := range sources {
		sourceData := map[string]interface{}{
			"type":       source.Type,
			"path":       source.Path,
			"precedence": source.Precedence,
			"loaded_at":  source.LoadedAt.Format(time.RFC3339),
			"keys":       source.Keys,
		}
		auditLog["sources"] = append(auditLog["sources"].([]map[string]interface{}), sourceData)

		// Track all keys
		for _, key := range source.Keys {
			allKeys[key] = true
			if config.IsSensitive(key) {
				redactedKeys = append(redactedKeys, key)
			}
		}
	}

	// Convert keys map to sorted slice
	keysSlice := make([]string, 0, len(allKeys))
	for key := range allKeys {
		keysSlice = append(keysSlice, key)
	}
	sort.Strings(keysSlice)
	sort.Strings(redactedKeys)

	auditLog["keys"] = keysSlice
	auditLog["redacted_keys"] = redactedKeys

	// Output as JSON
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(auditLog); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to emit audit log: %v\n", err)
	}
}

func init() {
	// Configuration file path flag
	rootCmd.PersistentFlags().StringP("config", "c", "", "config file path (overrides IDENTITY_BROKER_CONFIG_PATH)")

	// Logging configuration flags
	rootCmd.PersistentFlags().String("log-level", "", "log level: debug, info, warn, error")
	rootCmd.PersistentFlags().String("log-format", "", "log format: text, json")

	// Server configuration flags
	rootCmd.PersistentFlags().Int("server.enduser.port", 0, "end-user server port (default: 8000)")
	rootCmd.PersistentFlags().String("server.enduser.bind", "", "end-user server bind address (default: ::)")
	rootCmd.PersistentFlags().Int("server.admin.port", 0, "admin server port (default: 14000)")
	rootCmd.PersistentFlags().String("server.admin.bind", "", "admin server bind address (default: ::)")
	rootCmd.PersistentFlags().Duration("server.shutdown.timeout", 0, "graceful shutdown timeout (default: 30s)")
}
