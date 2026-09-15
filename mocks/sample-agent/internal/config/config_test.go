package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_BrokerAgentIDOverride(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "config.yaml"), []byte(`
server:
  port: 9002
  bind: "0.0.0.0"
oauth2:
  client_id: "original-client-id"
  client_secret: "secret"
  redirect_uri: "http://localhost:9002/oauth2/callback"
`), 0600); err != nil {
		t.Fatal(err)
	}

	t.Run("uses yaml client_id when env var is absent", func(t *testing.T) {
		t.Setenv("BROKER_AGENT_ID", "")
		cfg, err := Load(dir)
		if err != nil {
			t.Fatalf("Load() error: %v", err)
		}
		if cfg.OAuth2.ClientID != "original-client-id" {
			t.Errorf("ClientID = %q, want %q", cfg.OAuth2.ClientID, "original-client-id")
		}
	})

	t.Run("BROKER_AGENT_ID overrides yaml client_id", func(t *testing.T) {
		t.Setenv("BROKER_AGENT_ID", "550e8400-e29b-41d4-a716-446655440000")
		cfg, err := Load(dir)
		if err != nil {
			t.Fatalf("Load() error: %v", err)
		}
		if cfg.OAuth2.ClientID != "550e8400-e29b-41d4-a716-446655440000" {
			t.Errorf("ClientID = %q, want %q", cfg.OAuth2.ClientID, "550e8400-e29b-41d4-a716-446655440000")
		}
	})
}
