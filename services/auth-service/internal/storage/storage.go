package storage

import (
	"context"
	"time"

	"github.com/f4ke-n0name/autoparts-hub/services/auth-service/internal/domain/models"
)

type UserStorage interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
	GetUserByID(ctx context.Context, id string) (*models.User, error)
}

type TokenStorage interface {
	SaveRefreshToken(ctx context.Context, userId, token string, expiresAt time.Time) error
	GetRefreshToken(ctx context.Context, token string) (string, error)
	DeleteRefreshToken(ctx context.Context, token string) error
}
