# Pulsar Publisher for DICT Events

## Overview

The DictPublisher is an Apache Pulsar event publisher specifically designed for the Bridge service to publish DICT events asynchronously to enable event-driven processing by the Connect service.

## Architecture

### Components

1. **DictPublisher** (`dict_publisher.go`)
   - Main publisher implementation
   - Handles connection management
   - Provides type-safe publishing methods
   - Implements retry logic and error handling

2. **Event Structures** (`dict_events.go`)
   - Type-safe event definitions
   - Factory methods for creating events
   - ISO8601 timestamp formatting
   - Trace ID propagation

3. **Prometheus Metrics**
   - Message publication counters
   - Error tracking
   - Duration histograms

## Features

### 1. Producer Configuration

- **Producer Name**: `rsfn-bridge-producer`
- **Topic**: `rsfn-dict-res-out` (responses from Bacen)
- **Batching**: Enabled (max 100 messages or 10ms delay)
- **Compression**: LZ4
- **Non-blocking**: Async publishing with goroutines

### 2. Publishing Methods

#### PublishEntryCreated
```go
func (dp *DictPublisher) PublishEntryCreated(
    ctx context.Context,
    entry *entities.DictEntry,
    traceID string
) error
```
Publishes an event when a new DICT entry is created.

#### PublishEntryUpdated
```go
func (dp *DictPublisher) PublishEntryUpdated(
    ctx context.Context,
    entry *entities.DictEntry,
    traceID string
) error
```
Publishes an event when a DICT entry is updated.

#### PublishEntryDeleted
```go
func (dp *DictPublisher) PublishEntryDeleted(
    ctx context.Context,
    keyID string,
    traceID string
) error
```
Publishes an event when a DICT entry is deleted.

#### PublishError
```go
func (dp *DictPublisher) PublishError(
    ctx context.Context,
    err error,
    context string,
    traceID string
) error
```
Publishes an error event for monitoring and alerting.

### 3. Message Format

All messages are JSON-serialized with the following properties:

```json
{
  "event_type": "entry.created | entry.updated | entry.deleted | error",
  "source": "rsfn-bridge",
  "timestamp": "2025-10-27T10:30:00Z",
  "trace_id": "otel-trace-id"
}
```

### 4. Error Handling

- **Retry Logic**: 3 attempts with exponential backoff
- **Non-blocking**: Publishing happens in goroutines
- **Graceful Degradation**: Errors are logged but don't block HTTP responses
- **Context Awareness**: Respects context cancellation

### 5. Observability

#### Prometheus Metrics

1. **pulsar_messages_published_total**
   - Type: Counter
   - Labels: `event_type`
   - Description: Total number of successfully published messages

2. **pulsar_publish_errors_total**
   - Type: Counter
   - Description: Total number of publish errors after all retries

3. **pulsar_publish_duration_seconds**
   - Type: Histogram
   - Description: Duration of publish operations
   - Buckets: Default Prometheus buckets

## Configuration

### Environment Variables

```bash
# Pulsar broker URL
CONN_BRIDGE_PULSAR_BROKER_URL=pulsar://pulsar.lbpay.local:6650

# Connection timeout (optional, default: 10s)
CONN_BRIDGE_PULSAR_TIMEOUT=30s
```

### Programmatic Configuration

```go
config := &pulsar.DictPublisherConfig{
    BrokerURL:         "pulsar://localhost:6650",
    Topic:             "rsfn-dict-res-out",
    ProducerName:      "rsfn-bridge-producer",
    BatchingEnabled:   true,
    MaxMessages:       100,
    BatchingMaxDelay:  10 * time.Millisecond,
    CompressionType:   pulsar.LZ4,
    OperationTimeout:  30 * time.Second,
    ConnectionTimeout: 10 * time.Second,
}

publisher, err := pulsar.NewDictPublisher(config)
if err != nil {
    log.Fatal(err)
}
defer publisher.Close()
```

## Usage Examples

### Basic Usage

```go
// Create publisher
publisher, err := pulsar.NewDictPublisher(
    pulsar.DefaultDictPublisherConfig("pulsar://localhost:6650"),
)
if err != nil {
    return err
}
defer publisher.Close()

// Publish entry created event
entry := &entities.DictEntry{
    Key:         "user@example.com",
    Type:        entities.KeyTypeEmail,
    Participant: "12345678",
    // ... other fields
}

ctx := context.Background()
traceID := "otel-trace-id-123"

err = publisher.PublishEntryCreated(ctx, entry, traceID)
if err != nil {
    log.Errorf("Failed to publish event: %v", err)
}
```

### Integration with Use Cases

