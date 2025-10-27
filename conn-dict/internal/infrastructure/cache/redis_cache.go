package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// CacheStrategy represents different caching strategies
type CacheStrategy string

const (
	// StrategyCacheAside - Cache-Aside (Lazy Loading)
	StrategyCacheAside CacheStrategy = "cache_aside"
	// StrategyWriteThrough - Write-Through Cache
	StrategyWriteThrough CacheStrategy = "write_through"
	// StrategyWriteBehind - Write-Behind (Write-Back)
	StrategyWriteBehind CacheStrategy = "write_behind"
	// StrategyReadThrough - Read-Through Cache
	StrategyReadThrough CacheStrategy = "read_through"
	// StrategyWriteAround - Write-Around Cache
	StrategyWriteAround CacheStrategy = "write_around"
)

// Cache defines the interface for all caching strategies
type Cache interface {
	// Cache-Aside (Strategy 1)
	GetOrLoad(ctx context.Context, key string, loader func() (interface{}, error), ttl time.Duration) (interface{}, error)

	// Write-Through (Strategy 2)
	SetWithDB(ctx context.Context, key string, value interface{}, ttl time.Duration, dbWriter func() error) error

	// Write-Behind (Strategy 3)
	SetAsync(ctx context.Context, key string, value interface{}, ttl time.Duration, dbWriter func() error) error

	// Read-Through (Strategy 4)
	GetWithLoader(ctx context.Context, key string, loader func() (interface{}, error), ttl time.Duration) (interface{}, error)

	// Write-Around (Strategy 5)
	InvalidateAndWrite(ctx context.Context, key string, dbWriter func() error) error

	// Utilities
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Close() error

	// Flush pending writes (for Write-Behind strategy)
	FlushPendingWrites(ctx context.Context) error
}

// RedisCache implements the Cache interface with all 5 strategies
type RedisCache struct {
	client             *redis.Client
	logger             *logrus.Logger
	writeBehindQueue   map[string]*queuedWrite
	writeBehindMutex   sync.RWMutex
	writeBehindTicker  *time.Ticker
	writeBehindStop    chan bool
	writeBehindEnabled bool
}

// queuedWrite represents a queued database write operation
type queuedWrite struct {
	Key       string
	Value     interface{}
	TTL       time.Duration
	DBWriter  func() error
	Timestamp time.Time
}

// RedisCacheConfig holds Redis cache configuration
type RedisCacheConfig struct {
	Addr               string
	Password           string
	DB                 int
	PoolSize           int
	MinIdleConns       int
	MaxRetries         int
	DialTimeout        time.Duration
	ReadTimeout        time.Duration
	WriteTimeout       time.Duration
	WriteBehindEnabled bool
	WriteBehindBatch   time.Duration
}

// Prometheus metrics for cache operations
var (
	cacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "conn_dict",
			Subsystem: "cache",
			Name:      "hits_total",
			Help:      "Total number of cache hits by strategy",
		},
		[]string{"strategy"},
	)

	cacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "conn_dict",
			Subsystem: "cache",
			Name:      "misses_total",
			Help:      "Total number of cache misses by strategy",
		},
		[]string{"strategy"},
	)

	cacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "conn_dict",
			Subsystem: "cache",
			Name:      "operation_duration_seconds",
			Help:      "Cache operation duration in seconds",
			Buckets:   []float64{0.0001, 0.0005, 0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0},
		},
		[]string{"operation", "strategy"},
	)

	cacheErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "conn_dict",
			Subsystem: "cache",
			Name:      "errors_total",
			Help:      "Total number of cache errors by type",
		},
		[]string{"operation", "error_type"},
	)

	writeBehindQueueSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Namespace: "conn_dict",
			Subsystem: "cache",
			Name:      "write_behind_queue_size",
			Help:      "Current size of write-behind queue",
		},
	)

	writeBehindFlushDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Namespace: "conn_dict",
			Subsystem: "cache",
			Name:      "write_behind_flush_duration_seconds",
			Help:      "Write-behind flush duration in seconds",
			Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
	)
)

