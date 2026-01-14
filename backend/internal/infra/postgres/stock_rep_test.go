package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"inventory-backend/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestStockRepo_Change_InvalidProductID_ReturnsErrInsufficientStock(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockRepo{DB: db}

	_, _, gotErr := r.Change(context.Background(), uuid.New().String(), "not-a-uuid", 1, "x")

	if gotErr != domain.ErrInsufficientStock {
		t.Fatalf("error mismatch: got=%v want=%v", gotErr, domain.ErrInsufficientStock)
	}
}

func TestStockRepo_Change_InvalidActorUserID_ReturnsErrUnauthorized(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockRepo{DB: db}

	_, _, gotErr := r.Change(context.Background(), "not-a-uuid", uuid.New().String(), 1, "x")
	if gotErr != domain.ErrUnauthorized {
		t.Fatalf("error mismatch: got=%v want=%v", gotErr, domain.ErrUnauthorized)
	}
}

func TestStockRepo_Change_NotFound_WhenNoInventoryRow(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockRepo{DB: db}

	pid := uuid.New()
	aid := uuid.New()

	mock.ExpectBegin()

	mock.ExpectQuery(`SELECT quantity FROM inventories WHERE product_id = \$1 FOR UPDATE`).
		WithArgs(pid).
		WillReturnError(sql.ErrNoRows)

	mock.ExpectRollback()

	_, _, gotErr := r.Change(context.Background(), aid.String(), pid.String(), 1, "x")
	if gotErr != domain.ErrNotFound {
		t.Fatalf("error mismatch: got=%v want=%v", gotErr, domain.ErrNotFound)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestStockRepo_Change_InsufficientStock_WhenNextNegative(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockRepo{DB: db}

	pid := uuid.New()
	aid := uuid.New()

	mock.ExpectBegin()

	lockRows := sqlmock.NewRows([]string{"quantity"}).AddRow(1)
	mock.ExpectQuery(`SELECT quantity FROM inventories WHERE product_id = \$1 FOR UPDATE`).
		WithArgs(pid).
		WillReturnRows(lockRows)

	mock.ExpectRollback()

	_, _, gotErr := r.Change(context.Background(), aid.String(), pid.String(), -2, "ship")
	if gotErr != domain.ErrInsufficientStock {
		t.Fatalf("error mismatch: got=%v want=%v", gotErr, domain.ErrInsufficientStock)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestStockRepo_Change_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockRepo{DB: db}

	pid := uuid.New()
	aid := uuid.New()

	mock.ExpectBegin()

	lockRows := sqlmock.NewRows([]string{"quantity"}).AddRow(3)
	mock.ExpectQuery(`SELECT quantity FROM inventories WHERE product_id = \$1 FOR UPDATE`).
		WithArgs(pid).
		WillReturnRows(lockRows)

	now := time.Now()
	updRows := sqlmock.NewRows([]string{"product_id", "quantity", "updated_at"}).
		AddRow(pid, 5, now)

	mock.ExpectQuery(`UPDATE inventories SET quantity = \$2,updated_at = \$3 WHERE product_id = \$1 RETURNING product_id,quantity,updated_at`).
		WithArgs(pid, 5, sqlmock.AnyArg()).
		WillReturnRows(updRows)

	mvID := uuid.New()
	mvRows := sqlmock.NewRows([]string{"id", "product_id", "delta", "reason", "created_by", "created_at"}).
		AddRow(mvID, pid, 2, "restock", aid, now)

	mock.ExpectQuery(`INSERT INTO stock_movements\(product_id,delta,reason,created_by\)VALUES\(\$1,\$2,\$3,\$4\) RETURNING id,product_id,delta,reason,created_by,created_at`).
		WithArgs(pid, 2, "restock", aid).
		WillReturnRows(mvRows)

	mock.ExpectCommit()

	inv, mv, gotErr := r.Change(context.Background(), aid.String(), pid.String(), 2, "restock")
	if gotErr != nil {
		t.Fatalf("unexpected error: %v", gotErr)
	}
	if inv.ProductID != pid.String() {
		t.Fatalf("inv.ProductID mismatch: got=%s want=%s", inv.ProductID, pid.String())
	}
	if inv.Quantity != 5 {
		t.Fatalf("inv.Quantity mismatch: got=%d want=%d", inv.Quantity, 5)
	}
	if mv.ID != mvID.String() {
		t.Fatalf("mv.ID mismatch: got=%s want=%s", mv.ID, mvID.String())
	}
	if mv.Delta != 2 {
		t.Fatalf("mv.Delta mismatch: got=%d want=%d", mv.Delta, 2)
	}
	if mv.CreatedBy != aid.String() {
		t.Fatalf("mv.CreatedBy mismatch: got=%s want=%s", mv.CreatedBy, aid.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestStockRepo_Change_LockQueryOtherError_ReturnsThatError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockRepo{DB: db}

	pid := uuid.New()
	aid := uuid.New()

	mock.ExpectBegin()

	wantErr := errors.New("db down")
	mock.ExpectQuery(`SELECT quantity FROM inventories WHERE product_id = \$1 FOR UPDATE`).
		WithArgs(pid).
		WillReturnError(wantErr)

	mock.ExpectRollback()

	_, _, gotErr := r.Change(context.Background(), aid.String(), pid.String(), 1, "x")
	if !errors.Is(gotErr, wantErr) {
		t.Fatalf("error mismatch: got=%v want=%v", gotErr, wantErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}
