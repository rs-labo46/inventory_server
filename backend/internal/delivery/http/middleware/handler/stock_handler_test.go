package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"inventory-backend/internal/application/stock"
	"inventory-backend/internal/delivery/http/middleware"
	"inventory-backend/internal/delivery/http/middleware/handler"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
)

type stockRepoStub struct {
	change func(ctx context.Context, actorUserID string, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error)
}

func (s stockRepoStub) Change(ctx context.Context, actorUserID string, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
	return s.change(ctx, actorUserID, productID, delta, reason)
}

var _ repo.StockRepository = stockRepoStub{}

type authSessionRepoStub struct {
	findByID func(ctx context.Context, sessionID string) (domain.Session, error)
	create   func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error)
	delete   func(ctx context.Context, sessionID string) error
}

func (s authSessionRepoStub) FindByID(ctx context.Context, sessionID string) (domain.Session, error) {
	return s.findByID(ctx, sessionID)
}
func (s authSessionRepoStub) Create(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
	return s.create(ctx, userID, expiresAt)
}
func (s authSessionRepoStub) Delete(ctx context.Context, sessionID string) error {
	return s.delete(ctx, sessionID)
}

var _ repo.SessionRepository = authSessionRepoStub{}

func TestStockHandler_Inbound_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	h := handler.StockHandler{UC: stock.Usecase{}}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/stock/inbound", nil)
	rec := httptest.NewRecorder()

	h.Inbound(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestStockHandler_Inbound_Success_WithAuthMiddleware(t *testing.T) {
	t.Parallel()

	uc := stock.Usecase{
		Stock: stockRepoStub{
			change: func(ctx context.Context, actorUserID string, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
				if actorUserID != "u1" {
					t.Fatalf("actorUserID mismatch: got=%s want=%s", actorUserID, "u1")
				}
				if productID != "p1" {
					t.Fatalf("productID mismatch: got=%s want=%s", productID, "p1")
				}
				if delta != 5 {
					t.Fatalf("delta mismatch: got=%d want=%d", delta, 5)
				}
				if reason != "restock" {
					t.Fatalf("reason mismatch: got=%s want=%s", reason, "restock")
				}

				now := time.Date(2026, 1, 14, 12, 0, 0, 0, time.UTC)
				return domain.Inventory{ProductID: "p1", Quantity: 10}, domain.StockMovement{
					ID:        "m1",
					ProductID: "p1",
					Delta:     5,
					Reason:    "restock",
					CreatedBy: "u1",
					CreatedAt: now,
				}, nil
			},
		},
	}

	h := handler.StockHandler{UC: uc}

	authmw := middleware.RequireAuth{
		Sessions: authSessionRepoStub{
			findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
				if sessionID != "s1" {
					t.Fatalf("sessionID mismatch: got=%s want=%s", sessionID, "s1")
				}
				return domain.Session{ID: "s1", UserID: "u1", ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
			},
			create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
				return domain.Session{}, nil
			},
			delete: func(ctx context.Context, sessionID string) error { return nil },
		},
		CookieName: "session_id",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.Inbound(w, r)
	})

	body := `{"product_id":"p1","quantity":5,"reason":"restock"}`
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/stock/inbound", bytes.NewBufferString(body))
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "s1"})
	rec := httptest.NewRecorder()

	authmw.Wrap(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}

	respBody := rec.Body.String()
	if !json.Valid([]byte(respBody)) {
		t.Fatalf("not valid JSON: %s", respBody)
	}
	if !strings.Contains(respBody, `"inventory"`) || !strings.Contains(respBody, `"movement"`) {
		t.Fatalf("response should contain inventory/movement: %s", respBody)
	}
}

func TestStockHandler_Outbound_Success_WithAuthMiddleware(t *testing.T) {
	t.Parallel()

	uc := stock.Usecase{
		Stock: stockRepoStub{
			change: func(ctx context.Context, actorUserID string, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
				if delta != -2 {
					t.Fatalf("delta mismatch: got=%d want=%d", delta, -2)
				}
				return domain.Inventory{ProductID: productID, Quantity: 8}, domain.StockMovement{
					ID:        "m2",
					ProductID: productID,
					Delta:     -2,
					Reason:    reason,
					CreatedBy: actorUserID,
					CreatedAt: time.Now(),
				}, nil
			},
		},
	}

	h := handler.StockHandler{UC: uc}

	authmw := middleware.RequireAuth{
		Sessions: authSessionRepoStub{
			findByID: func(ctx context.Context, sessionID string) (domain.Session, error) {
				return domain.Session{ID: sessionID, UserID: "u1", ExpiresAt: time.Now().Add(10 * time.Minute)}, nil
			},
			create: func(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
				return domain.Session{}, nil
			},
			delete: func(ctx context.Context, sessionID string) error { return nil },
		},
		CookieName: "session_id",
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.Outbound(w, r)
	})

	body := `{"product_id":"p1","quantity":2,"reason":"ship"}`
	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/stock/outbound", bytes.NewBufferString(body))
	req.AddCookie(&http.Cookie{Name: "session_id", Value: "s1"})
	rec := httptest.NewRecorder()

	authmw.Wrap(next).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}
}
