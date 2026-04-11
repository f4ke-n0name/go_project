package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	inventory "github.com/f4ke-n0name/autoparts-hub/gen/inventory"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	grpcinventory "github.com/f4ke-n0name/autoparts-hub/services/inventory-service/internal/grpc/inventory"
)

type Server struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       string
}

func New(log *slog.Logger, port string, handler *grpcinventory.Handler) *Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(log),
		),
	)

	inventory.RegisterInventoryServiceServer(grpcServer, handler)
	reflection.Register(grpcServer)

	return &Server{log: log, grpcServer: grpcServer, port: port}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", s.port))
	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}
	s.log.Info("gRPC server started", slog.String("port", s.port))
	if err = s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}
	return nil
}

func (s *Server) Stop() {
	s.log.Info("stopping gRPC server")
	s.grpcServer.GracefulStop()
}

func loggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		log.Info("grpc call", slog.String("method", info.FullMethod))
		resp, err := handler(ctx, req)
		if err != nil {
			log.Error("grpc call failed", slog.String("method", info.FullMethod), logger.Err(err))
		}
		return resp, err
	}
}
