package interfaces

import (
	"context"
	"time"
)

// CircuitBreakerState represents the state of a circuit breaker
type CircuitBreakerState string

const (
	StateClosed   CircuitBreakerState = "CLOSED"
	StateOpen     CircuitBreakerState = "OPEN"
	StateHalfOpen CircuitBreakerState = "HALF_OPEN"
)

// CircuitBreakerStats represents circuit breaker statistics
type CircuitBreakerStats struct {
	State              CircuitBreakerState
	Requests           uint32
	TotalSuccesses     uint32
	TotalFailures      uint32
	ConsecutiveSuccess uint32
	ConsecutiveFailure uint32
	LastStateChange    time.Time
}

// CircuitBreaker defines the interface for circuit breaker pattern
type CircuitBreaker interface {
	// Execute executes a function with circuit breaker protection
	Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error)

	// GetState returns the current state of the circuit breaker
	GetState() CircuitBreakerState

	// GetStats returns circuit breaker statistics
	GetStats() CircuitBreakerStats

	// Reset resets the circuit breaker to closed state
	Reset()

	// IsOpen returns true if the circuit is open
	IsOpen() bool
}
