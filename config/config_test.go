package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_ReadsCustomConfigPathAndKeepsOptionalPathsEmpty(t *testing.T) {
	t.Parallel()

	configPath := filepath.Join(t.TempDir(), "custom.yaml")
	content := []byte(`
name: Custom Worthly
env: production
log:
  level: DEBUG
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}

	if cfg.Name != "Custom Worthly" {
		t.Fatalf("expected custom name, got %q", cfg.Name)
	}
	if cfg.Env != "production" {
		t.Fatalf("expected production env, got %q", cfg.Env)
	}
	if cfg.DB.Path != "" {
		t.Fatalf("expected empty db path, got %q", cfg.DB.Path)
	}
	if cfg.Log.Path != "" {
		t.Fatalf("expected empty log path, got %q", cfg.Log.Path)
	}
	if cfg.Log.Level != "DEBUG" {
		t.Fatalf("expected DEBUG log level, got %q", cfg.Log.Level)
	}
}
