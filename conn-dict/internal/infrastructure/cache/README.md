# Redis Cache Implementation - 5 Strategies

## Overview

This package implements a comprehensive Redis caching solution with **5 different caching strategies** optimized for different use cases in the DICT LBPay Connect service.

## Architecture

```
cache/
├── redis_cache.go          # Main cache implementation with all 5 strategies
├── redis_cache_test.go     # Comprehensive tests for all strategies
├── cache_metrics.go        # Prometheus metrics and observability
├── cache_keys.go           # Key patterns, TTLs, and builders
├── examples.go             # Usage examples for each strategy
├── redis_client.go         # Legacy client (kept for compatibility)
├── redis_repository.go     # Low-level Redis operations
└── README.md               # This file
```

## 5 Caching Strategies

### 1. Cache-Aside (Lazy Loading) ⚡

**Use Case**: Entry lookups, frequently accessed data

**Pattern**:
```
1. Check cache
2. If miss, query database
3. Store in cache
4. Return data
```

**Configuration**:
- TTL: 5 minutes
- Key pattern: `entry:{key_id}`

**Example**:
```go
result, err := cache.GetOrLoad(ctx, "entry:123", func() (interface{}, error) {
    // Load from database
    return db.GetEntry("123")
}, 5*time.Minute)
```

**Pros**: Simple, works well for read-heavy workloads
**Cons**: Cache miss penalty, potential thundering herd

---

### 2. Write-Through 📝

**Use Case**: Entry creation, critical data writes

**Pattern**:
```
1. Write to database first
2. Write to cache simultaneously
3. Return success
```

**Configuration**:
- TTL: 5 minutes
- Key pattern: `entry:{key_id}`

**Example**:
```go
err := cache.SetWithDB(ctx, "entry:123", entry, 5*time.Minute, func() error {
    // Write to database
    return db.CreateEntry(entry)
})
```

**Pros**: Data consistency, cache always up-to-date
**Cons**: Write latency (must write to both DB and cache)

---

### 3. Write-Behind (Write-Back) 🚀

**Use Case**: High-frequency updates (metrics, counters, analytics)

**Pattern**:
```
1. Write to cache immediately
2. Queue database write
3. Batch write to DB every 10 seconds
```

**Configuration**:
- TTL: 10 minutes
- Key pattern: `metrics:{participant_ispb}`
- Batch interval: 10 seconds (configurable)

**Example**:
```go
err := cache.SetAsync(ctx, "metrics:12345678", metrics, 10*time.Minute, func() error {
    // This will be called asynchronously in batch
    return db.UpdateMetrics(metrics)
})
```

**Pros**: Extremely fast writes, reduced DB load
**Cons**: Risk of data loss if cache fails, eventual consistency

---

### 4. Read-Through 📖

**Use Case**: Configuration data, reference data

**Pattern**:
```
1. Check cache
2. If miss, load from DB automatically
3. Cache result
4. Return data
```

**Configuration**:
- TTL: 1 hour
- Key pattern: `config:{key}`

**Example**:
```go
result, err := cache.GetWithLoader(ctx, "config:max_retries", func() (interface{}, error) {
    // Automatically loads from DB on miss
    return db.GetConfig("max_retries")
}, 1*time.Hour)
```

**Pros**: Simple for consumers, automatic cache population
**Cons**: Same as cache-aside, loader coupled to cache

---

### 5. Write-Around ⚠️

**Use Case**: Bulk operations (VSYNC), infrequently read data

**Pattern**:
```
1. Write directly to database
2. Invalidate cache
3. Don't populate cache immediately
```

**Configuration**:
- No TTL (invalidation-based)
- Key pattern: `bulk:{operation_id}`

**Example**:
```go
err := cache.InvalidateAndWrite(ctx, "entry:123", func() error {
    // Bulk write to database
    return db.BulkUpdate(entries)
})
// Cache will be repopulated on next read
```

**Pros**: Prevents cache pollution from bulk ops, saves memory
**Cons**: Cache miss on next read, not suitable for hot data

---

## Key Patterns

The package provides standard key patterns for different data types:

