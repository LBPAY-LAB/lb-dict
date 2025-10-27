package circuitbreaker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/lbpay-lab/conn-bridge/internal/domain/interfaces"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/sirupsen/logrus"
	"github.com/sony/gobreaker"
)

var (
	// Prometheus metrics
	circuitBreakerState = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "circuit_breaker_state",
			Help: "Current state of the circuit breaker (0=closed, 1=open, 2=half-open)",
		},
		[]string{"name"},
	)

	circuitBreakerFailuresTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_failures_total",
			Help: "Total number of circuit breaker failures",
		},
		[]string{"name"},
	)

	circuitBreakerSuccessesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_successes_total",
			Help: "Total number of circuit breaker successes",
		},
		[]string{"name"},
	)

	circuitBreakerRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_requests_total",
			Help: "Total number of circuit breaker requests",
		},
		[]string{"name", "state"},
	)

	circuitBreakerStateTransitionsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "circuit_breaker_state_transitions_total",
			Help: "Total number of circuit breaker state transitions",
		},
		[]string{"name", "from", "to"},
	)
)

// GoBreakerAdapter adapts gobreaker to our CircuitBreaker interface
type GoBreakerAdapter struct {
	breaker           *gobreaker.CircuitBreaker
	name              string
	logger            *logrus.Logger
	mu                sync.RWMutex
	lastStateChange   time.Time
	metricsRegistered bool
}

// Config holds the configuration for the circuit breaker
type Config struct {
	// Name of the circuit breaker (for logging and metrics)
	Name string

	// MaxRequests is the maximum number of requests allowed to pass through
	// when the circuit breaker is half-open (default: 3)
	MaxRequests uint32

	// Interval is the cyclic period of the closed state for the circuit breaker
	// to clear the internal counts. If Interval is 0, the circuit breaker doesn't
	// clear internal counts during the closed state. (default: 60s)
	Interval time.Duration

	// Timeout is the period of the open state, after which the state becomes half-open
	// (default: 60s)
	Timeout time.Duration

	// MaxFailures is the maximum number of consecutive failures before opening the circuit
	// (default: 5)
	MaxFailures uint32

	// ReadyToTrip is called with a copy of Counts whenever a request fails in the closed state.
	// If ReadyToTrip returns true, the circuit breaker will be placed into the open state.
	ReadyToTrip func(counts gobreaker.Counts) bool

	// OnStateChange is called whenever the state of the circuit breaker changes
	OnStateChange func(name string, from gobreaker.State, to gobreaker.State)

	// Logger for circuit breaker operations
	Logger *logrus.Logger
}

// NewGoBreakerAdapter creates a new circuit breaker adapter with the specified configuration
func NewGoBreakerAdapter(config *Config) interfaces.CircuitBreaker {
	if config == nil {
		config = DefaultConfig("default-circuit-breaker")
	}

	// Set defaults
	if config.Name == "" {
		config.Name = "circuit-breaker"
	}
	if config.MaxRequests == 0 {
		config.MaxRequests = 3
	}
	if config.Interval == 0 {
		config.Interval = 60 * time.Second
	}
	if config.Timeout == 0 {
		config.Timeout = 60 * time.Second
	}
	if config.MaxFailures == 0 {
		config.MaxFailures = 5
	}
	if config.Logger == nil {
		config.Logger = logrus.New()
	}

	adapter := &GoBreakerAdapter{
		name:            config.Name,
		logger:          config.Logger,
		lastStateChange: time.Now(),
	}

	// Set up ReadyToTrip function
	readyToTrip := config.ReadyToTrip
	if readyToTrip == nil {
		maxFailures := config.MaxFailures
		readyToTrip = func(counts gobreaker.Counts) bool {
			// Trip the circuit after MaxFailures consecutive failures
			return counts.ConsecutiveFailures >= maxFailures
		}
	}

	// Wrap ReadyToTrip to track failures
	wrappedReadyToTrip := func(counts gobreaker.Counts) bool {
		result := readyToTrip(counts)
		if result {
			adapter.logger.WithFields(logrus.Fields{
				"name":                 adapter.name,
				"consecutive_failures": counts.ConsecutiveFailures,
				"total_failures":       counts.TotalFailures,
				"requests":             counts.Requests,
			}).Warn("Circuit breaker ready to trip")
		}
		return result
	}

	// Set up OnStateChange callback
	onStateChange := config.OnStateChange
	wrappedOnStateChange := func(name string, from gobreaker.State, to gobreaker.State) {
		adapter.mu.Lock()
		adapter.lastStateChange = time.Now()
		adapter.mu.Unlock()

		// Log state transition
		adapter.logger.WithFields(logrus.Fields{
			"name": name,
			"from": from.String(),
			"to":   to.String(),
		}).Info("Circuit breaker state transition")

		// Update Prometheus metrics
		circuitBreakerStateTransitionsTotal.WithLabelValues(
			name,
			from.String(),
			to.String(),
		).Inc()

		// Update state gauge
		stateValue := adapter.stateToMetricValue(to)
		circuitBreakerState.WithLabelValues(name).Set(stateValue)

		// Call custom callback if provided
		if onStateChange != nil {
			onStateChange(name, from, to)
		}
	}

	// Create gobreaker settings
	settings := gobreaker.Settings{
		Name:          config.Name,
		MaxRequests:   config.MaxRequests,
		Interval:      config.Interval,
		Timeout:       config.Timeout,
		ReadyToTrip:   wrappedReadyToTrip,
		OnStateChange: wrappedOnStateChange,
	}

	adapter.breaker = gobreaker.NewCircuitBreaker(settings)

	// Initialize metrics
	circuitBreakerState.WithLabelValues(config.Name).Set(0) // Start in closed state

	adapter.logger.WithFields(logrus.Fields{
		"name":         config.Name,
		"max_requests": config.MaxRequests,
		"interval":     config.Interval,
		"timeout":      config.Timeout,
		"max_failures": config.MaxFailures,
	}).Info("Circuit breaker initialized")

	return adapter
}

