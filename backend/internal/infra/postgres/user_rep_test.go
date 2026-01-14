package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"inventory-backend/internal/domain"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
)

func TestUserRepo_FindByEmail_NotFound(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := UserRepo{DB: db}

	mock.ExpectQuery(`SELECT id, email, password_hash, created_at FROM users WHERE email = \$1`).
		WithArgs("x@example.com").
		WillReturnError(sql.ErrNoRows)

	_, gotErr := r.FindByEmail(context.Background(), "x@example.com")
	if gotErr != domain.ErrNotFound {
		t.Fatalf("error mismatch: got=%v want=%v", gotErr, domain.ErrNotFound)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestUserRepo_FindByEmail_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := UserRepo{DB: db}

	id := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at"}).
		AddRow(id, "x@example.com", "hash", now)

	mock.ExpectQuery(`SELECT id, email, password_hash, created_at FROM users WHERE email = \$1`).
		WithArgs("x@example.com").
		WillReturnRows(rows)

	u, err := r.FindByEmail(context.Background(), "x@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != id.String() {
		t.Fatalf("id mismatch: got=%s want=%s", u.ID, id.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestUserRepo_FindByID_InvalidUUID_ReturnsErrNotFound(t *testing.T) {
	t.Parallel()

	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := UserRepo{DB: db}

	_, gotErr := r.FindByID(context.Background(), "not-a-uuid")
	if gotErr != domain.ErrNotFound {
		t.Fatalf("error mismatch: got=%v want=%v", gotErr, domain.ErrNotFound)
	}
}

func TestUserRepo_Create_Success(t *testing.T) {
	t.Parallel()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := UserRepo{DB: db}

	id := uuid.New()
	now := time.Now()

	rows := sqlmock.NewRows([]string{"id", "email", "password_hash", "created_at"}).
		AddRow(id, "x@example.com", "hash", now)

	mock.ExpectQuery(`INSERT INTO users \(email, password_hash\) VALUES \(\$1, \$2\) RETURNING id, email, password_hash, created_at`).
		WithArgs("x@example.com", "hash").
		WillReturnRows(rows)

	u, err := r.Create(context.Background(), "x@example.com", "hash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != id.String() {
		t.Fatalf("id mismatch: got=%s want=%s", u.ID, id.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}
