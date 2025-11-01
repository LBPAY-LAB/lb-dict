package ratelimit

// Threshold constants for rate limit monitoring
const (
	// ThresholdWarningPercent is the WARNING threshold (20% remaining = 80% utilized)
	ThresholdWarningPercent = 20

	// ThresholdCriticalPercent is the CRITICAL threshold (10% remaining = 90% utilized)
	ThresholdCriticalPercent = 10
)

// ThresholdStatus represents the threshold status of a rate limit state
type ThresholdStatus string

const (
	// StatusOK indicates tokens are above WARNING threshold (>20% remaining)
	StatusOK ThresholdStatus = "OK"

	// StatusWarning indicates tokens are at or below 20% remaining
	StatusWarning ThresholdStatus = "WARNING"

	// StatusCritical indicates tokens are at or below 10% remaining
	StatusCritical ThresholdStatus = "CRITICAL"
)

// ThresholdAnalyzer analyzes PolicyState and determines threshold violations
type ThresholdAnalyzer struct{}

// NewThresholdAnalyzer creates a new ThresholdAnalyzer
func NewThresholdAnalyzer() *ThresholdAnalyzer {
	return &ThresholdAnalyzer{}
}

// AnalyzeState determines the threshold status based on available tokens
// Business rules:
// - CRITICAL: available_tokens <= 10% of capacity (90%+ utilized)
// - WARNING: available_tokens <= 20% of capacity (80%+ utilized)
// - OK: available_tokens > 20% of capacity (<80% utilized)
func (ta *ThresholdAnalyzer) AnalyzeState(state *PolicyState) ThresholdStatus {
	remainingPercent := state.GetRemainingPercent()

	if remainingPercent <= float64(ThresholdCriticalPercent) {
		return StatusCritical
	}

	if remainingPercent <= float64(ThresholdWarningPercent) {
		return StatusWarning
	}

	return StatusOK
}

// ShouldCreateAlert determines if an alert should be created for the given state
func (ta *ThresholdAnalyzer) ShouldCreateAlert(state *PolicyState) bool {
	status := ta.AnalyzeState(state)
	return status == StatusWarning || status == StatusCritical
}

// GetAlertSeverity returns the alert severity for the given state
// Returns empty string if no alert should be created
func (ta *ThresholdAnalyzer) GetAlertSeverity(state *PolicyState) AlertSeverity {
	status := ta.AnalyzeState(state)

	switch status {
	case StatusCritical:
		return SeverityCritical
	case StatusWarning:
		return SeverityWarning
	default:
		return ""
	}
}

// ShouldResolveAlert determines if an alert should be auto-resolved
// WARNING alerts resolve when utilization < 80%
// CRITICAL alerts resolve when utilization < 90%
func (ta *ThresholdAnalyzer) ShouldResolveAlert(severity AlertSeverity, state *PolicyState) bool {
	utilizationPercent := state.GetUtilizationPercent()

	switch severity {
	case SeverityWarning:
		return utilizationPercent < 80.0
	case SeverityCritical:
		return utilizationPercent < 90.0
	default:
		return false
	}
}

// CalculateThresholdTokens calculates the token count at a given threshold percentage
func (ta *ThresholdAnalyzer) CalculateThresholdTokens(capacity int, thresholdPercent int) int {
	return capacity * thresholdPercent / 100
}

// GetWarningThresholdTokens returns the token count at WARNING threshold (20%)
func (ta *ThresholdAnalyzer) GetWarningThresholdTokens(capacity int) int {
	return ta.CalculateThresholdTokens(capacity, ThresholdWarningPercent)
}

// GetCriticalThresholdTokens returns the token count at CRITICAL threshold (10%)
func (ta *ThresholdAnalyzer) GetCriticalThresholdTokens(capacity int) int {
	return ta.CalculateThresholdTokens(capacity, ThresholdCriticalPercent)
}
