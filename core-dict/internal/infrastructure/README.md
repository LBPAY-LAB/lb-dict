# Infrastructure Layer - Core DICT

This directory contains the infrastructure implementations for the Core DICT service, including caching, messaging, database access, and gRPC services.

## 📁 Directory Structure

```
infrastructure/
├── cache/                      # Redis caching and rate limiting
│   ├── redis_client.go         # Redis connection and operations
│   ├── cache_impl.go           # 5 caching strategies
│   ├── rate_limiter.go         # Token bucket & sliding window rate limiting
│   └── redis_client_test.go   # Unit tests
├── messaging/                  # Apache Pulsar event streaming
│   ├── pulsar_producer.go      # Event producer (domain events)
│   └── pulsar_consumer.go      # Event consumer (response events)
├── database/                   # PostgreSQL repositories
│   ├── postgres_connection.go  # Connection pool
│   ├── entry_repository_impl.go
│   ├── account_repository_impl.go
│   ├── claim_repository_impl.go
│   ├── audit_repository_impl.go
│   └── transaction_manager.go
└── grpc/                       # gRPC server and interceptors
    ├── grpc_server.go
    ├── core_dict_service_handler.go
    ├── auth_interceptor.go
    ├── logging_interceptor.go
    └── metrics_interceptor.go
```

---

## 🔴 Redis Cache

### Features
- Connection pooling with configurable size
- TLS/SSL support
- 5 caching strategies (Cache-Aside, Write-Through, Write-Behind, Read-Through, Write-Around)
- Token bucket and sliding window rate limiting
- Distributed locks (SetNX)
- Pipeline and transaction support

### Usage

#### Basic Redis Operations

```go
import "github.com/lbpay-lab/core-dict/internal/infrastructure/cache"

// Create client
config := cache.DefaultRedisConfig()
config.URL = "redis://localhost:6379/0"
client, err := cache.NewRedisClient(config)
defer client.Close()

// Set with TTL
client.Set(ctx, "key", "value", 5*time.Minute)

// Get
value, err := client.Get(ctx, "key")

// Distributed lock
acquired, err := client.SetNX(ctx, "lock:resource", "holder", 30*time.Second)
```

#### Cache-Aside Strategy

```go
cache := cache.NewCache(client, cache.CacheAside, "core-dict:")
handler := cache.NewCacheAsideHandler(cache)

var entry DictEntry
err := handler.GetOrLoad(ctx, cacheKey, &entry, 5*time.Minute,
    func(ctx context.Context) (interface{}, error) {
        return repo.GetByKey(ctx, key)
    },
)
```

#### Write-Through Strategy

```go
handler := cache.NewWriteThroughHandler(cache)

err := handler.Write(ctx, cacheKey, entry, 5*time.Minute,
    func(ctx context.Context, value interface{}) error {
        return repo.Save(ctx, value)
    },
)
```

#### Rate Limiting

```go
// IP-based rate limiting (100 req/s)
ipLimiter := cache.NewIPRateLimiter(client, 100)
allowed, err := ipLimiter.Allow(ctx, "192.168.1.1")
if !allowed {
    return errors.New("rate limit exceeded")
}

// Account-based rate limiting
accountLimiter := cache.NewAccountRateLimiter(client, 100)
allowed, err := accountLimiter.Allow(ctx, accountID)

// Multi-key rate limiting (IP + Account)
multiLimiter := cache.NewMultiKeyRateLimiter(ipLimiter, accountLimiter)
allowed, err := multiLimiter.Allow(ctx, ipAddress, accountID)
```

#### Sliding Window Rate Limiter

```go
limiter := cache.NewSlidingWindowRateLimiter(client, config)
allowed, err := limiter.Allow(ctx, "user123")

// Get current count in window
count, err := limiter.GetCount(ctx, "user123")
```

### Configuration

```go
type RedisConfig struct {
    URL              string        // redis://localhost:6379/0
    Password         string
    DB               int           // Default: 0
    PoolSize         int           // Default: 10
    MinIdleConns     int           // Default: 5
    MaxRetries       int           // Default: 3
    ConnMaxIdleTime  time.Duration // Default: 5min
    ConnMaxLifetime  time.Duration // Default: 30min
    DialTimeout      time.Duration // Default: 5s
    ReadTimeout      time.Duration // Default: 3s
    WriteTimeout     time.Duration // Default: 3s
    TLSEnabled       bool
    TLSSkipVerify    bool
}
```

