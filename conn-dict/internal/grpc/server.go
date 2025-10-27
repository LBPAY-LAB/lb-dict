package grpc

import (
	"context"
	"fmt"
	"net"
	"time"

	bridgev1 "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/conn_dict/v1"
	"github.com/lbpay-lab/conn-dict/internal/grpc/handlers"
	"github.com/lbpay-lab/conn-dict/internal/grpc/interceptors"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
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

	// Register ConnectService with all handlers
	// This service exposes Entry/Claim/Infraction operations to core-dict
	connectv1.RegisterConnectServiceServer(s.grpcServer, &connectServiceServer{
		entryHandler:      s.entryHandler,
		claimHandler:      s.claimHandler,
		infractionHandler: s.infractionHandler,
		logger:            s.logger,
	})
	s.logger.Info("Registered ConnectService with all handlers")

	// Register health check service
	s.healthServer = health.NewServer()

	// Set serving status for each registered service
	s.healthServer.SetServingStatus("dict.bridge.v1.BridgeService", grpc_health_v1.HealthCheckResponse_SERVING)
	s.healthServer.SetServingStatus("dict.connect.v1.ConnectService", grpc_health_v1.HealthCheckResponse_SERVING)

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

// connectServiceServer implements ConnectService by delegating to handlers
type connectServiceServer struct {
	connectv1.UnimplementedConnectServiceServer
	entryHandler      *handlers.EntryHandler
	claimHandler      *handlers.ClaimHandler
	infractionHandler *handlers.InfractionHandler
	logger            *logrus.Logger
}

// Entry Operations
// Note: These methods are NOT IMPLEMENTED yet because EntryHandler only implements
// BridgeService (internal operations), not ConnectService (core-dict operations).
// TODO: Create a separate query repository or use case for read-only Entry operations.
func (s *connectServiceServer) GetEntry(ctx context.Context, req *connectv1.GetEntryRequest) (*connectv1.GetEntryResponse, error) {
	s.logger.Warn("GetEntry not implemented - TODO: implement read-only entry queries")
	return nil, status.Error(codes.Unimplemented, "GetEntry not implemented yet - pending read-only query layer")
}

func (s *connectServiceServer) GetEntryByKey(ctx context.Context, req *connectv1.GetEntryByKeyRequest) (*connectv1.GetEntryByKeyResponse, error) {
	s.logger.Warn("GetEntryByKey not implemented - TODO: implement read-only entry queries")
	return nil, status.Error(codes.Unimplemented, "GetEntryByKey not implemented yet - pending read-only query layer")
}

func (s *connectServiceServer) ListEntries(ctx context.Context, req *connectv1.ListEntriesRequest) (*connectv1.ListEntriesResponse, error) {
	s.logger.Warn("ListEntries not implemented - TODO: implement read-only entry queries")
	return nil, status.Error(codes.Unimplemented, "ListEntries not implemented yet - pending read-only query layer")
}

// Claim Operations
func (s *connectServiceServer) CreateClaim(ctx context.Context, req *connectv1.CreateClaimRequest) (*connectv1.CreateClaimResponse, error) {
	resp, err := s.claimHandler.CreateClaim(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.CreateClaimResponse), nil
}

func (s *connectServiceServer) ConfirmClaim(ctx context.Context, req *connectv1.ConfirmClaimRequest) (*connectv1.ConfirmClaimResponse, error) {
	resp, err := s.claimHandler.ConfirmClaim(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.ConfirmClaimResponse), nil
}

func (s *connectServiceServer) CancelClaim(ctx context.Context, req *connectv1.CancelClaimRequest) (*connectv1.CancelClaimResponse, error) {
	resp, err := s.claimHandler.CancelClaim(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.CancelClaimResponse), nil
}

func (s *connectServiceServer) GetClaim(ctx context.Context, req *connectv1.GetClaimRequest) (*connectv1.GetClaimResponse, error) {
	resp, err := s.claimHandler.GetClaim(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.GetClaimResponse), nil
}

func (s *connectServiceServer) ListClaims(ctx context.Context, req *connectv1.ListClaimsRequest) (*connectv1.ListClaimsResponse, error) {
	resp, err := s.claimHandler.ListClaims(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.ListClaimsResponse), nil
}

// Infraction Operations
func (s *connectServiceServer) CreateInfraction(ctx context.Context, req *connectv1.CreateInfractionRequest) (*connectv1.CreateInfractionResponse, error) {
	resp, err := s.infractionHandler.CreateInfraction(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.CreateInfractionResponse), nil
}

func (s *connectServiceServer) InvestigateInfraction(ctx context.Context, req *connectv1.InvestigateInfractionRequest) (*connectv1.InvestigateInfractionResponse, error) {
	resp, err := s.infractionHandler.InvestigateInfraction(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.InvestigateInfractionResponse), nil
}

func (s *connectServiceServer) ResolveInfraction(ctx context.Context, req *connectv1.ResolveInfractionRequest) (*connectv1.ResolveInfractionResponse, error) {
	resp, err := s.infractionHandler.ResolveInfraction(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.ResolveInfractionResponse), nil
}

func (s *connectServiceServer) DismissInfraction(ctx context.Context, req *connectv1.DismissInfractionRequest) (*connectv1.DismissInfractionResponse, error) {
	resp, err := s.infractionHandler.DismissInfraction(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.DismissInfractionResponse), nil
}

func (s *connectServiceServer) GetInfraction(ctx context.Context, req *connectv1.GetInfractionRequest) (*connectv1.GetInfractionResponse, error) {
	resp, err := s.infractionHandler.GetInfraction(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.GetInfractionResponse), nil
}

func (s *connectServiceServer) ListInfractions(ctx context.Context, req *connectv1.ListInfractionsRequest) (*connectv1.ListInfractionsResponse, error) {
	resp, err := s.infractionHandler.ListInfractions(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp.(*connectv1.ListInfractionsResponse), nil
}

// Health Check
func (s *connectServiceServer) HealthCheck(ctx context.Context, req *emptypb.Empty) (*connectv1.HealthCheckResponse, error) {
	return &connectv1.HealthCheckResponse{
		Status: connectv1.HealthCheckResponse_HEALTH_STATUS_HEALTHY,
		// TODO: Add actual component health checks (postgresql, redis, temporal, pulsar)
	}, nil
}
