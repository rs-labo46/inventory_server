package config

import (
	"log"
	"os"
	"strings"
	"time"
)

type Config struct {
	AppEnv string

	HTTPAddr string
	DBURL    string

	AllowedOrigins []string

	SessionTTL            time.Duration
	SessionCookieName     string
	SessionCookieSecure   bool
	SessionCookieSameSite string
	SessionCookiePath     string
}

func MustLoad() Config {
	appEnv := getenv("APP_ENV", "development")

	httpAddr := getenv("HTTP_ADDR", ":8080")

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	allowed := getenv("ALLOWED_ORIGINS", "")
	allowedOrigins := splitComma(allowed)

	sessionTTLStr := getenv("SESSION_TTL", "8h")
	sessionTTL, err := time.ParseDuration(sessionTTLStr)
	if err != nil {
		log.Fatalf("SESSION_TTL is invalid: %v", err)
	}

	cookieName := getenv("SESSION_COOKIE_NAME", "session_id")
	cookieSecure := getenv("SESSION_COOKIE_SECURE", "false") == "true"
	cookieSameSite := getenv("SESSION_COOKIE_SAMESITE", "Lax")
	cookiePath := getenv("SESSION_COOKIE_PATH", "/")

	return Config{
		AppEnv: appEnv,

		HTTPAddr: httpAddr,
		DBURL:    dbURL,

		AllowedOrigins: allowedOrigins,

		SessionTTL:            sessionTTL,
		SessionCookieName:     cookieName,
		SessionCookieSecure:   cookieSecure,
		SessionCookieSameSite: cookieSameSite,
		SessionCookiePath:     cookiePath,
	}
}

func getenv(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}

func splitComma(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}
