package repo

import (
	"context"
	"inventory-backend/internal/domain"
)

type StockRepository interface {
	Change(ctx context.Context, actorUserID string, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error)
}
