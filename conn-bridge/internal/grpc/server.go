package grpc

import (
	"context"
	"fmt"
	"net"

	pb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// Server implements the Bridge gRPC server
type Server struct {
	pb.UnimplementedBridgeServiceServer
	logger      *logrus.Logger
	grpcServer  *grpc.Server
	port        int
	soapClient  SOAPClient
	xmlSigner   XMLSigner
}

// SOAPClient defines the interface for SOAP operations
type SOAPClient interface {
	SendSOAPRequest(ctx context.Context, endpoint string, soapEnvelope []byte) ([]byte, error)
	BuildSOAPEnvelope(bodyXML string, signedXML string) ([]byte, error)
	ParseSOAPResponse(soapResponse []byte) ([]byte, error)
	HealthCheck(ctx context.Context) error
}

// XMLSigner defines the interface for XML signing operations
type XMLSigner interface {
	SignXML(ctx context.Context, xmlData string) (string, error)
	HealthCheck(ctx context.Context) error
}

// NewServer creates a new Bridge gRPC server instance
func NewServer(logger *logrus.Logger, port int, soapClient SOAPClient, xmlSigner XMLSigner) *Server {
	return &Server{
		logger:     logger,
		port:       port,
		soapClient: soapClient,
		xmlSigner:  xmlSigner,
	}
}

// Start initializes and starts the gRPC server
func (s *Server) Start() error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))
	if err != nil {
		return fmt.Errorf("failed to listen on port %d: %w", s.port, err)
	}

	// Create gRPC server with interceptors
	s.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			s.loggingInterceptor,
			s.metricsInterceptor,
		),
	)

	// Register Bridge service
	pb.RegisterBridgeServiceServer(s.grpcServer, s)

	// Register health check service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("bridge.BridgeService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s.grpcServer, healthServer)

	// Register reflection for debugging
	reflection.Register(s.grpcServer)

	s.logger.Infof("Bridge gRPC server listening on port %d", s.port)

	// Start serving (blocking)
	if err := s.grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// Stop gracefully stops the gRPC server
func (s *Server) Stop() {
	if s.grpcServer != nil {
		s.logger.Info("Shutting down Bridge gRPC server...")
		s.grpcServer.GracefulStop()
	}
}

// loggingInterceptor logs all incoming requests
func (s *Server) loggingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	s.logger.Infof("gRPC request: method=%s", info.FullMethod)
	resp, err := handler(ctx, req)
	if err != nil {
		s.logger.Errorf("gRPC request failed: method=%s error=%v", info.FullMethod, err)
	}
	return resp, err
}

// metricsInterceptor collects metrics for requests
func (s *Server) metricsInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	// TODO: Implement Prometheus metrics collection
	return handler(ctx, req)
}