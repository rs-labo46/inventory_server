package postgres

import (
	"context"
	"database/sql"
	"inventory-backend/internal/domain"

	"github.com/google/uuid"
)

type StockMovementRepo struct {
	DB *sql.DB
}

func (r StockMovementRepo) List(ctx context.Context, productID string, limit int) ([]domain.StockMovement, error) {
	if limit <= 0 {
		limit = 50
	}

	//productIDが指定されている場合だけuuidに変換
	var pid uuid.UUID
	var hasPID bool
	if productID != "" {
		parsed, err := uuid.Parse(productID)
		if err != nil {
			return nil, domain.ErrInvalid
		}
		pid = parsed
		hasPID = true
	}
	//絞り込みあり/なし
	var (
		rows *sql.Rows
		err  error
	)
	if hasPID {
		const q = `SELECT id,product_id,delta,reason,created_by,created_at FROM stock_movements WHERE product_id = $1 ORDER BY created_at DESC LIMIT $2`
		rows, err = r.DB.QueryContext(ctx, q, pid, limit)
	} else {
		const q = `SELECT id,product_id,delta,reason,created_by,created_at FROM stock_movements ORDER BY created_at DESC LIMIT $1`
		rows, err = r.DB.QueryContext(ctx, q, limit)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]domain.StockMovement, 0)
	for rows.Next() {
		var (
			id        uuid.UUID
			p         uuid.UUID
			createdBy uuid.UUID
			m         domain.StockMovement
		)

		if err := rows.Scan(&id, &p, &m.Delta, &m.Reason, &createdBy, &m.CreatedAt); err != nil {
			return nil, err
		}
		m.ID = id.String()
		m.ProductID = p.String()
		m.CreatedBy = createdBy.String()
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}
