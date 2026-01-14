package stock

import (
	"context"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"
)

type Usecase struct {
	Stock repo.StockRepository
}

type ChangeResult struct {
	Inventory domain.Inventory
	Movement  domain.StockMovement
}

// プラス
func (u Usecase) Inbound(ctx context.Context, actorUserID string, productID string, qty int, reason string) (ChangeResult, error) {
	if qty <= 0 {
		return ChangeResult{}, domain.ErrInvalid
	}

	inv, mv, err := u.Stock.Change(ctx, actorUserID, productID, qty, reason)
	if err != nil {
		return ChangeResult{}, err
	}
	return ChangeResult{
		Inventory: inv,
		Movement:  mv,
	}, nil
}

// マイナス
func (u Usecase) Outbound(ctx context.Context, actorUserID string, productID string, qty int, reason string) (ChangeResult, error) {
	if qty <= 0 {
		return ChangeResult{}, domain.ErrInvalid
	}
	inv, mv, err := u.Stock.Change(ctx, actorUserID, productID, -qty, reason)
	if err != nil {
		return ChangeResult{}, err
	}
	return ChangeResult{
		Inventory: inv,
		Movement:  mv,
	}, nil
}
