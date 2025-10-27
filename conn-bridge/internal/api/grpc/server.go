package grpc

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/lbpay-lab/conn-bridge/internal/api/grpc/handlers"
	"github.com/lbpay-lab/conn-bridge/internal/application/usecases"
)

// Server represents the gRPC server
type Server struct {
	grpcServer *grpc.Server
	port       int
	handlers   *handlers.DictHandler
}

// Config holds the configuration for the gRPC server
type Config struct {
	Port                int
	CreateEntryUseCase  *usecases.CreateEntryUseCase
	QueryEntryUseCase   *usecases.QueryEntryUseCase
	DeleteEntryUseCase  *usecases.DeleteEntryUseCase
	CreateClaimUseCase  *usecases.CreateClaimUseCase
}

// NewServer creates a new gRPC server
func NewServer(config *Config) (*Server, error) {
	if config.Port <= 0 {
		return nil, fmt.Errorf("invalid port: %d", config.Port)
	}

	// Create gRPC server with interceptors
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(unaryInterceptor),
	)

	// Create handlers
	dictHandler := handlers.NewDictHandler(
		config.CreateEntryUseCase,
		config.QueryEntryUseCase,
		config.DeleteEntryUseCase,
		config.CreateClaimUseCase,
	)

	// Register services
	// TODO: Register proto-generated services
	// dictpb.RegisterDictServiceServer(grpcServer, dictHandler)

	// Register health check service
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Register reflection service for development
	reflection.Register(grpcServer)

	return &Server{
		grpcServer: grpcServer,
		port:       config.Port,
		handlers:   dictHandler,
	}, nil
}

// Start starts the gRPC server
func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	fmt.Printf("gRPC server listening on port %d\n", s.port)

	if err := s.grpcServer.Serve(listener); err != nil {
		return fmt.Errorf("failed to serve gRPC: %w", err)
	}

	return nil
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() {
	fmt.Println("Shutting down gRPC server...")
	s.grpcServer.GracefulStop()
}

// unaryInterceptor is a gRPC interceptor for logging and tracing
func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// TODO: Add logging, metrics, tracing
	// For now, just pass through
	return handler(ctx, req)
}
