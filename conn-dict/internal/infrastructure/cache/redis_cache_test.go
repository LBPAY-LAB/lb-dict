package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntry represents a sample entry for testing
type TestEntry struct {
	ID        string    `json:"id"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at"`
}

func setupTestCache(t *testing.T) *RedisCache {
	logger := logrus.New()
	logger.SetLevel(logrus.DebugLevel)

	config := RedisCacheConfig{
		Addr:               "localhost:6379",
		Password:           "",
		DB:                 1, // Use DB 1 for testing
		PoolSize:           10,
		MinIdleConns:       2,
		MaxRetries:         3,
		DialTimeout:        5 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		WriteBehindEnabled: true,
		WriteBehindBatch:   2 * time.Second, // Shorter interval for testing
	}

	cache, err := NewRedisCache(config, logger)
	require.NoError(t, err, "Failed to create Redis cache")

	return cache
}

func cleanupTestCache(t *testing.T, cache *RedisCache) {
	ctx := context.Background()
	// Clean up test keys
	_ = cache.DeletePattern(ctx, "test:*")
	_ = cache.Close()
}

// ===========================
// STRATEGY 1: Cache-Aside Tests
// ===========================

func TestCacheAside_HitScenario(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:entry:cache-aside-hit"

	// Pre-populate cache
	entry := &TestEntry{
		ID:        "1",
		Key:       "test-key",
		Value:     "cached-value",
		CreatedAt: time.Now(),
	}
	err := cache.Set(ctx, key, entry, 5*time.Minute)
	require.NoError(t, err)

	// Test cache-aside with cache hit (loader should NOT be called)
	loaderCalled := false
	loader := func() (interface{}, error) {
		loaderCalled = true
		return nil, errors.New("loader should not be called on cache hit")
	}

	result, err := cache.GetOrLoad(ctx, key, loader, 5*time.Minute)
	require.NoError(t, err)
	assert.False(t, loaderCalled, "Loader should not be called on cache hit")

	// Verify result
	resultEntry := result.(*TestEntry)
	assert.Equal(t, "test-key", resultEntry.Key)
	assert.Equal(t, "cached-value", resultEntry.Value)

	t.Log("✓ Cache-Aside (Hit): Retrieved from cache successfully")
}

func TestCacheAside_MissScenario(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:entry:cache-aside-miss"

	// Loader that simulates database fetch
	loaderCalled := false
	entry := &TestEntry{
		ID:        "2",
		Key:       "test-key",
		Value:     "loaded-from-db",
		CreatedAt: time.Now(),
	}
	loader := func() (interface{}, error) {
		loaderCalled = true
		return entry, nil
	}

	// Test cache-aside with cache miss (loader SHOULD be called)
	result, err := cache.GetOrLoad(ctx, key, loader, 5*time.Minute)
	require.NoError(t, err)
	assert.True(t, loaderCalled, "Loader should be called on cache miss")

	// Verify result
	resultEntry := result.(*TestEntry)
	assert.Equal(t, "loaded-from-db", resultEntry.Value)

	// Verify it was cached
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists, "Result should be cached after load")

	t.Log("✓ Cache-Aside (Miss): Loaded from DB and cached successfully")
}

// ===========================
// STRATEGY 2: Write-Through Tests
// ===========================

func TestWriteThrough_Success(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:entry:write-through"

	entry := &TestEntry{
		ID:        "3",
		Key:       "test-key",
		Value:     "write-through-value",
		CreatedAt: time.Now(),
	}

	// DB writer that simulates database write
	dbWritten := false
	dbWriter := func() error {
		dbWritten = true
		return nil
	}

	// Test write-through
	err := cache.SetWithDB(ctx, key, entry, 5*time.Minute, dbWriter)
	require.NoError(t, err)
	assert.True(t, dbWritten, "Database should be written first")

	// Verify it was cached
	var cachedEntry TestEntry
	err = cache.Get(ctx, key, &cachedEntry)
	require.NoError(t, err)
	assert.Equal(t, "write-through-value", cachedEntry.Value)

	t.Log("✓ Write-Through: Wrote to DB and cache successfully")
}

func TestWriteThrough_DBFailure(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:entry:write-through-fail"

	entry := &TestEntry{
		ID:    "4",
		Key:   "test-key",
		Value: "should-not-be-cached",
	}

	// DB writer that fails
	dbWriter := func() error {
		return errors.New("database write failed")
	}

	// Test write-through with DB failure
	err := cache.SetWithDB(ctx, key, entry, 5*time.Minute, dbWriter)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database write failed")

	// Verify it was NOT cached (because DB write failed)
	exists, _ := cache.Exists(ctx, key)
	assert.False(t, exists, "Entry should not be cached if DB write fails")

	t.Log("✓ Write-Through: Correctly failed on DB error, cache not populated")
}

