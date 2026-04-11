package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/catalog-service/config"
	appgrpc "github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/app/grpc"
	grpccatalog "github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/grpc/catalog"
	catalogsvc "github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/services/catalog"
	"github.com/f4ke-n0name/autoparts-hub/services/catalog-service/internal/storage/postgres"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("./.env")
	cfg := config.MustLoad()
	log := logger.SetupLogger(os.Getenv("ENV"))
	log.Info("starting catalog-service", slog.String("port", cfg.GRPC.Port))

	ctx := context.Background()
	store, err := postgres.New(ctx, cfg.DB.DSN())
	if err != nil {
		log.Error("failed to connect to postgres", logger.Err(err))
		os.Exit(1)
	}
	defer store.Close()

	service := catalogsvc.New(log, store, store, store)
	handler := grpccatalog.New(log, service)
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

	log.Info("catalog-service stopped")
}
