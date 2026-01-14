package postgres

import (
	"context"
	"database/sql"
)

func WithTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	tx, err := db.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
	})
	if err != nil {
		return err
	}
	//失敗したら取り消し
	if err := fn(tx); err != nil {
		_ = tx.Rollback()
		return err
	}
	//成功したらコミット
	return tx.Commit()
}
