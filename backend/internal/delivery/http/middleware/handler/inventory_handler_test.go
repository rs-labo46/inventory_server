package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"inventory-backend/internal/application/inventory"
	"inventory-backend/internal/delivery/http/middleware/handler"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
)

type productRepoStub2 struct {
	createWithZero func(ctx context.Context, name string) (domain.Product, domain.Inventory, error)
}

func (p productRepoStub2) CreateWithZeroInventory(ctx context.Context, name string) (domain.Product, domain.Inventory, error) {
	return p.createWithZero(ctx, name)
}

var _ repo.ProductRepository = productRepoStub2{}

type inventoryRepoStub2 struct {
	list            func(ctx context.Context) ([]repo.InventoryView, error)
	findByProductID func(ctx context.Context, productID string) (domain.Inventory, error)
}

func (i inventoryRepoStub2) List(ctx context.Context) ([]repo.InventoryView, error) {
	return i.list(ctx)
}
func (i inventoryRepoStub2) FindByProductID(ctx context.Context, productID string) (domain.Inventory, error) {
	return i.findByProductID(ctx, productID)
}

var _ repo.InventoryRepository = inventoryRepoStub2{}

func TestInventoryHandler_List_MethodNotAllowed(t *testing.T) {
	t.Parallel()

	h := handler.InventoryHandler{UC: inventory.Usecase{}}

	req := httptest.NewRequest(http.MethodPost, "http://example.com/api/inventories", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusMethodNotAllowed)
	}
}

func TestInventoryHandler_List_Success(t *testing.T) {
	t.Parallel()

	uc := inventory.Usecase{
		Products: productRepoStub2{
			createWithZero: func(ctx context.Context, name string) (domain.Product, domain.Inventory, error) {
				return domain.Product{}, domain.Inventory{}, nil
			},
		},
		Inventories: inventoryRepoStub2{
			list: func(ctx context.Context) ([]repo.InventoryView, error) {
				return []repo.InventoryView{
					{ProductID: "p1", ProductName: "Coffee", Quantity: 10},
				}, nil
			},
			findByProductID: func(ctx context.Context, productID string) (domain.Inventory, error) {
				return domain.Inventory{}, nil
			},
		},
	}

	h := handler.InventoryHandler{UC: uc}

	req := httptest.NewRequest(http.MethodGet, "http://example.com/api/inventories", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status mismatch: got=%d want=%d", rec.Code, http.StatusOK)
	}

}
