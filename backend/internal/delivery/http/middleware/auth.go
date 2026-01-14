package middleware

import (
	"context"
	"inventory-backend/internal/delivery/http/responder"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
	"net/http"
	"strings"
	"time"
)

// セッションクッキーが有効な時だけとおす
type RequireAuth struct {
	Sessions   repo.SessionRepository
	CookieName string
}

func (m RequireAuth) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		if !strings.HasPrefix(r.URL.Path, "/api/") {
			next.ServeHTTP(w, r)
			return
		}

		if r.URL.Path == "/api/auth/login" {
			next.ServeHTTP(w, r)
			return
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		c, err := r.Cookie(m.CookieName)
		if err != nil || c.Value == "" {
			responder.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "ログインしてください")
			return
		}

		sess, err := m.Sessions.FindByID(r.Context(), c.Value)
		if err != nil {
			responder.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "ログインしてください")
			return
		}

		if time.Now().After(sess.ExpiresAt) {
			responder.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "ログインしてください")
			return
		}

		ctx := context.WithValue(r.Context(), UserIDKey, sess.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserID(ctx context.Context) (string, error) {
	v := ctx.Value(UserIDKey)
	s, ok := v.(string)
	if !ok || s == "" {
		return "", domain.ErrUnauthorized
	}
	return s, nil
}
