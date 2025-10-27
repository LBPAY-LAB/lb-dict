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

	"github.com/lbpay-lab/conn-dict/internal/activities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/database"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/grpc"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/pulsar"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/lbpay-lab/conn-dict/internal/workflows"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

// Prometheus metrics for worker
var (
	workerTasksProcessed = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "conn_dict",
			Subsystem: "worker",
			Name:      "tasks_processed_total",
			Help:      "Total number of tasks processed by worker",
		},
		[]string{"task_type", "status"},
	)

	workerTaskDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "conn_dict",
			Subsystem: "worker",
			Name:      "task_duration_seconds",
			Help:      "Task execution duration in seconds",
			Buckets:   []float64{0.1, 0.5, 1.0, 5.0, 10.0, 30.0, 60.0, 120.0, 300.0},
		},
		[]string{"task_type"},
	)

	workerHealthStatus = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "conn_dict",
			Subsystem: "worker",
			Name:      "health_status",
			Help:      "Worker health status (1 = healthy, 0 = unhealthy)",
		},
	)

	workerActiveWorkflows = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "conn_dict",
			Subsystem: "worker",
			Name:      "active_workflows",
			Help:      "Number of active workflow executions",
		},
	)

	workerActiveActivities = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "conn_dict",
			Subsystem: "worker",
			Name:      "active_activities",
			Help:      "Number of active activity executions",
		},
	)
)

