package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"

	corev1 "github.com/lbpay-lab/dict-contracts/gen/proto/core/v1"
)

// ServerConfig holds gRPC server configuration
type ServerConfig struct {
	Port                int
	MaxConnectionIdle   time.Duration
	MaxConnectionAge    time.Duration
	MaxConnectionAgeGrace time.Duration
	KeepAliveTime       time.Duration
	KeepAliveTimeout    time.Duration
}

// DefaultServerConfig returns default gRPC server configuration
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Port:                  9090,
		MaxConnectionIdle:     15 * time.Minute,
		MaxConnectionAge:      30 * time.Minute,
		MaxConnectionAgeGrace: 5 * time.Minute,
		KeepAliveTime:         5 * time.Minute,
		KeepAliveTimeout:      20 * time.Second,
	}
}

// Server represents the gRPC server
type Server struct {
	config         *ServerConfig
	grpcServer     *grpc.Server
	entryHandler   corev1.CoreDictServiceServer
	authInterceptor   *AuthInterceptor
	loggingInterceptor *LoggingInterceptor
	metricsInterceptor *MetricsInterceptor
	recoveryInterceptor *RecoveryInterceptor
	rateLimitInterceptor *RateLimitInterceptor
}

// NewServer creates a new gRPC server instance
func NewServer(
	config *ServerConfig,
	entryHandler corev1.CoreDictServiceServer,
	authInterceptor *AuthInterceptor,
	loggingInterceptor *LoggingInterceptor,
	metricsInterceptor *MetricsInterceptor,
	recoveryInterceptor *RecoveryInterceptor,
	rateLimitInterceptor *RateLimitInterceptor,
) *Server {
	if config == nil {
		config = DefaultServerConfig()
	}

	return &Server{
		config:               config,
		entryHandler:         entryHandler,
		authInterceptor:      authInterceptor,
		loggingInterceptor:   loggingInterceptor,
		metricsInterceptor:   metricsInterceptor,
		recoveryInterceptor:  recoveryInterceptor,
		rateLimitInterceptor: rateLimitInterceptor,
	}
}

// Start starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.config.Port, err)
	}

	// Create gRPC server with interceptors
	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.recoveryInterceptor.Unary(),      // 1. Recovery (first to catch all panics)
			s.loggingInterceptor.Unary(),       // 2. Logging (log all requests)
			s.authInterceptor.Unary(),          // 3. Authentication (verify JWT)
			s.metricsInterceptor.Unary(),       // 4. Metrics (collect metrics)
			s.rateLimitInterceptor.Unary(),     // 5. Rate limiting (enforce limits)
		),
		grpc.KeepaliveParams(keepalive.ServerParameters{
			MaxConnectionIdle:     s.config.MaxConnectionIdle,
			MaxConnectionAge:      s.config.MaxConnectionAge,
			MaxConnectionAgeGrace: s.config.MaxConnectionAgeGrace,
			Time:                  s.config.KeepAliveTime,
			Timeout:               s.config.KeepAliveTimeout,
		}),
		grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
			MinTime:             5 * time.Second,
			PermitWithoutStream: true,
		}),
	)

	// Register services
	corev1.RegisterCoreDictServiceServer(s.grpcServer, s.entryHandler)

	// Enable reflection for grpcurl/grpcui
	reflection.Register(s.grpcServer)

	// Start server in goroutine
	go func() {
		fmt.Printf("gRPC server starting on port %d...\n", s.config.Port)
		if err := s.grpcServer.Serve(listener); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	fmt.Println("Shutting down gRPC server...")
	s.grpcServer.GracefulStop()
	fmt.Println("gRPC server stopped")

	return nil
}

// Stop stops the gRPC server gracefully
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.grpcServer.GracefulStop()
	}
}