// NewRedisCache creates a new Redis cache with all strategies enabled
func NewRedisCache(config RedisCacheConfig, logger *logrus.Logger) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:         config.Addr,
		Password:     config.Password,
		DB:           config.DB,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		MaxRetries:   config.MaxRetries,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Infof("Redis cache connected: addr=%s, db=%d, pool=%d", config.Addr, config.DB, config.PoolSize)

	cache := &RedisCache{
		client:             client,
		logger:             logger,
		writeBehindQueue:   make(map[string]*queuedWrite),
		writeBehindEnabled: config.WriteBehindEnabled,
	}

	// Start write-behind worker if enabled
	if config.WriteBehindEnabled {
		batchInterval := config.WriteBehindBatch
		if batchInterval == 0 {
			batchInterval = 10 * time.Second // Default: 10 seconds
		}
		cache.startWriteBehindWorker(batchInterval)
		logger.Infof("Write-behind cache worker started: batch_interval=%v", batchInterval)
	}

	return cache, nil
}

// ===========================
// STRATEGY 1: Cache-Aside (Lazy Loading)
// Use Case: Entry lookups
// TTL: 5 minutes
// Key pattern: entry:{key_id}
// ===========================

// GetOrLoad implements Cache-Aside pattern
// 1. Check cache
// 2. If miss, query database via loader
// 3. Store in cache
// 4. Return data
func (rc *RedisCache) GetOrLoad(ctx context.Context, key string, loader func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	start := time.Now()
	defer func() {
		cacheOperationDuration.WithLabelValues("get_or_load", string(StrategyCacheAside)).Observe(time.Since(start).Seconds())
	}()

	// Step 1: Check cache
	var result interface{}
	err := rc.Get(ctx, key, &result)
	if err == nil {
		cacheHitsTotal.WithLabelValues(string(StrategyCacheAside)).Inc()
		rc.logger.Debugf("[Cache-Aside] Cache hit: key=%s", key)
		return result, nil
	}

	if err != ErrCacheMiss {
		cacheErrorsTotal.WithLabelValues("get_or_load", "get_error").Inc()
		rc.logger.Warnf("[Cache-Aside] Cache error: key=%s, error=%v", key, err)
	}

	// Step 2: Cache miss - load from database
	cacheMissesTotal.WithLabelValues(string(StrategyCacheAside)).Inc()
	rc.logger.Debugf("[Cache-Aside] Cache miss: key=%s, loading from source", key)

	result, err = loader()
	if err != nil {
		cacheErrorsTotal.WithLabelValues("get_or_load", "loader_error").Inc()
		return nil, fmt.Errorf("loader failed: %w", err)
	}

	// Step 3: Store in cache
	if err := rc.Set(ctx, key, result, ttl); err != nil {
		cacheErrorsTotal.WithLabelValues("get_or_load", "set_error").Inc()
		rc.logger.Warnf("[Cache-Aside] Failed to cache result: key=%s, error=%v", key, err)
		// Return result anyway (cache failure is not critical)
	}

	return result, nil
}

// ===========================
// STRATEGY 2: Write-Through
// Use Case: Entry creation
// TTL: 5 minutes
// Key pattern: entry:{key_id}
// ===========================

// SetWithDB implements Write-Through pattern
// 1. Write to database first
// 2. Write to cache simultaneously
// 3. Return success
func (rc *RedisCache) SetWithDB(ctx context.Context, key string, value interface{}, ttl time.Duration, dbWriter func() error) error {
	start := time.Now()
	defer func() {
		cacheOperationDuration.WithLabelValues("set_with_db", string(StrategyWriteThrough)).Observe(time.Since(start).Seconds())
	}()

	rc.logger.Debugf("[Write-Through] Writing: key=%s", key)

	// Step 1: Write to database first (ensures data consistency)
	if err := dbWriter(); err != nil {
		cacheErrorsTotal.WithLabelValues("set_with_db", "db_write_error").Inc()
		return fmt.Errorf("database write failed: %w", err)
	}

	// Step 2: Write to cache
	if err := rc.Set(ctx, key, value, ttl); err != nil {
		cacheErrorsTotal.WithLabelValues("set_with_db", "cache_write_error").Inc()
		rc.logger.Warnf("[Write-Through] Cache write failed: key=%s, error=%v", key, err)
		// Database write succeeded, cache failure is not critical
	}

	return nil
}

// ===========================
// STRATEGY 3: Write-Behind (Write-Back)
// Use Case: High-frequency updates (metrics, counters)
// TTL: 10 minutes
// Key pattern: metrics:{participant_ispb}
// ===========================