func main() {
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

	logger.Info("Starting Temporal worker...")

	// Get Temporal server address from environment
	temporalAddress := os.Getenv("TEMPORAL_ADDRESS")
	if temporalAddress == "" {
		temporalAddress = "localhost:7233"
	}

	// Get namespace from environment
	namespace := os.Getenv("TEMPORAL_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	// Create Temporal client
	temporalClient, err := client.Dial(client.Options{
		HostPort:  temporalAddress,
		Namespace: namespace,
		// Note: Temporal SDK uses its own logger interface, incompatible with logrus
		// We use logrus for application logging, Temporal SDK will use default logger
	})
	if err != nil {
		log.Fatalf("Failed to create Temporal client: %v", err)
	}
	defer temporalClient.Close()

	logger.WithFields(logrus.Fields{
		"temporal_address": temporalAddress,
		"namespace":        namespace,
	}).Info("Connected to Temporal server")

	// Get task queue from environment
	taskQueue := os.Getenv("TEMPORAL_TASK_QUEUE")
	if taskQueue == "" {
		taskQueue = "conn-dict-task-queue"
	}

	// Get worker concurrency settings from environment
	maxConcurrentActivities := getEnvAsInt("MAX_CONCURRENT_ACTIVITIES", 200)
	maxConcurrentWorkflows := getEnvAsInt("MAX_CONCURRENT_WORKFLOWS", 100)

	// Create Temporal worker with optimized settings per requirements
	w := worker.New(temporalClient, taskQueue, worker.Options{
		MaxConcurrentActivityExecutionSize:     maxConcurrentActivities,
		MaxConcurrentWorkflowTaskExecutionSize: maxConcurrentWorkflows,
		MaxConcurrentActivityTaskPollers:       20,
		MaxConcurrentWorkflowTaskPollers:       10,
		EnableSessionWorker:                    true,
		MaxConcurrentSessionExecutionSize:      50,
	})

	logger.WithFields(logrus.Fields{
		"task_queue":                taskQueue,
		"max_concurrent_activities": maxConcurrentActivities,
		"max_concurrent_workflows":  maxConcurrentWorkflows,
	}).Info("Worker configuration")

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
	syncReportRepo := repositories.NewSyncReportRepository(postgresClient, logger)

	// Initialize Pulsar producer
	pulsarConfig := pulsar.ProducerConfig{
		URL:            getEnvOrDefault("PULSAR_URL", "pulsar://localhost:6650"),
		Topic:          getEnvOrDefault("PULSAR_TOPIC", "persistent://public/default/dict-events"),
		ProducerName:   getEnvOrDefault("PULSAR_PRODUCER_NAME", "conn-dict-worker"),
		MaxReconnect:   10,
		ConnectTimeout: 30 * time.Second,
	}

	pulsarProducer, err := pulsar.NewProducer(pulsarConfig, logger)
	if err != nil {
		log.Fatalf("Failed to initialize Pulsar producer: %v", err)
	}
	defer pulsarProducer.Close()

	logger.Info("Pulsar producer initialized successfully")

	// Register workflows
	w.RegisterWorkflow(workflows.ClaimWorkflow)
	logger.Info("Registered ClaimWorkflow")

	// Register Entry workflow with waiting period (30 days)
	w.RegisterWorkflow(workflows.DeleteEntryWithWaitingPeriodWorkflow)
	logger.Info("Registered DeleteEntryWithWaitingPeriodWorkflow")

	// Register Infraction workflow
	w.RegisterWorkflow(workflows.InvestigateInfractionWorkflow)
	logger.Info("Registered InvestigateInfractionWorkflow")

	// Register VSYNC workflows
	w.RegisterWorkflow(workflows.VSyncWorkflow)
	w.RegisterWorkflow(workflows.VSyncSchedulerWorkflow)
	logger.Info("Registered VSYNC workflows (Sync, Scheduler)")

	// Register Claim activities
	claimActivities := activities.NewClaimActivities(logger, claimRepo, pulsarProducer)
	w.RegisterActivity(claimActivities.CreateClaimActivity)
	w.RegisterActivity(claimActivities.SubmitClaimToBacenActivity)
	w.RegisterActivity(claimActivities.UpdateClaimStatusActivity)
	w.RegisterActivity(claimActivities.NotifyDonorActivity)
	w.RegisterActivity(claimActivities.CompleteClaimActivity)
	w.RegisterActivity(claimActivities.CancelClaimActivity)
	w.RegisterActivity(claimActivities.ExpireClaimActivity)
	w.RegisterActivity(claimActivities.GetClaimStatusActivity)
	w.RegisterActivity(claimActivities.ValidateClaimEligibilityActivity)
	w.RegisterActivity(claimActivities.SendClaimConfirmationActivity)
	w.RegisterActivity(claimActivities.UpdateEntryOwnershipActivity)
	w.RegisterActivity(claimActivities.PublishClaimEventActivity)
	logger.Info("Registered Claim activities (including Sprint 1 activities: Create, SubmitToBacen, UpdateStatus)")

	// Register Entry activities
	entryActivities := activities.NewEntryActivities(logger, entryRepo, pulsarProducer)
	w.RegisterActivity(entryActivities.CreateEntryActivity)
	w.RegisterActivity(entryActivities.UpdateEntryActivity)
	w.RegisterActivity(entryActivities.DeleteEntryActivity)
	w.RegisterActivity(entryActivities.ActivateEntryActivity)
	w.RegisterActivity(entryActivities.DeactivateEntryActivity)
	w.RegisterActivity(entryActivities.GetEntryStatusActivity)
	w.RegisterActivity(entryActivities.ValidateEntryActivity)
	w.RegisterActivity(entryActivities.UpdateEntryOwnershipActivity)
	logger.Info("Registered Entry activities")

	// Register Infraction activities
	infractionActivities := activities.NewInfractionActivities(logger, infractionRepo, pulsarProducer)
	w.RegisterActivity(infractionActivities.CreateInfractionActivity)
	w.RegisterActivity(infractionActivities.InvestigateInfractionActivity)
	w.RegisterActivity(infractionActivities.ResolveInfractionActivity)
	w.RegisterActivity(infractionActivities.DismissInfractionActivity)
	w.RegisterActivity(infractionActivities.EscalateInfractionActivity)
	w.RegisterActivity(infractionActivities.AddEvidenceActivity)
	w.RegisterActivity(infractionActivities.GetInfractionStatusActivity)
	w.RegisterActivity(infractionActivities.ValidateInfractionEligibilityActivity)
	w.RegisterActivity(infractionActivities.NotifyReportedParticipantActivity)
	w.RegisterActivity(infractionActivities.NotifyBacenActivity)
	w.RegisterActivity(infractionActivities.PublishInfractionEventActivity)
	logger.Info("Registered Infraction activities")

	// Initialize Bridge gRPC client for VSYNC
	bridgeAddress := getEnvOrDefault("BRIDGE_ADDRESS", "localhost:9094")
	bridgeClient, err := grpc.NewBridgeClient(&grpc.BridgeClientConfig{
		Address:        bridgeAddress,
		ConnectTimeout: 10 * time.Second,
		RequestTimeout: 30 * time.Second,
	}, logger)
	if err != nil {
		logger.WithError(err).Warn("Failed to initialize Bridge client - VSYNC will not work")
		// Don't fail startup - Bridge may not be available in dev environment
	} else {
		logger.WithField("bridge_address", bridgeAddress).Info("Bridge client initialized successfully")
	}

	// Register VSYNC activities
	vsyncActivities := activities.NewVSyncActivities(logger, entryRepo, syncReportRepo, bridgeClient)
	w.RegisterActivity(vsyncActivities.FetchBacenEntriesActivity)
	w.RegisterActivity(vsyncActivities.CompareEntriesActivity)
	w.RegisterActivity(vsyncActivities.GenerateSyncReportActivity)
	logger.Info("Registered VSYNC activities (Fetch, Compare, GenerateReport)")

	logger.Info("Registered all activities (Claim, Entry, Infraction, VSYNC)")

	// Start HTTP server for metrics and health checks
	metricsPort := getEnvAsInt("METRICS_PORT", 9093)
	healthPort := getEnvAsInt("HEALTH_PORT", 8081)

	// Start metrics server
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
	healthServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", healthPort),
		Handler:      createHealthCheckHandler(logger, postgresClient, temporalClient),
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

	// Set initial health status
	workerHealthStatus.Set(1)

	// Start worker in a goroutine
	errChan := make(chan error, 1)
	go func() {
		logger.Infof("Starting worker on task queue: %s", taskQueue)
		if err := w.Run(worker.InterruptCh()); err != nil {
			errChan <- err
		}
	}()

	logger.Info("Worker started successfully")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errChan:
		logger.Errorf("Worker error: %v", err)
		workerHealthStatus.Set(0)
	case sig := <-sigChan:
		logger.Infof("Received signal: %v", sig)
	}

	// Graceful shutdown
	logger.Info("Shutting down worker...")
	workerHealthStatus.Set(0)

	// Stop accepting new tasks
	w.Stop()

	// Wait for current tasks to complete (max 30s)
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	logger.Info("Waiting for active tasks to complete (max 30s)...")

	// Shutdown health check server
	if err := healthServer.Shutdown(shutdownCtx); err != nil {
		logger.Warnf("Health check server shutdown error: %v", err)
	}

	logger.Info("Worker stopped successfully")
}

// createHealthCheckHandler creates HTTP handler for health checks
func createHealthCheckHandler(logger *logrus.Logger, pgClient *database.PostgresClient, temporalClient client.Client) http.Handler {
	mux := http.NewServeMux()

	// Liveness probe - checks if worker is running
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"healthy","service":"conn-dict-worker"}`)
	})

	// Readiness probe - checks if worker can process tasks
	mux.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		// Check PostgreSQL health
		if err := pgClient.HealthCheck(ctx); err != nil {
			logger.WithError(err).Warn("PostgreSQL health check failed")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"not_ready","reason":"database_unhealthy","error":"%s"}`, err.Error())
			return
		}

		// Check Temporal connection
		_, err := temporalClient.CheckHealth(ctx, &client.CheckHealthRequest{})
		if err != nil {
			logger.WithError(err).Warn("Temporal health check failed")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, `{"status":"not_ready","reason":"temporal_unhealthy","error":"%s"}`, err.Error())
			return
		}

		// All checks passed
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"status":"ready","service":"conn-dict-worker"}`)
	})

	return mux
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