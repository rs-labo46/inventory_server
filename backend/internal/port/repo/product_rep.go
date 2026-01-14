package repo

import (
	"context"
	"inventory-backend/internal/domain"
)

type ProductRepository interface {
	CreateWithZeroInventory(ctx context.Context, name string) (domain.Product, domain.Inventory, error)
}
