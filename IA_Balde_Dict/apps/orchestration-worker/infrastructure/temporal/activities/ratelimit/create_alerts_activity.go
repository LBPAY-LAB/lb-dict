package ratelimit

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
	domainRL "github.com/lb-conn/connector-dict/domain/ratelimit"
)

// CreateAlertsActivity creates alerts for threshold violations
type CreateAlertsActivity struct {
	alertRepo ports.AlertRepository
}

// CreateAlertsInput contains the input for creating alerts
type CreateAlertsInput struct {
	Violations []ThresholdViolation `json:"violations"`
}

// CreateAlertsResult contains the result of creating alerts
type CreateAlertsResult struct {
	CreatedCount int `json:"created_count"`
	SkippedCount int `json:"skipped_count"`
}

// NewCreateAlertsActivity creates a new CreateAlertsActivity
func NewCreateAlertsActivity(alertRepo ports.AlertRepository) *CreateAlertsActivity {
	return &CreateAlertsActivity{
		alertRepo: alertRepo,
	}
}

// Execute creates alerts for detected threshold violations
func (a *CreateAlertsActivity) Execute(ctx context.Context, input CreateAlertsInput) (*CreateAlertsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("CreateAlertsActivity started", "violations", len(input.Violations))

	if len(input.Violations) == 0 {
		logger.Info("No violations to create alerts for")
		return &CreateAlertsResult{CreatedCount: 0, SkippedCount: 0}, nil
	}

	createdCount := 0
	skippedCount := 0

	for _, violation := range input.Violations {
		// Check if unresolved alert already exists for this endpoint
		existingAlerts, err := a.alertRepo.GetUnresolvedByEndpoint(ctx, violation.EndpointID)
		if err != nil {
			logger.Warn("Failed to check existing alerts",
				"endpoint_id", violation.EndpointID,
				"error", err)
		}

		// Skip if alert with same or higher severity already exists
		skipAlert := false
		for _, existing := range existingAlerts {
			if existing.Severity == violation.Severity {
				logger.Info("Alert already exists with same severity, skipping",
					"endpoint_id", violation.EndpointID,
					"severity", violation.Severity)
				skipAlert = true
				break
			}
			if existing.Severity == domainRL.SeverityCritical && violation.Severity == domainRL.SeverityWarning {
				logger.Info("CRITICAL alert already exists, skipping WARNING",
					"endpoint_id", violation.EndpointID)
				skipAlert = true
				break
			}
		}

		if skipAlert {
			skippedCount++
			continue
		}

		// Create new alert
		alert, err := domainRL.NewAlert(
			violation.EndpointID,
			violation.Severity,
			violation.State,
		)
		if err != nil {
			logger.Error("Failed to create alert object",
				"endpoint_id", violation.EndpointID,
				"error", err)
			skippedCount++
			continue
		}

		// Save alert to database
		if err := a.alertRepo.Save(ctx, alert); err != nil {
			logger.Error("Failed to save alert",
				"endpoint_id", violation.EndpointID,
				"error", err)
			skippedCount++
			continue
		}

		logger.Info("Created alert",
			"alert_id", alert.ID,
			"endpoint_id", violation.EndpointID,
			"severity", violation.Severity,
			"utilization", fmt.Sprintf("%.2f%%", alert.UtilizationPercent))

		createdCount++
	}

	result := &CreateAlertsResult{
		CreatedCount: createdCount,
		SkippedCount: skippedCount,
	}

	logger.Info("CreateAlertsActivity completed",
		"created", createdCount,
		"skipped", skippedCount)

	return result, nil
}
