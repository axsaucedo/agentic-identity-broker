// Package main is the entry point for the Agentic Identity Broker application.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "identity-broker",
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

	// TODO: Initialize and run application services

	fmt.Println("\nAgentic Identity Broker starting...")
	fmt.Printf("Log Level: %s\n", cfg.Log.Level)
	fmt.Printf("Log Format: %s\n", cfg.Log.Format)

	return nil
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

// sortKeys returns a sorted list of keys from a map.
func sortKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
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
}
