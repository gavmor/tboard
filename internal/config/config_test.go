package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigFile(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "test_config.yaml")

	content := `
theme: dracula
default_port: 443
targets:
  - host: 1.1.1.1:53
  - host: custom.domain.org
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to create temp config file: %v", err)
	}

	cfg, err := LoadConfigFile(configPath)
	if err != nil {
		t.Fatalf("expected no error loading config file, got %v", err)
	}

	if cfg.Theme != "dracula" {
		t.Errorf("expected theme 'dracula', got %s", cfg.Theme)
	}
	if cfg.DefaultPort != 443 {
		t.Errorf("expected default port 443, got %d", cfg.DefaultPort)
	}
}
