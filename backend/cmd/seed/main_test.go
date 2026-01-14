package main

import (
	"os"
	"testing"
)

func TestGetenv_Default(t *testing.T) {
	os.Unsetenv("X_TEST_ENV")
	got := getenv("X_TEST_ENV", "fallback")
	if got != "fallback" {
		t.Fatalf("got=%s want=%s", got, "fallback")
	}
}

func TestGetenv_Override(t *testing.T) {
	os.Setenv("X_TEST_ENV", "value")
	t.Cleanup(func() { os.Unsetenv("X_TEST_ENV") })

	got := getenv("X_TEST_ENV", "fallback")
	if got != "value" {
		t.Fatalf("got=%s want=%s", got, "value")
	}
}