| Pattern | Format | TTL | Strategy | Use Case |
|---------|--------|-----|----------|----------|
| Entry | `entry:{id}` | 5min | Cache-Aside | Entry lookups |
| Metrics | `metrics:{ispb}` | 10min | Write-Behind | Analytics/counters |
| Config | `config:{name}` | 1hour | Read-Through | Configuration |
| Participant | `participant:{ispb}` | 15min | Cache-Aside | Participant data |
| Dict Key | `dict:key:{key}` | 5min | Cache-Aside | Key lookups |
| Dict ISPB | `dict:ispb:{ispb}` | 5min | Cache-Aside | ISPB lookups |
| VSYNC | `vsync:{id}` | none | Write-Around | Bulk sync |

## Configuration

### Redis Connection

```go
config := RedisCacheConfig{
    Addr:               "localhost:6379",
    Password:           "",
    DB:                 0,
    PoolSize:           100,          // Connection pool size
    MinIdleConns:       10,           // Minimum idle connections
    MaxRetries:         3,            // Retry on failure
    DialTimeout:        5 * time.Second,
    ReadTimeout:        3 * time.Second,
    WriteTimeout:       3 * time.Second,
    WriteBehindEnabled: true,         // Enable write-behind strategy
    WriteBehindBatch:   10 * time.Second, // Batch interval
}

cache, err := NewRedisCache(config, logger)
```

### Environment Variables

```bash
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=100
REDIS_MIN_IDLE_CONNS=10
REDIS_WRITE_BEHIND_ENABLED=true
REDIS_WRITE_BEHIND_BATCH=10s
```

## Prometheus Metrics

The cache exports comprehensive Prometheus metrics:

### Counters
- `conn_dict_cache_hits_total{strategy}` - Total cache hits by strategy
- `conn_dict_cache_misses_total{strategy}` - Total cache misses by strategy
- `conn_dict_cache_errors_total{operation,error_type}` - Total errors

### Histograms
- `conn_dict_cache_operation_duration_seconds{operation,strategy}` - Operation duration
- `conn_dict_cache_write_behind_flush_duration_seconds` - Flush duration

### Gauges
- `conn_dict_cache_write_behind_queue_size` - Current write-behind queue size

### Example Queries

```promql
# Cache hit rate by strategy
rate(conn_dict_cache_hits_total[5m]) /
(rate(conn_dict_cache_hits_total[5m]) + rate(conn_dict_cache_misses_total[5m]))

# P95 operation latency
histogram_quantile(0.95, rate(conn_dict_cache_operation_duration_seconds_bucket[5m]))

# Write-behind queue depth
conn_dict_cache_write_behind_queue_size
```

## Usage Examples

### Basic Setup

```go
import "github.com/lbpay-lab/conn-dict/internal/infrastructure/cache"

// Initialize cache
cache, err := cache.NewRedisCache(config, logger)
if err != nil {
    log.Fatal(err)
}
defer cache.Close()

// Use key builder for standard patterns
keyBuilder := cache.NewCacheKeyBuilder()
```

### Strategy Selection Guide

| Scenario | Strategy | Reason |
|----------|----------|--------|
| User looks up entry | Cache-Aside | Read-heavy, tolerate miss |
| User creates entry | Write-Through | Need consistency |
| Update counter/metrics | Write-Behind | High frequency writes |
| Load config at startup | Read-Through | Infrequent, auto-load |
| Bulk VSYNC operation | Write-Around | Avoid cache pollution |

### Integration with Service Layer

```go
type EntryService struct {
    cache      cache.Cache
    db         database.DB
    keyBuilder *cache.CacheKeyBuilder
}

func (s *EntryService) GetEntry(ctx context.Context, id string) (*Entry, error) {
    cacheKey := s.keyBuilder.BuildEntryKey(id)

    // Cache-Aside pattern
    result, err := s.cache.GetOrLoad(ctx, cacheKey, func() (interface{}, error) {
        return s.db.GetEntry(id)
    }, 5*time.Minute)

    if err != nil {
        return nil, err
    }

    return result.(*Entry), nil
}

func (s *EntryService) CreateEntry(ctx context.Context, entry *Entry) error {
    cacheKey := s.keyBuilder.BuildEntryKey(entry.ID)

    // Write-Through pattern
    return s.cache.SetWithDB(ctx, cacheKey, entry, 5*time.Minute, func() error {
        return s.db.InsertEntry(entry)
    })
}

func (s *EntryService) UpdateMetrics(ctx context.Context, ispb string, metrics *Metrics) error {
    cacheKey := s.keyBuilder.BuildMetricsKey(ispb)

    // Write-Behind pattern
    return s.cache.SetAsync(ctx, cacheKey, metrics, 10*time.Minute, func() error {
        return s.db.UpdateMetrics(metrics)
    })
}
```

