package grpc_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lbpay-lab/core-dict/internal/infrastructure/grpc"
)

func TestRetryPolicy_ExponentialBackoff(t *testing.T) {
	policy := grpc.NewRetryPolicy(grpc.RetryConfig{
		MaxRetries: 3,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		Multiplier: 2.0,
		Jitter:     0.0, // No jitter for predictable testing
	})

	// Test delay calculation for each attempt
	delays := []time.Duration{
		policy.NextDelay(0), // 100ms
		policy.NextDelay(1), // 200ms
		policy.NextDelay(2), // 400ms
		policy.NextDelay(3), // 800ms
	}

	assert.Equal(t, 100*time.Millisecond, delays[0])
	assert.Equal(t, 200*time.Millisecond, delays[1])
	assert.Equal(t, 400*time.Millisecond, delays[2])
	assert.Equal(t, 800*time.Millisecond, delays[3])
}

func TestRetryPolicy_MaxRetries(t *testing.T) {
	policy := grpc.NewRetryPolicy(grpc.RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		Multiplier: 2.0,
	})

	attempts := 0
	err := policy.Execute(context.Background(), func() error {
		attempts++
		return status.Error(codes.Unavailable, "service unavailable")
	})

	assert.Error(t, err)
	assert.Equal(t, 4, attempts) // Initial attempt + 3 retries
}

func TestRetryPolicy_SuccessOnRetry(t *testing.T) {
	policy := grpc.NewRetryPolicy(grpc.RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		Multiplier: 2.0,
	})

	attempts := 0
	err := policy.Execute(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			return status.Error(codes.Unavailable, "service unavailable")
		}
		return nil
	})

	assert.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestRetryPolicy_NonRetryableError(t *testing.T) {
	policy := grpc.NewRetryPolicy(grpc.RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
		MaxDelay:   100 * time.Millisecond,
		Multiplier: 2.0,
	})

	attempts := 0
	err := policy.Execute(context.Background(), func() error {
		attempts++
		return status.Error(codes.InvalidArgument, "bad request")
	})

	assert.Error(t, err)
	assert.Equal(t, 1, attempts) // Should not retry
}

func TestRetryPolicy_ContextCancellation(t *testing.T) {
	policy := grpc.NewRetryPolicy(grpc.RetryConfig{
		MaxRetries: 10,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   1 * time.Second,
		Multiplier: 2.0,
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	attempts := 0
	err := policy.Execute(ctx, func() error {
		attempts++
		return status.Error(codes.Unavailable, "unavailable")
	})

	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
	assert.LessOrEqual(t, attempts, 1) // Should stop quickly
}

func TestRetryPolicy_IsRetryable(t *testing.T) {
	policy := grpc.NewRetryPolicy(grpc.RetryConfig{
		MaxRetries: 3,
		BaseDelay:  10 * time.Millisecond,
	})

	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "circuit open",
			err:      grpc.ErrCircuitOpen,
			expected: false,
		},
		{
			name:     "unavailable",
			err:      status.Error(codes.Unavailable, "unavailable"),
			expected: true,
		},
		{
			name:     "deadline exceeded",
			err:      status.Error(codes.DeadlineExceeded, "timeout"),
			expected: true,
		},
		{
			name:     "resource exhausted",
			err:      status.Error(codes.ResourceExhausted, "too many requests"),
			expected: true,
		},
		{
			name:     "invalid argument",
			err:      status.Error(codes.InvalidArgument, "bad request"),
			expected: false,
		},
		{
			name:     "not found",
			err:      status.Error(codes.NotFound, "not found"),
			expected: false,
		},
		{
			name:     "non-grpc error",
			err:      errors.New("generic error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := policy.IsRetryable(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRetryPolicy_MaxDelayCap(t *testing.T) {
	policy := grpc.NewRetryPolicy(grpc.RetryConfig{
		MaxRetries: 10,
		BaseDelay:  100 * time.Millisecond,
		MaxDelay:   500 * time.Millisecond,
		Multiplier: 2.0,
		Jitter:     0.0,
	})

	// After many attempts, delay should be capped at MaxDelay
	delay := policy.NextDelay(10)
	assert.LessOrEqual(t, delay, 500*time.Millisecond)
}
