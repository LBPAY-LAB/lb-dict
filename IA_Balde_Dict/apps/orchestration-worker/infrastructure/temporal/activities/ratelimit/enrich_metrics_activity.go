package ratelimit

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
	domainRL "github.com/lb-conn/connector-dict/domain/ratelimit"
)

// EnrichMetricsActivity calculates and enriches policy states with computed metrics
type EnrichMetricsActivity struct {
	stateRepo  ports.StateRepository
	calculator *domainRL.Calculator
}

// EnrichMetricsResult contains the result of enriching metrics
type EnrichMetricsResult struct {
	EnrichedCount int `json:"enriched_count"`
	SkippedCount  int `json:"skipped_count"`
}

// NewEnrichMetricsActivity creates a new EnrichMetricsActivity
func NewEnrichMetricsActivity(
	stateRepo ports.StateRepository,
	calculator *domainRL.Calculator,
) *EnrichMetricsActivity {
	return &EnrichMetricsActivity{
		stateRepo:  stateRepo,
		calculator: calculator,
	}
}

// Execute enriches policy states with calculated metrics
func (a *EnrichMetricsActivity) Execute(ctx context.Context) (*EnrichMetricsResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("EnrichMetricsActivity started")

	// Get all latest states
	states, err := a.stateRepo.GetLatestAll(ctx)
	if err != nil {
		logger.Error("Failed to get latest states", "error", err)
		return nil, fmt.Errorf("failed to get latest states: %w", err)
	}

	logger.Info("Retrieved latest states for enrichment", "count", len(states))

	enrichedCount := 0
	skippedCount := 0

	// Enrich each state with calculated metrics
	for _, state := range states {
		// Get previous state for consumption rate calculation
		previousState, err := a.stateRepo.GetPreviousState(ctx, state.EndpointID, state.CreatedAt)
		if err != nil {
			logger.Warn("Failed to get previous state",
				"endpoint_id", state.EndpointID,
				"error", err)
			// Continue without previous state - some metrics will be 0
		}

		// Calculate and populate metrics
		if err := a.calculator.EnrichStateWithMetrics(state, previousState); err != nil {
			logger.Warn("Failed to enrich metrics",
				"endpoint_id", state.EndpointID,
				"error", err)
			skippedCount++
			continue
		}

		// Update state in database with enriched metrics
		if err := a.stateRepo.Save(ctx, state); err != nil {
			logger.Error("Failed to save enriched state",
				"endpoint_id", state.EndpointID,
				"error", err)
			skippedCount++
			continue
		}

		enrichedCount++
	}

	result := &EnrichMetricsResult{
		EnrichedCount: enrichedCount,
		SkippedCount:  skippedCount,
	}

	logger.Info("EnrichMetricsActivity completed",
		"enriched", enrichedCount,
		"skipped", skippedCount)

	return result, nil
}
