package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/boxingoctopus/snackmates/api/internal/config"
)

func TestLoadFromEnvOnly(t *testing.T) {
	t.Setenv("CONFIG_FILE", "")
	t.Setenv("API_PORT", "9090")
	t.Setenv("JWT_SECRET", "test-secret")

	cfg, err := config.Load(config.Options{})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "9090" {
		t.Fatalf("Port = %q, want 9090", cfg.Port)
	}
	if cfg.JWTSecret != "test-secret" {
		t.Fatalf("JWTSecret = %q, want test-secret", cfg.JWTSecret)
	}
}

func TestLoadFromFileWithEnvOverride(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	content := `
[server]
port = "7000"
jwt_secret = "file-secret"

[database]
url = "postgres://file/file"
`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	t.Setenv("API_PORT", "8000")

	cfg, err := config.Load(config.Options{ConfigFile: path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8000" {
		t.Fatalf("Port = %q, want env override 8000", cfg.Port)
	}
	if cfg.JWTSecret != "file-secret" {
		t.Fatalf("JWTSecret = %q, want file-secret", cfg.JWTSecret)
	}
	if cfg.DatabaseURL != "postgres://file/file" {
		t.Fatalf("DatabaseURL = %q, want postgres://file/file", cfg.DatabaseURL)
	}
}

func TestLoadInvalidFile(t *testing.T) {
	_, err := config.Load(config.Options{ConfigFile: "/does/not/exist.toml"})
	if err == nil {
		t.Fatal("expected error for missing config file")
	}
}
