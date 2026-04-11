package auth

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/pkg/jwt"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/auth-service/internal/domain/models"
	"github.com/f4ke-n0name/autoparts-hub/services/auth-service/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	log        *slog.Logger
	users      storage.UserStorage
	tokens     storage.TokenStorage
	jwtManager *jwt.Manager
}

func New(log *slog.Logger, users storage.UserStorage, tokens storage.TokenStorage, jwtManager *jwt.Manager) *Service {
	return &Service{
		log:        log,
		users:      users,
		tokens:     tokens,
		jwtManager: jwtManager,
	}
}

func (s *Service) Register(ctx context.Context, firstName, lastName, email, password string) (accessToken, refreshToken string, err error) {
	const op = "auth_service.Register"
	log := s.log.With(slog.String("op", op), slog.String("email", email))
	log.Info("Register user")

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", fmt.Errorf("%s: %w", op, err)
	}

	user := &models.User{
		FirstName:    firstName,
		LastName:     lastName,
		Email:        email,
		PasswordHash: string(passwordHash),
		Role:         models.RoleBuyer,
	}

	if err := s.users.CreateUser(ctx, user); err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrAlreadyExists) {
			log.Warn("User already exists")
			return "", "", pkgerrors.ErrAlreadyExists
		}
		log.Error("Failed to create user", err.Error())
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	return s.issueTokens(ctx, op, user.ID, string(user.Role))
}

func (s *Service) Login(ctx context.Context, email, password string) (accessToken, refreshToken string, err error) {
	const op = "auth_service.Login"
	log := s.log.With(slog.String("op", op), slog.String("email", email))

	log.Info("Login user")

	user, err := s.users.GetUserByEmail(ctx, email)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			return "", "", pkgerrors.ErrNotFound
		}
		log.Error("Failed to get user", err.Error())
		return "", "", fmt.Errorf("%s: %w", op, err)
	}
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		log.Warn("Password is invalid")
		return "", "", pkgerrors.ErrUnauthorized
	}
	return s.issueTokens(ctx, op, user.ID, string(user.Role))
}

func (s *Service) RefreshToken(ctx context.Context, refreshToken string) (newAccessToken string, newRefreshToken string, err error) {
	const op = "auth_service.RefreshToken"
	log := s.log.With(slog.String("op", op))

	claims, err := s.jwtManager.ParseToken(refreshToken)
	if err != nil {
		log.Error("Failed to parse refresh token")
		return "", "", pkgerrors.ErrUnauthorized
	}

	userID, err := s.tokens.GetRefreshToken(ctx, refreshToken)
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) {
			log.Warn("Refresh token not found or expired")
			return "", "", pkgerrors.ErrUnauthorized
		}
		log.Error("Failed to get refresh token", logger.Err(err))
		return "", "", fmt.Errorf("%s: get refresh token: %w", op, err)
	}

	if err = s.tokens.DeleteRefreshToken(ctx, refreshToken); err != nil {
		log.Error("Failed to delete refresh token", logger.Err(err))
		return "", "", fmt.Errorf("%s: delete refresh token: %w", op, err)
	}

	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		log.Error("failed to get user", logger.Err(err))
		return "", "", fmt.Errorf("%s: get user: %w", op, err)
	}
	_ = claims
	return s.issueTokens(ctx, op, user.ID, string(user.Role))
}

func (s *Service) ValidateToken(ctx context.Context, accessToken string) (userID, role string, err error) {
	const op = "auth_service.ValidateToken"

	claims, err := s.jwtManager.ParseToken(accessToken)
	if err != nil {
		s.log.Warn("Invalid access token", slog.String("op", op), logger.Err(err))
		return "", "", pkgerrors.ErrUnauthorized
	}

	return claims.UserID, claims.Role, nil
}

func (s *Service) issueTokens(ctx context.Context, op string, userID, role string) (string, string, error) {
	accessToken, err := s.jwtManager.GenerateAccessToken(userID, role)
	if err != nil {
		return "", "", fmt.Errorf("%s generate access token: %w", op, err)
	}
	refreshToken, err := s.jwtManager.GenerateRefreshToken(userID)

	if err != nil {
		return "", "", fmt.Errorf("%s generate refresh token: %w", op, err)
	}

	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err = s.tokens.SaveRefreshToken(ctx, userID, refreshToken, expiresAt); err != nil {
		return "", "", fmt.Errorf("%s save refresh token: %w", op, err)
	}
	return accessToken, refreshToken, nil
}
