package middleware

import (
	"net/http"
)

type OriginChecker struct {
	Allowed map[string]struct{}
}

func NewOriginChecker(origins []string) OriginChecker {
	m := make(map[string]struct{}, len(origins))
	for _, o := range origins {
		m[o] = struct{}{}
	}
	return OriginChecker{Allowed: m}
}
func (c OriginChecker) Check(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		switch r.Method {
		case http.MethodGet, http.MethodHead:
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		if origin == "" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if _, ok := c.Allowed[origin]; !ok {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