```go
// In CreateEntryUseCase
func (uc *CreateEntryUseCase) Execute(ctx context.Context, req *Request) error {
    // Create entry in Bacen
    entry, err := uc.bacenClient.CreateEntry(ctx, req)
    if err != nil {
        return err
    }

    // Publish event (non-blocking)
    traceID := getTraceIDFromContext(ctx)
    if err := uc.publisher.PublishEntryCreated(ctx, entry, traceID); err != nil {
        // Log but don't fail the request
        uc.logger.WithError(err).Warn("Failed to publish entry created event")
    }

    return nil
}
```

## Event Schemas

### EntryCreatedEvent
```json
{
  "event_type": "entry.created",
  "source": "rsfn-bridge",
  "timestamp": "2025-10-27T10:30:00Z",
  "trace_id": "trace-123",
  "entry": {
    "key": "user@example.com",
    "type": "EMAIL",
    "participant": "12345678",
    "account": {
      "ispb": "12345678",
      "branch": "0001",
      "number": "123456",
      "type": "CHECKING"
    },
    "owner": {
      "type": "PERSON",
      "document": "12345678900",
      "name": "John Doe"
    },
    "status": "ACTIVE",
    "created_at": "2025-10-27T10:30:00Z"
  }
}
```

### EntryDeletedEvent
```json
{
  "event_type": "entry.deleted",
  "source": "rsfn-bridge",
  "timestamp": "2025-10-27T10:30:00Z",
  "trace_id": "trace-123",
  "key_id": "user@example.com",
  "key_type": "EMAIL",
  "participant": "12345678"
}
```

### ErrorEvent
```json
{
  "event_type": "error",
  "source": "rsfn-bridge",
  "timestamp": "2025-10-27T10:30:00Z",
  "trace_id": "trace-123",
  "error_code": "ERROR",
  "error_message": "Failed to create entry: timeout",
  "context": "CreateEntry operation"
}
```

## Testing

### Unit Tests

Run unit tests:
```bash
cd /Users/jose.silva.lb/LBPay/IA_Dict/conn-bridge
go test ./internal/infrastructure/pulsar/... -v
```

### Integration Tests

For integration testing with a real Pulsar instance, use Docker:

```bash
# Start Pulsar
docker run -d -p 6650:6650 -p 8080:8080 \
  --name pulsar \
  apachepulsar/pulsar:3.0.0 \
  bin/pulsar standalone

# Run tests
go test ./internal/infrastructure/pulsar/... -v -tags=integration

# Stop Pulsar
docker stop pulsar && docker rm pulsar
```

## Performance Considerations

1. **Batching**: Messages are batched (up to 100 messages or 10ms) for efficiency
2. **Compression**: LZ4 compression reduces network bandwidth
3. **Async Publishing**: Non-blocking goroutines prevent HTTP response delays
4. **Connection Pooling**: Single persistent connection per producer

## Error Scenarios

### Pulsar Unavailable
- Initial connection failure: Service fails to start
- Runtime disconnection: Messages fail with retries, then error counter increments
- Recovery: Automatic reconnection on next publish attempt

### Context Cancellation
- Publish operation respects context cancellation
- In-flight messages are not sent
- Graceful shutdown ensures no message loss

### Message Validation Failures
- Nil entry/error checks before publishing
- JSON serialization errors are logged and counted
- Client code receives error immediately

## Monitoring and Alerting

### Key Metrics to Monitor

1. **pulsar_messages_published_total**
   - Alert if rate drops significantly
   - Track by event_type for insights

2. **pulsar_publish_errors_total**
   - Alert if error rate > 1%
   - Indicates Pulsar connectivity issues

3. **pulsar_publish_duration_seconds**
   - Alert if p99 > 100ms
   - May indicate network or Pulsar performance issues

### Example Prometheus Queries

```promql
# Message publish rate
rate(pulsar_messages_published_total[5m])

# Error rate percentage
rate(pulsar_publish_errors_total[5m]) /
rate(pulsar_messages_published_total[5m]) * 100

# P99 publish duration
histogram_quantile(0.99,
  rate(pulsar_publish_duration_seconds_bucket[5m]))
```

## Thread Safety

The DictPublisher is thread-safe:
- Uses sync.RWMutex for close state management
- Safe concurrent calls to publishing methods
- Goroutines handle async publishing safely

## Graceful Shutdown

```go
// On application shutdown
publisher.Close()

// This will:
// 1. Mark publisher as closed
// 2. Flush pending batched messages
// 3. Close the producer
// 4. Close the Pulsar client connection
```

## Future Enhancements

1. **Dead Letter Queue**: For messages that fail after all retries
2. **Schema Registry**: For schema evolution and validation
3. **Multi-tenancy**: Support for multiple topics/tenants
4. **Message Deduplication**: Prevent duplicate event publishing
5. **Observability**: Distributed tracing with OpenTelemetry spans
