package config

import "testing"

func TestLoadRequiresSessionSecret(t *testing.T) {
	t.Setenv("SESSION_SECRET", "")
	t.Setenv("ADMIN_INIT_USERNAME", "")
	t.Setenv("ADMIN_INIT_PASSWORD", "")

	if _, err := Load("1.0.3"); err == nil {
		t.Fatal("expected error when SESSION_SECRET is missing")
	}
}

func TestLoadValidConfig(t *testing.T) {
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("ADMIN_INIT_USERNAME", "admin")
	t.Setenv("ADMIN_INIT_PASSWORD", "admin12345678")
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_PATH", "")

	cfg, err := Load("1.0.3")
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if cfg.Version != "1.0.3" {
		t.Fatalf("unexpected version: %s", cfg.Version)
	}
	if cfg.DBPath == "" {
		t.Fatal("expected default db path")
	}
}

func TestLoadRejectsHalfAdminConfig(t *testing.T) {
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("ADMIN_INIT_USERNAME", "admin")
	t.Setenv("ADMIN_INIT_PASSWORD", "")

	if _, err := Load("1.0.3"); err == nil {
		t.Fatal("expected error when only one admin env is set")
	}
}
