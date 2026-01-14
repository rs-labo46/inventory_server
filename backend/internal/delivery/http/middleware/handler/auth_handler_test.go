package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	appauth "inventory-backend/internal/application/auth"
	"inventory-backend/internal/delivery/http/middleware/handler"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"

	"golang.org/x/crypto/bcrypt"
)

type userRepoStub struct {
	findByEmail func(ctx context.Context, email string) (domain.User, error)
	findByID    func(ctx context.Context, id string) (domain.User, error)
	create      func(ctx context.Context, email string, passwordHash string) (domain.User, error)
}

func (u userRepoStub) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	return u.findByEmail(ctx, email)
}
func (u userRepoStub) FindByID(ctx context.Context, id string) (domain.User, error) {
	return u.findByID(ctx, id)
}
func (u userRepoStub) Create(ctx context.Context, email string, passwordHash string) (domain.User, error) {
	return u.create(ctx, email, passwordHash)
}

var _ repo.UserRepository = userRepoStub{}

type sessionRepoStub struct {
	create   func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error)
	findByID func(ctx context.Context, sessionID string) (domain.Session, error)
	delete   func(ctx context.Context, sessionID string) error
}

func (s sessionRepoStub) Create(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
	return s.create(ctx, userID, expiresAt)
}
func (s sessionRepoStub) FindByID(ctx context.Context, sessionID string) (domain.Session, error) {
	return s.findByID(ctx, sessionID)
}
func (s sessionRepoStub) Delete(ctx context.Context, sessionID string) error {
	return s.delete(ctx, sessionID)
}

var _ repo.SessionRepository = sessionRepoStub{}

func TestAuthHandler_Login_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	h := handler.AuthHandler{
		Auth:           appauth.Usecase{},
		CookieName:     "session_id",
		CookiePath:     "/",
		CookieSameSite: "Lax",
	}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/auth/login", nil)
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestAuthHandler_Login_BadJSON(t *testing.T) {
	t.Parallel()

	h := handler.AuthHandler{
		Auth:           appauth.Usecase{},
		CookieName:     "session_id",
		CookiePath:     "/",
		CookieSameSite: "Lax",
	}

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/login", bytes.NewBufferString("{bad json"))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Login_MissingFields(t *testing.T) {
	t.Parallel()

	h := handler.AuthHandler{
		Auth:           appauth.Usecase{},
		CookieName:     "session_id",
		CookiePath:     "/",
		CookieSameSite: "Lax",
	}

	body := `{"email":"","password":""}`
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/login", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}

func TestAuthHandler_Login_Unauthorized(t *testing.T) {
	t.Parallel()

	uc := appauth.Usecase{
		Users: userRepoStub{
			findByEmail: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{}, domain.ErrNotFound
			},
			findByID: func(ctx context.Context, id string) (domain.User, error) {
				return domain.User{}, domain.ErrNotFound
			},
			create: func(ctx context.Context, email string, passwordHash string) (domain.User, error) {
				return domain.User{}, nil
			},
		},
		Sessions: sessionRepoStub{
			create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
				return domain.Session{}, nil
			},
			findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
				return domain.Session{}, nil
			},
			delete: func(ctx context.Context, sessionID string) error { return nil },
		},
		SessionTTL: 1 * time.Hour,
	}

	h := handler.AuthHandler{
		Auth:           uc,
		CookieName:     "session_id",
		CookiePath:     "/",
		CookieSecure:   false,
		CookieSameSite: "Lax",
	}

	body := `{"email":"x@example.com","password":"pw"}`
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/login", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthHandler_Login_Success_SetsCookie(t *testing.T) {
	t.Parallel()

	hash, err := bcrypt.GenerateFromPassword([]byte("pw"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("bcrypt error: %v", err)
	}

	uc := appauth.Usecase{
		Users: userRepoStub{
			findByEmail: func(ctx context.Context, email string) (domain.User, error) {
				return domain.User{
					ID:           "u1",
					Email:        email,
					PasswordHash: string(hash),
				}, nil
			},
			findByID: func(ctx context.Context, id string) (domain.User, error) {
				return domain.User{}, domain.ErrNotFound
			},
			create: func(ctx context.Context, email string, passwordHash string) (domain.User, error) {
				return domain.User{}, nil
			},
		},
		Sessions: sessionRepoStub{
			create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
				return domain.Session{ID: "s1", UserID: userID, ExpiresAt: expiresAt}, nil
			},
			findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
				return domain.Session{}, nil
			},
			delete: func(ctx context.Context, sessionID string) error { return nil },
		},
		SessionTTL: 1 * time.Hour,
	}

	h := handler.AuthHandler{
		Auth:           uc,
		CookieName:     "session_id",
		CookiePath:     "/",
		CookieSecure:   false,
		CookieSameSite: "Lax",
	}

	body := `{"email":"x@example.com","password":"pw"}`
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/login", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}

	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session_id" {
			found = true
			if c.Value != "s1" {
				t.Fatalf("cookie value mismatch: got=%s want=%s", c.Value, "s1")
			}
			if c.Path != "/" {
				t.Fatalf("cookie path mismatch: got=%s want=%s", c.Path, "/")
			}
		}
	}
	if !found {
		t.Fatalf("session cookie was not set")
	}
	var lr struct {
		User struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}

	if err := json.NewDecoder(rec.Body).Decode(&lr); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if lr.User.ID != "u1" || lr.User.Email != "x@example.com" {
		t.Fatalf("response mismatch: got id=%s email=%s", lr.User.ID, lr.User.Email)
	}

}

func TestAuthHandler_Logout_DeletesCookie(t *testing.T) {
	t.Parallel()

	uc := appauth.Usecase{
		Sessions: sessionRepoStub{
			create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
				return domain.Session{}, nil
			},
			findByID: func(ctx context.Context, sessionID string) (domain.Session, error) { return domain.Session{}, nil },
			delete:   func(ctx context.Context, sessionID string) error { return nil },
		},
		Users: userRepoStub{
			findByEmail: func(ctx context.Context, email string) (domain.User, error) { return domain.User{}, nil },
			findByID:    func(ctx context.Context, id string) (domain.User, error) { return domain.User{}, nil },
			create: func(ctx context.Context, email string, passwordHash string) (domain.User, error) {
				return domain.User{}, nil
			},
		},
	}

	h := handler.AuthHandler{
		Auth:           uc,
		CookieName:     "session_id",
		CookiePath:     "/",
		CookieSecure:   false,
		CookieSameSite: "Lax",
	}

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "s1"})
	rec := httptest.NewRecorder()

	h.Logout(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}

	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "session_id" {
			found = true
			if c.MaxAge != -1 {
				t.Fatalf("cookie MaxAge mismatch: got=%d want=%d", c.MaxAge, -1)
			}
		}
	}
	if !found {
		t.Fatalf("delete cookie was not set")
	}
}
