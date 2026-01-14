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

func TestSessionRepo_Create_InvalidUserID_ReturnsErrInvalid(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := SessionRepo{DB: db}

	_, err = r.Create(context.Background(), "not-a-uuid", time.Now())
	if err != domain.ErrInvalid {
		t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrInvalid)
	}
}

func TestSessionRepo_Create_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := SessionRepo{DB: db}

	sid := uuid.New()
	uid := uuid.New()
	expires := time.Now().Add(1 * time.Hour)
	created := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "expires_at", "created_at"}).
		AddRow(sid, uid, expires, created)

	mock.ExpectQuery(`INSERT INTO sessions \(user_id, expires_at\) VALUES \(\$1, \$2\) RETURNING id, user_id, expires_at, created_at`).
		WithArgs(uid, expires).
		WillReturnRows(rows)

	s, err := r.Create(context.Background(), uid.String(), expires)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != sid.String() || s.UserID != uid.String() {
		t.Fatalf("session mismatch: %+v", s)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestSessionRepo_FindByID_InvalidUUID_ReturnsErrNotFound(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := SessionRepo{DB: db}

	_, err = r.FindByID(context.Background(), "not-a-uuid")
	if err != domain.ErrNotFound {
		t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrNotFound)
	}
}

func TestSessionRepo_FindByID_NoRows_ReturnsErrNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := SessionRepo{DB: db}
	sid := uuid.New()

	mock.ExpectQuery(`SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = \$1`).
		WithArgs(sid).
		WillReturnError(sql.ErrNoRows)

	_, err = r.FindByID(context.Background(), sid.String())
	if err != domain.ErrNotFound {
		t.Fatalf("error mismatch: got=%v want=%v", err, domain.ErrNotFound)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestSessionRepo_FindByID_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := SessionRepo{DB: db}

	sid := uuid.New()
	uid := uuid.New()
	expires := time.Now().Add(1 * time.Hour)
	created := time.Now()

	rows := sqlmock.NewRows([]string{"id", "user_id", "expires_at", "created_at"}).
		AddRow(sid, uid, expires, created)

	mock.ExpectQuery(`SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = \$1`).
		WithArgs(sid).
		WillReturnRows(rows)

	s, err := r.FindByID(context.Background(), sid.String())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.ID != sid.String() || s.UserID != uid.String() {
		t.Fatalf("session mismatch: %+v", s)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}

func TestSessionRepo_Delete_InvalidUUID_ReturnsNil(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := SessionRepo{DB: db}

	if err := r.Delete(context.Background(), "not-a-uuid"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSessionRepo_Delete_ExecError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock.New error: %v", err)
	}
	defer func() { _ = db.Close() }()

	r := SessionRepo{DB: db}
	sid := uuid.New()

	wantErr := errors.New("exec failed")
	mock.ExpectExec(`DELETE FROM sessions`).
		WithArgs(sid).
		WillReturnError(wantErr)

	err = r.Delete(context.Background(), sid.String())
	if !errors.Is(err, wantErr) {
		t.Fatalf("error mismatch: got=%v want=%v", err, wantErr)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations not met: %v", err)
	}
}
