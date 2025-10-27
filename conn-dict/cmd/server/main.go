package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/application/adapters"
	"github.com/lbpay-lab/conn-dict/internal/application/usecases"
	"github.com/lbpay-lab/conn-dict/internal/grpc"
	"github.com/lbpay-lab/conn-dict/internal/grpc/handlers"
	"github.com/lbpay-lab/conn-dict/internal/grpc/services"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/cache"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/database"
	grpcInfra "github.com/lbpay-lab/conn-dict/internal/infrastructure/grpc"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/pulsar"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
	"go.temporal.io/sdk/client"
)

// Prometheus metrics for gRPC server
var (
	serverRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc_server",
			Name:      "requests_total",
			Help:      "Total number of gRPC requests received",
		},
		[]string{"method", "status"},
	)

	serverRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc_server",
			Name:      "request_duration_seconds",
			Help:      "gRPC request duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0, 2.0, 5.0},
		},
		[]string{"method"},
	)

	serverHealthStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc_server",
			Name:      "health_status",
			Help:      "Server health status (1 = healthy, 0 = unhealthy)",
		},
	)

	serverUptime = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc_server",
			Name:      "uptime_seconds",
			Help:      "Server uptime in seconds",
		},
	)
)

func main() {
	startTime := time.Now()

	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	logger.Info("Starting Connect gRPC server...")

	// Initialize PostgreSQL client
	pgConfig := &database.PostgresConfig{
		Host:     getEnvOrDefault("POSTGRES_HOST", "localhost"),
		Port:     getEnvAsInt("POSTGRES_PORT", 5432),
		User:     getEnvOrDefault("POSTGRES_USER", "dict_user"),
		Password: getEnvOrDefault("POSTGRES_PASSWORD", "dict_password"),
		Database: getEnvOrDefault("POSTGRES_DB", "dict_db"),
		SSLMode:  getEnvOrDefault("POSTGRES_SSLMODE", "disable"),
		MaxConns: int32(getEnvAsInt("POSTGRES_MAX_CONNS", 25)),
		MinConns: int32(getEnvAsInt("POSTGRES_MIN_CONNS", 5)),
	}

	postgresClient, err := database.NewPostgresClient(pgConfig, logger)
	if err != nil {
		log.Fatalf("Failed to initialize PostgreSQL client: %v", err)
	}
	defer postgresClient.Close()

	logger.Info("PostgreSQL client initialized successfully")

	// Health check PostgreSQL
	ctx := context.Background()
	if err := postgresClient.HealthCheck(ctx); err != nil {
		log.Fatalf("PostgreSQL health check failed: %v", err)
	}
	logger.Info("PostgreSQL health check passed")

	// Initialize repositories
	claimRepo := repositories.NewClaimRepository(postgresClient, logger)
	entryRepo := repositories.NewEntryRepository(postgresClient, logger)
	infractionRepo := repositories.NewInfractionRepository(postgresClient, logger)

	logger.Info("Repositories initialized successfully")

	// Initialize Redis cache
	redisAddr := fmt.Sprintf("%s:%d",
		getEnvOrDefault("REDIS_HOST", "localhost"),
		getEnvAsInt("REDIS_PORT", 6379))

	redisConfig := cache.RedisConfig{
		Addr:         redisAddr,
		Password:     getEnvOrDefault("REDIS_PASSWORD", ""),
		DB:           getEnvAsInt("REDIS_DB", 0),
		MaxRetries:   getEnvAsInt("REDIS_MAX_RETRIES", 3),
		DialTimeout:  30 * time.Second,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		PoolSize:     getEnvAsInt("REDIS_POOL_SIZE", 20),
		MinIdleConns: getEnvAsInt("REDIS_MIN_IDLE_CONNS", 5),
	}

	redisClient, err := cache.NewRedisClient(redisConfig, logger)
	if err != nil {
		log.Fatalf("Failed to initialize Redis client: %v", err)
	}
	defer redisClient.Close()

	logger.Info("Redis client initialized successfully")

	// Health check Redis - connection already tested in NewRedisClient

	// Initialize Pulsar event publisher
	pulsarURL := getEnvOrDefault("PULSAR_URL", "pulsar://localhost:6650")
	pulsarTopic := getEnvOrDefault("PULSAR_TOPIC", "persistent://public/default/dict-events")

	pulsarProducer, err := pulsar.NewProducer(pulsar.ProducerConfig{URL: pulsarURL, Topic: pulsarTopic, ProducerName: "conn-dict-server"}, logger)
	if err != nil {
		log.Fatalf("Failed to initialize Pulsar event publisher: %v", err)
	}
	defer pulsarProducer.Close()

	logger.Info("Pulsar producer initialized successfully")

	// Initialize Temporal client
	temporalAddress := getEnvOrDefault("TEMPORAL_ADDRESS", "localhost:7233")
	temporalNamespace := getEnvOrDefault("TEMPORAL_NAMESPACE", "default")

	temporalClient, err := client.Dial(client.Options{
		HostPort:  temporalAddress,
		Namespace: temporalNamespace,
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	logger.WithFields(logrus.Fields{
		"temporal_address": temporalAddress,
		"namespace":        temporalNamespace,
	}).Info("Connected to Temporal server")

	// Initialize Bridge gRPC client
	bridgeAddress := getEnvOrDefault("BRIDGE_GRPC_ADDRESS", "localhost:9094")
	bridgeClientConfig := &grpcInfra.BridgeClientConfig{
		Address:        bridgeAddress,
		ConnectTimeout: 10 * time.Second,
		RequestTimeout: 30 * time.Second,
	}

	bridgeClient, err := grpcInfra.NewBridgeClient(bridgeClientConfig, logger)
	if err != nil {
		log.Fatalf("Failed to initialize Bridge gRPC client: %v", err)
	}
	defer bridgeClient.Close()

	logger.WithField("bridge_address", bridgeAddress).Info("Bridge gRPC client initialized successfully")

	// Initialize adapters
	entryRepoAdapter := adapters.NewEntryRepositoryAdapter(entryRepo)
	cacheAdapter := adapters.NewCacheAdapter(redisClient)
	eventPublisherAdapter := adapters.NewEventPublisherAdapter(pulsarProducer, logger)

	// Initialize tracer
	tracer := otel.Tracer("conn-dict/server")

	// Initialize use cases
	entryUseCase := usecases.NewEntryUseCase(
		bridgeClient,
		entryRepoAdapter,
		cacheAdapter,
		eventPublisherAdapter,
		logger,
		tracer,
	)

	// Initialize gRPC handlers
	entryHandler := handlers.NewEntryHandler(entryUseCase, logger, tracer)

	// Initialize QueryHandler for read-only Entry operations
	queryHandler := handlers.NewQueryHandler(
		entryRepo,
		redisClient,
		logger,
		tracer,
	)

	logger.Info("QueryHandler initialized successfully")

	// Initialize Claim and Infraction services (direct repository access for now)
	claimService := services.NewClaimService(temporalClient, claimRepo, logger)
	infractionService := services.NewInfractionService(temporalClient, infractionRepo, logger)

	// Initialize Claim and Infraction handlers
	claimHandler := handlers.NewClaimHandler(claimService, logger, tracer)
	infractionHandler := handlers.NewInfractionHandler(infractionService, logger, tracer)

	logger.Info("Use cases, services, and handlers initialized successfully")

	// Create gRPC server
	grpcPort := getEnvAsInt("GRPC_PORT", 9092)
	devMode := getEnvOrDefault("DEV_MODE", "true") == "true"

	serverConfig := &grpc.ServerConfig{
		Port:              grpcPort,
		DevMode:           devMode,
		EntryHandler:      entryHandler,
		ClaimHandler:      claimHandler,
		InfractionHandler: infractionHandler,
		QueryHandler:      queryHandler,
	}

	grpcServerInstance := grpc.NewServer(logger, serverConfig)

	logger.WithFields(logrus.Fields{
		"port":     grpcPort,
		"dev_mode": devMode,
	}).Info("gRPC server configured")

	// Start metrics server
	metricsPort := getEnvAsInt("METRICS_PORT", 9091)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())

		metricsAddr := fmt.Sprintf(":%d", metricsPort)
		logger.Infof("Starting metrics server on %s", metricsAddr)

		if err := http.ListenAndServe(metricsAddr, mux); err != nil {
			logger.Errorf("Metrics server failed: %v", err)
		}
	}()

	// Start health check server
	healthPort := getEnvAsInt("HEALTH_PORT", 8080)
	healthServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", healthPort),
		Handler:      createHealthCheckHandler(logger, postgresClient, redisClient, temporalClient, bridgeClient),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	go func() {
		logger.Infof("Starting health check server on :%d", healthPort)
		if err := healthServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Health check server failed: %v", err)
		}
	}()

	// Start uptime tracker
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			serverUptime.Set(time.Since(startTime).Seconds())
		}
	}()

	// Set initial health status
	serverHealthStatus.Set(1)

	// Start gRPC server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Infof("Starting gRPC server on port %d", grpcPort)
		if err := grpcServerInstance.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	logger.Info("gRPC server started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		logger.Errorf("gRPC server error: %v", err)
		serverHealthStatus.Set(0)
	case sig := <-sigChan:
		logger.Infof("Received signal: %v", sig)
	}

	// Graceful shutdown
	logger.Info("Shutting down gRPC server...")
	serverHealthStatus.Set(0)

	// Stop gRPC server
	grpcServerInstance.Stop()

	// Shutdown health check server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		logger.Warnf("Health check server shutdown error: %v", err)
	}

	logger.Info("gRPC server stopped successfully")
}

