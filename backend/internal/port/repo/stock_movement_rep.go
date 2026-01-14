package repo

import (
	"context"
	"inventory-backend/internal/domain"
)

type StockMovementRepository interface {
	List(ctx context.Context, productID string, limit int) ([]domain.StockMovement, error)
}
