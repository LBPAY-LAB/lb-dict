package grpc

import (
	"fmt"
	"sync"
	"time"
)

// State represents the circuit breaker state
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

func (s State) String() string {
	switch s {
	case StateClosed:
		return "CLOSED"
	case StateOpen:
		return "OPEN"
	case StateHalfOpen:
		return "HALF_OPEN"
	default:
		return "UNKNOWN"
	}
}

// CircuitBreaker implements the circuit breaker pattern to prevent cascading failures
type CircuitBreaker struct {
	state         State
	failureCount  int
	successCount  int
	threshold     int           // Number of consecutive failures to open circuit
	timeout       time.Duration // Duration to wait before attempting half-open
	halfOpenTests int           // Number of successful tests needed in half-open to close
	lastFailTime  time.Time
	mu            sync.Mutex
	onStateChange func(from, to State) // Callback for state changes
}

// CircuitBreakerConfig holds configuration for the circuit breaker
type CircuitBreakerConfig struct {
	Threshold     int           // Default: 5
	Timeout       time.Duration // Default: 60s
	HalfOpenTests int           // Default: 1
	OnStateChange func(from, to State)
}

// NewCircuitBreaker creates a new circuit breaker with the given configuration
func NewCircuitBreaker(config CircuitBreakerConfig) *CircuitBreaker {
	if config.Threshold <= 0 {
		config.Threshold = 5
	}
	if config.Timeout <= 0 {
		config.Timeout = 60 * time.Second
	}
	if config.HalfOpenTests <= 0 {
		config.HalfOpenTests = 1
	}

	return &CircuitBreaker{
		state:         StateClosed,
		threshold:     config.Threshold,
		timeout:       config.Timeout,
		halfOpenTests: config.HalfOpenTests,
		onStateChange: config.OnStateChange,
	}
}

// Call executes the given function with circuit breaker protection
func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Check if we should attempt recovery from open state
	if cb.state == StateOpen {
		if time.Since(cb.lastFailTime) > cb.timeout {
			cb.setState(StateHalfOpen)
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}

	// In half-open state, allow limited requests
	if cb.state == StateHalfOpen {
		cb.mu.Unlock()
		err := fn()
		cb.mu.Lock()
		defer cb.mu.Unlock()

		if err != nil {
			cb.recordFailureLocked()
			return err
		}

		cb.recordSuccessLocked()
		return nil
	}

	cb.mu.Unlock()

	// Normal operation (closed state)
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.recordFailureLocked()
		return err
	}

	cb.recordSuccessLocked()
	return nil
}

// RecordSuccess manually records a successful operation
func (cb *CircuitBreaker) RecordSuccess() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.recordSuccessLocked()
}

// RecordFailure manually records a failed operation
func (cb *CircuitBreaker) RecordFailure() {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	cb.recordFailureLocked()
}

// GetState returns the current state of the circuit breaker
func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// GetMetrics returns current metrics of the circuit breaker
func (cb *CircuitBreaker) GetMetrics() map[string]interface{} {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return map[string]interface{}{
		"state":         cb.state.String(),
		"failure_count": cb.failureCount,
		"success_count": cb.successCount,
		"threshold":     cb.threshold,
		"last_fail":     cb.lastFailTime,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *CircuitBreaker) Reset() {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	oldState := cb.state
	cb.state = StateClosed
	cb.failureCount = 0
	cb.successCount = 0
	cb.lastFailTime = time.Time{}

	if oldState != StateClosed && cb.onStateChange != nil {
		cb.onStateChange(oldState, StateClosed)
	}
}

// recordSuccessLocked handles success recording (must be called with lock held)
func (cb *CircuitBreaker) recordSuccessLocked() {
	cb.failureCount = 0

	if cb.state == StateHalfOpen {
		cb.successCount++
		if cb.successCount >= cb.halfOpenTests {
			cb.setState(StateClosed)
		}
	}
}

// recordFailureLocked handles failure recording (must be called with lock held)
func (cb *CircuitBreaker) recordFailureLocked() {
	cb.failureCount++
	cb.successCount = 0
	cb.lastFailTime = time.Now()

	if cb.state == StateClosed && cb.failureCount >= cb.threshold {
		cb.setState(StateOpen)
	} else if cb.state == StateHalfOpen {
		// Any failure in half-open immediately reopens the circuit
		cb.setState(StateOpen)
	}
}

// setState changes the circuit breaker state and triggers callback
func (cb *CircuitBreaker) setState(newState State) {
	if cb.state == newState {
		return
	}

	oldState := cb.state
	cb.state = newState

	// Reset counters on state change
	if newState == StateClosed {
		cb.failureCount = 0
		cb.successCount = 0
	} else if newState == StateHalfOpen {
		cb.successCount = 0
	}

	// Trigger callback if set
	if cb.onStateChange != nil {
		// Call callback without holding lock to prevent deadlocks
		go cb.onStateChange(oldState, newState)
	}
}

// String returns a string representation of the circuit breaker
func (cb *CircuitBreaker) String() string {
	cb.mu.Lock()
	defer cb.mu.Unlock()

	return fmt.Sprintf(
		"CircuitBreaker{state=%s, failures=%d/%d, successes=%d, lastFail=%v}",
		cb.state.String(),
		cb.failureCount,
		cb.threshold,
		cb.successCount,
		cb.lastFailTime.Format(time.RFC3339),
	)
}
