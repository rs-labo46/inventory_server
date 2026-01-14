package postgres

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestProductRepo_CreateWithZeroInventory_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := ProductRepo{DB: db}

	mock.ExpectBegin()

	pid := uuid.New()
	now := time.Now()

	rows1 := sqlmock.NewRows([]string{"id", "name", "created_at"}).
		AddRow(pid, "Coffee", now)

	mock.ExpectQuery(`INSERT INTO products\(name\) VALUES\(\$1\) RETURNING id,name,created_at`).
		WithArgs("Coffee").
		WillReturnRows(rows1)

	rows2 := sqlmock.NewRows([]string{"product_id", "quantity", "updated_at"}).
		AddRow(pid, 0, now)

	mock.ExpectQuery(`INSERT INTO inventories\(product_id,quantity,updated_at\) VALUES\(\$1,\$2,\$3\) RETURNING product_id, quantity, updated_at`).
		WithArgs(pid, 0, sqlmock.AnyArg()).
		WillReturnRows(rows2)

	mock.ExpectCommit()

	p, inv, err := r.CreateWithZeroInventory(context.Background(), "Coffee")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID != pid.String() || p.Name != "Coffee" {
		t.Fatalf("product mismatch: %+v", p)
	}
	if inv.ProductID != pid.String() || inv.Quantity != 0 {
		t.Fatalf("inventory mismatch: %+v", inv)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestProductRepo_CreateWithZeroInventory_FirstInsertError_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := ProductRepo{DB: db}

	mock.ExpectBegin()

	wantErr := errors.New("insert products failed")
	mock.ExpectQuery(`INSERT INTO products\(name\) VALUES\(\$1\) RETURNING id,name,created_at`).
		WithArgs("Coffee").
		WillReturnError(wantErr)

	mock.ExpectRollback()

	_, _, err = r.CreateWithZeroInventory(context.Background(), "Coffee")
	if !errors.Is(err, wantErr) {
		t.Fatalf("error mismatch: got=%v want=%v", err, wantErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestProductRepo_CreateWithZeroInventory_SecondInsertError_Rollback(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := ProductRepo{DB: db}

	mock.ExpectBegin()

	pid := uuid.New()
	now := time.Now()

	rows1 := sqlmock.NewRows([]string{"id", "name", "created_at"}).
		AddRow(pid, "Coffee", now)

	mock.ExpectQuery(`INSERT INTO products\(name\) VALUES\(\$1\) RETURNING id,name,created_at`).
		WithArgs("Coffee").
		WillReturnRows(rows1)

	wantErr := errors.New("insert inventories failed")
	mock.ExpectQuery(`INSERT INTO inventories\(product_id,quantity,updated_at\) VALUES\(\$1,\$2,\$3\) RETURNING product_id, quantity, updated_at`).
		WithArgs(pid, 0, sqlmock.AnyArg()).
		WillReturnError(wantErr)

	mock.ExpectRollback()

	_, _, err = r.CreateWithZeroInventory(context.Background(), "Coffee")
	if !errors.Is(err, wantErr) {
		t.Fatalf("error mismatch: got=%v want=%v", err, wantErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}
