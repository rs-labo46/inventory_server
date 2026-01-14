package repo

import (
	"context"
	"inventory-backend/internal/domain"
)

type InventoryView struct {
	ProductID   string
	ProductName string
	Quantity    int
}

type InventoryRepository interface {
	List(ctx context.Context) ([]InventoryView, error)
	FindByProductID(ctx context.Context, productID string) (domain.Inventory, error)
}
