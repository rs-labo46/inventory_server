package postgres

import (
	"context"
	"database/sql"
	"inventory-backend/internal/domain"
	"time"

	"github.com/google/uuid"
)

type ProductRepo struct {
	DB *sql.DB
}

func (r ProductRepo) CreateWithZeroInventory(ctx context.Context, name string) (domain.Product, domain.Inventory, error) {
	var p domain.Product
	var inv domain.Inventory

	err := WithTx(ctx, r.DB, func(tx *sql.Tx) error {
		const q1 = `INSERT INTO products(name) VALUES($1) RETURNING id,name,created_at`

		var id uuid.UUID
		if err := tx.QueryRowContext(ctx, q1, name).Scan(&id, &p.Name, &p.CreatedAt); err != nil {
			return err
		}
		p.ID = id.String()

		//在庫を数量0で作る
		const q2 = `INSERT INTO inventories(product_id,quantity,updated_at) VALUES($1,$2,$3) RETURNING product_id, quantity, updated_at`
		var pid uuid.UUID
		if err := tx.QueryRowContext(ctx, q2, id, 0, time.Now()).Scan(&pid, &inv.Quantity, &inv.UpdatedAt); err != nil {
			return err
		}
		inv.ProductID = pid.String()
		return nil
	})
	if err != nil {
		return domain.Product{}, domain.Inventory{}, err
	}
	return p, inv, nil
}
