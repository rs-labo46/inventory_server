package stock

import (
	"context"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
)

type MovementUsecase struct {
	Movements repo.StockMovementRepository
}

func (u MovementUsecase) List(ctx context.Context, productID string, limit int) ([]domain.StockMovement, error) {
	if limit <= 0 {
		limit = 50
	}
	return u.Movements.List(ctx, productID, limit)
}
