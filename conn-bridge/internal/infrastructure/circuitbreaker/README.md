# Circuit Breaker Implementation

## Overview

This package implements the Circuit Breaker pattern using `github.com/sony/gobreaker` to protect against cascading failures when calling the Bacen API.

## Architecture

### Components

1. **GoBreakerAdapter**: Main adapter that implements the `interfaces.CircuitBreaker` interface
2. **CircuitBreakerClient**: Wrapper for Bacen HTTP client that integrates circuit breaker protection

### Circuit Breaker States

- **CLOSED**: Normal operation, all requests pass through
- **OPEN**: Fast-fail mode, requests are rejected immediately without calling Bacen
- **HALF_OPEN**: Recovery testing mode, limited requests are allowed to test if service recovered

## Configuration

### Default Configuration

```go
MaxFailures: 5                  // 5 consecutive failures trigger circuit open
Timeout:     60 * time.Second   // Wait 60s before attempting half-open
MaxRequests: 3                  // Allow 3 requests in half-open state
Interval:    60 * time.Second   // Clear internal counts every 60s
```

### Environment Variables

Configure circuit breaker behavior via environment variables:

- `CONN_BRIDGE_CIRCUIT_BREAKER_NAME`: Circuit breaker name (default: "bacen-circuit-breaker")
- `CONN_BRIDGE_CIRCUIT_BREAKER_MAX_FAILURES`: Max consecutive failures (default: 5)
- `CONN_BRIDGE_CIRCUIT_BREAKER_TIMEOUT`: Timeout before half-open in seconds (default: 60s)
- `CONN_BRIDGE_CIRCUIT_BREAKER_MAX_REQUESTS`: Max requests in half-open state (default: 3)

### Example Configuration

```bash
export CONN_BRIDGE_CIRCUIT_BREAKER_NAME="bacen-circuit-breaker"
export CONN_BRIDGE_CIRCUIT_BREAKER_MAX_FAILURES=5
export CONN_BRIDGE_CIRCUIT_BREAKER_TIMEOUT=60s
export CONN_BRIDGE_CIRCUIT_BREAKER_MAX_REQUESTS=3
```

## Usage

### Basic Usage

```go
import (
    "github.com/lbpay-lab/conn-bridge/internal/infrastructure/circuitbreaker"
)

// Create circuit breaker with default config
config := circuitbreaker.DefaultConfig("my-breaker")
breaker := circuitbreaker.NewGoBreakerAdapter(config)

// Execute function with circuit breaker protection
result, err := breaker.Execute(ctx, func() (interface{}, error) {
    // Your code here
    return someOperation()
})
```

### Custom Configuration

```go
config := &circuitbreaker.Config{
    Name:        "custom-breaker",
    MaxRequests: 5,
    Interval:    30 * time.Second,
    Timeout:     45 * time.Second,
    MaxFailures: 10,
    Logger:      logrus.New(),
}

breaker := circuitbreaker.NewGoBreakerAdapter(config)
```

### Integration with Bacen Client

The circuit breaker is automatically integrated with the Bacen HTTP client:

```go
// In container initialization
baseBacenClient := bacen.NewHTTPClient(config)
circuitBreaker := circuitbreaker.NewGoBreakerAdapter(cbConfig)

// Wrap with circuit breaker
bacenClient := bacen.NewCircuitBreakerClient(baseBacenClient, circuitBreaker, logger)
```

## Observability

### Prometheus Metrics

The circuit breaker exports the following Prometheus metrics:

#### State Gauge
```
circuit_breaker_state{name="bacen-circuit-breaker"}
```
Values:
- `0`: CLOSED (normal operation)
- `1`: OPEN (fast-fail mode)
- `2`: HALF_OPEN (recovery mode)

#### Counters

```
circuit_breaker_failures_total{name="bacen-circuit-breaker"}
circuit_breaker_successes_total{name="bacen-circuit-breaker"}
circuit_breaker_requests_total{name="bacen-circuit-breaker",state="closed|open|half_open"}
circuit_breaker_state_transitions_total{name="bacen-circuit-breaker",from="closed",to="open"}
```

### Logging

State transitions and important events are logged with structured fields:

```json
{
  "level": "info",
  "msg": "Circuit breaker state transition",
  "name": "bacen-circuit-breaker",
  "from": "closed",
  "to": "open",
  "time": "2024-10-27T10:00:00Z"
}
```

## Behavior

### State Transitions

1. **CLOSED → OPEN**: After 5 consecutive failures
   - All subsequent requests fail fast without calling Bacen
   - Timer starts for timeout period (60s)

2. **OPEN → HALF_OPEN**: After timeout period expires
   - Circuit allows up to 3 test requests
   - If any fails, immediately returns to OPEN
   - If all succeed, transitions to CLOSED

3. **HALF_OPEN → CLOSED**: After successful test requests
   - Normal operation resumes
   - Failure counters reset

4. **HALF_OPEN → OPEN**: If any test request fails
   - Returns to fast-fail mode
   - Timer resets for another timeout period

### Fast-Fail Behavior

When circuit is OPEN:
- Requests return immediately with error: "circuit breaker is open"
- No network calls are made to Bacen
- This prevents overwhelming a failing service

## Thread Safety

The circuit breaker implementation is thread-safe and can be safely used from multiple goroutines concurrently.

## Testing

Run tests with:

```bash
go test ./internal/infrastructure/circuitbreaker/... -v
go test ./internal/infrastructure/bacen/... -v
```

## Performance Considerations

- Minimal overhead when circuit is CLOSED
- Zero network calls when circuit is OPEN (fast-fail)
- Automatic recovery attempts via HALF_OPEN state
- Metrics collection has negligible performance impact

## Limitations

- gobreaker v0.5.0 doesn't expose a public Reset() method
- To upgrade to v2, change import to `github.com/sony/gobreaker/v2`

## References

- [Circuit Breaker Pattern - Martin Fowler](https://martinfowler.com/bliki/CircuitBreaker.html)
- [gobreaker GitHub](https://github.com/sony/gobreaker)
- [Prometheus Client Go](https://github.com/prometheus/client_golang)
