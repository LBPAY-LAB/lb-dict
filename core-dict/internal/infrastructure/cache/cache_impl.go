// Package cache provides caching implementations with multiple strategies
package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// CacheStrategy represents different caching strategies
type CacheStrategy string

const (
	// CacheAside (Lazy Loading) - Read from cache, if miss then read from DB and populate cache
	CacheAside CacheStrategy = "cache-aside"

	// WriteThrough - Write to cache and DB synchronously
	WriteThrough CacheStrategy = "write-through"

	// WriteBehind (Write Back) - Write to cache first, DB asynchronously
	WriteBehind CacheStrategy = "write-behind"

	// ReadThrough - Cache reads from DB automatically on miss
	ReadThrough CacheStrategy = "read-through"

	// WriteAround - Write to DB, invalidate cache (don't update cache on write)
	WriteAround CacheStrategy = "write-around"
)

// Cache interface defines caching operations
type Cache interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string, dest interface{}) error

	// Set stores a value in cache
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error

	// Delete removes a key from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// Clear removes all keys matching a pattern
	Clear(ctx context.Context, pattern string) error

	// GetStrategy returns the current caching strategy
	GetStrategy() CacheStrategy

	// Close closes the cache connection
	Close() error
}

// cacheImpl implements the Cache interface using Redis
type cacheImpl struct {
	client   *RedisClient
	strategy CacheStrategy
	prefix   string
}

// NewCache creates a new cache instance with the specified strategy
func NewCache(client *RedisClient, strategy CacheStrategy, prefix string) Cache {
	if prefix == "" {
		prefix = "core-dict:"
	}
	return &cacheImpl{
		client:   client,
		strategy: strategy,
		prefix:   prefix,
	}
}

