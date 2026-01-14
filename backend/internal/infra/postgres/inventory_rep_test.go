package postgres

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"inventory-backend/internal/domain"
	"inventory-backend/internal/port/repo"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestInventoryRepo_List_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := InventoryRepo{DB: db}

	pid := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "name", "quantity"}).
		AddRow(pid, "Coffee", 10)

	mock.ExpectQuery(`SELECT p.id,p.name,i.quantity FROM inventories i JOIN products p ON p.id = i.product_id ORDER BY p.created_at DESC`).
		WillReturnRows(rows)

	got, err := r.List(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("len mismatch: got=%d want=%d", len(got), 1)
	}
	if got[0].ProductID != pid.String() {
		t.Fatalf("ProductID mismatch: got=%s want=%s", got[0].ProductID, pid.String())
	}
	if got[0].ProductName != "Coffee" || got[0].Quantity != 10 {
		t.Fatalf("row mismatch: %+v", got[0])
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestInventoryRepo_List_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := InventoryRepo{DB: db}

	wantErr := errors.New("query failed")
	mock.ExpectQuery(`SELECT p.id,p.name,i.quantity FROM inventories i JOIN products p ON p.id = i.product_id ORDER BY p.created_at DESC`).
		WillReturnError(wantErr)

	_, err = r.List(context.Background())
	if !errors.Is(err, wantErr) {
		t.Fatalf("error mismatch: got=%v want=%v", err, wantErr)
	}
}

func TestInventoryRepo_List_ScanError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := InventoryRepo{DB: db}

	rows := sqlmock.NewRows([]string{"id", "name", "quantity"}).
		AddRow("not-uuid", "Coffee", 10)

	mock.ExpectQuery(`SELECT p.id,p.name,i.quantity FROM inventories i JOIN products p ON p.id = i.product_id ORDER BY p.created_at DESC`).
		WillReturnRows(rows)

	_, err = r.List(context.Background())
	if err == nil {
		t.Fatalf("expected error but got nil")
	}
}

func TestInventoryRepo_FindByProductID_InvalidUUID_ReturnsErrNotFound(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := InventoryRepo{DB: db}

	_, err = r.FindByProductID(context.Background(), "not-a-uuid")
	if err != domain.ErrNotFound {
		t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrNotFound)
	}
}

func TestInventoryRepo_FindByProductID_NoRows_ReturnsErrNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := InventoryRepo{DB: db}
	pid := uuid.New()

	mock.ExpectQuery(`SELECT product_id, quantity, updated_at FROM inventories WHERE product_id = \$1`).
		WithArgs(pid).
		WillReturnError(sql.ErrNoRows)

	_, err = r.FindByProductID(context.Background(), pid.String())
	if err != domain.ErrNotFound {
		t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrNotFound)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestInventoryRepo_FindByProductID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := InventoryRepo{DB: db}
	pid := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"product_id", "quantity", "updated_at"}).
		AddRow(pid, 7, now)

	mock.ExpectQuery(`SELECT product_id, quantity, updated_at FROM inventories WHERE product_id = \$1`).
		WithArgs(pid).
		WillReturnRows(rows)

	inv, err := r.FindByProductID(context.Background(), pid.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inv.ProductID != pid.String() || inv.Quantity != 7 {
		t.Fatalf("inventory mismatch: %+v", inv)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

var _ repo.InventoryView = repo.InventoryView{}
