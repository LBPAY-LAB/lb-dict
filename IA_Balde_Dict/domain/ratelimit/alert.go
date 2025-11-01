package ratelimit

import (
	"fmt"
	"time"
)

// AlertSeverity represents the severity level of a rate limit alert
type AlertSeverity string

const (
	// SeverityWarning indicates 20% or less tokens remaining (80%+ utilized)
	SeverityWarning AlertSeverity = "WARNING"

	// SeverityCritical indicates 10% or less tokens remaining (90%+ utilized)
	SeverityCritical AlertSeverity = "CRITICAL"
)

// Alert represents a rate limit threshold violation alert
type Alert struct {
	// ID is the database primary key
	ID int64

	// EndpointID references the policy
	EndpointID string

	// Alert severity and threshold
	Severity         AlertSeverity
	ThresholdPercent int // 10 or 20

	// Token bucket state at alert time
	AvailableTokens    int
	Capacity           int
	UtilizationPercent float64

	// Calculated metrics at alert time
	ConsumptionRatePerMinute    float64
	RecoveryETASeconds          int
	ExhaustionProjectionSeconds int
	PSPCategory                 string

	// Alert message
	Message string

	// Alert resolution tracking
	Resolved        bool
	ResolvedAt      *time.Time
	ResolutionNotes string

	// CreatedAt is when the alert was created
	CreatedAt time.Time
}

// NewAlert creates a new Alert with validation
func NewAlert(
	endpointID string,
	severity AlertSeverity,
	state *PolicyState,
) (*Alert, error) {
	if err := validateSeverity(severity); err != nil {
		return nil, err
	}

	thresholdPercent, err := getThresholdPercent(severity)
	if err != nil {
		return nil, err
	}

	utilizationPercent := state.GetUtilizationPercent()

	// Validate business rule: severity matches utilization
	if !validateSeverityThreshold(severity, utilizationPercent) {
		return nil, fmt.Errorf(
			"severity %s requires utilization >= %d%%, got %.2f%%",
			severity,
			100-thresholdPercent,
			utilizationPercent,
		)
	}

	message := generateAlertMessage(endpointID, severity, state)

	return &Alert{
		EndpointID:                  endpointID,
		Severity:                    severity,
		ThresholdPercent:            thresholdPercent,
		AvailableTokens:             state.AvailableTokens,
		Capacity:                    state.Capacity,
		UtilizationPercent:          utilizationPercent,
		ConsumptionRatePerMinute:    state.ConsumptionRatePerMinute,
		RecoveryETASeconds:          state.RecoveryETASeconds,
		ExhaustionProjectionSeconds: state.ExhaustionProjectionSeconds,
		PSPCategory:                 state.PSPCategory,
		Message:                     message,
		Resolved:                    false,
		CreatedAt:                   time.Now().UTC(),
	}, nil
}

// Resolve marks the alert as resolved
func (a *Alert) Resolve(notes string) {
	now := time.Now().UTC()
	a.Resolved = true
	a.ResolvedAt = &now
	a.ResolutionNotes = notes
}

// Validate performs validation on Alert fields
func (a *Alert) Validate() error {
	if err := validateSeverity(a.Severity); err != nil {
		return err
	}

	if a.ThresholdPercent != 10 && a.ThresholdPercent != 20 {
		return ErrInvalidThreshold
	}

	if a.AvailableTokens < 0 {
		return NewValidationError("AvailableTokens", a.AvailableTokens, "cannot be negative")
	}

	if a.Capacity <= 0 {
		return NewValidationError("Capacity", a.Capacity, "must be greater than zero")
	}

	if !validateSeverityThreshold(a.Severity, a.UtilizationPercent) {
		return fmt.Errorf(
			"severity %s requires utilization >= %d%%, got %.2f%%",
			a.Severity,
			100-a.ThresholdPercent,
			a.UtilizationPercent,
		)
	}

	if a.PSPCategory != "" {
		if err := validatePSPCategory(a.PSPCategory); err != nil {
			return err
		}
	}

	return nil
}

// Helper functions

func validateSeverity(severity AlertSeverity) error {
	if severity != SeverityWarning && severity != SeverityCritical {
		return ErrInvalidSeverity
	}
	return nil
}

func getThresholdPercent(severity AlertSeverity) (int, error) {
	switch severity {
	case SeverityWarning:
		return 20, nil
	case SeverityCritical:
		return 10, nil
	default:
		return 0, ErrInvalidSeverity
	}
}

func validateSeverityThreshold(severity AlertSeverity, utilizationPercent float64) bool {
	switch severity {
	case SeverityWarning:
		return utilizationPercent >= 80.0
	case SeverityCritical:
		return utilizationPercent >= 90.0
	default:
		return false
	}
}

func generateAlertMessage(endpointID string, severity AlertSeverity, state *PolicyState) string {
	return fmt.Sprintf(
		"[%s] Rate limit %s: Endpoint %s has %d/%d tokens remaining (%.2f%% utilized). "+
			"Recovery ETA: %ds, Exhaustion projection: %ds, Consumption rate: %.2f tokens/min",
		severity,
		map[AlertSeverity]string{
			SeverityWarning:  "WARNING threshold breached",
			SeverityCritical: "CRITICAL threshold breached",
		}[severity],
		endpointID,
		state.AvailableTokens,
		state.Capacity,
		state.GetUtilizationPercent(),
		state.RecoveryETASeconds,
		state.ExhaustionProjectionSeconds,
		state.ConsumptionRatePerMinute,
	)
}
