package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCalculator_CalculateConsumptionRate(t *testing.T) {
	calculator := NewCalculator()

	t.Run("calculates consumption rate correctly", func(t *testing.T) {
		now := time.Now().UTC()
		fiveMinutesAgo := now.Add(-5 * time.Minute)

		previousState := &PolicyState{
			EndpointID:        "ENTRIES_WRITE",
			AvailableTokens:   30000,
			Capacity:          36000,
			ResponseTimestamp: fiveMinutesAgo,
		}

		currentState := &PolicyState{
			EndpointID:        "ENTRIES_WRITE",
			AvailableTokens:   29000,
			Capacity:          36000,
			ResponseTimestamp: now,
		}

		rate, err := calculator.CalculateConsumptionRate(currentState, previousState)
		require.NoError(t, err)

		// (30000 - 29000) / 5 minutes = 200 tokens/minute
		assert.InDelta(t, 200.0, rate, 0.01)
	})

	t.Run("returns error when no previous state", func(t *testing.T) {
		currentState := &PolicyState{
			EndpointID:        "ENTRIES_WRITE",
			AvailableTokens:   29000,
			Capacity:          36000,
			ResponseTimestamp: time.Now().UTC(),
		}

		_, err := calculator.CalculateConsumptionRate(currentState, nil)
		assert.ErrorIs(t, err, ErrInsufficientData)
	})

	t.Run("returns zero for negative rate (tokens increasing)", func(t *testing.T) {
		now := time.Now().UTC()
		fiveMinutesAgo := now.Add(-5 * time.Minute)

		previousState := &PolicyState{
			AvailableTokens:   29000,
			ResponseTimestamp: fiveMinutesAgo,
		}

		currentState := &PolicyState{
			AvailableTokens:   30000, // Increased (refill > consumption)
			ResponseTimestamp: now,
		}

		rate, err := calculator.CalculateConsumptionRate(currentState, previousState)
		require.NoError(t, err)
		assert.Equal(t, 0.0, rate)
	})
}

func TestCalculator_CalculateRecoveryETA(t *testing.T) {
	calculator := NewCalculator()

	t.Run("calculates recovery ETA correctly", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 30000,
			Capacity:        36000,
			RefillTokens:    1200,
			RefillPeriodSec: 60,
		}

		eta, err := calculator.CalculateRecoveryETA(state)
		require.NoError(t, err)

		// Tokens needed: 36000 - 30000 = 6000
		// Refill rate: 1200 / 60 = 20 tokens/sec
		// ETA: 6000 / 20 = 300 seconds
		assert.Equal(t, 300, eta)
	})

	t.Run("returns zero when at full capacity", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 36000,
			Capacity:        36000,
			RefillTokens:    1200,
			RefillPeriodSec: 60,
		}

		eta, err := calculator.CalculateRecoveryETA(state)
		require.NoError(t, err)
		assert.Equal(t, 0, eta)
	})

	t.Run("returns error for invalid refill rate", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 30000,
			Capacity:        36000,
			RefillTokens:    0, // Invalid
			RefillPeriodSec: 60,
		}

		_, err := calculator.CalculateRecoveryETA(state)
		assert.ErrorIs(t, err, ErrInvalidRefillRate)
	})
}

func TestCalculator_CalculateExhaustionProjection(t *testing.T) {
	calculator := NewCalculator()

	t.Run("calculates exhaustion projection correctly", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens:          30000,
			ConsumptionRatePerMinute: 300, // 300 tokens/min = 5 tokens/sec
		}

		projection, err := calculator.CalculateExhaustionProjection(state)
		require.NoError(t, err)

		// 30000 tokens / 5 tokens/sec = 6000 seconds
		assert.Equal(t, 6000, projection)
	})

	t.Run("returns zero when no consumption", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens:          30000,
			ConsumptionRatePerMinute: 0,
		}

		projection, err := calculator.CalculateExhaustionProjection(state)
		require.NoError(t, err)
		assert.Equal(t, 0, projection)
	})

	t.Run("returns zero when already exhausted", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens:          0,
			ConsumptionRatePerMinute: 300,
		}

		projection, err := calculator.CalculateExhaustionProjection(state)
		require.NoError(t, err)
		assert.Equal(t, 0, projection)
	})
}

func TestCalculator_EnrichStateWithMetrics(t *testing.T) {
	calculator := NewCalculator()

	t.Run("enriches state with all metrics", func(t *testing.T) {
		now := time.Now().UTC()
		fiveMinutesAgo := now.Add(-5 * time.Minute)

		previousState := &PolicyState{
			AvailableTokens:   30000,
			ResponseTimestamp: fiveMinutesAgo,
		}

		currentState := &PolicyState{
			AvailableTokens:   29000,
			Capacity:          36000,
			RefillTokens:      1200,
			RefillPeriodSec:   60,
			ResponseTimestamp: now,
		}

		err := calculator.EnrichStateWithMetrics(currentState, previousState)
		require.NoError(t, err)

		// Consumption rate should be calculated
		assert.Greater(t, currentState.ConsumptionRatePerMinute, 0.0)

		// Recovery ETA should be calculated
		assert.Greater(t, currentState.RecoveryETASeconds, 0)

		// Exhaustion projection should be calculated
		assert.Greater(t, currentState.ExhaustionProjectionSeconds, 0)

		// 404 rate should be set (placeholder = 0)
		assert.Equal(t, 0.0, currentState.Error404Rate)
	})
}
