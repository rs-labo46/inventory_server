package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"inventory-backend/internal/application/inventory"
	"inventory-backend/internal/delivery/http/middleware/handler"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
)

type productRepoStub struct {
	createWithZero func(ctx context.Context, name string) (domain.Product, domain.Inventory, error)
}

func (p productRepoStub) CreateWithZeroInventory(ctx context.Context, name string) (domain.Product, domain.Inventory, error) {
	return p.createWithZero(ctx, name)
}

var _ repo.ProductRepository = productRepoStub{}

type inventoryRepoStub struct {
	list            func(ctx context.Context) ([]repo.InventoryView, error)
	findByProductID func(ctx context.Context, productID string) (domain.Inventory, error)
}

func (i inventoryRepoStub) List(ctx context.Context) ([]repo.InventoryView, error) {
	return i.list(ctx)
}

func (i inventoryRepoStub) FindByProductID(ctx context.Context, productID string) (domain.Inventory, error) {
	return i.findByProductID(ctx, productID)
}

var _ repo.InventoryRepository = inventoryRepoStub{}

func TestProductHandler_Create_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	h := handler.ProductHandler{UC: inventory.Usecase{}}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/products", nil)
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestProductHandler_Create_BadJSON(t *testing.T) {
	t.Parallel()

	h := handler.ProductHandler{UC: inventory.Usecase{}}

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/products", bytes.NewBufferString("{bad json"))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}

func TestProductHandler_Create_EmptyName(t *testing.T) {
	t.Parallel()

	h := handler.ProductHandler{UC: inventory.Usecase{}}

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/products", bytes.NewBufferString(`{"name":""}`))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusBadRequest)
	}
}

func TestProductHandler_Create_Success(t *testing.T) {
	t.Parallel()

	uc := inventory.Usecase{
		Products: productRepoStub{
			createWithZero: func(ctx context.Context, name string) (domain.Product, domain.Inventory, error) {
				if name != "coffee" {
					t.Fatalf("name mismatch: got=%s want=%s", name, "coffee")
				}
				return domain.Product{ID: "p1", Name: "coffee"}, domain.Inventory{ProductID: "p1", Quantity: 0}, nil
			},
		},
		Inventories: inventoryRepoStub{
			list: func(ctx context.Context) ([]repo.InventoryView, error) { return nil, nil },
			findByProductID: func(ctx context.Context, productID string) (domain.Inventory, error) {
				return domain.Inventory{}, nil
			},
		},
	}

	h := handler.ProductHandler{UC: uc}

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/products", bytes.NewBufferString(`{"name":"coffee"}`))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusCreated)
	}

	body := rec.Body.String()
	if !json.Valid([]byte(body)) {
		t.Fatalf("not valid JSON: %s", body)
	}
	if !strings.Contains(body, `"product"`) || !strings.Contains(body, `"inventory"`) {
		t.Fatalf("response should contain product/inventory: %s", body)
	}
	if !strings.Contains(body, `"id"`) || !strings.Contains(body, `"p1"`) {
		t.Fatalf("response should contain product id: %s", body)
	}
}
