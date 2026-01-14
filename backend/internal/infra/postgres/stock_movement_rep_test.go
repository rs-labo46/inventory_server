package postgres

import (
	"context"
	"testing"
	"time"

	"inventory-backend/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestStockMovementRepo_List_DefaultLimit_NoFilter(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockMovementRepo{DB: db}

	now := time.Now()
	id := uuid.New()
	pid := uuid.New()
	createdBy := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "product_id", "delta", "reason", "created_by", "created_at"}).
		AddRow(id, pid, 1, "restock", createdBy, now)

	mock.ExpectQuery(`SELECT id,product_id,delta,reason,created_by,created_at FROM stock_movements ORDER BY created_at DESC LIMIT \$1`).
		WithArgs(50).
		WillReturnRows(rows)

	got, err := r.List(context.Background(), "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len mismatch: got=%d want=%d", len(got), 1)
	}
	if got[0].ID != id.String() {
		t.Fatalf("id mismatch: got=%s want=%s", got[0].ID, id.String())
	}
	if got[0].ProductID != pid.String() {
		t.Fatalf("productID mismatch: got=%s want=%s", got[0].ProductID, pid.String())
	}
	if got[0].CreatedBy != createdBy.String() {
		t.Fatalf("createdBy mismatch: got=%s want=%s", got[0].CreatedBy, createdBy.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestStockMovementRepo_List_WithProductID(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockMovementRepo{DB: db}

	now := time.Now()
	id := uuid.New()
	pid := uuid.New()
	createdBy := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "product_id", "delta", "reason", "created_by", "created_at"}).
		AddRow(id, pid, 2, "ship", createdBy, now)

	mock.ExpectQuery(`SELECT id,product_id,delta,reason,created_by,created_at FROM stock_movements WHERE product_id = \$1 ORDER BY created_at DESC LIMIT \$2`).
		WithArgs(pid, 10).
		WillReturnRows(rows)

	got, err := r.List(context.Background(), pid.String(), 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len mismatch: got=%d want=%d", len(got), 1)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestStockMovementRepo_List_InvalidProductID_ReturnsErrInvalid(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockMovementRepo{DB: db}

	_, gotErr := r.List(context.Background(), "not-a-uuid", 10)
	if gotErr != domain.ErrInvalid {
		t.Fatalf("error mismatch: got=%v want=%v", gotErr, domain.ErrInvalid)
	}
}

func TestStockMovementRepo_List_RowsScanError(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := StockMovementRepo{DB: db}

	rows := sqlmock.NewRows([]string{"id", "product_id", "delta", "reason", "created_by", "created_at"}).
		AddRow(uuid.New(), uuid.New(), 1, "x", uuid.New(), "not-time")

	mock.ExpectQuery(`SELECT id,product_id,delta,reason,created_by,created_at FROM stock_movements ORDER BY created_at DESC LIMIT \$1`).
		WithArgs(1).
		WillReturnRows(rows)

	_, gotErr := r.List(context.Background(), "", 1)
	if gotErr == nil {
		t.Fatalf("expected error but got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}
