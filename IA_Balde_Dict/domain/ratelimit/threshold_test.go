package ratelimit

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestThresholdAnalyzer_AnalyzeState(t *testing.T) {
	analyzer := NewThresholdAnalyzer()

	t.Run("returns OK when tokens > 20%", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 9000,  // 25% of capacity
			Capacity:        36000,
		}

		status := analyzer.AnalyzeState(state)
		assert.Equal(t, StatusOK, status)
	})

	t.Run("returns WARNING when tokens <= 20%", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 7200,  // Exactly 20% of capacity
			Capacity:        36000,
		}

		status := analyzer.AnalyzeState(state)
		assert.Equal(t, StatusWarning, status)
	})

	t.Run("returns CRITICAL when tokens <= 10%", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 3600,  // Exactly 10% of capacity
			Capacity:        36000,
		}

		status := analyzer.AnalyzeState(state)
		assert.Equal(t, StatusCritical, status)
	})

	t.Run("returns CRITICAL for very low tokens", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 100,   // ~0.3% of capacity
			Capacity:        36000,
		}

		status := analyzer.AnalyzeState(state)
		assert.Equal(t, StatusCritical, status)
	})
}

func TestThresholdAnalyzer_ShouldCreateAlert(t *testing.T) {
	analyzer := NewThresholdAnalyzer()

	t.Run("returns true for WARNING state", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 7000,  // ~19% (WARNING)
			Capacity:        36000,
		}

		assert.True(t, analyzer.ShouldCreateAlert(state))
	})

	t.Run("returns true for CRITICAL state", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 3000,  // ~8% (CRITICAL)
			Capacity:        36000,
		}

		assert.True(t, analyzer.ShouldCreateAlert(state))
	})

	t.Run("returns false for OK state", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 30000, // ~83% (OK)
			Capacity:        36000,
		}

		assert.False(t, analyzer.ShouldCreateAlert(state))
	})
}

func TestThresholdAnalyzer_GetAlertSeverity(t *testing.T) {
	analyzer := NewThresholdAnalyzer()

	t.Run("returns WARNING severity", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 6000,  // ~17% (WARNING)
			Capacity:        36000,
		}

		severity := analyzer.GetAlertSeverity(state)
		assert.Equal(t, SeverityWarning, severity)
	})

	t.Run("returns CRITICAL severity", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 2000,  // ~5.5% (CRITICAL)
			Capacity:        36000,
		}

		severity := analyzer.GetAlertSeverity(state)
		assert.Equal(t, SeverityCritical, severity)
	})

	t.Run("returns empty for OK state", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 30000, // ~83% (OK)
			Capacity:        36000,
		}

		severity := analyzer.GetAlertSeverity(state)
		assert.Empty(t, severity)
	})
}

func TestThresholdAnalyzer_ShouldResolveAlert(t *testing.T) {
	analyzer := NewThresholdAnalyzer()

	t.Run("resolves WARNING when utilization < 80%", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 8000,  // ~22% remaining = 78% utilized
			Capacity:        36000,
		}

		assert.True(t, analyzer.ShouldResolveAlert(SeverityWarning, state))
	})

	t.Run("does not resolve WARNING when utilization >= 80%", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 7000,  // ~19% remaining = 81% utilized
			Capacity:        36000,
		}

		assert.False(t, analyzer.ShouldResolveAlert(SeverityWarning, state))
	})

	t.Run("resolves CRITICAL when utilization < 90%", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 4000,  // ~11% remaining = 89% utilized
			Capacity:        36000,
		}

		assert.True(t, analyzer.ShouldResolveAlert(SeverityCritical, state))
	})

	t.Run("does not resolve CRITICAL when utilization >= 90%", func(t *testing.T) {
		state := &PolicyState{
			AvailableTokens: 3000,  // ~8% remaining = 92% utilized
			Capacity:        36000,
		}

		assert.False(t, analyzer.ShouldResolveAlert(SeverityCritical, state))
	})
}

func TestThresholdAnalyzer_CalculateThresholdTokens(t *testing.T) {
	analyzer := NewThresholdAnalyzer()

	t.Run("calculates WARNING threshold tokens", func(t *testing.T) {
		tokens := analyzer.GetWarningThresholdTokens(36000)
		assert.Equal(t, 7200, tokens) // 20% of 36000
	})

	t.Run("calculates CRITICAL threshold tokens", func(t *testing.T) {
		tokens := analyzer.GetCriticalThresholdTokens(36000)
		assert.Equal(t, 3600, tokens) // 10% of 36000
	})
}

func TestAlert_Creation(t *testing.T) {
	t.Run("creates WARNING alert successfully", func(t *testing.T) {
		state := &PolicyState{
			EndpointID:      "ENTRIES_WRITE",
			AvailableTokens: 6000,  // ~17% (WARNING)
			Capacity:        36000,
			PSPCategory:     "A",
			CreatedAt:       time.Now().UTC(),
		}

		alert, err := NewAlert("ENTRIES_WRITE", SeverityWarning, state)
		require.NoError(t, err)
		assert.Equal(t, SeverityWarning, alert.Severity)
		assert.Equal(t, 20, alert.ThresholdPercent)
		assert.Greater(t, alert.UtilizationPercent, 80.0)
	})

	t.Run("creates CRITICAL alert successfully", func(t *testing.T) {
		state := &PolicyState{
			EndpointID:      "ENTRIES_WRITE",
			AvailableTokens: 2000,  // ~5.5% (CRITICAL)
			Capacity:        36000,
			PSPCategory:     "A",
			CreatedAt:       time.Now().UTC(),
		}

		alert, err := NewAlert("ENTRIES_WRITE", SeverityCritical, state)
		require.NoError(t, err)
		assert.Equal(t, SeverityCritical, alert.Severity)
		assert.Equal(t, 10, alert.ThresholdPercent)
		assert.Greater(t, alert.UtilizationPercent, 90.0)
	})

	t.Run("returns error for invalid severity match", func(t *testing.T) {
		state := &PolicyState{
			EndpointID:      "ENTRIES_WRITE",
			AvailableTokens: 30000, // ~83% (OK - not WARNING)
			Capacity:        36000,
		}

		_, err := NewAlert("ENTRIES_WRITE", SeverityWarning, state)
		assert.Error(t, err)
	})
}