// ===========================
// STRATEGY 3: Write-Behind Tests
// ===========================

func TestWriteBehind_AsyncWrite(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:metrics:write-behind"

	entry := &TestEntry{
		ID:        "5",
		Key:       "metrics-key",
		Value:     "async-value",
		CreatedAt: time.Now(),
	}

	// DB writer that simulates database write
	dbWritten := false
	dbWriter := func() error {
		dbWritten = true
		return nil
	}

	// Test write-behind (async)
	err := cache.SetAsync(ctx, key, entry, 10*time.Minute, dbWriter)
	require.NoError(t, err)

	// Verify it was cached immediately
	var cachedEntry TestEntry
	err = cache.Get(ctx, key, &cachedEntry)
	require.NoError(t, err)
	assert.Equal(t, "async-value", cachedEntry.Value)

	// DB should NOT be written yet (it's queued)
	assert.False(t, dbWritten, "DB should not be written immediately (async)")

	// Wait for flush (2 seconds batch interval)
	time.Sleep(3 * time.Second)

	// Now DB should be written
	assert.True(t, dbWritten, "DB should be written after batch flush")

	t.Log("✓ Write-Behind: Cached immediately, DB written async")
}

func TestWriteBehind_ManualFlush(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()

	// Queue multiple writes
	writeCount := 0
	for i := 0; i < 5; i++ {
		key := fmt.Sprintf("test:metrics:flush:%d", i)
		entry := &TestEntry{
			ID:    fmt.Sprintf("%d", i),
			Value: fmt.Sprintf("value-%d", i),
		}

		dbWriter := func() error {
			writeCount++
			return nil
		}

		err := cache.SetAsync(ctx, key, entry, 10*time.Minute, dbWriter)
		require.NoError(t, err)
	}

	// Flush manually
	err := cache.FlushPendingWrites(ctx)
	require.NoError(t, err)
	assert.Equal(t, 5, writeCount, "All 5 writes should be flushed")

	t.Log("✓ Write-Behind: Manual flush processed all queued writes")
}

// ===========================
// STRATEGY 4: Read-Through Tests
// ===========================

func TestReadThrough_HitScenario(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:config:read-through-hit"

	// Pre-populate cache
	config := &TestEntry{
		ID:    "config-1",
		Key:   "max_retries",
		Value: "5",
	}
	err := cache.Set(ctx, key, config, 1*time.Hour)
	require.NoError(t, err)

	// Test read-through with cache hit
	loaderCalled := false
	loader := func() (interface{}, error) {
		loaderCalled = true
		return nil, errors.New("loader should not be called on cache hit")
	}

	result, err := cache.GetWithLoader(ctx, key, loader, 1*time.Hour)
	require.NoError(t, err)
	assert.False(t, loaderCalled, "Loader should not be called on cache hit")

	resultConfig := result.(*TestEntry)
	assert.Equal(t, "5", resultConfig.Value)

	t.Log("✓ Read-Through (Hit): Retrieved from cache successfully")
}

func TestReadThrough_MissScenario(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:config:read-through-miss"

	// Loader that simulates config fetch
	loaderCalled := false
	config := &TestEntry{
		ID:    "config-2",
		Key:   "timeout",
		Value: "30s",
	}
	loader := func() (interface{}, error) {
		loaderCalled = true
		return config, nil
	}

	// Test read-through with cache miss
	result, err := cache.GetWithLoader(ctx, key, loader, 1*time.Hour)
	require.NoError(t, err)
	assert.True(t, loaderCalled, "Loader should be called on cache miss")

	resultConfig := result.(*TestEntry)
	assert.Equal(t, "30s", resultConfig.Value)

	// Verify automatic caching
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists, "Config should be automatically cached")

	t.Log("✓ Read-Through (Miss): Loaded and automatically cached")
}

// ===========================
// STRATEGY 5: Write-Around Tests
// ===========================

func TestWriteAround_BulkOperation(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:entry:write-around"

	// Pre-populate cache with old data
	oldEntry := &TestEntry{
		ID:    "old",
		Value: "old-value",
	}
	err := cache.Set(ctx, key, oldEntry, 5*time.Minute)
	require.NoError(t, err)

	// DB writer for bulk operation
	dbWritten := false
	dbWriter := func() error {
		dbWritten = true
		// Simulate bulk write to DB
		return nil
	}

	// Test write-around (invalidate cache, write to DB)
	err = cache.InvalidateAndWrite(ctx, key, dbWriter)
	require.NoError(t, err)
	assert.True(t, dbWritten, "DB should be written")

	// Verify cache was invalidated
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists, "Cache should be invalidated (not exist)")

	t.Log("✓ Write-Around: DB written, cache invalidated")
}

