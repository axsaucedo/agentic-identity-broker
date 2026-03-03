// Package config provides configuration loading for the MCP server mock.
package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the configuration for the MCP server mock.
type Config struct {
	Server ServerConfig `yaml:"server"`
}

// ServerConfig holds HTTP server settings.
type ServerConfig struct {
	Port int    `yaml:"port"`
	Bind string `yaml:"bind"`
}

// Load reads the YAML config file from the given directory.
func Load(configDir string) (*Config, error) {
	data, err := os.ReadFile(fmt.Sprintf("%s/config.yaml", configDir))
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if cfg.Server.Bind == "" {
		cfg.Server.Bind = "0.0.0.0"
	}
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 9003
	}

	return &cfg, nil
}