// SetAsync implements Write-Behind pattern
// 1. Write to cache immediately
// 2. Queue database write
// 3. Batch write to DB every 10 seconds
func (rc *RedisCache) SetAsync(ctx context.Context, key string, value interface{}, ttl time.Duration, dbWriter func() error) error {
	start := time.Now()
	defer func() {
		cacheOperationDuration.WithLabelValues("set_async", string(StrategyWriteBehind)).Observe(time.Since(start).Seconds())
	}()

	if !rc.writeBehindEnabled {
		return fmt.Errorf("write-behind strategy is not enabled")
	}

	rc.logger.Debugf("[Write-Behind] Async write: key=%s", key)

	// Step 1: Write to cache immediately
	if err := rc.Set(ctx, key, value, ttl); err != nil {
		cacheErrorsTotal.WithLabelValues("set_async", "cache_write_error").Inc()
		return fmt.Errorf("cache write failed: %w", err)
	}

	// Step 2: Queue database write
	rc.writeBehindMutex.Lock()
	rc.writeBehindQueue[key] = &queuedWrite{
		Key:       key,
		Value:     value,
		TTL:       ttl,
		DBWriter:  dbWriter,
		Timestamp: time.Now(),
	}
	queueSize := len(rc.writeBehindQueue)
	rc.writeBehindMutex.Unlock()

	writeBehindQueueSize.Set(float64(queueSize))
	rc.logger.Debugf("[Write-Behind] Queued write: key=%s, queue_size=%d", key, queueSize)

	return nil
}

// startWriteBehindWorker starts background worker for write-behind pattern
func (rc *RedisCache) startWriteBehindWorker(batchInterval time.Duration) {
	rc.writeBehindTicker = time.NewTicker(batchInterval)
	rc.writeBehindStop = make(chan bool)

	go func() {
		for {
			select {
			case <-rc.writeBehindTicker.C:
				ctx := context.Background()
				if err := rc.FlushPendingWrites(ctx); err != nil {
					rc.logger.Errorf("[Write-Behind] Flush failed: %v", err)
				}
			case <-rc.writeBehindStop:
				rc.logger.Info("[Write-Behind] Worker stopped")
				return
			}
		}
	}()
}

// FlushPendingWrites flushes all pending writes to database
func (rc *RedisCache) FlushPendingWrites(ctx context.Context) error {
	start := time.Now()
	defer func() {
		writeBehindFlushDuration.Observe(time.Since(start).Seconds())
	}()

	rc.writeBehindMutex.Lock()
	queue := rc.writeBehindQueue
	rc.writeBehindQueue = make(map[string]*queuedWrite)
	rc.writeBehindMutex.Unlock()

	if len(queue) == 0 {
		return nil
	}

	rc.logger.Infof("[Write-Behind] Flushing %d pending writes", len(queue))

	// Batch write to database
	errors := make([]error, 0)
	successCount := 0

	for key, qw := range queue {
		if err := qw.DBWriter(); err != nil {
			errors = append(errors, fmt.Errorf("key=%s: %w", key, err))
			cacheErrorsTotal.WithLabelValues("flush_pending", "db_write_error").Inc()
			rc.logger.Errorf("[Write-Behind] DB write failed: key=%s, error=%v", key, err)
		} else {
			successCount++
		}
	}

	writeBehindQueueSize.Set(0)
	rc.logger.Infof("[Write-Behind] Flush completed: success=%d, errors=%d, duration=%v",
		successCount, len(errors), time.Since(start))

	if len(errors) > 0 {
		return fmt.Errorf("flush completed with %d errors", len(errors))
	}

	return nil
}

// ===========================
// STRATEGY 4: Read-Through
// Use Case: Configuration data
// TTL: 1 hour
// Key pattern: config:{key}
// ===========================

