package config

import (
	"testing"
	"time"
)

func TestMustLoad_Defaults(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")

	cfg := MustLoad()

	if cfg.AppEnv != "development" {
		t.Fatalf("AppEnv got=%s want=%s", cfg.AppEnv, "development")
	}
	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("HTTPAddr got=%s want=%s", cfg.HTTPAddr, ":8080")
	}
	if cfg.DBURL == "" {
		t.Fatalf("DBURL should not be empty")
	}
	if len(cfg.AllowedOrigins) != 0 {
		t.Fatalf("AllowedOrigins should be empty: got=%v", cfg.AllowedOrigins)
	}

	if cfg.SessionTTL != 8*time.Hour {
		t.Fatalf("SessionTTL got=%v want=%v", cfg.SessionTTL, 8*time.Hour)
	}
	if cfg.SessionCookieName != "session_id" {
		t.Fatalf("CookieName got=%s want=%s", cfg.SessionCookieName, "session_id")
	}
	if cfg.SessionCookieSecure != false {
		t.Fatalf("CookieSecure got=%v want=%v", cfg.SessionCookieSecure, false)
	}
	if cfg.SessionCookieSameSite != "Lax" {
		t.Fatalf("CookieSameSite got=%s want=%s", cfg.SessionCookieSameSite, "Lax")
	}
	if cfg.SessionCookiePath != "/" {
		t.Fatalf("CookiePath got=%s want=%s", cfg.SessionCookiePath, "/")
	}
}

func TestMustLoad_AllowedOrigins_SplitAndTrim(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/db?sslmode=disable")
	t.Setenv("ALLOWED_ORIGINS", "http://a, http://b  , ,http://c")

	cfg := MustLoad()

	if len(cfg.AllowedOrigins) != 3 {
		t.Fatalf("AllowedOrigins len got=%d want=%d (%v)", len(cfg.AllowedOrigins), 3, cfg.AllowedOrigins)
	}
	if cfg.AllowedOrigins[0] != "http://a" || cfg.AllowedOrigins[1] != "http://b" || cfg.AllowedOrigins[2] != "http://c" {
		t.Fatalf("AllowedOrigins mismatch: %v", cfg.AllowedOrigins)
	}
}