// getEnvAsInt retrieves environment variable as integer with default value
func getEnvAsInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}

	return value
}

// getEnvOrDefault retrieves environment variable with default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// createHealthCheckHandler creates HTTP handler for health checks
func createHealthCheckHandler(
	logger *logrus.Logger,
	pgClient *database.PostgresClient,
	redisClient *cache.RedisClient,
	temporalClient client.Client,
	bridgeClient *grpcInfra.BridgeClient,
) http.Handler {
	mux := http.NewServeMux()

	// Liveness probe - checks if server is running
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"conn-dict-server"}`)
	})

	// Readiness probe - checks if server can handle requests
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		healthChecks := make(map[string]error)

		// Check PostgreSQL health
		healthChecks["postgresql"] = pgClient.HealthCheck(ctx)

		// Check Redis health (simple existence check)
		_, redisErr := redisClient.Exists(ctx, "health:check")
		healthChecks["redis"] = redisErr

		// Check Temporal connection
		_, temporalErr := temporalClient.CheckHealth(ctx, &client.CheckHealthRequest{})
		healthChecks["temporal"] = temporalErr

		// Bridge client health is checked lazily (on first request)
		// We don't check it here to avoid unnecessary overhead

		// Determine overall health status
		allHealthy := true
		for service, err := range healthChecks {
			if err != nil {
				logger.WithError(err).Warnf("%s health check failed", service)
				allHealthy = false
			}
		}

		if !allHealthy {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"not_ready","reason":"dependencies_unhealthy"}`)
			return
		}

		// All checks passed
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","service":"conn-dict-server"}`)
	})

	// Detailed status endpoint with all dependency checks
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
		defer cancel()

		type serviceStatus struct {
			Name    string `json:"name"`
			Healthy bool   `json:"healthy"`
			Error   string `json:"error,omitempty"`
		}

		statuses := []serviceStatus{
			{Name: "postgresql", Healthy: true},
			{Name: "redis", Healthy: true},
			{Name: "temporal", Healthy: true},
			{Name: "bridge", Healthy: true},
		}

		// Check PostgreSQL
		if err := pgClient.HealthCheck(ctx); err != nil {
			statuses[0].Healthy = false
			statuses[0].Error = err.Error()
		}

		// Check Redis (simple existence check)
		if _, err := redisClient.Exists(ctx, "health:check"); err != nil {
			statuses[1].Healthy = false
			statuses[1].Error = err.Error()
		}

		// Check Temporal
		if _, err := temporalClient.CheckHealth(ctx, &client.CheckHealthRequest{}); err != nil {
			statuses[2].Healthy = false
			statuses[2].Error = err.Error()
		}

		// Bridge client status (connected = healthy)
		statuses[3].Healthy = true

		// Determine overall status
		overallHealthy := true
		for _, s := range statuses {
			if !s.Healthy {
				overallHealthy = false
				break
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if overallHealthy {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		// Simple JSON response
		fmt.Fprintf(w, `{"status":"%s","service":"conn-dict-server"}`,
			func() string {
				if overallHealthy {
					return "healthy"
				}
				return "degraded"
			}())
	})

	return mux
}