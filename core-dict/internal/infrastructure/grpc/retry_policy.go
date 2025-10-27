package grpc

import (
	"context"
	"math"
	"math/rand"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RetryPolicy defines the retry behavior for gRPC calls
type RetryPolicy struct {
	maxRetries int
	baseDelay  time.Duration
	maxDelay   time.Duration
	multiplier float64
	jitter     float64 // Jitter percentage (0.0 - 1.0)
}

// RetryConfig holds configuration for retry policy
type RetryConfig struct {
	MaxRetries int           // Default: 3
	BaseDelay  time.Duration // Default: 100ms
	MaxDelay   time.Duration // Default: 2s
	Multiplier float64       // Default: 2.0
	Jitter     float64       // Default: 0.2 (20%)
}

// NewRetryPolicy creates a new retry policy with the given configuration
func NewRetryPolicy(config RetryConfig) *RetryPolicy {
	if config.MaxRetries <= 0 {
		config.MaxRetries = 3
	}
	if config.BaseDelay <= 0 {
		config.BaseDelay = 100 * time.Millisecond
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 2 * time.Second
	}
	if config.Multiplier <= 0 {
		config.Multiplier = 2.0
	}
	if config.Jitter <= 0 {
		config.Jitter = 0.2
	}

	return &RetryPolicy{
		maxRetries: config.MaxRetries,
		baseDelay:  config.BaseDelay,
		maxDelay:   config.MaxDelay,
		multiplier: config.Multiplier,
		jitter:     config.Jitter,
	}
}

// Execute runs the given function with retry logic
func (rp *RetryPolicy) Execute(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt <= rp.maxRetries; attempt++ {
		// Check context cancellation before each attempt
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Don't retry if error is not retryable
		if !rp.IsRetryable(err) {
			return err
		}

		// Don't retry on last attempt
		if attempt >= rp.maxRetries {
			break
		}

		// Calculate delay with exponential backoff and jitter
		delay := rp.NextDelay(attempt)

		// Wait before next retry (with context cancellation support)
		select {
		case <-time.After(delay):
			// Continue to next attempt
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	return lastErr
}

// IsRetryable determines if an error should trigger a retry
func (rp *RetryPolicy) IsRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for circuit breaker open (not retryable)
	if err == ErrCircuitOpen {
		return false
	}

	// Extract gRPC status
	st, ok := status.FromError(err)
	if !ok {
		// Non-gRPC errors are not retryable by default
		return false
	}

	// Retry on specific gRPC codes
	switch st.Code() {
	case codes.Unavailable:
		return true
	case codes.DeadlineExceeded:
		return true
	case codes.ResourceExhausted:
		return true
	case codes.Aborted:
		return true
	case codes.Internal:
		// Internal errors may be transient
		return true
	default:
		return false
	}
}

// NextDelay calculates the delay before the next retry attempt
func (rp *RetryPolicy) NextDelay(attempt int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}

	// Calculate exponential backoff: baseDelay * multiplier^attempt
	delay := float64(rp.baseDelay) * math.Pow(rp.multiplier, float64(attempt))

	// Cap at max delay
	if delay > float64(rp.maxDelay) {
		delay = float64(rp.maxDelay)
	}

	// Add jitter: ±jitter%
	jitterAmount := delay * rp.jitter
	jitterRange := jitterAmount * 2
	jitterOffset := (rand.Float64() * jitterRange) - jitterAmount
	delay += jitterOffset

	// Ensure non-negative
	if delay < 0 {
		delay = 0
	}

	return time.Duration(delay)
}

// GetConfig returns the retry policy configuration
func (rp *RetryPolicy) GetConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: rp.maxRetries,
		BaseDelay:  rp.baseDelay,
		MaxDelay:   rp.maxDelay,
		Multiplier: rp.multiplier,
		Jitter:     rp.jitter,
	}
}

// CalculateMaxDuration returns the maximum possible duration for all retries
func (rp *RetryPolicy) CalculateMaxDuration() time.Duration {
	var totalDelay time.Duration

	for attempt := 0; attempt < rp.maxRetries; attempt++ {
		delay := float64(rp.baseDelay) * math.Pow(rp.multiplier, float64(attempt))
		if delay > float64(rp.maxDelay) {
			delay = float64(rp.maxDelay)
		}
		// Add max jitter
		delay += delay * rp.jitter
		totalDelay += time.Duration(delay)
	}

	return totalDelay
}
