package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"

	auth "github.com/f4ke-n0name/autoparts-hub/gen/auth"
	"github.com/f4ke-n0name/autoparts-hub/pkg/logger"
	grpcauth "github.com/f4ke-n0name/autoparts-hub/services/auth-service/internal/grpc/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type Server struct {
	log        *slog.Logger
	grpcServer *grpc.Server
	port       string
}

func New(log *slog.Logger, port string, handler *grpcauth.Handler) *Server {
	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingInterceptor(log),
		),
	)

	auth.RegisterAuthServer(grpcServer, handler)

	reflection.Register(grpcServer)

	return &Server{
		log:        log,
		grpcServer: grpcServer,
		port:       port,
	}
}

func (s *Server) Run() error {
	lis, err := net.Listen("tcp", ":"+s.port)

	if err != nil {
		return fmt.Errorf("listen tcp: %w", err)
	}

	s.log.Info("grpc server listening on port " + s.port + " ...")

	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("serve grpc: %w", err)
	}
	return nil
}

func (s *Server) Stop() {
	s.log.Info("grpc server shutting down...")
	s.grpcServer.GracefulStop()
}

func loggingInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Info("grpc call", slog.String("method", info.FullMethod))
		resp, err := handler(ctx, req)
		if err != nil {
			log.Error("grpc call failed", slog.String("method", info.FullMethod), logger.Err(err))
		}
		return resp, err
	}
}
