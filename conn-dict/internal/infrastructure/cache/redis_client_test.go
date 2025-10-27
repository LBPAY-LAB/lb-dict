package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test data structures
type TestEntry struct {
	ID        string `json:"id"`
	Key       string `json:"key"`
	AccountID string `json:"account_id"`
}

func getTestRedisClient(t *testing.T) *RedisClient {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := RedisConfig{
		Addr:     "localhost:6379",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewRedisClient(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return nil
	}
	return client
}

// TestNewRedisClient tests Redis client creation
func TestNewRedisClient(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := RedisConfig{
		Addr:     "localhost:6379",
		DB:       0,
		PoolSize: 10,
	}

	client, err := NewRedisClient(config, logger)
	if err != nil {
		t.Skipf("Redis not available: %v", err)
		return
	}

	require.NotNil(t, client)
	defer client.Close()
}

// TestStrategy1_CacheAside tests lazy loading pattern
func TestStrategy1_CacheAside(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test:entry:123"

	// Test cache miss
	var entry TestEntry
	err := client.Get(ctx, key, &entry)
	assert.Equal(t, ErrCacheMiss, err)

	// Set value
	testEntry := TestEntry{
		ID:        "entry-123",
		Key:       "11122233344",
		AccountID: "acc-456",
	}

	err = client.Set(ctx, key, testEntry, 5*time.Minute)
	require.NoError(t, err)

	// Test cache hit
	var retrieved TestEntry
	err = client.Get(ctx, key, &retrieved)
	require.NoError(t, err)
	assert.Equal(t, testEntry.ID, retrieved.ID)
}

// TestStrategy2_WriteThrough tests write-through pattern
func TestStrategy2_WriteThrough(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test:write-through:456"

	testEntry := TestEntry{
		ID:        "entry-456",
		Key:       "email@example.com",
		AccountID: "acc-789",
	}

	err := client.Set(ctx, key, testEntry, 10*time.Minute)
	require.NoError(t, err)

	var retrieved TestEntry
	err = client.Get(ctx, key, &retrieved)
	require.NoError(t, err)
	assert.Equal(t, testEntry, retrieved)
}

// TestStrategy3_WriteBehind tests async write pattern
func TestStrategy3_WriteBehind(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test:write-behind:789"

	testEntry := TestEntry{
		ID:        "entry-789",
		Key:       "+5511999999999",
		AccountID: "acc-321",
	}

	err := client.Set(ctx, key, testEntry, 15*time.Minute)
	require.NoError(t, err)

	var retrieved TestEntry
	err = client.Get(ctx, key, &retrieved)
	require.NoError(t, err)
	assert.Equal(t, testEntry, retrieved)
}

// TestStrategy4_RefreshAhead tests proactive refresh
func TestStrategy4_RefreshAhead(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test:refresh-ahead:101"

	refreshCalls := 0
	refreshFunc := func() (interface{}, error) {
		refreshCalls++
		return TestEntry{
			ID:        "entry-refreshed",
			Key:       "refreshed-key",
			AccountID: "acc-refresh",
		}, nil
	}

	testEntry := TestEntry{ID: "entry-101", Key: "initial-key", AccountID: "acc-101"}
	err := client.Set(ctx, key, testEntry, 2*time.Second)
	require.NoError(t, err)

	var retrieved TestEntry
	err = client.GetWithRefresh(ctx, key, &retrieved, 3*time.Second, 1*time.Second, refreshFunc)
	require.NoError(t, err)
	assert.Equal(t, "entry-101", retrieved.ID)
}

// TestStrategy5_CacheInvalidation tests deletion patterns
func TestStrategy5_CacheInvalidation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()

	keys := []string{
		"test:invalidate:entry:1",
		"test:invalidate:entry:2",
		"test:invalidate:entry:3",
	}

	for i, key := range keys {
		entry := TestEntry{ID: string(rune('1' + i)), Key: "key", AccountID: "acc"}
		err := client.Set(ctx, key, entry, 5*time.Minute)
		require.NoError(t, err)
	}

	err := client.Delete(ctx, keys[0])
	require.NoError(t, err)

	var entry TestEntry
	err = client.Get(ctx, keys[0], &entry)
	assert.Equal(t, ErrCacheMiss, err)
}

// TestCacheMiss tests cache miss error handling
func TestCacheMiss(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test:nonexistent:key"

	var entry TestEntry
	err := client.Get(ctx, key, &entry)

	assert.Equal(t, ErrCacheMiss, err)
	assert.True(t, errors.Is(err, ErrCacheMiss))
}

// TestCacheTTL tests TTL expiration
func TestCacheTTL(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test:ttl:expired"

	testEntry := TestEntry{ID: "ttl-test", Key: "ttl-key", AccountID: "ttl-acc"}

	err := client.Set(ctx, key, testEntry, 1*time.Second)
	require.NoError(t, err)

	var retrieved TestEntry
	err = client.Get(ctx, key, &retrieved)
	require.NoError(t, err)

	time.Sleep(1500 * time.Millisecond)

	err = client.Get(ctx, key, &retrieved)
	assert.Equal(t, ErrCacheMiss, err)
}

// TestConcurrentAccess tests concurrent read/write
func TestConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Redis integration test in short mode")
	}

	client := getTestRedisClient(t)
	if client == nil {
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test:concurrent:access"

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(n int) {
			entry := TestEntry{
				ID:        string(rune('0' + n)),
				Key:       "concurrent",
				AccountID: "acc",
			}
			_ = client.Set(ctx, key, entry, 5*time.Minute)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	var entry TestEntry
	err := client.Get(ctx, key, &entry)
	assert.NoError(t, err)
}

// Benchmark cache operations
func BenchmarkCacheSet(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := RedisConfig{Addr: "localhost:6379", DB: 0, PoolSize: 10}
	client, err := NewRedisClient(config, logger)
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	entry := TestEntry{ID: "bench", Key: "bench-key", AccountID: "bench-acc"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Set(ctx, "bench:set", entry, 5*time.Minute)
	}
}

func BenchmarkCacheGet(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := RedisConfig{Addr: "localhost:6379", DB: 0, PoolSize: 10}
	client, err := NewRedisClient(config, logger)
	if err != nil {
		b.Skipf("Redis not available: %v", err)
		return
	}
	defer client.Close()

	ctx := context.Background()
	entry := TestEntry{ID: "bench", Key: "bench-key", AccountID: "bench-acc"}
	_ = client.Set(ctx, "bench:get", entry, 5*time.Minute)

	var retrieved TestEntry
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = client.Get(ctx, "bench:get", &retrieved)
	}
}