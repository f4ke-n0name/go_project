package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/services/auth-service/internal/domain/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	db *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Storage, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to database: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("fail to ping database: %w", err)
	}
	return &Storage{db: pool}, nil
}

func (s *Storage) CreateUser(ctx context.Context, user *models.User) error {
	query := `
			INSERT INTO users(first_name, last_name, email, password_hash, role)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id, created_at, updated_at`

	err := s.db.QueryRow(ctx, query, user.FirstName, user.LastName, user.Email, user.PasswordHash, user.Role).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return pkgerrors.ErrAlreadyExists
		}
		return fmt.Errorf("create user: %w", err)
	}

	return nil
}

func (s *Storage) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
			SELECT id, first_name, last_name, email, password_hash, role, created_at, updated_at 
			FROM users 
			WHERE email = $1`
	user := &models.User{}
	err := s.db.QueryRow(ctx, query, email).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}

	return user, nil
}

func (s *Storage) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	query := `
			SELECT id, first_name, last_name, email, password_hash, role, created_at, updated_at 
			FROM users 
			WHERE id = $1`
	user := &models.User{}
	err := s.db.QueryRow(ctx, query, id).Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.PasswordHash, &user.Role, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pkgerrors.ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

func (s *Storage) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `
			INSERT INTO refresh_tokens(user_id, token, expires_at)
			VALUES ($1, $2, $3)
			`
	_, err := s.db.Exec(ctx, query, userID, token, expiresAt)
	if err != nil {
		return fmt.Errorf("save refresh token: %w", err)
	}
	return nil
}

func (s *Storage) GetRefreshToken(ctx context.Context, token string) (string, error) {
	query := `
			SELECT user_id
			FROM refresh_tokens
			WHERE token = $1 AND expires_at > now()`
	var userID string
	err := s.db.QueryRow(ctx, query, token).Scan(&userID)
	if err != nil {
		return "", fmt.Errorf("get refresh token: %w", err)
	}
	return userID, nil
}

func (s *Storage) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `
			DELETE FROM refresh_tokens
			WHERE token = $1`

	_, err := s.db.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("delete refresh token: %w", err)
	}

	return nil
}

func (s *Storage) Close() {
	s.db.Close()
}

func isUniqueViolation(err error) bool {
	return err != nil &&
		(fmt.Sprintf("%s", err) == "ERROR: duplicate key value violates unique constraint" ||
			containsErrCode(err, "23505"))
}

func containsErrCode(err error, code string) bool {
	return err != nil && len(fmt.Sprintf("%s", err)) > 0 &&
		fmt.Sprintf("%s", err) != "" &&
		fmt.Sprintf("%v", err) != "" &&
		pgErrCode(err) == code
}

func pgErrCode(err error) string {
	type pgErr interface {
		SQLState() string
	}
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState()
	}
	return ""
}
