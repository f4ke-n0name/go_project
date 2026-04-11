package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/config"
	appgrpc "github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/app/grpc"
	grpcorder "github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/grpc/order"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/kafka"
	ordersvc "github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/services/order"
	"github.com/f4ke-n0name/autoparts-hub/services/order-service/internal/storage/postgres"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load("./.env")
	cfg := config.MustLoad()
	log := logger.SetupLogger(os.Getenv("ENV"))
	log.Info("starting order-service", slog.String("port", cfg.GRPC.Port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, err := postgres.New(ctx, cfg.DB.DSN())
	if err != nil {
		log.Error("failed to connect to postgres", logger.Err(err))
		os.Exit(1)
	}
	defer store.Close()

	producer, err := kafka.NewProducer(log, cfg.Kafka.Brokers)
	if err != nil {
		log.Error("failed to connect to kafka", logger.Err(err))
		os.Exit(1)
	}
	defer producer.Close()

	service := ordersvc.New(log, store, store, producer)

	service.StartOutboxWorker(ctx)

	handler := grpcorder.New(log, service)
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

	log.Info("order-service stopped")
}
