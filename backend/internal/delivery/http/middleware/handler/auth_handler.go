package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"inventory-backend/internal/application/auth"
	"inventory-backend/internal/delivery/http/responder"
	"inventory-backend/internal/domain"
)

type AuthHandler struct {
	Auth auth.Usecase

	CookieName     string
	CookieSecure   bool
	CookieSameSite string
	CookiePath     string
}

type loginReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRes struct {
	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	} `json:"user"`
}

type okRes struct {
	OK bool `json:"ok"`
}

func (h AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responder.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "許可されていないメソッドです")

		return
	}

	var req loginReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "JSONが不正です")
		return
	}
	if req.Email == "" || req.Password == "" {
		responder.WriteError(w, http.StatusBadRequest, "INVALID_REQUEST", "emailとpasswordは必須です")
		return
	}

	result, err := h.Auth.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == domain.ErrUnauthorized {
			responder.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "認証に失敗しました")
			return
		}

		log.Printf("login internal error: %v", err)
		responder.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーです")
		return
	}

	h.setSessionCookie(w, result.SessionID, result.ExpiresAt)

	var res loginRes
	res.User.ID = result.UserID
	res.User.Email = result.Email
	responder.WriteJSON(w, http.StatusOK, res)
}

func (h AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		responder.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "許可されていないメソッドです")
		return
	}

	sid := h.getSessionID(r)
	_ = h.Auth.Logout(r.Context(), sid)

	http.SetCookie(w, &http.Cookie{
		Name:     h.CookieName,
		Value:    "",
		Path:     h.CookiePath,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: h.parseSameSite(),
		MaxAge:   -1,
	})
	responder.WriteJSON(w, http.StatusOK, okRes{OK: true})
}

func (h AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		responder.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "許可されていないメソッドです")
		return
	}

	sid := h.getSessionID(r)
	if sid == "" {
		responder.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "ログインしてください")
		return
	}

	me, err := h.Auth.Me(r.Context(), sid)
	if err != nil {
		responder.WriteError(w, http.StatusUnauthorized, "UNAUTHENTICATED", "ログインしてください")
		return
	}

	var res loginRes
	res.User.ID = me.UserID
	res.User.Email = me.Email
	responder.WriteJSON(w, http.StatusOK, res)
}

func (h AuthHandler) getSessionID(r *http.Request) string {
	c, err := r.Cookie(h.CookieName)
	if err != nil {
		return ""
	}
	return c.Value
}

func (h AuthHandler) setSessionCookie(w http.ResponseWriter, sessionID string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     h.CookieName,
		Value:    sessionID,
		Path:     h.CookiePath,
		HttpOnly: true,
		Secure:   h.CookieSecure,
		SameSite: h.parseSameSite(),
		Expires:  expiresAt,
	})
}

func (h AuthHandler) parseSameSite() http.SameSite {
	switch h.CookieSameSite {
	case "Strict":
		return http.SameSiteStrictMode
	case "None":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}