## Testing

Run all tests:
```bash
# Start Redis
docker run -d -p 6379:6379 redis:7-alpine

# Run tests
cd conn-dict
go test -v ./internal/infrastructure/cache/...

# Run specific strategy test
go test -v -run TestCacheAside ./internal/infrastructure/cache/
```

## Performance Characteristics

| Strategy | Read Latency | Write Latency | Consistency | Use When |
|----------|--------------|---------------|-------------|----------|
| Cache-Aside | Low (hit) / High (miss) | N/A | Eventual | Read >> Write |
| Write-Through | Low | High | Strong | Write consistency critical |
| Write-Behind | Low | Very Low | Eventual | High write frequency |
| Read-Through | Low (hit) / High (miss) | N/A | Eventual | Auto-load needed |
| Write-Around | High (first read) | Low | Eventual | Bulk operations |

## Best Practices

### 1. Connection Pool Sizing
```go
// For production
PoolSize: 100       // Max concurrent connections
MinIdleConns: 10    // Always ready connections
```

### 2. TTL Selection
- **Hot data** (frequently accessed): 5-15 minutes
- **Warm data** (moderately accessed): 15-60 minutes
- **Cold data** (rarely accessed): 1-24 hours
- **Static data** (configs): 1-24 hours

### 3. Error Handling
```go
result, err := cache.GetOrLoad(ctx, key, loader, ttl)
if err != nil {
    // Log but don't fail - degrade gracefully
    logger.Warn("Cache unavailable, using DB directly")
    return loader()
}
```

### 4. Graceful Degradation
Always implement fallback to database when cache is unavailable.

### 5. Cache Warming
```go
// Warm critical data on startup
func (s *Service) WarmCache(ctx context.Context) error {
    criticalKeys := []string{"config:max_retries", "config:timeout"}
    for _, key := range criticalKeys {
        _, err := s.cache.GetWithLoader(ctx, key, s.loadConfig, 1*time.Hour)
        if err != nil {
            return err
        }
    }
    return nil
}
```

## Monitoring

### Health Checks
```go
func (c *RedisCache) HealthCheck(ctx context.Context) error {
    return c.client.Ping(ctx).Err()
}
```

### Dashboard Metrics
Monitor these in Grafana:
- Cache hit rate (target: >80%)
- P95 latency (target: <10ms)
- Error rate (target: <0.1%)
- Write-behind queue size (target: <100)

## Troubleshooting

### High Cache Miss Rate
- Increase TTL
- Pre-warm cache on startup
- Check if keys are being invalidated too frequently

### High Write-Behind Queue Size
- Increase batch frequency
- Check database performance
- Verify DB writers are not failing

### Memory Issues
- Review TTL settings
- Consider using Write-Around for bulk operations
- Monitor key count and size

## Migration Guide

### From Old RedisClient
```go
// Old
client := cache.NewRedisClient(config, logger)
client.Get(ctx, key, &dest)

// New
cache := cache.NewRedisCache(config, logger)
cache.GetOrLoad(ctx, key, loader, ttl)
```

## References

- [Redis Best Practices](https://redis.io/docs/manual/patterns/)
- [Caching Strategies](https://aws.amazon.com/caching/best-practices/)
- [Cache Patterns](https://docs.microsoft.com/en-us/azure/architecture/patterns/cache-aside)

## Support

For issues or questions:
- Check logs: `logger.SetLevel(logrus.DebugLevel)`
- Review metrics: `http://localhost:9090/metrics`
- Open issue in project repository