### Environment Variables

```bash
REDIS_URL=redis://localhost:6379/0
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5
REDIS_MAX_RETRIES=3
CACHE_ENTRY_TTL=300s
CACHE_ACCOUNT_TTL=600s
```

---

## 📨 Pulsar Messaging

### Features
- Async and sync event publishing
- Compression (LZ4)
- Message batching (100 msgs or 10ms)
- Event ordering via partition key
- Dead Letter Queue (DLQ) support
- Multi-topic consumer
- Specialized producers (Key events, Claim events)

### Usage

#### Event Producer

```go
import "github.com/lbpay-lab/core-dict/internal/infrastructure/messaging"

// Create producer
config := messaging.DefaultProducerConfig()
config.BrokerURL = "pulsar://localhost:6650"
producer, err := messaging.NewEventProducer(config)
defer producer.Close()

// Publish event
event := &messaging.DomainEvent{
    EventID:       "evt_123",
    EventType:     "KeyCreated",
    AggregateID:   "12345678901",
    AggregateType: "DictEntry",
    Data: map[string]interface{}{
        "key": "12345678901",
        "key_type": "CPF",
    },
}
err = producer.PublishEvent(ctx, event)

// Async publish
producer.PublishEventAsync(ctx, event, func(msgID pulsar.MessageID, err error) {
    if err != nil {
        log.Printf("Failed: %v", err)
    }
})

// Flush pending messages
producer.Flush()
```

#### Specialized Key Event Producer

```go
keyProducer, err := messaging.NewKeyEventProducer("pulsar://localhost:6650")
defer keyProducer.Close()

// Publish key created
keyProducer.PublishKeyCreated(ctx, key, keyType, accountISPB)

// Publish key updated
keyProducer.PublishKeyUpdated(ctx, key, map[string]interface{}{
    "old_status": "ACTIVE",
    "new_status": "BLOCKED",
})

// Publish key deleted
keyProducer.PublishKeyDeleted(ctx, key, "user_requested")
```

#### Event Consumer

```go
config := messaging.DefaultConsumerConfig()
config.BrokerURL = "pulsar://localhost:6650"
consumer, err := messaging.NewEventConsumer(config)
defer consumer.Close()

// Register event handlers
consumer.RegisterHandler("KeyCreatedResponse", func(ctx context.Context, event *messaging.DomainEvent) error {
    key := event.Data["key"].(string)
    success := event.Data["success"].(bool)

    if success {
        log.Printf("Key %s created successfully", key)
    }
    return nil
})

// Start consuming (blocks)
go consumer.Start(ctx)
```

#### Response Event Consumer

```go
responseConsumer, err := messaging.NewResponseEventConsumer("pulsar://localhost:6650")
defer responseConsumer.Close()

// Handle responses from RSFN
responseConsumer.OnKeyCreatedResponse(func(ctx context.Context, key string, success bool, errorMsg string) error {
    if success {
        // Update local DB
    } else {
        log.Printf("Error: %s", errorMsg)
    }
    return nil
})

go responseConsumer.Start(ctx)
```

#### Dead Letter Queue

```go
dlqPolicy := &messaging.DeadLetterPolicy{
    MaxRedeliveryCount: 5,
    DeadLetterTopic: "persistent://core-dict/dlq/failed-events",
}
consumer, err := messaging.NewConsumerWithDLQ(config, dlqPolicy)
```

### Configuration

```go
type PulsarProducerConfig struct {
    BrokerURL          string                  // pulsar://localhost:6650
    Topic              string
    ProducerName       string
    CompressionType    pulsar.CompressionType  // Default: LZ4
    BatchingMaxMessages uint                   // Default: 100
    BatchingMaxDelay   time.Duration          // Default: 10ms
    SendTimeout        time.Duration          // Default: 30s
}
```

### Environment Variables

