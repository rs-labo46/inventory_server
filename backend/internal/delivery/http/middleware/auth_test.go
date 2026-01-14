package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
)

type sessionRepoStub struct {
	findByID func(ctx context.Context, sessionID string) (domain.Session, error)
}

func (s sessionRepoStub) Create(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
	return domain.Session{}, nil
}

func (s sessionRepoStub) FindByID(ctx context.Context, sessionID string) (domain.Session, error) {
	return s.findByID(ctx, sessionID)
}

func (s sessionRepoStub) Delete(ctx context.Context, sessionID string) error { return nil }

var _ repo.SessionRepository = sessionRepoStub{}

func TestRequireAuth_Wrap_Allows_WhenSessionValid(t *testing.T) {
	mw := RequireAuth{
		Sessions: sessionRepoStub{
			findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
				return domain.Session{
					ID:        sessionID,
					UserID:    "u1",
					ExpiresAt: time.Now().Add(10 * time.Minute),
				}, nil
			},
		},
		CookieName: "session_id",
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/inventories", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "s1"})
	rec := httptest.NewRecorder()

	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true

		gotUserID, err := GetUserID(r.Context())
		if err != nil {
			t.Fatalf("GetUserID error: %v", err)
		}
		if gotUserID != "u1" {
			t.Fatalf("userID mismatch: got=%s want=%s", gotUserID, "u1")
		}

		w.WriteHeader(http.StatusOK)
	})

	mw.Wrap(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}
	if !called {
		t.Fatalf("next handler should be called")
	}
}

func TestRequireAuth_Wrap_OptionsReturns204(t *testing.T) {
	mw := RequireAuth{
		Sessions: sessionRepoStub{
			findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
				return domain.Session{}, domain.ErrNotFound
			},
		},
		CookieName: "session_id",
	}

	req := httptest.NewRequest(http.MethodOptions, "http://example.com/api/inventories", nil)
	rec := httptest.NewRecorder()

	mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("OPTIONSはnextに行かない想定")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusNoContent)
	}
}

func TestRequireAuth_Wrap_NoCookieUnauthorized(t *testing.T) {
	mw := RequireAuth{
		Sessions: sessionRepoStub{
			findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
				return domain.Session{}, domain.ErrNotFound
			},
		},
		CookieName: "session_id",
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/inventories", nil)
	rec := httptest.NewRecorder()

	mw.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("cookie無しはnextに到達しない想定")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusUnauthorized)
	}
	if !strings.Contains(rec.Body.String(), `"UNAUTHENTICATED"`) {
		t.Fatalf("body should contain UNAUTHENTICATED: %s", rec.Body.String())
	}
}

func TestGetUserID_OK(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, "u1")
	got, err := GetUserID(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "u1" {
		t.Fatalf("userID mismatch: got=%s want=%s", got, "u1")
	}
}

func TestGetUserID_TypeMismatchUnauthorized(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserIDKey, 123)
	_, err := GetUserID(ctx)
	if err != domain.ErrUnauthorized {
		t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrUnauthorized)
	}
}
