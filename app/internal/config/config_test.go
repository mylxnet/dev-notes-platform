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

func TestLoadCookieSecureMode(t *testing.T) {
	t.Setenv("SESSION_SECRET", "0123456789abcdef0123456789abcdef")
	t.Setenv("ADMIN_INIT_USERNAME", "admin")
	t.Setenv("ADMIN_INIT_PASSWORD", "admin12345678")
	t.Setenv("DATA_DIR", t.TempDir())
	t.Setenv("DB_PATH", "")

	cases := map[string]string{
		"":        "auto",
		"auto":    "auto",
		"true":    "always",
		"false":   "never",
		"  TRUE ": "always",
	}
	for raw, want := range cases {
		t.Setenv("COOKIE_SECURE", raw)
		cfg, err := Load("1.3.1")
		if err != nil {
			t.Fatalf("COOKIE_SECURE=%q 应通过，实际报错 %v", raw, err)
		}
		if cfg.CookieSecureMode != want {
			t.Fatalf("COOKIE_SECURE=%q 期望 %s，实际 %s", raw, want, cfg.CookieSecureMode)
		}
	}

	t.Setenv("COOKIE_SECURE", "maybe")
	if _, err := Load("1.3.1"); err == nil {
		t.Fatal("非法 COOKIE_SECURE 应该启动失败")
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
