# Task 4: Database Connection Patterns Analysis

## Overview
Analysis of database and persistence patterns in connector-dict to guide the VSync implementation's data layer design.

## Current Database Architecture

### Infrastructure Components

#### 1. PostgreSQL (for Temporal)
- **Purpose**: Temporal workflow state management
- **Container**: `postgres:12`
- **Configuration**: Via docker-compose only
- **Usage**: Indirect (through Temporal SDK)

#### 2. Redis Cache
- **Purpose**: Response caching, idempotency
- **Container**: `redis:7.2`
- **Library**: `github.com/redis/go-redis/v9`
- **Usage**: Direct via Redis client

### Key Finding: No Direct PostgreSQL Usage
**Important Discovery**: The connector-dict does NOT directly use PostgreSQL for business data. All persistence is handled through:
1. **Temporal**: Workflow state and history
2. **Redis**: Caching and idempotency
3. **BACEN Bridge**: Source of truth for DICT data

## Redis Implementation Patterns

### 1. Redis Client Configuration
```go
// Setup
func NewRedis(opts *redis.Options, prefix string, provider interfaces.Provider) *Redis {
    client := redis.NewClient(opts)
    return &Redis{
        client: client,
        prefix: prefix,
        tracer: provider.Tracer(),
        logger: provider.Logger(),
    }
}

// Configuration from environment
type Config struct {
    RedisAddr   string `env:"REDIS_ADDR"`    // localhost:6379
    RedisDB     int    `env:"REDIS_DB"`      // 0
    RedisPrefix string `env:"REDIS_PREFIX"`  // dict:
}
```

### 2. Key Management Pattern
```go
// Consistent key generation
func (r *Redis) KeyFromHash(hash string) string {
    if r.prefix == "" {
        return fmt.Sprintf("hash:%s", hash)
    }
    return fmt.Sprintf("%s:hash:%s", r.prefix, hash)
}
```

### 3. Cache Operations
```go
// Basic operations
Get(ctx context.Context, key string) ([]byte, bool, error)
Set(ctx context.Context, key string, val []byte, ttl time.Duration) error
SetIfAbsent(ctx context.Context, key string, val []byte, ttl time.Duration) (bool, error)
Delete(ctx context.Context, key string) error
RefreshTTL(ctx context.Context, key string, ttl time.Duration) (bool, error)
```

### 4. Error Handling with Envelopes
```go
// Envelope pattern for cached responses
type Envelope struct {
    IsError   bool            `json:"is_error"`
    Value     json.RawMessage `json:"value,omitempty"`
    Error     json.RawMessage `json:"error,omitempty"`
}

// Store success or error
SetWithError(ctx context.Context, key string, data any, isError bool, ttl time.Duration) error
GetCachedWithError(ctx context.Context, op, reqID string, data any) error
```

### 5. Observability Integration
```go
// All operations include tracing and logging
ctx, span := r.tracer.StartSpanWithAttributes(ctx, "RedisCache.Get", map[string]interface{}{
    "cache.key": key,
})
defer span.End()

r.logger.InfoWithOperation(ctx, "RedisCache.Get", "cache hit",
    interfaces.String("key", key),
    interfaces.String("latency_ms", fmt.Sprintf("%d", duration)))
```

## Temporal as Database

### Workflow State Persistence
Temporal handles all workflow state persistence:
- Workflow execution history
- Activity results
- Timer states
- Child workflow states

### Pattern for Data Persistence
```go
// Data is persisted as workflow state
type ClaimWorkflowState struct {
    ClaimID   string
    Status    string
    CreatedAt time.Time
    UpdatedAt time.Time
}

// Query pattern
workflow.SetQueryHandler(ctx, "getClaimStatus", func() (string, error) {
    return state.Status, nil
})
```

## VSync Database Requirements

### PostgreSQL Implementation Needed
Unlike the main connector-dict, VSync REQUIRES PostgreSQL for:
1. **CID Storage**: Persistent storage of generated CIDs
2. **Sync State**: Track sync status with BACEN
3. **Audit Trail**: History of all operations
4. **Batch Processing**: Queue for daily sync

### Recommended Database Stack

#### 1. PostgreSQL Setup
```go
// Using pgx v5 (recommended)
import "github.com/jackc/pgx/v5/pgxpool"

type DatabaseConfig struct {
    Host     string `env:"DB_HOST" default:"localhost"`
    Port     int    `env:"DB_PORT" default:"5432"`
    User     string `env:"DB_USER" default:"vsync"`
    Password string `env:"DB_PASSWORD" required:"true"`
    Database string `env:"DB_NAME" default:"vsync"`
    SSLMode  string `env:"DB_SSL_MODE" default:"disable"`
}

func NewPostgresPool(cfg DatabaseConfig) (*pgxpool.Pool, error) {
    connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
        cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Database, cfg.SSLMode)

    config, err := pgxpool.ParseConfig(connStr)
    if err != nil {
        return nil, err
    }

    // Connection pool settings
    config.MaxConns = 20
    config.MinConns = 5
    config.MaxConnLifetime = time.Hour
    config.MaxConnIdleTime = time.Minute * 30

    return pgxpool.NewWithConfig(context.Background(), config)
}
```

#### 2. Migration Tool
```go
// Using golang-migrate
import "github.com/golang-migrate/migrate/v4"

migrations/
├── 000001_create_cid_table.up.sql
├── 000001_create_cid_table.down.sql
├── 000002_add_sync_status.up.sql
└── 000002_add_sync_status.down.sql
```

