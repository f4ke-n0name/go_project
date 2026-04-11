package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/f4ke-n0name/autoparts-hub/pkg/jwt"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/auth-service/config"
	appgrpc "github.com/f4ke-n0name/autoparts-hub/services/auth-service/internal/app/grpc"
	grpcauth "github.com/f4ke-n0name/autoparts-hub/services/auth-service/internal/grpc/auth"
	authsvc "github.com/f4ke-n0name/autoparts-hub/services/auth-service/internal/services/auth"
	"github.com/f4ke-n0name/autoparts-hub/services/auth-service/internal/storage/postgres"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("./.env")
	cfg := config.MustLoad()
	log := logger.SetupLogger(os.Getenv("ENV"))
	log.Info("starting auth-service",
		slog.String("port", cfg.GRPC.Port),
	)
	ctx := context.Background()
	store, err := postgres.New(ctx, cfg.DB.DSN())
	if err != nil {
		log.Error("failed to connect to postgres", logger.Err(err))
		os.Exit(1)
	}
	defer store.Close()

	jwtManager := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.Access, cfg.JWT.Refresh)
	service := authsvc.New(log, store, store, jwtManager)
	handler := grpcauth.New(log, service)
	grpcServer := appgrpc.New(log, cfg.GRPC.Port, handler)
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		if err := grpcServer.Run(); err != nil {
			log.Error("gRPC server error", logger.Err(err))
			os.Exit(1)
		}
	}()

	sig := <-quit
	log.Info("shutting down", slog.String("signal", sig.String()))

	grpcServer.Stop()

	log.Info("auth-service stopped")
}
