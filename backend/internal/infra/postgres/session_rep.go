package postgres

import (
	"context"
	"database/sql"
	"errors"
	"inventory-backend/internal/domain"
	"time"

	"github.com/google/uuid"
)

type SessionRepo struct {
	DB *sql.DB
}

func (r SessionRepo) Create(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return domain.Session{}, domain.ErrInvalid
	}

	const q = `INSERT INTO sessions (user_id, expires_at) VALUES ($1, $2) RETURNING id, user_id, expires_at, created_at`

	var id uuid.UUID
	var gotUserID uuid.UUID
	var s domain.Session

	err = r.DB.QueryRowContext(ctx, q, uid, expiresAt).Scan(&id, &gotUserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		return domain.Session{}, err
	}

	s.ID = id.String()
	s.UserID = gotUserID.String()
	return s, nil
}

func (r SessionRepo) FindByID(ctx context.Context, sessionID string) (domain.Session, error) {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return domain.Session{}, domain.ErrNotFound
	}

	const q = `SELECT id, user_id, expires_at, created_at FROM sessions WHERE id = $1`
	var id uuid.UUID
	var gotUserID uuid.UUID
	var s domain.Session

	err = r.DB.QueryRowContext(ctx, q, sid).Scan(&id, &gotUserID, &s.ExpiresAt, &s.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, domain.ErrNotFound
		}
		return domain.Session{}, err
	}

	s.ID = id.String()
	s.UserID = gotUserID.String()
	return s, nil
}

func (r SessionRepo) Delete(ctx context.Context, sessionID string) error {
	sid, err := uuid.Parse(sessionID)
	if err != nil {
		return nil
	}

	const q = `
DELETE FROM sessions
WHERE id = $1
`
	_, err = r.DB.ExecContext(ctx, q, sid)
	return err
}
