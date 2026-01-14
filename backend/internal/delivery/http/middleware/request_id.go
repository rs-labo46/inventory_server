package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

// 追跡
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 16)
		_, _ = rand.Read(b)
		id := hex.EncodeToString(b)

		w.Header().Set("X-Request-Id", id)

		next.ServeHTTP(w, r)
	})
}
