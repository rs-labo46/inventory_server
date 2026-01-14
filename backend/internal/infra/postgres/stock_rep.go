package postgres

import (
	"context"
	"database/sql"
	"errors"
	"inventory-backend/internal/domain"
	"time"

	"github.com/google/uuid"
)

type StockRepo struct {
	DB *sql.DB
}

func (r StockRepo) Change(ctx context.Context, actorUserID string, productID string, delta int, reason string) (domain.Inventory, domain.StockMovement, error) {
	pid, err := uuid.Parse(productID)
	if err != nil {
		return domain.Inventory{}, domain.StockMovement{}, domain.ErrInsufficientStock
	}

	aid, err := uuid.Parse(actorUserID)
	if err != nil {
		return domain.Inventory{}, domain.StockMovement{}, domain.ErrUnauthorized
	}

	var inv domain.Inventory
	var mv domain.StockMovement

	//在庫更新と履歴追加
	err = WithTx(ctx, r.DB, func(tx *sql.Tx) error {
		//同時更新を防ぐ
		const qLock = `SELECT quantity FROM inventories WHERE product_id = $1 FOR UPDATE`

		var current int
		if err := tx.QueryRowContext(ctx, qLock, pid).Scan(&current); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return domain.ErrNotFound
			}
			return err
		}

		//新い数量を計算する。
		next := current + delta

		//出庫時マイナスになったら失敗
		if next < 0 {
			return domain.ErrInsufficientStock
		}

		//inventoriesを更新
		const qUpd = `UPDATE inventories SET quantity = $2,updated_at = $3 WHERE product_id = $1 RETURNING product_id,quantity,updated_at`

		var gotPID uuid.UUID
		if err := tx.QueryRowContext(ctx, qUpd, pid, next, time.Now()).Scan(&gotPID, &inv.Quantity, &inv.UpdatedAt); err != nil {
			return err
		}
		inv.ProductID = gotPID.String()

		//履歴を入れる
		const qInsMv = `INSERT INTO stock_movements(product_id,delta,reason,created_by)VALUES($1,$2,$3,$4) RETURNING id,product_id,delta,reason,created_by,created_at`

		var id uuid.UUID
		var mpid uuid.UUID
		var createdBy uuid.UUID

		if err := tx.QueryRowContext(ctx, qInsMv, pid, delta, reason, aid).Scan(&id, &mpid, &mv.Delta, &mv.Reason, &createdBy, &mv.CreatedAt); err != nil {
			return err
		}
		mv.ID = id.String()
		mv.ProductID = mpid.String()
		mv.CreatedBy = createdBy.String()

		return nil
	})
	if err != nil {
		return domain.Inventory{}, domain.StockMovement{}, err
	}
	return inv, mv, nil
}
