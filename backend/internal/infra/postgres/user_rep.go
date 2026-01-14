package postgres

import (
	"context"
	"database/sql"
	"errors"
	"inventory-backend/internal/domain"

	"github.com/google/uuid"
)

type UserRepo struct {
	DB *sql.DB
}

func (r UserRepo) FindByEmail(ctx context.Context, email string) (domain.User, error) {
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE email = $1`

	var id uuid.UUID
	var u domain.User

	err := r.DB.QueryRowContext(ctx, q, email).Scan(&id, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}

	u.ID = id.String()
	return u, nil
}

func (r UserRepo) FindByID(ctx context.Context, id string) (domain.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return domain.User{}, domain.ErrNotFound
	}
	const q = `SELECT id, email, password_hash, created_at FROM users WHERE id = $1`

	var gotID uuid.UUID
	var u domain.User

	err = r.DB.QueryRowContext(ctx, q, uid).Scan(&gotID, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	u.ID = gotID.String()
	return u, nil
}

func (r UserRepo) Create(ctx context.Context, email string, passwordHash string) (domain.User, error) {
	const q = `INSERT INTO users (email, password_hash) VALUES ($1, $2) RETURNING id, email, password_hash, created_at`

	var id uuid.UUID
	var u domain.User
	err := r.DB.QueryRowContext(ctx, q, email, passwordHash).Scan(&id, &u.Email, &u.PasswordHash, &u.CreatedAt)
	if err != nil {
		return domain.User{}, err
	}
	u.ID = id.String()
	return u, nil
}
