package setup

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"google.golang.org/grpc"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/database/repositories/ratelimit"
	grpcRL "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/grpc/ratelimit"
	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/metrics/ratelimit"
	pulsarRL "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/pulsar/ratelimit"
	activitiesRL "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities/ratelimit"
	workflowsRL "github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/workflows/ratelimit"
	domainRL "github.com/lb-conn/connector-dict/domain/ratelimit"
)

// RateLimitConfig contains configuration for rate limit monitoring
type RateLimitConfig struct {
	PulsarURL   string
	PulsarTopic string
	CronSchedule string
}

// SetupRateLimitMonitoring sets up all components for rate limit monitoring
func SetupRateLimitMonitoring(
	ctx context.Context,
	temporalClient client.Client,
	temporalWorker worker.Worker,
	bridgeGRPCConn *grpc.ClientConn,
	dbPool *pgxpool.Pool,
	prometheusRegistry prometheus.Registerer,
	config RateLimitConfig,
) error {
	// Create repositories
	policyRepo := ratelimit.NewPolicyRepository(dbPool)
	stateRepo := ratelimit.NewStateRepository(dbPool)
	alertRepo := ratelimit.NewAlertRepository(dbPool)

	// Create Bridge gRPC client
	bridgeClient := grpcRL.NewBridgeRateLimitClient(bridgeGRPCConn)

	// Create domain services
	calculator := domainRL.NewCalculator()
	thresholdAnalyzer := domainRL.NewThresholdAnalyzer()

	// Create Pulsar publisher
	alertPublisher, err := pulsarRL.NewAlertPublisher(pulsarRL.AlertPublisherConfig{
		PulsarURL: config.PulsarURL,
		Topic:     config.PulsarTopic,
	})
	if err != nil {
		return fmt.Errorf("failed to create alert publisher: %w", err)
	}

	// Create Prometheus metrics exporter
	metricsExporter := ratelimit.NewExporter(prometheusRegistry, stateRepo, alertRepo)

	// Create Temporal activities
	getPoliciesActivity := activitiesRL.NewGetPoliciesActivity(
		bridgeClient,
		policyRepo,
		stateRepo,
	)

	enrichMetricsActivity := activitiesRL.NewEnrichMetricsActivity(
		stateRepo,
		calculator,
	)

	analyzeThresholdsActivity := activitiesRL.NewAnalyzeThresholdsActivity(
		stateRepo,
		thresholdAnalyzer,
	)

	createAlertsActivity := activitiesRL.NewCreateAlertsActivity(alertRepo)

	autoResolveAlertsActivity := activitiesRL.NewAutoResolveAlertsActivity(
		alertRepo,
		stateRepo,
	)

	cleanupOldDataActivity := activitiesRL.NewCleanupOldDataActivity(stateRepo)

	publishAlertEventActivity := activitiesRL.NewPublishAlertEventActivity(
		alertPublisher.GetProducer(),
	)

	// Register activities with Temporal worker
	temporalWorker.RegisterActivity(getPoliciesActivity.Execute)
	temporalWorker.RegisterActivity(enrichMetricsActivity.Execute)
	temporalWorker.RegisterActivity(analyzeThresholdsActivity.Execute)
	temporalWorker.RegisterActivity(createAlertsActivity.Execute)
	temporalWorker.RegisterActivity(autoResolveAlertsActivity.Execute)
	temporalWorker.RegisterActivity(cleanupOldDataActivity.Execute)
	temporalWorker.RegisterActivity(publishAlertEventActivity.Execute)

	// Create and register workflow
	monitorPoliciesWorkflow := workflowsRL.NewMonitorPoliciesWorkflow()
	temporalWorker.RegisterWorkflow(monitorPoliciesWorkflow.Execute)

	// Start cron workflow
	workflowOptions := client.StartWorkflowOptions{
		ID:           "dict-rate-limit-monitor-cron",
		TaskQueue:    "dict-workflows",
		CronSchedule: config.CronSchedule, // "*/5 * * * *" = every 5 minutes
	}

	_, err = temporalClient.ExecuteWorkflow(ctx, workflowOptions, monitorPoliciesWorkflow.Execute)
	if err != nil {
		return fmt.Errorf("failed to start rate limit monitoring workflow: %w", err)
	}

	// Start periodic metrics update (every 30 seconds)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := metricsExporter.UpdateMetrics(ctx); err != nil {
					// Log error but don't stop the ticker
					fmt.Printf("Failed to update metrics: %v\n", err)
				}
			}
		}
	}()

	return nil
}