#### 3. Repository Pattern
```go
type CIDRepository struct {
    db     *pgxpool.Pool
    tracer interfaces.Tracer
    logger interfaces.Logger
}

func (r *CIDRepository) Create(ctx context.Context, cid *domain.CID) error {
    ctx, span := r.tracer.Start(ctx, "CIDRepository.Create")
    defer span.End()

    query := `
        INSERT INTO cids (
            key_value, key_type, cid, ispb, branch,
            account_number, account_type, tax_id,
            created_at, sync_status
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

    _, err := r.db.Exec(ctx, query,
        cid.KeyValue, cid.KeyType, cid.CID, cid.ISPB,
        cid.Branch, cid.AccountNumber, cid.AccountType,
        cid.TaxID, cid.CreatedAt, cid.SyncStatus)

    if err != nil {
        r.logger.Error(ctx, "Failed to create CID", err)
        return err
    }

    return nil
}
```

## Database Schema for VSync

### Core Tables
```sql
-- CID storage
CREATE TABLE cids (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_value       VARCHAR(77) NOT NULL,
    key_type        VARCHAR(10) NOT NULL,
    cid             VARCHAR(255) NOT NULL UNIQUE,
    ispb            VARCHAR(8) NOT NULL,
    branch          VARCHAR(10),
    account_number  VARCHAR(20) NOT NULL,
    account_type    VARCHAR(4) NOT NULL,
    tax_id          VARCHAR(14) NOT NULL,
    created_at      TIMESTAMP NOT NULL,
    updated_at      TIMESTAMP NOT NULL DEFAULT NOW(),
    sync_status     VARCHAR(20) NOT NULL DEFAULT 'pending',
    synced_at       TIMESTAMP,
    UNIQUE(key_value, ispb)
);

-- Sync batch tracking
CREATE TABLE sync_batches (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    batch_date      DATE NOT NULL UNIQUE,
    total_records   INTEGER NOT NULL DEFAULT 0,
    synced_records  INTEGER NOT NULL DEFAULT 0,
    status          VARCHAR(20) NOT NULL DEFAULT 'pending',
    started_at      TIMESTAMP,
    completed_at    TIMESTAMP,
    error_message   TEXT
);

-- Audit log
CREATE TABLE audit_logs (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    correlation_id  VARCHAR(255) NOT NULL,
    action          VARCHAR(50) NOT NULL,
    key_value       VARCHAR(77),
    cid             VARCHAR(255),
    payload         JSONB,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_cids_key_value ON cids(key_value);
CREATE INDEX idx_cids_sync_status ON cids(sync_status);
CREATE INDEX idx_cids_created_at ON cids(created_at);
CREATE INDEX idx_audit_logs_correlation_id ON audit_logs(correlation_id);
```

## Connection Pool Best Practices

### 1. Pool Configuration
```go
// Recommended settings for VSync
MaxConns: 20          // Maximum connections
MinConns: 5           // Minimum idle connections
MaxConnLifetime: 1h   // Connection lifetime
MaxConnIdleTime: 30m  // Idle timeout
```

### 2. Health Checks
```go
func (db *Database) Health(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()

    if err := db.pool.Ping(ctx); err != nil {
        return fmt.Errorf("database health check failed: %w", err)
    }
    return nil
}
```

### 3. Graceful Shutdown
```go
func (db *Database) Close() {
    db.pool.Close()
}
```

## Transaction Patterns

### Basic Transaction
```go
func (r *CIDRepository) CreateBatch(ctx context.Context, cids []*domain.CID) error {
    tx, err := r.db.Begin(ctx)
    if err != nil {
        return err
    }
    defer tx.Rollback(ctx)

    for _, cid := range cids {
        if err := r.createInTx(ctx, tx, cid); err != nil {
            return err
        }
    }

    return tx.Commit(ctx)
}
```

## Observability for Database

### 1. Query Logging
```go
// Log slow queries
if duration > 100*time.Millisecond {
    r.logger.Warn(ctx, "Slow query detected",
        interfaces.String("query", query),
        interfaces.Int64("duration_ms", duration.Milliseconds()))
}
```

### 2. Metrics
```go
// Track connection pool metrics
metrics := db.pool.Stat()
gauge.Set("db.connections.active", float64(metrics.AcquiredConns()))
gauge.Set("db.connections.idle", float64(metrics.IdleConns()))
```

## Recommendations for VSync

### 1. Use PostgreSQL Directly
- VSync needs persistent storage for CIDs
- Cannot rely only on Redis/Temporal
- PostgreSQL provides ACID guarantees

### 2. Implement Repository Pattern
- Clean separation of concerns
- Easier testing with mocks
- Consistent with existing patterns

### 3. Use pgx/v5 Library
- Best performance for PostgreSQL
- Native PostgreSQL types support
- Built-in connection pooling

### 4. Migration Strategy
- Use golang-migrate for schema management
- Version control all migrations
- Automated migration on startup

### 5. Cache Strategy
- Use Redis for hot data (recent CIDs)
- PostgreSQL for complete history
- TTL: 24 hours for cache entries

## Conclusion

While connector-dict doesn't directly use PostgreSQL (relying on Temporal and Redis), the VSync system requires a proper database for:
1. **Persistent CID storage**
2. **Sync state tracking**
3. **Audit trail**
4. **Batch processing**

The recommended approach:
- **PostgreSQL** with pgx/v5 for persistence
- **Redis** for caching and performance
- **Repository pattern** for data access
- **golang-migrate** for schema management
- **Observability** throughout the data layer

This architecture provides the durability and querying capabilities needed for VSync while following the established patterns in the connector-dict codebase.