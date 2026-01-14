package stock_test

import (
	"context"
	"testing"

	"inventory-backend/internal/application/stock"
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

func TestMovementUsecase_List(t *testing.T) {
	t.Parallel()

	t.Run("limit<=0 は50に補正", func(t *testing.T) {
		t.Parallel()

		uc := stock.MovementUsecase{
			Movements: movementRepoStub{
				list: func(ctx context.Context, productID string, limit int) ([]domain.StockMovement, error) {
					if limit != 50 {
						t.Fatalf("limit mismatch: got=%d want=%d", limit, 50)
					}
					return []domain.StockMovement{}, nil
				},
			},
		}

		_, err := uc.List(context.Background(), "p1", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("limit>0 はそのまま渡される", func(t *testing.T) {
		t.Parallel()

		uc := stock.MovementUsecase{
			Movements: movementRepoStub{
				list: func(ctx context.Context, productID string, limit int) ([]domain.StockMovement, error) {
					if limit != 10 {
						t.Fatalf("limit mismatch: got=%d want=%d", limit, 10)
					}
					return []domain.StockMovement{}, nil
				},
			},
		}

		_, err := uc.List(context.Background(), "p1", 10)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