```bash
PULSAR_URL=pulsar://localhost:6650
PULSAR_TOPIC_KEY_EVENTS=persistent://dict/events/key-events
PULSAR_TOPIC_CLAIM_EVENTS=persistent://dict/events/claim-events
PULSAR_TOPIC_DICT_REQ_OUT=persistent://lb-conn/dict/rsfn-dict-req-out
PULSAR_TOPIC_DICT_RES_IN=persistent://lb-conn/dict/rsfn-dict-res-in
PULSAR_CONSUMER_SUBSCRIPTION=core-dict-sub
PULSAR_CONSUMER_TYPE=shared
```

---

## 🗄️ Database (PostgreSQL)

### Features
- Connection pooling (pgx)
- Transaction management
- Row-Level Security (RLS)
- Prepared statements
- Batch operations

### Usage

```go
// Create connection
config := &database.PostgresConfig{
    URL: "postgres://user:pass@localhost:5432/lbpay_core_dict",
    MaxOpenConns: 100,
    MaxIdleConns: 10,
}
pool, err := database.NewPostgresConnection(config)
defer pool.Close()

// Use repository
repo := database.NewEntryRepository(pool)

// Find by key
entry, err := repo.GetByKey(ctx, "12345678901")

// Transaction
tx, err := pool.Begin(ctx)
defer tx.Rollback(ctx)

err = repo.Create(ctx, tx, entry)
err = tx.Commit(ctx)
```

---

## 🔧 gRPC Server

### Features
- JWT authentication interceptor
- Request/response logging
- Prometheus metrics
- Panic recovery
- Rate limiting

### Usage

```go
// Create gRPC server
server := grpc.NewGRPCServer(&grpc.ServerConfig{
    Port: 9090,
    Interceptors: []grpc.UnaryServerInterceptor{
        grpc.AuthInterceptor(jwtSecret),
        grpc.LoggingInterceptor(logger),
        grpc.MetricsInterceptor(),
        grpc.RateLimitInterceptor(limiter),
    },
})

// Register services
pb.RegisterCoreDictServiceServer(server, handler)

// Start server
server.Start()
```

---

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
go test ./internal/infrastructure/...

# Run with coverage
go test -cover ./internal/infrastructure/...

# Run specific package
go test ./internal/infrastructure/cache
```

### Integration Tests

Requires running infrastructure:
```bash
# Start Redis and Pulsar
docker-compose up -d redis pulsar

# Run tests
go test -tags=integration ./internal/infrastructure/...
```

---

## 📊 Performance

### Redis
- **Connection pool**: 10 connections (configurable)
- **Pipeline**: Batch operations for lower latency
- **TTL**: Automatic expiration (5min entries, 10min accounts)

### Pulsar
- **Batching**: 100 messages or 10ms (whichever comes first)
- **Compression**: LZ4 (best balance speed/ratio)
- **Async publishing**: Non-blocking
- **Message ordering**: Guaranteed by partition key

### Rate Limiting
- **Lua scripts**: Atomic operations (no race conditions)
- **Redis ZSET**: Efficient sliding window
- **Distributed**: Works with multiple instances

---

## 🔍 Monitoring

### Metrics Exposed

#### Redis
- `redis_connections_active`
- `redis_connections_idle`
- `redis_hits_total`
- `redis_misses_total`
- `redis_operations_duration_seconds`

#### Pulsar
- `pulsar_messages_sent_total`
- `pulsar_messages_received_total`
- `pulsar_publish_latency_seconds`
- `pulsar_consumer_lag`

#### Rate Limiter
- `rate_limit_requests_total{allowed="true|false"}`
- `rate_limit_active_limiters`

Access metrics: `http://localhost:9091/metrics`

---

## 🚀 Example Application

See `examples/redis_pulsar_example.go` for a complete example demonstrating:
- Redis caching with Cache-Aside pattern
- Rate limiting by IP
- Pulsar event publishing (sync and async)
- Integration of cache + events

Run example:
```bash
# Start infrastructure
docker-compose up -d

# Run example
go run examples/redis_pulsar_example.go
```

---

## 📚 References

- [Redis Go Client](https://github.com/redis/go-redis)
- [Apache Pulsar Go Client](https://github.com/apache/pulsar-client-go)
- [PostgreSQL pgx](https://github.com/jackc/pgx)
- [gRPC Go](https://grpc.io/docs/languages/go/)

---

**Last Updated**: 2025-10-27
**Maintainer**: DevOps Core Team
