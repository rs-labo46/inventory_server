package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRequestID_SetsRequestIDHeader(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "http://example.com/healthz", nil)
	rec := httptest.NewRecorder()

	h := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}

	id := rec.Header().Get("X-Request-Id")
	if id == "" {
		t.Fatalf("X-Request-Id should be set")
	}

	if len(id) != 32 {
		t.Fatalf("X-Request-Id length mismatch: got=%d want=%d", len(id), 32)
	}
}

func TestRecover_ConvertsPanicTo500(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/x", nil)
	rec := httptest.NewRecorder()

	h := Recover(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusInternalServerError)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "application/json") {
		t.Fatalf("Content-Type should contain application/json: got=%s", ct)
	}
	if !strings.Contains(rec.Body.String(), `"INTERNAL_ERROR"`) {
		t.Fatalf("body should contain INTERNAL_ERROR: %s", rec.Body.String())
	}
}

func TestOriginChecker_OptionsReturns204(t *testing.T) {
	oc := NewOriginChecker([]string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodOptions, "http://example.com/api/x", nil)
	rec := httptest.NewRecorder()

	h := oc.Check(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("OPTIONSはnextに到達しない想定")
	}))

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusNoContent)
	}
}

func TestOriginChecker_GetPassThrough(t *testing.T) {
	oc := NewOriginChecker([]string{})

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/x", nil)
	rec := httptest.NewRecorder()

	called := false
	h := oc.Check(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}
	if !called {
		t.Fatalf("next handler should be called for GET")
	}
}

func TestOriginChecker_PostWithoutOriginForbidden(t *testing.T) {
	oc := NewOriginChecker([]string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/x", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()

	h := oc.Check(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("Origin無しはnextに到達しない想定")
	}))

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusForbidden)
	}
}

func TestOriginChecker_PostWithDisallowedOriginForbidden(t *testing.T) {
	oc := NewOriginChecker([]string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/x", strings.NewReader(`{}`))
	req.Header.Set("Origin", "http://evil.example.com")
	rec := httptest.NewRecorder()

	h := oc.Check(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("許可されてないOriginはnextに到達しない想定")
	}))

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusForbidden)
	}
}

func TestOriginChecker_PostWithAllowedOriginPassThrough(t *testing.T) {
	oc := NewOriginChecker([]string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/x", strings.NewReader(`{}`))
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()

	called := false
	h := oc.Check(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}
	if !called {
		t.Fatalf("next handler should be called for allowed origin")
	}
}
