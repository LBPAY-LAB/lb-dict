package ratelimit

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/infrastructure/temporal/activities/ratelimit"
)

// MonitorPoliciesWorkflow monitors rate limit policies every 5 minutes
type MonitorPoliciesWorkflow struct{}

// NewMonitorPoliciesWorkflow creates a new MonitorPoliciesWorkflow
func NewMonitorPoliciesWorkflow() *MonitorPoliciesWorkflow {
	return &MonitorPoliciesWorkflow{}
}

// Execute runs the monitoring workflow
func (w *MonitorPoliciesWorkflow) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("MonitorPoliciesWorkflow started")

	// Activity options with retry policy
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    5,
			NonRetryableErrorTypes: []string{
				"BridgeAuthError",
				"BridgePermissionError",
			},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Activity 1: Get policies from Bridge
	logger.Info("Executing GetPoliciesActivity")
	var policiesResult ratelimit.GetPoliciesResult
	err := workflow.ExecuteActivity(ctx, "GetPoliciesActivity").Get(ctx, &policiesResult)
	if err != nil {
		logger.Error("GetPoliciesActivity failed", "error", err)
		return fmt.Errorf("GetPoliciesActivity failed: %w", err)
	}

	logger.Info("GetPoliciesActivity completed",
		"category", policiesResult.PSPCategory,
		"policies", policiesResult.PolicyCount,
		"states", policiesResult.StateCount)

	// Activity 2: Enrich metrics (non-critical - continue on error)
	logger.Info("Executing EnrichMetricsActivity")
	var enrichResult ratelimit.EnrichMetricsResult
	err = workflow.ExecuteActivity(ctx, "EnrichMetricsActivity").Get(ctx, &enrichResult)
	if err != nil {
		logger.Warn("EnrichMetricsActivity failed (non-critical)", "error", err)
		// Continue workflow - metrics enrichment is not critical
	} else {
		logger.Info("EnrichMetricsActivity completed",
			"enriched", enrichResult.EnrichedCount,
			"skipped", enrichResult.SkippedCount)
	}

	// Activity 3: Analyze thresholds
	logger.Info("Executing AnalyzeThresholdsActivity")
	var thresholdResults ratelimit.ThresholdAnalysisResult
	err = workflow.ExecuteActivity(ctx, "AnalyzeThresholdsActivity").Get(ctx, &thresholdResults)
	if err != nil {
		logger.Error("AnalyzeThresholdsActivity failed", "error", err)
		return fmt.Errorf("AnalyzeThresholdsActivity failed: %w", err)
	}

	logger.Info("AnalyzeThresholdsActivity completed",
		"total", thresholdResults.TotalStates,
		"ok", thresholdResults.OKCount,
		"warning", thresholdResults.WarningCount,
		"critical", thresholdResults.CriticalCount,
		"violations", len(thresholdResults.Violations))

	// Activity 4: Create alerts if violations detected
	if len(thresholdResults.Violations) > 0 {
		logger.Info("Executing CreateAlertsActivity", "violations", len(thresholdResults.Violations))

		createAlertsInput := ratelimit.CreateAlertsInput{
			Violations: thresholdResults.Violations,
		}

		var createAlertsResult ratelimit.CreateAlertsResult
		err = workflow.ExecuteActivity(ctx, "CreateAlertsActivity", createAlertsInput).Get(ctx, &createAlertsResult)
		if err != nil {
			logger.Error("CreateAlertsActivity failed", "error", err)
			// Continue - don't fail entire workflow
		} else {
			logger.Info("CreateAlertsActivity completed",
				"created", createAlertsResult.CreatedCount,
				"skipped", createAlertsResult.SkippedCount)

			// Activity 5: Publish alert events to Pulsar (if alerts created)
			if createAlertsResult.CreatedCount > 0 {
				logger.Info("Alerts created, publishing to Pulsar")
				// Note: PublishAlertEventActivity would need the actual alert objects
				// For now, this is a placeholder - implementation would need to fetch created alerts
			}
		}
	} else {
		logger.Info("No violations detected, skipping alert creation")
	}

	// Activity 6: Auto-resolve alerts (if tokens recovered)
	logger.Info("Executing AutoResolveAlertsActivity")
	var autoResolveResult ratelimit.AutoResolveResult
	err = workflow.ExecuteActivity(ctx, "AutoResolveAlertsActivity").Get(ctx, &autoResolveResult)
	if err != nil {
		logger.Warn("AutoResolveAlertsActivity failed (non-critical)", "error", err)
	} else {
		logger.Info("AutoResolveAlertsActivity completed",
			"total_resolved", autoResolveResult.TotalResolved,
			"endpoints", autoResolveResult.EndpointCount)
	}

	// Activity 7: Cleanup old data (conditional - run once per day at 03:00 AM)
	if shouldRunCleanup(workflow.Now(ctx)) {
		logger.Info("Executing CleanupOldDataActivity")
		var cleanupResult ratelimit.CleanupResult
		err = workflow.ExecuteActivity(ctx, "CleanupOldDataActivity").Get(ctx, &cleanupResult)
		if err != nil {
			logger.Warn("CleanupOldDataActivity failed (non-critical)", "error", err)
		} else {
			logger.Info("CleanupOldDataActivity completed",
				"deleted_count", cleanupResult.DeletedCount,
				"cutoff_date", cleanupResult.CutoffDate)
		}
	}

	logger.Info("MonitorPoliciesWorkflow completed successfully")
	return nil
}

// shouldRunCleanup determines if cleanup should run (once per day at 03:00 AM)
func shouldRunCleanup(now time.Time) bool {
	hour := now.Hour()
	minute := now.Minute()

	// Run cleanup if current time is between 03:00 and 03:05 (5-minute window)
	return hour == 3 && minute >= 0 && minute < 5
}
