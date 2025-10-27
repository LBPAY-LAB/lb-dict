package main

import (
	"context"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	corev1 "github.com/lbpay-lab/dict-contracts/gen/proto/core/v1"
	grpchandler "github.com/lbpay-lab/core-dict/internal/infrastructure/grpc"
)

const (
	defaultPort     = "9090"
	defaultLogLevel = "info"
)

func main() {
	// ============================================================
	// 1. CONFIGURAÇÃO
	// ============================================================
	port := getEnv("GRPC_PORT", defaultPort)
	logLevel := getEnv("LOG_LEVEL", defaultLogLevel)
	useMockMode := getEnv("CORE_DICT_USE_MOCK_MODE", "true") == "true"

	// ============================================================
	// 2. LOGGER
	// ============================================================
	logger := setupLogger(logLevel)
	logger.Info("Starting Core DICT gRPC Server",
		"port", port,
		"mock_mode", useMockMode,
		"version", "1.0.0",
	)

	// ============================================================
	// 3. gRPC SERVER
	// ============================================================
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Error("Failed to listen", "error", err, "port", port)
		os.Exit(1)
	}

	// gRPC server options
	opts := []grpc.ServerOption{
		grpc.UnaryInterceptor(loggingInterceptor(logger)),
		grpc.MaxRecvMsgSize(10 * 1024 * 1024), // 10MB
		grpc.MaxSendMsgSize(10 * 1024 * 1024), // 10MB
	}

	grpcServer := grpc.NewServer(opts...)

	// ============================================================
	// 4. REGISTER SERVICES
	// ============================================================

	// 4a. Core DICT Service Handler
	var cleanup *Cleanup
	if useMockMode {
		logger.Warn("⚠️  MOCK MODE ENABLED - Using mock responses for all RPCs")
		logger.Warn("⚠️  Set CORE_DICT_USE_MOCK_MODE=false to enable real business logic")

		// Create handler with nil dependencies (mock mode doesn't need them)
		handler := grpchandler.NewCoreDictServiceHandler(
			true, // useMockMode = true
			nil, nil, nil, nil, nil, nil, nil, nil, nil, // commands (not needed in mock mode)
			nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, // queries (not needed in mock mode)
			logger,
		)
		corev1.RegisterCoreDictServiceServer(grpcServer, handler)
		logger.Info("✅ CoreDictService registered (MOCK MODE)")
	} else {
		logger.Info("🚀 REAL MODE ENABLED - Initializing all dependencies...")

		// Initialize all dependencies and create handler
		handler, cleanupResources, err := initializeRealHandler(logger)
		if err != nil {
			logger.Error("❌ Failed to initialize Real Mode", "error", err)
			logger.Error("💡 Tip: Set CORE_DICT_USE_MOCK_MODE=true to use mock mode for testing")
			os.Exit(1)
		}
		cleanup = cleanupResources

		// Register handler
		corev1.RegisterCoreDictServiceServer(grpcServer, handler)
		logger.Info("✅ CoreDictService registered (REAL MODE)")
	}

	// 4b. Health Check Service
	healthServer := health.NewServer()
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("dict.core.v1.CoreDictService", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	logger.Info("✅ Health Check service registered")

	// 4c. Reflection (for grpcurl)
	reflection.Register(grpcServer)
	logger.Info("✅ gRPC Reflection enabled (for grpcurl)")

	// ============================================================
	// 5. START SERVER
	// ============================================================
	go func() {
		logger.Info("🚀 gRPC server listening", "address", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			logger.Error("Failed to serve", "error", err)
			os.Exit(1)
		}
	}()

	// ============================================================
	// 6. GRACEFUL SHUTDOWN
	// ============================================================
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("🛑 Shutting down gRPC server...")

	// Graceful stop with timeout
	stopped := make(chan struct{})
	go func() {
		grpcServer.GracefulStop()
		close(stopped)
	}()

	select {
	case <-stopped:
		logger.Info("✅ Server stopped gracefully")
	case <-time.After(30 * time.Second):
		logger.Warn("⏰ Graceful shutdown timed out, forcing stop")
		grpcServer.Stop()
	}

	// Cleanup resources (only in Real Mode)
	if cleanup != nil {
		cleanup.Close(logger)
	}
}

// ============================================================
// HELPERS
// ============================================================

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func setupLogger(level string) *slog.Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: logLevel,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return slog.New(handler)
}

// loggingInterceptor logs all gRPC requests/responses
func loggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Call handler
		resp, err := handler(ctx, req)

		// Log
		duration := time.Since(start)
		if err != nil {
			logger.Error("gRPC request failed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
				"error", err,
			)
		} else {
			logger.Info("gRPC request completed",
				"method", info.FullMethod,
				"duration_ms", duration.Milliseconds(),
			)
		}

		return resp, err
	}
}