// Execute executes a function with circuit breaker protection
func (cb *GoBreakerAdapter) Execute(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	// Track request
	state := cb.breaker.State()
	circuitBreakerRequestsTotal.WithLabelValues(cb.name, state.String()).Inc()

	// Check if circuit is open and fail fast
	if state == gobreaker.StateOpen {
		cb.logger.WithFields(logrus.Fields{
			"name":  cb.name,
			"state": "open",
		}).Warn("Circuit breaker is open, rejecting request")
		return nil, fmt.Errorf("circuit breaker is open")
	}

	// Execute function through circuit breaker
	result, err := cb.breaker.Execute(func() (interface{}, error) {
		return fn()
	})

	// Track success or failure
	if err != nil {
		circuitBreakerFailuresTotal.WithLabelValues(cb.name).Inc()
		cb.logger.WithFields(logrus.Fields{
			"name":  cb.name,
			"error": err.Error(),
		}).Debug("Circuit breaker request failed")
	} else {
		circuitBreakerSuccessesTotal.WithLabelValues(cb.name).Inc()
		cb.logger.WithFields(logrus.Fields{
			"name": cb.name,
		}).Debug("Circuit breaker request succeeded")
	}

	return result, err
}

// GetState returns the current state of the circuit breaker
func (cb *GoBreakerAdapter) GetState() interfaces.CircuitBreakerState {
	state := cb.breaker.State()
	switch state {
	case gobreaker.StateClosed:
		return interfaces.StateClosed
	case gobreaker.StateOpen:
		return interfaces.StateOpen
	case gobreaker.StateHalfOpen:
		return interfaces.StateHalfOpen
	default:
		return interfaces.StateClosed
	}
}

// GetStats returns circuit breaker statistics
func (cb *GoBreakerAdapter) GetStats() interfaces.CircuitBreakerStats {
	counts := cb.breaker.Counts()

	cb.mu.RLock()
	lastStateChange := cb.lastStateChange
	cb.mu.RUnlock()

	return interfaces.CircuitBreakerStats{
		State:              cb.GetState(),
		Requests:           counts.Requests,
		TotalSuccesses:     counts.TotalSuccesses,
		TotalFailures:      counts.TotalFailures,
		ConsecutiveSuccess: counts.ConsecutiveSuccesses,
		ConsecutiveFailure: counts.ConsecutiveFailures,
		LastStateChange:    lastStateChange,
	}
}

// Reset resets the circuit breaker to closed state
func (cb *GoBreakerAdapter) Reset() {
	// Note: gobreaker v0.5.0 doesn't expose a public reset method
	// This is a limitation of the current version
	// In production, you might want to:
	// 1. Upgrade to gobreaker v2 which has a Reset() method
	// 2. Use a custom wrapper that recreates the breaker
	// 3. Wait for the timeout to naturally transition states

	cb.logger.WithFields(logrus.Fields{
		"name": cb.name,
	}).Warn("Reset called but not implemented in gobreaker v0.5.0")
}

// IsOpen returns true if the circuit is open
func (cb *GoBreakerAdapter) IsOpen() bool {
	return cb.breaker.State() == gobreaker.StateOpen
}

// stateToMetricValue converts gobreaker.State to a numeric value for Prometheus
func (cb *GoBreakerAdapter) stateToMetricValue(state gobreaker.State) float64 {
	switch state {
	case gobreaker.StateClosed:
		return 0
	case gobreaker.StateOpen:
		return 1
	case gobreaker.StateHalfOpen:
		return 2
	default:
		return 0
	}
}

// DefaultConfig returns a default circuit breaker configuration
// matching the requirements: 5 max failures, 60s timeout, 3 half-open requests
func DefaultConfig(name string) *Config {
	return &Config{
		Name:        name,
		MaxRequests: 3,  // Half-open max requests
		Interval:    60 * time.Second,
		Timeout:     60 * time.Second, // Before attempting half-open
		MaxFailures: 5,  // Consecutive errors before opening
		Logger:      logrus.New(),
	}
}

// NewConfigFromEnv creates a circuit breaker configuration from environment variables
func NewConfigFromEnv(name string, maxFailures uint32, timeout time.Duration, maxRequests uint32, logger *logrus.Logger) *Config {
	if logger == nil {
		logger = logrus.New()
	}

	return &Config{
		Name:        name,
		MaxRequests: maxRequests,
		Interval:    60 * time.Second,
		Timeout:     timeout,
		MaxFailures: maxFailures,
		Logger:      logger,
	}
}

// ExecuteWithBreaker is a convenience wrapper for executing functions with error handling
func (cb *GoBreakerAdapter) ExecuteWithBreaker(fn func() error) error {
	_, err := cb.Execute(context.Background(), func() (interface{}, error) {
		return nil, fn()
	})
	return err
}
