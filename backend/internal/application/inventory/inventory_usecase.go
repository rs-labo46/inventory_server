package inventory

import (
	"context"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
)

type Usecase struct {
	Products    repo.ProductRepository
	Inventories repo.InventoryRepository
}

type CreateProductResult struct {
	Product   domain.Product
	Inventory domain.Inventory
}

func (u Usecase) CreateProduct(ctx context.Context, name string) (CreateProductResult, error) {
	if name == "" {
		return CreateProductResult{}, domain.ErrInvalid
	}

	p, inv, err := u.Products.CreateWithZeroInventory(ctx, name)
	if err != nil {
		return CreateProductResult{}, err
	}
	return CreateProductResult{Product: p, Inventory: inv}, nil
}

func (u Usecase) ListInventories(ctx context.Context) ([]repo.InventoryView, error) {
	return u.Inventories.List(ctx)
}