// GetWithLoader implements Read-Through pattern
// 1. Check cache
// 2. If miss, load from DB automatically via loader
// 3. Cache result
// 4. Return data
func (rc *RedisCache) GetWithLoader(ctx context.Context, key string, loader func() (interface{}, error), ttl time.Duration) (interface{}, error) {
	start := time.Now()
	defer func() {
		cacheOperationDuration.WithLabelValues("get_with_loader", string(StrategyReadThrough)).Observe(time.Since(start).Seconds())
	}()

	// Step 1: Check cache
	var result interface{}
	err := rc.Get(ctx, key, &result)
	if err == nil {
		cacheHitsTotal.WithLabelValues(string(StrategyReadThrough)).Inc()
		rc.logger.Debugf("[Read-Through] Cache hit: key=%s", key)
		return result, nil
	}

	if err != ErrCacheMiss {
		cacheErrorsTotal.WithLabelValues("get_with_loader", "get_error").Inc()
		rc.logger.Warnf("[Read-Through] Cache error: key=%s, error=%v", key, err)
	}

	// Step 2: Cache miss - load from database automatically
	cacheMissesTotal.WithLabelValues(string(StrategyReadThrough)).Inc()
	rc.logger.Debugf("[Read-Through] Cache miss: key=%s, loading from DB", key)

	result, err = loader()
	if err != nil {
		cacheErrorsTotal.WithLabelValues("get_with_loader", "loader_error").Inc()
		return nil, fmt.Errorf("loader failed: %w", err)
	}

	// Step 3: Cache result automatically
	if err := rc.Set(ctx, key, result, ttl); err != nil {
		cacheErrorsTotal.WithLabelValues("get_with_loader", "set_error").Inc()
		rc.logger.Warnf("[Read-Through] Failed to cache result: key=%s, error=%v", key, err)
	}

	return result, nil
}

// ===========================
// STRATEGY 5: Write-Around
// Use Case: Bulk operations (VSYNC)
// Pattern: Write directly to DB, invalidate cache
// ===========================

// InvalidateAndWrite implements Write-Around pattern
// 1. Write directly to database
// 2. Invalidate cache
// 3. Don't populate cache immediately (cache populated on next read)
func (rc *RedisCache) InvalidateAndWrite(ctx context.Context, key string, dbWriter func() error) error {
	start := time.Now()
	defer func() {
		cacheOperationDuration.WithLabelValues("invalidate_and_write", string(StrategyWriteAround)).Observe(time.Since(start).Seconds())
	}()

	rc.logger.Debugf("[Write-Around] Writing to DB and invalidating cache: key=%s", key)

	// Step 1: Write directly to database
	if err := dbWriter(); err != nil {
		cacheErrorsTotal.WithLabelValues("invalidate_and_write", "db_write_error").Inc()
		return fmt.Errorf("database write failed: %w", err)
	}

	// Step 2: Invalidate cache (don't populate)
	if err := rc.Delete(ctx, key); err != nil {
		cacheErrorsTotal.WithLabelValues("invalidate_and_write", "invalidate_error").Inc()
		rc.logger.Warnf("[Write-Around] Cache invalidation failed: key=%s, error=%v", key, err)
		// Database write succeeded, cache invalidation failure is not critical
	}

	rc.logger.Debugf("[Write-Around] Completed: key=%s", key)
	return nil
}

// ===========================
// Utility Methods
// ===========================

// Get retrieves a value from cache
func (rc *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return ErrCacheMiss
	}
	if err != nil {
		return fmt.Errorf("redis get error: %w", err)
	}

	// Unmarshal JSON
	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Set stores a value in cache with TTL
func (rc *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Marshal to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := rc.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (rc *RedisCache) Delete(ctx context.Context, key string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	return nil
}

// DeletePattern removes all keys matching a pattern (uses pipeline for efficiency)
func (rc *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	iter := rc.client.Scan(ctx, 0, pattern, 100).Iterator()
	pipe := rc.client.Pipeline()

	count := 0
	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
		count++
	}
	if err := iter.Err(); err != nil {
		return err
	}

	if count > 0 {
		if _, err := pipe.Exec(ctx); err != nil {
			return fmt.Errorf("redis pipeline exec error: %w", err)
		}
		rc.logger.Debugf("Cache pattern invalidated: pattern=%s, keys=%d", pattern, count)
	}

	return nil
}

// Exists checks if a key exists
func (rc *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Close closes the Redis client and stops write-behind worker
func (rc *RedisCache) Close() error {
	if rc.writeBehindEnabled {
		// Stop write-behind worker
		if rc.writeBehindStop != nil {
			close(rc.writeBehindStop)
		}
		if rc.writeBehindTicker != nil {
			rc.writeBehindTicker.Stop()
		}

		// Flush any pending writes
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := rc.FlushPendingWrites(ctx); err != nil {
			rc.logger.Errorf("Failed to flush pending writes on shutdown: %v", err)
		}
	}

	return rc.client.Close()
}

// ErrCacheMiss is returned when a key is not found in cache
var ErrCacheMiss = fmt.Errorf("cache miss")
