package ratelimit

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
)

// AutoResolveAlertsActivity automatically resolves alerts when tokens recover
type AutoResolveAlertsActivity struct {
	alertRepo ports.AlertRepository
	stateRepo ports.StateRepository
}

// AutoResolveResult contains the result of auto-resolving alerts
type AutoResolveResult struct {
	TotalResolved int `json:"total_resolved"`
	EndpointCount int `json:"endpoint_count"`
}

// NewAutoResolveAlertsActivity creates a new AutoResolveAlertsActivity
func NewAutoResolveAlertsActivity(
	alertRepo ports.AlertRepository,
	stateRepo ports.StateRepository,
) *AutoResolveAlertsActivity {
	return &AutoResolveAlertsActivity{
		alertRepo: alertRepo,
		stateRepo: stateRepo,
	}
}

// Execute auto-resolves alerts based on current token levels
func (a *AutoResolveAlertsActivity) Execute(ctx context.Context) (*AutoResolveResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("AutoResolveAlertsActivity started")

	// Get all latest states
	states, err := a.stateRepo.GetLatestAll(ctx)
	if err != nil {
		logger.Error("Failed to get latest states", "error", err)
		return nil, fmt.Errorf("failed to get latest states: %w", err)
	}

	logger.Info("Checking auto-resolve for states", "count", len(states))

	totalResolved := 0
	endpointCount := 0

	// For each state, try to auto-resolve alerts
	for _, state := range states {
		// Call database function auto_resolve_alerts()
		// This function resolves alerts based on business rules:
		// - WARNING alerts resolved when utilization < 80%
		// - CRITICAL alerts resolved when utilization < 90%
		resolved, err := a.alertRepo.AutoResolve(
			ctx,
			state.EndpointID,
			state.AvailableTokens,
			state.Capacity,
		)
		if err != nil {
			logger.Warn("Failed to auto-resolve alerts",
				"endpoint_id", state.EndpointID,
				"error", err)
			continue
		}

		if resolved > 0 {
			logger.Info("Auto-resolved alerts",
				"endpoint_id", state.EndpointID,
				"count", resolved,
				"available", state.AvailableTokens,
				"capacity", state.Capacity,
				"utilization", fmt.Sprintf("%.2f%%", state.GetUtilizationPercent()))

			totalResolved += resolved
			endpointCount++
		}
	}

	result := &AutoResolveResult{
		TotalResolved: totalResolved,
		EndpointCount: endpointCount,
	}

	logger.Info("AutoResolveAlertsActivity completed",
		"total_resolved", totalResolved,
		"endpoints", endpointCount)

	return result, nil
}
