package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/inventory-service/config"
	appgrpc "github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/app/grpc"
	grpcinventory "github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/grpc/inventory"
	"github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/kafka"
	inventorysvc "github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/services/inventory"
	"github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/storage/postgres"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("./.env")
	cfg := config.MustLoad()
	log := logger.SetupLogger(os.Getenv("ENV"))
	log.Info("starting inventory-service", slog.String("port", cfg.GRPC.Port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := postgres.New(ctx, cfg.DB.DSN())
	if err != nil {
		log.Error("failed to connect to postgres", logger.Err(err))
		os.Exit(1)
	}
	defer store.Close()

	service := inventorysvc.New(log, store)
	consumer, err := kafka.NewConsumer(log, cfg.Kafka.Brokers, cfg.Kafka.GroupID, service)
	if err != nil {
		log.Error("failed to create kafka consumer", logger.Err(err))
		os.Exit(1)
	}
	defer consumer.Close()

	consumer.Start(ctx)
	handler := grpcinventory.New(log, service)
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

	cancel()
	grpcServer.Stop()

	log.Info("inventory-service stopped")
}
