package auth

import (
	"context"
	"log/slog"

	"github.com/f4ke-n0name/autoparts-hub/gen/auth"
	pkgerrors "github.com/f4ke-n0name/autoparts-hub/pkg/errors"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService interface {
	Register(context context.Context, firstName, lastName, email, password string) (accessToken, refreshToken string, err error)
	Login(context context.Context, email, password string) (accessToken, refreshToken string, err error)
	RefreshToken(ctx context.Context, refreshToken string) (newAccessToken, newRefreshToken string, err error)
	ValidateToken(ctx context.Context, accessToken string) (userID, role string, err error)
}

type Handler struct {
	auth.UnimplementedAuthServer
	log     *slog.Logger
	service AuthService
}

func New(log *slog.Logger, service AuthService) *Handler {
	return &Handler{
		log:     log,
		service: service,
	}
}

func (h *Handler) Register(ctx context.Context, req *auth.RegisterRequest) (*auth.RegisterResponse, error) {
	const op = "grpc.auth.Handler.Register"
	log := h.log.With(slog.String("op", op))
	if err := validateRegister(req); err != nil {
		return nil, err
	}

	accessToken, refreshToken, err := h.service.Register(ctx, req.GetFirstName(), req.GetLastName(), req.GetEmail(), req.GetPassword())
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "User with this email already exists")
		}
		log.Error("Failed to register user", logger.Err(err))
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &auth.RegisterResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (h *Handler) Login(ctx context.Context, req *auth.LoginRequest) (*auth.LoginResponse, error) {
	const op = "grpc.auth.Handler.Login"
	log := h.log.With(slog.String("op", op))
	if req.GetEmail() == "" || req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "Email and password are required")
	}
	accessToken, refreshToken, err := h.service.Login(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrNotFound) || pkgerrors.Is(err, pkgerrors.ErrUnauthorized) {
			return nil, status.Error(codes.Unauthenticated, "Invalid email or password")
		}
		log.Error("Failed to login", logger.Err(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}

	return &auth.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (h *Handler) RefreshToken(ctx context.Context, req *auth.RefreshRequest) (*auth.RefreshResponse, error) {
	const op = "grpc.auth.RefreshToken"
	log := h.log.With(slog.String("op", op))

	if req.GetRefreshToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "refresh token is required")
	}

	accessToken, refreshToken, err := h.service.RefreshToken(ctx, req.GetRefreshToken())
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrUnauthorized) {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired refresh token")
		}
		log.Error("failed to refresh token", logger.Err(err))
		return nil, status.Error(codes.Internal, "internal error")
	}
	return &auth.RefreshResponse{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (h *Handler) ValidateToken(ctx context.Context, req *auth.ValidateRequest) (*auth.ValidateResponse, error) {
	const op = "grpc.auth.Handler.ValidateToken"
	log := h.log.With(slog.String("op", op))

	if req.GetAccessToken() == "" {
		return nil, status.Error(codes.InvalidArgument, "Access token is required")
	}

	userID, role, err := h.service.ValidateToken(ctx, req.GetAccessToken())
	if err != nil {
		if pkgerrors.Is(err, pkgerrors.ErrUnauthorized) {
			return &auth.ValidateResponse{Valid: false}, nil
		}
		log.Error("Failed to validate token", logger.Err(err))
		return nil, status.Error(codes.Internal, "Internal error")
	}
	return &auth.ValidateResponse{UserId: userID, Role: role, Valid: true}, nil
}

func validateRegister(req *auth.RegisterRequest) error {
	if req.GetFirstName() == "" {
		return status.Error(codes.InvalidArgument, "first_name is required")
	}
	if req.GetLastName() == "" {
		return status.Error(codes.InvalidArgument, "last_name is required")
	}
	if req.GetEmail() == "" {
		return status.Error(codes.InvalidArgument, "email is required")
	}
	if req.GetPassword() == "" {
		return status.Error(codes.InvalidArgument, "password is required")
	}
	if len(req.GetPassword()) < 6 {
		return status.Error(codes.InvalidArgument, "password must be at least 6 characters")
	}
	return nil
}
