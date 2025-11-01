package ratelimit

import (
	"time"
)

// Calculator provides pure functions for calculating rate limit metrics
type Calculator struct{}

// NewCalculator creates a new Calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// CalculateConsumptionRate calculates tokens per minute consumption rate
// based on current and previous state snapshots
//
// Formula: (previousTokens - currentTokens) / elapsedMinutes
//
// Returns:
//   - consumption rate in tokens/minute
//   - error if insufficient data or division by zero
func (c *Calculator) CalculateConsumptionRate(
	currentState *PolicyState,
	previousState *PolicyState,
) (float64, error) {
	if previousState == nil {
		return 0, ErrInsufficientData
	}

	// Calculate elapsed time in minutes
	elapsed := currentState.ResponseTimestamp.Sub(previousState.ResponseTimestamp)
	elapsedMinutes := elapsed.Minutes()

	if elapsedMinutes <= 0 {
		return 0, ErrDivisionByZero
	}

	// Calculate token consumption (can be negative if refilling faster than consuming)
	tokenConsumed := previousState.AvailableTokens - currentState.AvailableTokens

	// Calculate consumption rate per minute
	consumptionRate := float64(tokenConsumed) / elapsedMinutes

	// Negative rate means tokens are increasing (refill > consumption)
	// Return 0 for negative rates to indicate no net consumption
	if consumptionRate < 0 {
		return 0, nil
	}

	return consumptionRate, nil
}

// CalculateRecoveryETA calculates estimated seconds until full token recovery
//
// Formula: (capacity - available_tokens) / refill_rate
// where refill_rate = refill_tokens / refill_period_sec
//
// Returns:
//   - estimated seconds to full recovery
//   - error if invalid parameters
func (c *Calculator) CalculateRecoveryETA(state *PolicyState) (int, error) {
	if state.Capacity <= 0 {
		return 0, ErrInvalidCapacity
	}

	if state.RefillPeriodSec <= 0 || state.RefillTokens <= 0 {
		return 0, ErrInvalidRefillRate
	}

	// If already at full capacity, no recovery needed
	if state.AvailableTokens >= state.Capacity {
		return 0, nil
	}

	// Calculate tokens needed to reach full capacity
	tokensNeeded := state.Capacity - state.AvailableTokens

	// Calculate refill rate (tokens per second)
	refillRate := float64(state.RefillTokens) / float64(state.RefillPeriodSec)

	// Calculate seconds to recovery
	recoverySeconds := float64(tokensNeeded) / refillRate

	return int(recoverySeconds), nil
}

// CalculateExhaustionProjection calculates estimated seconds until token exhaustion
// if current consumption trend continues
//
// Formula: available_tokens / (consumption_rate_per_minute / 60)
//
// Returns:
//   - estimated seconds to exhaustion
//   - 0 if no consumption (tokens stable or increasing)
//   - error if invalid parameters
func (c *Calculator) CalculateExhaustionProjection(state *PolicyState) (int, error) {
	// If consumption rate is zero or negative, tokens won't exhaust
	if state.ConsumptionRatePerMinute <= 0 {
		return 0, nil // No exhaustion projected
	}

	if state.AvailableTokens <= 0 {
		return 0, nil // Already exhausted
	}

	// Convert consumption rate to tokens per second
	consumptionRatePerSecond := state.ConsumptionRatePerMinute / 60.0

	if consumptionRatePerSecond <= 0 {
		return 0, ErrDivisionByZero
	}

	// Calculate seconds until exhaustion
	exhaustionSeconds := float64(state.AvailableTokens) / consumptionRatePerSecond

	return int(exhaustionSeconds), nil
}

// Calculate404Rate calculates the percentage of 404 errors in recent requests
//
// This requires historical request data which may not be available in all contexts.
// For now, this is a placeholder that returns 0.
//
// TODO: Implement when request history tracking is available
func (c *Calculator) Calculate404Rate(
	endpointID string,
	since time.Time,
	until time.Time,
) (float64, error) {
	// Placeholder implementation
	// In production, this would query request logs or metrics to calculate 404 rate
	return 0, nil
}

// DetectCategoryChange detects if PSP category changed between states
func (c *Calculator) DetectCategoryChange(
	currentState *PolicyState,
	previousState *PolicyState,
) bool {
	if previousState == nil {
		return false
	}

	// Empty category is considered "not set", not a change
	if currentState.PSPCategory == "" && previousState.PSPCategory == "" {
		return false
	}

	return currentState.PSPCategory != previousState.PSPCategory
}

// EnrichStateWithMetrics calculates and populates all calculated metrics in a PolicyState
// This is a convenience function to populate all metrics at once
func (c *Calculator) EnrichStateWithMetrics(
	currentState *PolicyState,
	previousState *PolicyState,
) error {
	// Calculate consumption rate
	if previousState != nil {
		consumptionRate, err := c.CalculateConsumptionRate(currentState, previousState)
		if err != nil && err != ErrInsufficientData {
			return err
		}
		currentState.ConsumptionRatePerMinute = consumptionRate
	}

	// Calculate recovery ETA
	recoveryETA, err := c.CalculateRecoveryETA(currentState)
	if err != nil {
		return err
	}
	currentState.RecoveryETASeconds = recoveryETA

	// Calculate exhaustion projection (only if we have consumption rate)
	if currentState.ConsumptionRatePerMinute > 0 {
		exhaustionProjection, err := c.CalculateExhaustionProjection(currentState)
		if err != nil {
			return err
		}
		currentState.ExhaustionProjectionSeconds = exhaustionProjection
	}

	// Calculate 404 rate (placeholder for now)
	// In production, this would use actual request history
	currentState.Error404Rate = 0

	return nil
}
