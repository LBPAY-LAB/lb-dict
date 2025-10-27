package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
	"github.com/stretchr/testify/assert"
)

func TestNewGoBreakerAdapter(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name:   "Default config",
			config: DefaultConfig("test-breaker"),
		},
		{
			name: "Custom config",
			config: &Config{
				Name:        "custom-breaker",
				MaxRequests: 5,
				Interval:    30 * time.Second,
				Timeout:     45 * time.Second,
				MaxFailures: 10,
			},
		},
		{
			name:   "Nil config should use defaults",
			config: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adapter := NewGoBreakerAdapter(tt.config)
			assert.NotNil(t, adapter)

			// Check initial state is closed
			assert.Equal(t, "CLOSED", string(adapter.GetState()))
		})
	}
}

func TestGoBreakerAdapter_Execute_Success(t *testing.T) {
	adapter := NewGoBreakerAdapter(DefaultConfig("test-breaker"))

	result, err := adapter.Execute(context.Background(), func() (interface{}, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)

	// Verify stats
	stats := adapter.GetStats()
	assert.Equal(t, uint32(1), stats.TotalSuccesses)
	assert.Equal(t, uint32(0), stats.TotalFailures)
}

func TestGoBreakerAdapter_Execute_Failure(t *testing.T) {
	adapter := NewGoBreakerAdapter(DefaultConfig("test-breaker"))

	expectedErr := errors.New("test error")
	result, err := adapter.Execute(context.Background(), func() (interface{}, error) {
		return nil, expectedErr
	})

	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify stats
	stats := adapter.GetStats()
	assert.Equal(t, uint32(0), stats.TotalSuccesses)
	assert.Equal(t, uint32(1), stats.TotalFailures)
}

func TestGoBreakerAdapter_CircuitOpens(t *testing.T) {
	config := &Config{
		Name:        "test-breaker",
		MaxRequests: 3,
		Interval:    60 * time.Second,
		Timeout:     1 * time.Second, // Short timeout for testing
		MaxFailures: 3,                // Open after 3 failures
	}

	adapter := NewGoBreakerAdapter(config)

	// Initial state should be closed
	assert.False(t, adapter.IsOpen())

	// Generate failures to open the circuit
	for i := 0; i < 3; i++ {
		_, err := adapter.Execute(context.Background(), func() (interface{}, error) {
			return nil, errors.New("failure")
		})
		assert.Error(t, err)
	}

	// Circuit should now be open
	assert.True(t, adapter.IsOpen())

	// Next request should fail fast without executing function
	executed := false
	_, err := adapter.Execute(context.Background(), func() (interface{}, error) {
		executed = true
		return "should not execute", nil
	})

	assert.Error(t, err)
	assert.False(t, executed, "Function should not be executed when circuit is open")
}

func TestGoBreakerAdapter_HalfOpen(t *testing.T) {
	config := &Config{
		Name:        "test-breaker",
		MaxRequests: 2,
		Interval:    60 * time.Second,
		Timeout:     100 * time.Millisecond, // Very short timeout for testing
		MaxFailures: 2,
	}

	adapter := NewGoBreakerAdapter(config)

	// Generate failures to open the circuit
	for i := 0; i < 2; i++ {
		_, _ = adapter.Execute(context.Background(), func() (interface{}, error) {
			return nil, errors.New("failure")
		})
	}

	// Circuit should be open
	assert.True(t, adapter.IsOpen())

	// Wait for timeout to transition to half-open
	time.Sleep(150 * time.Millisecond)

	// First request should be allowed (half-open)
	result, err := adapter.Execute(context.Background(), func() (interface{}, error) {
		return "success", nil
	})

	assert.NoError(t, err)
	assert.Equal(t, "success", result)
}

func TestGoBreakerAdapter_GetStats(t *testing.T) {
	adapter := NewGoBreakerAdapter(DefaultConfig("test-breaker"))

	// Execute some successful requests
	for i := 0; i < 3; i++ {
		_, _ = adapter.Execute(context.Background(), func() (interface{}, error) {
			return "success", nil
		})
	}

	// Execute some failed requests
	for i := 0; i < 2; i++ {
		_, _ = adapter.Execute(context.Background(), func() (interface{}, error) {
			return nil, errors.New("failure")
		})
	}

	stats := adapter.GetStats()
	assert.Equal(t, uint32(3), stats.TotalSuccesses)
	assert.Equal(t, uint32(2), stats.TotalFailures)
	assert.Equal(t, uint32(5), stats.Requests)
	assert.Equal(t, uint32(2), stats.ConsecutiveFailure)
}

func TestGoBreakerAdapter_StateTransitions(t *testing.T) {
	stateChanges := make([]string, 0)

	config := &Config{
		Name:        "test-breaker",
		MaxRequests: 1,
		Interval:    60 * time.Second,
		Timeout:     100 * time.Millisecond,
		MaxFailures: 2,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			stateChanges = append(stateChanges, to.String())
		},
	}

	adapter := NewGoBreakerAdapter(config)

	// Generate failures to trigger state change
	for i := 0; i < 2; i++ {
		_, _ = adapter.Execute(context.Background(), func() (interface{}, error) {
			return nil, errors.New("failure")
		})
	}

	// Should have transitioned to open
	assert.Contains(t, stateChanges, "open")
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig("test-breaker")

	assert.Equal(t, "test-breaker", config.Name)
	assert.Equal(t, uint32(3), config.MaxRequests)
	assert.Equal(t, 60*time.Second, config.Interval)
	assert.Equal(t, 60*time.Second, config.Timeout)
	assert.Equal(t, uint32(5), config.MaxFailures)
}

func TestNewConfigFromEnv(t *testing.T) {
	logger := logrus.New()
	config := NewConfigFromEnv("test-breaker", 10, 30*time.Second, 5, logger)

	assert.Equal(t, "test-breaker", config.Name)
	assert.Equal(t, uint32(10), config.MaxFailures)
	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, uint32(5), config.MaxRequests)
	assert.Equal(t, 60*time.Second, config.Interval)
	assert.NotNil(t, config.Logger)
}

func TestExecuteWithBreaker(t *testing.T) {
	adapter := NewGoBreakerAdapter(DefaultConfig("test-breaker")).(*GoBreakerAdapter)

	// Test successful execution
	err := adapter.ExecuteWithBreaker(func() error {
		return nil
	})
	assert.NoError(t, err)

	// Test failed execution
	expectedErr := errors.New("test error")
	err = adapter.ExecuteWithBreaker(func() error {
		return expectedErr
	})
	assert.Error(t, err)
}
