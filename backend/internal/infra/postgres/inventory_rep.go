package postgres

import (
	"context"
	"database/sql"
	"errors"
	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"

	"github.com/google/uuid"
)

type InventoryRepo struct {
	DB *sql.DB
}

func (r InventoryRepo) List(ctx context.Context) ([]repo.InventoryView, error) {
	const q = `SELECT p.id,p.name,i.quantity FROM inventories i JOIN products p ON p.id = i.product_id ORDER BY p.created_at DESC`
	rows, err := r.DB.QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]repo.InventoryView, 0)
	for rows.Next() {
		var pid uuid.UUID
		var v repo.InventoryView
		if err := rows.Scan(&pid, &v.ProductName, &v.Quantity); err != nil {
			return nil, err
		}
		v.ProductID = pid.String()
		out = append(out, v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (r InventoryRepo) FindByProductID(ctx context.Context, productID string) (domain.Inventory, error) {
	uid, err := uuid.Parse(productID)
	if err != nil {
		return domain.Inventory{}, domain.ErrNotFound
	}

	const q = `SELECT product_id, quantity, updated_at FROM inventories WHERE product_id = $1`
	var pid uuid.UUID
	var inv domain.Inventory

	err = r.DB.QueryRowContext(ctx, q, uid).Scan(&pid, &inv.Quantity, &inv.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Inventory{}, domain.ErrNotFound
		}
		return domain.Inventory{}, err
	}
	inv.ProductID = pid.String()
	return inv, nil
}
