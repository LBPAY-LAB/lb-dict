package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	bridgev1 "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	"github.com/lbpay-lab/conn-dict/internal/grpc/handlers"
	"github.com/lbpay-lab/conn-dict/internal/grpc/interceptors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server implements the Connect gRPC server
type Server struct {
	logger            *logrus.Logger
	grpcServer        *grpc.Server
	port              int
	entryHandler      *handlers.EntryHandler
	claimHandler      *handlers.ClaimHandler
	infractionHandler *handlers.InfractionHandler
	healthServer      *health.Server
	devMode           bool
}

// ServerConfig holds configuration for the gRPC server
type ServerConfig struct {
	Port              int
	DevMode           bool
	EntryHandler      *handlers.EntryHandler
	ClaimHandler      *handlers.ClaimHandler
	InfractionHandler *handlers.InfractionHandler
}

// NewServer creates a new Connect gRPC server instance
func NewServer(logger *logrus.Logger, config *ServerConfig) *Server {
	if config.Port == 0 {
		config.Port = 9092
	}

	return &Server{
		logger:            logger,
		port:              config.Port,
		entryHandler:      config.EntryHandler,
		claimHandler:      config.ClaimHandler,
		infractionHandler: config.InfractionHandler,
		devMode:           config.DevMode,
	}
}

// Start initializes and starts the gRPC server
func (s *Server) Start(ctx context.Context) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	// Create gRPC server with chained interceptors
	// Order matters: Recovery → Logging → Tracing → Metrics
	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptors.RecoveryInterceptor(s.logger),      // Panic recovery (first - catch all panics)
			interceptors.LoggingInterceptor(s.logger),       // Request logging
			interceptors.TracingInterceptor("conn-dict"),    // OpenTelemetry tracing
			interceptors.MetricsInterceptor(),               // Prometheus metrics (last)
		),
	)

	// Register BridgeService with entry handler
	// This service handles Create/Get/Update/Delete Entry operations to Bacen RSFN
	bridgev1.RegisterBridgeServiceServer(s.grpcServer, s.entryHandler)
	s.logger.Info("Registered BridgeService with EntryHandler")

	// NOTE: ClaimHandler and InfractionHandler are READY but cannot be registered yet
	// because proto files are not generated. Once dict-contracts generates the proto code:
	// 1. Import: corev1 "github.com/lbpay-lab/dict-contracts/gen/proto/core/v1"
	// 2. Register: corev1.RegisterClaimServiceServer(s.grpcServer, s.claimHandler)
	// 3. Register: corev1.RegisterInfractionServiceServer(s.grpcServer, s.infractionHandler)
	// 4. Update handlers to embed pb.UnimplementedClaimServiceServer and pb.UnimplementedInfractionServiceServer
	//
	// For now, handlers are instantiated but not registered.
	if s.claimHandler != nil {
		s.logger.Info("ClaimHandler initialized (pending proto generation for registration)")
	}
	if s.infractionHandler != nil {
		s.logger.Info("InfractionHandler initialized (pending proto generation for registration)")
	}

	// Register health check service
	s.healthServer = health.NewServer()

	// Set serving status for each registered service
	s.healthServer.SetServingStatus("dict.bridge.v1.BridgeService", grpc_health_v1.HealthCheckResponse_SERVING)
	// TODO: Add health check for ClaimService and InfractionService when implemented
	// s.healthServer.SetServingStatus("dict.core.v1.ClaimService", grpc_health_v1.HealthCheckResponse_SERVING)
	// s.healthServer.SetServingStatus("dict.core.v1.InfractionService", grpc_health_v1.HealthCheckResponse_SERVING)

	// Set overall server status
	s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	grpc_health_v1.RegisterHealthServer(s.grpcServer, s.healthServer)
	s.logger.Info("Registered health check service")

	// Register reflection service (dev mode or always for debugging)
	if s.devMode {
		reflection.Register(s.grpcServer)
		s.logger.Info("Registered reflection service (dev mode)")
	}

	s.logger.WithFields(logrus.Fields{
		"port":     s.port,
		"dev_mode": s.devMode,
	}).Info("Connect gRPC server starting")

	// Start serving (blocking)
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	if s.grpcServer == nil {
		return
	}

	s.logger.Info("Shutting down Connect gRPC server...")

	// Mark as not serving
	if s.healthServer != nil {
		s.healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	}

	// Graceful stop with timeout
	stopped := make(chan struct{})
	go func() {
		s.grpcServer.GracefulStop()
		close(stopped)
	}()

	// Wait for graceful stop or timeout
	select {
	case <-stopped:
		s.logger.Info("gRPC server stopped gracefully")
	case <-time.After(30 * time.Second):
		s.logger.Warn("gRPC server graceful stop timeout, forcing stop")
		s.grpcServer.Stop()
	}
}