func TestWriteAround_CacheRepopulationOnRead(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:entry:write-around-read"

	// Simulate write-around
	newEntry := &TestEntry{
		ID:    "new",
		Value: "new-value-from-db",
	}
	dbWriter := func() error {
		return nil
	}

	err := cache.InvalidateAndWrite(ctx, key, dbWriter)
	require.NoError(t, err)

	// Now simulate cache population on next read (using cache-aside)
	loader := func() (interface{}, error) {
		// Simulate fetching the new value from DB
		return newEntry, nil
	}

	result, err := cache.GetOrLoad(ctx, key, loader, 5*time.Minute)
	require.NoError(t, err)

	resultEntry := result.(*TestEntry)
	assert.Equal(t, "new-value-from-db", resultEntry.Value)

	// Verify cache is now populated
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists, "Cache should be repopulated on read")

	t.Log("✓ Write-Around: Cache repopulated on subsequent read")
}

// ===========================
// Utility Tests
// ===========================

func TestDeletePattern_MultipleKeys(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()

	// Create multiple keys with pattern
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("test:pattern:entry:%d", i)
		entry := &TestEntry{ID: fmt.Sprintf("%d", i)}
		err := cache.Set(ctx, key, entry, 5*time.Minute)
		require.NoError(t, err)
	}

	// Delete all keys matching pattern
	err := cache.DeletePattern(ctx, "test:pattern:entry:*")
	require.NoError(t, err)

	// Verify all keys are deleted
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("test:pattern:entry:%d", i)
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists, fmt.Sprintf("Key %s should be deleted", key))
	}

	t.Log("✓ DeletePattern: Successfully deleted all matching keys")
}

func TestCache_TTLExpiration(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()
	key := "test:entry:ttl"

	entry := &TestEntry{
		ID:    "ttl-test",
		Value: "expires-soon",
	}

	// Set with short TTL
	err := cache.Set(ctx, key, entry, 1*time.Second)
	require.NoError(t, err)

	// Verify it exists
	exists, err := cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Wait for expiration
	time.Sleep(2 * time.Second)

	// Verify it expired
	exists, err = cache.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists, "Key should be expired")

	t.Log("✓ TTL: Key expired correctly")
}

// ===========================
// Integration Test
// ===========================

func TestAllStrategies_Integration(t *testing.T) {
	cache := setupTestCache(t)
	defer cleanupTestCache(t, cache)

	ctx := context.Background()

	t.Run("Full Workflow", func(t *testing.T) {
		// 1. Cache-Aside: Lookup entry (miss, load from DB)
		entryKey := "test:integration:entry:1"
		entry := &TestEntry{ID: "1", Key: "user-123", Value: "John Doe"}

		result, err := cache.GetOrLoad(ctx, entryKey, func() (interface{}, error) {
			return entry, nil
		}, 5*time.Minute)
		require.NoError(t, err)
		assert.Equal(t, "John Doe", result.(*TestEntry).Value)
		t.Log("  1. Cache-Aside: Entry loaded and cached ✓")

		// 2. Write-Through: Create new entry
		newEntryKey := "test:integration:entry:2"
		newEntry := &TestEntry{ID: "2", Key: "user-456", Value: "Jane Smith"}

		err = cache.SetWithDB(ctx, newEntryKey, newEntry, 5*time.Minute, func() error {
			return nil // Simulate DB write
		})
		require.NoError(t, err)
		t.Log("  2. Write-Through: New entry written to DB and cache ✓")

		// 3. Write-Behind: Update metrics
		metricsKey := "test:integration:metrics:participant-123"
		metrics := &TestEntry{ID: "m1", Value: "counter:42"}

		err = cache.SetAsync(ctx, metricsKey, metrics, 10*time.Minute, func() error {
			return nil // Simulate async DB write
		})
		require.NoError(t, err)
		t.Log("  3. Write-Behind: Metrics cached, DB write queued ✓")

		// 4. Read-Through: Fetch config
		configKey := "test:integration:config:max-retries"
		config := &TestEntry{ID: "c1", Key: "max_retries", Value: "5"}

		result, err = cache.GetWithLoader(ctx, configKey, func() (interface{}, error) {
			return config, nil
		}, 1*time.Hour)
		require.NoError(t, err)
		t.Log("  4. Read-Through: Config loaded and cached ✓")

		// 5. Write-Around: Bulk sync operation
		bulkKey := "test:integration:entry:bulk"
		err = cache.InvalidateAndWrite(ctx, bulkKey, func() error {
			return nil // Simulate bulk DB write
		})
		require.NoError(t, err)
		t.Log("  5. Write-Around: Bulk operation completed, cache invalidated ✓")

		t.Log("\n✅ All 5 caching strategies tested successfully!")
	})
}
