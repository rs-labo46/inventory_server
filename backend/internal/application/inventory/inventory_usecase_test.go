package inventory_test

import (
	"context"
	"errors"
	"inventory-backend/internal/application/inventory"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
	"testing"
)

type productRepoStub struct {
	createWithZero func(ctx context.Context, name string) (domain.Product, domain.Inventory, error)
}

func (p productRepoStub) CreateWithZeroInventory(ctx context.Context, name string) (domain.Product, domain.Inventory, error) {
	return p.createWithZero(ctx, name)
}

var _ repo.ProductRepository = productRepoStub{}

type inventoryRepoStub struct {
	list func(ctx context.Context) ([]repo.InventoryView, error)

	findByProductID func(ctx context.Context, productID string) (domain.Inventory, error)
}

func (i inventoryRepoStub) List(ctx context.Context) ([]repo.InventoryView, error) {
	if i.list == nil {
		return nil, errors.New("List is not stubbed")
	}
	return i.list(ctx)
}

func (i inventoryRepoStub) FindByProductID(ctx context.Context, productID string) (domain.Inventory, error) {
	if i.findByProductID == nil {
		return domain.Inventory{}, errors.New("FindByProductID is not stubbed")
	}
	return i.findByProductID(ctx, productID)
}

var _ repo.InventoryRepository = inventoryRepoStub{}

func TestInventoryUsecase_CreateProduct(t *testing.T) {
	t.Parallel()

	t.Run("nameが空なら ErrInvalid", func(t *testing.T) {
		t.Parallel()

		uc := inventory.Usecase{
			Products: productRepoStub{
				createWithZero: func(ctx context.Context, name string) (domain.Product, domain.Inventory, error) {
					t.Fatalf("nameが空なら repo は呼ばれない想定")
					return domain.Product{}, domain.Inventory{}, nil
				},
			},
			Inventories: inventoryRepoStub{
				list: func(ctx context.Context) ([]repo.InventoryView, error) { return nil, nil },
			},
		}

		_, err := uc.CreateProduct(context.Background(), "")
		if !errors.Is(err, domain.ErrInvalid) {
			t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrInvalid)
		}
	})

	t.Run("正常系: repoの結果が詰め替えられる", func(t *testing.T) {
		t.Parallel()

		wantProduct := domain.Product{}
		wantInv := domain.Inventory{}

		uc := inventory.Usecase{
			Products: productRepoStub{
				createWithZero: func(ctx context.Context, name string) (domain.Product, domain.Inventory, error) {
					if name != "coffee" {
						t.Fatalf("name mismatch: got=%s want=%s", name, "coffee")
					}
					return wantProduct, wantInv, nil
				},
			},
			Inventories: inventoryRepoStub{
				list: func(ctx context.Context) ([]repo.InventoryView, error) { return nil, nil },
			},
		}

		got, err := uc.CreateProduct(context.Background(), "coffee")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Product != wantProduct {
			t.Fatalf("product mismatch")
		}
		if got.Inventory != wantInv {
			t.Fatalf("inventory mismatch")
		}
	})
}

func TestInventoryUsecase_ListInventories(t *testing.T) {
	t.Parallel()

	t.Run("pass-throughで一覧を返す", func(t *testing.T) {
		t.Parallel()

		want := []repo.InventoryView{}

		uc := inventory.Usecase{
			Products: productRepoStub{
				createWithZero: func(ctx context.Context, name string) (domain.Product, domain.Inventory, error) {
					return domain.Product{}, domain.Inventory{}, nil
				},
			},
			Inventories: inventoryRepoStub{
				list: func(ctx context.Context) ([]repo.InventoryView, error) {
					return want, nil
				},
			},
		}

		got, err := uc.ListInventories(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == nil {
			t.Fatalf("got should not be nil")
		}
	})
}
