package middleware

import (
	"inventory-backend/internal/delivery/http/responder"
	"log"
	"net/http"
)

func Recover(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if v := recover(); v != nil {
				log.Printf("panic:%v", v)
				responder.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "サーバーエラーです")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
