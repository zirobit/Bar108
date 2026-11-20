package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestProductionRejectsWeakJWTSecret(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("DB_USER", "test")
	t.Setenv("DB_PASSWORD", "test")
	t.Setenv("DB_NAME", "test")
	t.Setenv("JWT_SECRET", "short")
	if _, err := Load(); err == nil {
		t.Fatal("expected weak production JWT secret to be rejected")
	}
}

func TestRejectsInvalidJWTExpiry(t *testing.T) {
	t.Setenv("APP_ENV", "test")
	t.Setenv("DB_USER", "test")
	t.Setenv("DB_PASSWORD", "test")
	t.Setenv("DB_NAME", "test")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("JWT_EXPIRY_HOURS", "0")
	if _, err := Load(); err == nil {
		t.Fatal("expected nonpositive JWT expiry to be rejected")
	}
}

func TestSecretFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(path, []byte("file-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("BAR108_TEST_SECRET_FILE", path)
	secret, err := getSecretEnv("BAR108_TEST_SECRET")
	if err != nil || secret != "file-secret" {
		t.Fatalf("got %q, %v; want file-secret", secret, err)
	}
}
