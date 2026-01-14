package repo

import (
	"context"
	"inventory-backend/internal/domain"
	"time"
)

type SessionRepository interface {
	Create(ctx context.Context, userID string, expiresAt time.Time) (domain.Session, error)
	FindByID(ctx context.Context, sessionID string) (domain.Session, error)
	Delete(ctx context.Context, sessionID string) error
}
