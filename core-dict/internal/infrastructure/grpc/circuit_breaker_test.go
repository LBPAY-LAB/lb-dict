package grpc

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_ClosedState(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold:     3,
		Timeout:       1 * time.Second,
		HalfOpenTests: 1,
	})

	// Circuit should start in closed state
	if cb.GetState() != StateClosed {
		t.Errorf("Expected initial state CLOSED, got %s", cb.GetState())
	}

	// Successful call should keep circuit closed
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state CLOSED after success, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_OpenState(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold:     3,
		Timeout:       100 * time.Millisecond,
		HalfOpenTests: 1,
	})

	testErr := errors.New("test error")

	// Trigger threshold failures
	for i := 0; i < 3; i++ {
		cb.Call(func() error { return testErr })
	}

	// Circuit should now be open
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state OPEN after %d failures, got %s", 3, cb.GetState())
	}

	// Calls should fail immediately with ErrCircuitOpen
	err := cb.Call(func() error { return nil })
	if err != ErrCircuitOpen {
		t.Errorf("Expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenState(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold:     2,
		Timeout:       50 * time.Millisecond,
		HalfOpenTests: 1,
	})

	testErr := errors.New("test error")

	// Trigger circuit to open
	for i := 0; i < 2; i++ {
		cb.Call(func() error { return testErr })
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("Expected state OPEN, got %s", cb.GetState())
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	// Next call should transition to half-open
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("Expected success in half-open, got %v", err)
	}

	// Successful test should close the circuit
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state CLOSED after successful half-open test, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold:     2,
		Timeout:       50 * time.Millisecond,
		HalfOpenTests: 1,
	})

	testErr := errors.New("test error")

	// Trigger circuit to open
	for i := 0; i < 2; i++ {
		cb.Call(func() error { return testErr })
	}

	// Wait for timeout to allow half-open
	time.Sleep(60 * time.Millisecond)

	// Failed test should reopen the circuit
	err := cb.Call(func() error { return testErr })
	if err == nil {
		t.Error("Expected error, got nil")
	}

	// Circuit should be open again
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state OPEN after half-open failure, got %s", cb.GetState())
	}
}

func TestCircuitBreaker_Reset(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold:     2,
		Timeout:       1 * time.Second,
		HalfOpenTests: 1,
	})

	testErr := errors.New("test error")

	// Trigger circuit to open
	for i := 0; i < 2; i++ {
		cb.Call(func() error { return testErr })
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("Expected state OPEN, got %s", cb.GetState())
	}

	// Reset should close the circuit immediately
	cb.Reset()

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state CLOSED after reset, got %s", cb.GetState())
	}

	// Should accept calls again
	err := cb.Call(func() error { return nil })
	if err != nil {
		t.Errorf("Expected no error after reset, got %v", err)
	}
}

func TestCircuitBreaker_Metrics(t *testing.T) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold:     3,
		Timeout:       1 * time.Second,
		HalfOpenTests: 1,
	})

	metrics := cb.GetMetrics()

	if metrics["state"] != "CLOSED" {
		t.Errorf("Expected state CLOSED in metrics, got %v", metrics["state"])
	}

	if metrics["failure_count"] != 0 {
		t.Errorf("Expected failure_count 0, got %v", metrics["failure_count"])
	}

	if metrics["threshold"] != 3 {
		t.Errorf("Expected threshold 3, got %v", metrics["threshold"])
	}
}

func BenchmarkCircuitBreaker_ClosedState(b *testing.B) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold:     5,
		Timeout:       1 * time.Second,
		HalfOpenTests: 1,
	})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Call(func() error { return nil })
	}
}

func BenchmarkCircuitBreaker_OpenState(b *testing.B) {
	cb := NewCircuitBreaker(CircuitBreakerConfig{
		Threshold:     5,
		Timeout:       1 * time.Second,
		HalfOpenTests: 1,
	})

	// Trigger circuit to open
	testErr := errors.New("test error")
	for i := 0; i < 5; i++ {
		cb.Call(func() error { return testErr })
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cb.Call(func() error { return nil })
	}
}
