package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"inventory-backend/internal/application/stock"
	"inventory-backend/internal/delivery/http/middleware/handler"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
)

type movementRepoStub struct {
	list func(ctx context.Context, productID string, limit int) ([]domain.StockMovement, error)
}

func (m movementRepoStub) List(ctx context.Context, productID string, limit int) ([]domain.StockMovement, error) {
	return m.list(ctx, productID, limit)
}

var _ repo.StockMovementRepository = movementRepoStub{}

func TestStockMovementHandler_List_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	h := handler.StockMovementHandler{UC: stock.MovementUsecase{}}

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/stock/movements", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestStockMovementHandler_List_InvalidLimit(t *testing.T) {
	t.Parallel()

	h := handler.StockMovementHandler{UC: stock.MovementUsecase{}}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/stock/movements?limit=abc", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}

func TestStockMovementHandler_List_LimitTooLarge(t *testing.T) {
	t.Parallel()

	h := handler.StockMovementHandler{UC: stock.MovementUsecase{}}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/stock/movements?limit=201", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}

func TestStockMovementHandler_List_Success(t *testing.T) {
	t.Parallel()

	uc := stock.MovementUsecase{
		Movements: movementRepoStub{
			list: func(ctx context.Context, productID string, limit int) ([]domain.StockMovement, error) {
				if productID != "p1" {
					t.Fatalf("productID mismatch: got=%s want=%s", productID, "p1")
				}
				if limit != 50 {
					t.Fatalf("limit mismatch: got=%d want=%d", limit, 50)
				}
				return []domain.StockMovement{
					{
						ID:        "m1",
						ProductID: "p1",
						Delta:     1,
						Reason:    "restock",
						CreatedBy: "u1",
						CreatedAt: time.Now(),
					},
				}, nil
			},
		},
	}

	h := handler.StockMovementHandler{UC: uc}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/stock/movements?product_id=p1", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !json.Valid([]byte(body)) {
		t.Fatalf("not valid JSON: %s", body)
	}
	if !strings.Contains(body, `"items"`) {
		t.Fatalf("response should contain items: %s", body)
	}
}