// Get retrieves a value from cache (implements Cache-Aside pattern)
func (c *cacheImpl) Get(ctx context.Context, key string, dest interface{}) error {
	fullKey := c.makeKey(key)

	data, err := c.client.Get(ctx, fullKey)
	if err != nil {
		if err == ErrCacheMiss {
			return ErrCacheMiss
		}
		return fmt.Errorf("cache get failed: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		return fmt.Errorf("failed to unmarshal cached data: %w", err)
	}

	return nil
}

// Set stores a value in cache
func (c *cacheImpl) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	fullKey := c.makeKey(key)

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := c.client.Set(ctx, fullKey, data, ttl); err != nil {
		return fmt.Errorf("cache set failed: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (c *cacheImpl) Delete(ctx context.Context, key string) error {
	fullKey := c.makeKey(key)
	return c.client.Del(ctx, fullKey)
}

// Exists checks if a key exists in cache
func (c *cacheImpl) Exists(ctx context.Context, key string) (bool, error) {
	fullKey := c.makeKey(key)
	return c.client.Exists(ctx, fullKey)
}

// Clear removes all keys matching a pattern
func (c *cacheImpl) Clear(ctx context.Context, pattern string) error {
	fullPattern := c.makeKey(pattern)
	keys, err := c.client.Keys(ctx, fullPattern)
	if err != nil {
		return fmt.Errorf("failed to get keys: %w", err)
	}

	if len(keys) == 0 {
		return nil
	}

	return c.client.Del(ctx, keys...)
}

// GetStrategy returns the current caching strategy
func (c *cacheImpl) GetStrategy() CacheStrategy {
	return c.strategy
}

// Close closes the cache connection
func (c *cacheImpl) Close() error {
	return c.client.Close()
}

// makeKey creates a full cache key with prefix
func (c *cacheImpl) makeKey(key string) string {
	return c.prefix + key
}

// CacheAsideHandler implements Cache-Aside (Lazy Loading) pattern
type CacheAsideHandler struct {
	cache Cache
}

// NewCacheAsideHandler creates a Cache-Aside handler
func NewCacheAsideHandler(cache Cache) *CacheAsideHandler {
	return &CacheAsideHandler{cache: cache}
}

// GetOrLoad retrieves from cache or loads from DB on miss
func (c *CacheAsideHandler) GetOrLoad(
	ctx context.Context,
	key string,
	dest interface{},
	ttl time.Duration,
	loader func(ctx context.Context) (interface{}, error),
) error {
	// Try to get from cache
	err := c.cache.Get(ctx, key, dest)
	if err == nil {
		return nil // Cache hit
	}

	if err != ErrCacheMiss {
		// Cache error, fallback to loader
		data, loadErr := loader(ctx)
		if loadErr != nil {
			return loadErr
		}
		// Try to unmarshal into dest
		if jsonData, ok := data.([]byte); ok {
			return json.Unmarshal(jsonData, dest)
		}
		// Direct assignment for same type
		*dest.(*interface{}) = data
		return nil
	}

	// Cache miss - load from DB
	data, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("loader failed: %w", err)
	}

	// Store in cache (best-effort, don't fail if cache set fails)
	_ = c.cache.Set(ctx, key, data, ttl)

	// Unmarshal into dest
	if jsonData, ok := data.([]byte); ok {
		return json.Unmarshal(jsonData, dest)
	}
	// Direct assignment
	*dest.(*interface{}) = data
	return nil
}

// WriteThroughHandler implements Write-Through pattern
type WriteThroughHandler struct {
	cache Cache
}

// NewWriteThroughHandler creates a Write-Through handler
func NewWriteThroughHandler(cache Cache) *WriteThroughHandler {
	return &WriteThroughHandler{cache: cache}
}

// Write writes to cache and DB synchronously
func (w *WriteThroughHandler) Write(
	ctx context.Context,
	key string,
	value interface{},
	ttl time.Duration,
	writer func(ctx context.Context, value interface{}) error,
) error {
	// Write to DB first
	if err := writer(ctx, value); err != nil {
		return fmt.Errorf("DB write failed: %w", err)
	}

	// Write to cache
	if err := w.cache.Set(ctx, key, value, ttl); err != nil {
		// Cache write failed, but DB write succeeded
		// Log error but don't fail the operation
		return fmt.Errorf("cache write failed (DB write succeeded): %w", err)
	}

	return nil
}

// WriteBehindHandler implements Write-Behind (Write-Back) pattern
type WriteBehindHandler struct {
	cache    Cache
	writeQ   chan writeOperation
	stopChan chan struct{}
}

type writeOperation struct {
	ctx    context.Context
	key    string
	value  interface{}
	writer func(ctx context.Context, value interface{}) error
}

// NewWriteBehindHandler creates a Write-Behind handler with async DB writes
func NewWriteBehindHandler(cache Cache, workers int) *WriteBehindHandler {
	h := &WriteBehindHandler{
		cache:    cache,
		writeQ:   make(chan writeOperation, 1000),
		stopChan: make(chan struct{}),
	}

	// Start background workers
	for i := 0; i < workers; i++ {
		go h.worker()
	}

	return h
}

// Write writes to cache immediately, DB asynchronously
func (w *WriteBehindHandler) Write(
	ctx context.Context,
	key string,
	value interface{},
	ttl time.Duration,
	writer func(ctx context.Context, value interface{}) error,
) error {
	// Write to cache immediately
	if err := w.cache.Set(ctx, key, value, ttl); err != nil {
		return fmt.Errorf("cache write failed: %w", err)
	}

	// Queue DB write for async processing
	select {
	case w.writeQ <- writeOperation{ctx: ctx, key: key, value: value, writer: writer}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// worker processes async DB writes
func (w *WriteBehindHandler) worker() {
	for {
		select {
		case op := <-w.writeQ:
			// Write to DB
			if err := op.writer(op.ctx, op.value); err != nil {
				// TODO: implement retry logic or DLQ
				// For now, just log the error
				fmt.Printf("async DB write failed for key %s: %v\n", op.key, err)
			}
		case <-w.stopChan:
			return
		}
	}
}

// Stop stops the write-behind workers
func (w *WriteBehindHandler) Stop() {
	close(w.stopChan)
}

// ReadThroughHandler implements Read-Through pattern
type ReadThroughHandler struct {
	cache Cache
}

// NewReadThroughHandler creates a Read-Through handler
func NewReadThroughHandler(cache Cache) *ReadThroughHandler {
	return &ReadThroughHandler{cache: cache}
}

// Read reads from cache, auto-loads from DB on miss
func (r *ReadThroughHandler) Read(
	ctx context.Context,
	key string,
	dest interface{},
	ttl time.Duration,
	loader func(ctx context.Context) (interface{}, error),
) error {
	// Try cache first
	err := r.cache.Get(ctx, key, dest)
	if err == nil {
		return nil
	}

	if err != ErrCacheMiss {
		return fmt.Errorf("cache read failed: %w", err)
	}

	// Cache miss - load from DB
	data, err := loader(ctx)
	if err != nil {
		return fmt.Errorf("DB load failed: %w", err)
	}

	// Store in cache
	if err := r.cache.Set(ctx, key, data, ttl); err != nil {
		// Cache set failed, but we have the data
		// Continue with the data
	}

	// Unmarshal into dest
	if jsonData, ok := data.([]byte); ok {
		return json.Unmarshal(jsonData, dest)
	}
	*dest.(*interface{}) = data
	return nil
}

// WriteAroundHandler implements Write-Around pattern
type WriteAroundHandler struct {
	cache Cache
}

// NewWriteAroundHandler creates a Write-Around handler
func NewWriteAroundHandler(cache Cache) *WriteAroundHandler {
	return &WriteAroundHandler{cache: cache}
}

// Write writes to DB and invalidates cache (doesn't update cache)
func (w *WriteAroundHandler) Write(
	ctx context.Context,
	key string,
	value interface{},
	writer func(ctx context.Context, value interface{}) error,
) error {
	// Write to DB
	if err := writer(ctx, value); err != nil {
		return fmt.Errorf("DB write failed: %w", err)
	}

	// Invalidate cache (delete key)
	if err := w.cache.Delete(ctx, key); err != nil {
		// Cache invalidation failed, log but don't fail
		fmt.Printf("cache invalidation failed for key %s: %v\n", key, err)
	}

	return nil
}

// CacheKeyBuilder helps build standardized cache keys
type CacheKeyBuilder struct {
	prefix string
}

// NewCacheKeyBuilder creates a new cache key builder
func NewCacheKeyBuilder(prefix string) *CacheKeyBuilder {
	return &CacheKeyBuilder{prefix: prefix}
}

// Build builds a cache key from parts
func (b *CacheKeyBuilder) Build(parts ...string) string {
	key := b.prefix
	for _, part := range parts {
		key += ":" + part
	}
	return key
}

// EntryKey builds a cache key for a DICT entry
func (b *CacheKeyBuilder) EntryKey(key string) string {
	return b.Build("entry", key)
}

// AccountKey builds a cache key for an account
func (b *CacheKeyBuilder) AccountKey(ispb, branch, account, accountType string) string {
	return b.Build("account", ispb, branch, account, accountType)
}

// ClaimKey builds a cache key for a claim
func (b *CacheKeyBuilder) ClaimKey(claimID string) string {
	return b.Build("claim", claimID)
}

// StatisticsKey builds a cache key for statistics
func (b *CacheKeyBuilder) StatisticsKey() string {
	return b.Build("statistics")
}
