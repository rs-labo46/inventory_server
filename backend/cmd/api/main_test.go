package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"inventory-backend/internal/delivery/http/middleware"
	"inventory-backend/internal/delivery/http/middleware/handler"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
)

type sessionRepoStub struct{}

func (s sessionRepoStub) Create(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
	return domain.Session{}, nil
}
func (s sessionRepoStub) FindByID(ctx context.Context, sessionID string) (domain.Session, error) {
	return domain.Session{ID: sessionID, UserID: "u1", ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
}
func (s sessionRepoStub) Delete(ctx context.Context, sessionID string) error { return nil }

var _ repo.SessionRepository = sessionRepoStub{}

func TestBuildHTTPHandler_Healthz_OK(t *testing.T) {
	originChecker := middleware.NewOriginChecker([]string{})

	authMW := middleware.RequireAuth{
		Sessions:   sessionRepoStub{},
		CookieName: "session_id",
	}

	h := buildHTTPHandler(
		handler.AuthHandler{CookieName: "session_id", CookiePath: "/", CookieSameSite: "Lax"},
		handler.ProductHandler{},
		handler.InventoryHandler{},
		handler.StockHandler{},
		handler.StockMovementHandler{},
		authMW,
		originChecker,
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/healthz", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}

	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatalf("X-Request-Id should be set")
	}

	if !strings.Contains(rec.Body.String(), "ok") && !strings.Contains(rec.Body.String(), "OK") {
		t.Fatalf("response should contain ok: %s", rec.Body.String())
	}
}

func TestBuildHTTPHandler_ProtectedRoute_Unauthenticated(t *testing.T) {
	originChecker := middleware.NewOriginChecker([]string{})

	authMW := middleware.RequireAuth{
		Sessions:   sessionRepoStub{},
		CookieName: "session_id",
	}

	h := buildHTTPHandler(
		handler.AuthHandler{CookieName: "session_id", CookiePath: "/", CookieSameSite: "Lax"},
		handler.ProductHandler{},
		handler.InventoryHandler{},
		handler.StockHandler{},
		handler.StockMovementHandler{},
		authMW,
		originChecker,
	)

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/inventories", nil)
	rec := httptest.NewRecorder()

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusUnauthorized)
	}

	if !strings.Contains(rec.Body.String(), `"error"`) {
		t.Fatalf("response should contain error: %s", rec.Body.String())
	}
}
