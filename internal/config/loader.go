// Package config implements the configuration loading adapter.
package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/agentic-identity-broker/agentic-identity-broker/internal/domain/config"
	"github.com/agentic-identity-broker/agentic-identity-broker/internal/ports"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Loader implements the ConfigPort interface using Viper, Cobra, and godotenv.
// Each Loader instance has its own Viper instance for proper isolation.
type Loader struct {
	v       *viper.Viper
	sources []ports.ConfigSource
	cmd     *cobra.Command
}

// NewLoader creates a new configuration loader with instance-scoped Viper.
// This prevents test conflicts from global Viper state.
func NewLoader() *Loader {
	return &Loader{
		v:       viper.New(),
		sources: make([]ports.ConfigSource, 0),
	}
}

// SetCommand sets the Cobra command for CLI flag binding.
// Must be called before GetConfig if CLI flags should be used.
func (l *Loader) SetCommand(cmd *cobra.Command) {
	l.cmd = cmd
}

// GetConfig returns the fully loaded and validated configuration.
// Implements ports.ConfigPort interface.
func (l *Loader) GetConfig(ctx context.Context) (*ports.Config, error) {
	// Phase 1: Set defaults
	l.setDefaults()

	// Phase 2: Load .env files
	if err := l.loadEnvFiles(); err != nil {
		return nil, err
	}

	// Phase 3: Load YAML (User Story 2)
	if err := l.loadYAML(); err != nil {
		return nil, err
	}

	// Phase 4: Expand environment variables (User Story 2)
	if err := l.expandEnvVars(); err != nil {
		return nil, err
	}

	// Phase 5: Bind CLI flags (User Story 3)
	if err := l.bindFlags(); err != nil {
		return nil, err
	}

	// Unmarshal to Config struct
	var cfg ports.Config
	if err := l.v.Unmarshal(&cfg); err != nil {
		return nil, &config.ConfigError{
			Field:    "config",
			Expected: "valid configuration structure",
			Err:      err,
		}
	}

	// Phase 6: Validate configuration (User Story 4)
	if err := Validate(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// GetSources returns metadata about all configuration sources used.
// Implements ports.ConfigPort interface.
func (l *Loader) GetSources() []ports.ConfigSource {
	return l.sources
}

// Reload reloads configuration from all sources.
// NOT IMPLEMENTED in initial scope (no hot-reloading).
func (l *Loader) Reload(ctx context.Context) error {
	return fmt.Errorf("reload not implemented in initial scope")
}

// setDefaults sets default configuration values.
// Records source metadata for audit logging.
func (l *Loader) setDefaults() {
	l.v.SetDefault("log.level", string(config.LogLevelInfo))
	l.v.SetDefault("log.format", string(config.LogFormatText))

	// Record defaults source
	l.sources = append(l.sources, ports.ConfigSource{
		Type:       ports.SourceTypeDefault,
		Path:       "defaults",
		Precedence: 0,
		LoadedAt:   time.Now(),
		Keys:       []string{"log.level", "log.format"},
	})
}

// loadEnvFiles loads .env files in correct precedence order.
// Order: .env → .env.local → .env.{environment} → .env.{environment}.local
// Records source metadata for each loaded file.
func (l *Loader) loadEnvFiles() error {
	// Determine environment
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	// Define .env file loading order
	envFiles := []string{
		".env",
		".env.local",
		fmt.Sprintf(".env.%s", env),
		fmt.Sprintf(".env.%s.local", env),
	}

	// Load each .env file if it exists
	for _, filename := range envFiles {
		if err := l.loadEnvFile(filename); err != nil {
			// Continue if file doesn't exist
			if os.IsNotExist(err) {
				continue
			}
			return err
		}
	}

	return nil
}

// loadEnvFile loads a single .env file and records source metadata.
func (l *Loader) loadEnvFile(filename string) error {
	// Get absolute path
	absPath, err := filepath.Abs(filename)
	if err != nil {
		return &config.ConfigError{
			Field:    "env_file",
			Value:    filename,
			Expected: "valid file path",
			Err:      err,
		}
	}

	// Check if file exists
	if _, err := os.Stat(absPath); err != nil {
		return err
	}

	// Load .env file
	envMap, err := godotenv.Read(absPath)
	if err != nil {
		return &config.ConfigError{
			Field:    "env_file",
			Value:    absPath,
			Expected: "valid .env file format",
			Err:      err,
		}
	}

	// Set environment variables in Viper
	keys := make([]string, 0, len(envMap))
	for key, value := range envMap {
		// Strip IDENTITY_BROKER_ prefix and convert to Viper format
		viperKey := key
		if strings.HasPrefix(strings.ToUpper(key), "IDENTITY_BROKER_") {
			viperKey = strings.TrimPrefix(strings.ToUpper(key), "IDENTITY_BROKER_")
		}
		// Convert to lowercase with dots
		viperKey = strings.ToLower(strings.ReplaceAll(viperKey, "_", "."))
		l.v.Set(viperKey, value)
		keys = append(keys, viperKey)
	}

	// Record source metadata
	l.sources = append(l.sources, ports.ConfigSource{
		Type:       ports.SourceTypeEnvFile,
		Path:       absPath,
		Precedence: 1,
		LoadedAt:   time.Now(),
		Keys:       keys,
	})

	return nil
}

// loadYAML loads configuration from a YAML file.
// File path is determined by --config flag or IDENTITY_BROKER_CONFIG_PATH env var.
// If neither is set, looks for config.yaml in the current directory.
// Records source metadata for audit logging.
func (l *Loader) loadYAML() error {
	// Determine config file path
	configPath := os.Getenv("IDENTITY_BROKER_CONFIG_PATH")
	if configPath == "" {
		configPath = "config.yaml"
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		// Config file is optional
		return nil
	}

	// Get absolute path
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return &config.ConfigError{
			Field:    "config_file",
			Value:    configPath,
			Expected: "valid file path",
			Err:      err,
		}
	}

	// Set config file in Viper
	l.v.SetConfigFile(absPath)

	// Read config file
	if err := l.v.ReadInConfig(); err != nil {
		// Check for specific error types
		if os.IsPermission(err) {
			return &config.ConfigError{
				Field:    "config_file",
				Value:    absPath,
				Expected: "readable file with proper permissions",
				Source:   "yaml",
				Err:      err,
			}
		}
		return &config.ConfigError{
			Field:    "config_file",
			Value:    absPath,
			Expected: "valid YAML syntax",
			Source:   "yaml",
			Err:      err,
		}
	}

	// Get all keys from the config file
	keys := l.v.AllKeys()

	// Record source metadata
	l.sources = append(l.sources, ports.ConfigSource{
		Type:       ports.SourceTypeYAML,
		Path:       absPath,
		Precedence: 2,
		LoadedAt:   time.Now(),
		Keys:       keys,
	})

	return nil
}

// expandEnvVars expands environment variable references in configuration values.
// Supports ${VAR_NAME} syntax with circular reference detection.
// Validates against command injection patterns.
func (l *Loader) expandEnvVars() error {
	// Get all configuration keys
	allKeys := l.v.AllKeys()

	for _, key := range allKeys {
		value := l.v.GetString(key)
		if value == "" {
			continue
		}

		// Check if value contains environment variable reference
		if !strings.Contains(value, "${") {
			continue
		}

		// Validate against command injection patterns
		if err := l.validateSecurePattern(value); err != nil {
			return &config.ConfigError{
				Field:    key,
				Value:    value,
				Expected: "environment variable reference without command injection patterns",
				Source:   "yaml",
				Err:      err,
			}
		}

		// Expand environment variables with circular detection
		expanded, err := l.expandWithCircularCheck(value, make(map[string]bool), 0)
		if err != nil {
			return &config.ConfigError{
				Field:    key,
				Value:    value,
				Expected: "resolvable environment variable reference",
				Source:   "yaml",
				Err:      err,
			}
		}

		// Update value in Viper
		l.v.Set(key, expanded)
	}

	return nil
}

// expandWithCircularCheck expands environment variables with circular reference detection.
// maxDepth limits recursion to prevent infinite loops.
// visited tracks variables seen in the current expansion chain.
func (l *Loader) expandWithCircularCheck(value string, visited map[string]bool, depth int) (string, error) {
	const maxDepth = 10

	// Check recursion depth
	if depth >= maxDepth {
		return "", fmt.Errorf("environment variable expansion exceeded maximum depth of %d", maxDepth)
	}

	// Track expansion errors
	var expansionErrors []string

	// Expand ${VAR} patterns
	result := os.Expand(value, func(varName string) string {
		// Check for circular reference
		if visited[varName] {
			// Build circular chain for error message
			chain := []string{}
			for v := range visited {
				chain = append(chain, v)
			}
			chain = append(chain, varName)
			expansionErrors = append(expansionErrors, fmt.Sprintf("CIRCULAR:%s→%s", strings.Join(chain, "→"), varName))
			return ""
		}

		// Get environment variable value
		varValue, exists := os.LookupEnv(varName)
		if !exists {
			expansionErrors = append(expansionErrors, fmt.Sprintf("UNDEFINED:%s", varName))
			return ""
		}

		// Check if variable value contains more references
		if strings.Contains(varValue, "${") {
			// Mark variable as visited
			newVisited := make(map[string]bool)
			for k, v := range visited {
				newVisited[k] = v
			}
			newVisited[varName] = true

			// Recursively expand
			expanded, err := l.expandWithCircularCheck(varValue, newVisited, depth+1)
			if err != nil {
				expansionErrors = append(expansionErrors, fmt.Sprintf("NESTED:%v", err))
				return ""
			}
			return expanded
		}

		return varValue
	})

	// Check for expansion errors
	if len(expansionErrors) > 0 {
		for _, errMsg := range expansionErrors {
			if strings.HasPrefix(errMsg, "CIRCULAR:") {
				chain := strings.TrimPrefix(errMsg, "CIRCULAR:")
				return "", fmt.Errorf("circular reference detected: %s", chain)
			}
			if strings.HasPrefix(errMsg, "UNDEFINED:") {
				varName := strings.TrimPrefix(errMsg, "UNDEFINED:")
				return "", fmt.Errorf("environment variable '%s' is not set", varName)
			}
			if strings.HasPrefix(errMsg, "NESTED:") {
				nested := strings.TrimPrefix(errMsg, "NESTED:")
				return "", fmt.Errorf("nested expansion error: %s", nested)
			}
		}
	}

	return result, nil
}

// validateSecurePattern validates that configuration values don't contain
// command injection patterns like $(command) or backticks.
func (l *Loader) validateSecurePattern(value string) error {
	// Reject command substitution patterns
	if strings.Contains(value, "$(") || strings.Contains(value, "`") {
		return fmt.Errorf("command substitution patterns $(command) and backticks are not allowed")
	}

	// Check for shell metacharacters that could indicate injection attempts
	dangerousChars := []string{";", "|", "&", ">", "<", "\n", "\r"}
	for _, char := range dangerousChars {
		if strings.Contains(value, char) {
			return fmt.Errorf("potentially dangerous shell metacharacter '%s' detected", char)
		}
	}

	return nil
}

// bindFlags binds Cobra command flags to Viper configuration.
// CLI flags have the highest precedence and override all other sources.
func (l *Loader) bindFlags() error {
	if l.cmd == nil {
		// No command set, skip flag binding
		return nil
	}

	// Track which keys came from CLI flags
	cliKeys := make([]string, 0)

	// Bind config file path flag
	if l.cmd.Flags().Changed("config") {
		configPath, _ := l.cmd.Flags().GetString("config")
		os.Setenv("IDENTITY_BROKER_CONFIG_PATH", configPath)
	}

	// Bind log.level flag
	if l.cmd.Flags().Changed("log-level") {
		logLevel, _ := l.cmd.Flags().GetString("log-level")
		l.v.Set("log.level", logLevel)
		cliKeys = append(cliKeys, "log.level")
	}

	// Bind log.format flag
	if l.cmd.Flags().Changed("log-format") {
		logFormat, _ := l.cmd.Flags().GetString("log-format")
		l.v.Set("log.format", logFormat)
		cliKeys = append(cliKeys, "log.format")
	}

	// Record CLI source if any flags were set
	if len(cliKeys) > 0 {
		l.sources = append(l.sources, ports.ConfigSource{
			Type:       ports.SourceTypeCLI,
			Path:       "cli",
			Precedence: 3,
			LoadedAt:   time.Now(),
			Keys:       cliKeys,
		})
	}

	return nil
}
