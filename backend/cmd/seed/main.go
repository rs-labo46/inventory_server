package main

import (
	"context"
	"inventory-backend/internal/infra/postgres"
	"log"
	"os"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is required")
	}

	email := getenv("SEED_EMAIL", "admin@example.com")
	pass := getenv("SEED_PASSWORD", "admin0001")

	hash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("bcrypt failed: %v", err)
	}

	db, err := postgres.Open(dsn)
	if err != nil {
		log.Fatalf("sql.Open failed: %v", err)
	}
	defer db.Close()

	if err := postgres.Ping(context.Background(), db); err != nil {
		log.Fatalf("db ping failed: %v", err)
	}

	const q = `INSERT INTO users (email, password_hash) VALUES ($1, $2) ON CONFLICT (email) DO NOTHING`

	if _, err := db.ExecContext(context.Background(), q, email, string(hash)); err != nil {
		log.Fatalf("insert failed: %v", err)
	}
	log.Printf("seed user: %s / %s", email, pass)
}

func getenv(key string, fallback string) string {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	return v
}
