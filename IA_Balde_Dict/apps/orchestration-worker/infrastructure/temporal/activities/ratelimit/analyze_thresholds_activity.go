package ratelimit

import (
	"context"
	"fmt"

	"go.temporal.io/sdk/activity"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
	domainRL "github.com/lb-conn/connector-dict/domain/ratelimit"
)

// AnalyzeThresholdsActivity checks for threshold violations
type AnalyzeThresholdsActivity struct {
	stateRepo         ports.StateRepository
	thresholdAnalyzer *domainRL.ThresholdAnalyzer
}

// ThresholdViolation represents a detected threshold violation
type ThresholdViolation struct {
	EndpointID string                 `json:"endpoint_id"`
	Severity   domainRL.AlertSeverity `json:"severity"`
	State      *domainRL.PolicyState  `json:"state"`
}

// ThresholdAnalysisResult contains the result of threshold analysis
type ThresholdAnalysisResult struct {
	TotalStates   int                   `json:"total_states"`
	OKCount       int                   `json:"ok_count"`
	WarningCount  int                   `json:"warning_count"`
	CriticalCount int                   `json:"critical_count"`
	Violations    []ThresholdViolation  `json:"violations"`
}

// NewAnalyzeThresholdsActivity creates a new AnalyzeThresholdsActivity
func NewAnalyzeThresholdsActivity(
	stateRepo ports.StateRepository,
	thresholdAnalyzer *domainRL.ThresholdAnalyzer,
) *AnalyzeThresholdsActivity {
	return &AnalyzeThresholdsActivity{
		stateRepo:         stateRepo,
		thresholdAnalyzer: thresholdAnalyzer,
	}
}

// Execute analyzes policy states for threshold violations
func (a *AnalyzeThresholdsActivity) Execute(ctx context.Context) (*ThresholdAnalysisResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("AnalyzeThresholdsActivity started")

	// Get all latest states
	states, err := a.stateRepo.GetLatestAll(ctx)
	if err != nil {
		logger.Error("Failed to get latest states", "error", err)
		return nil, fmt.Errorf("failed to get latest states: %w", err)
	}

	logger.Info("Analyzing thresholds for states", "count", len(states))

	result := &ThresholdAnalysisResult{
		TotalStates: len(states),
		Violations:  make([]ThresholdViolation, 0),
	}

	// Analyze each state
	for _, state := range states {
		status := a.thresholdAnalyzer.AnalyzeState(state)

		switch status {
		case domainRL.StatusOK:
			result.OKCount++

		case domainRL.StatusWarning:
			result.WarningCount++
			result.Violations = append(result.Violations, ThresholdViolation{
				EndpointID: state.EndpointID,
				Severity:   domainRL.SeverityWarning,
				State:      state,
			})
			logger.Warn("WARNING threshold breached",
				"endpoint_id", state.EndpointID,
				"available", state.AvailableTokens,
				"capacity", state.Capacity,
				"utilization", fmt.Sprintf("%.2f%%", state.GetUtilizationPercent()))

		case domainRL.StatusCritical:
			result.CriticalCount++
			result.Violations = append(result.Violations, ThresholdViolation{
				EndpointID: state.EndpointID,
				Severity:   domainRL.SeverityCritical,
				State:      state,
			})
			logger.Error("CRITICAL threshold breached",
				"endpoint_id", state.EndpointID,
				"available", state.AvailableTokens,
				"capacity", state.Capacity,
				"utilization", fmt.Sprintf("%.2f%%", state.GetUtilizationPercent()))
		}
	}

	logger.Info("AnalyzeThresholdsActivity completed",
		"total", result.TotalStates,
		"ok", result.OKCount,
		"warning", result.WarningCount,
		"critical", result.CriticalCount,
		"violations", len(result.Violations))

	return result, nil
}
