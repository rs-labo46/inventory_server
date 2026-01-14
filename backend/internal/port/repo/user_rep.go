package repo

import (
	"context"
	"inventory-backend/internal/domain"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (domain.User, error)
	FindByID(ctx context.Context, id string) (domain.User, error)
	Create(ctx context.Context, email string, passwordHash string) (domain.User, error)
}
