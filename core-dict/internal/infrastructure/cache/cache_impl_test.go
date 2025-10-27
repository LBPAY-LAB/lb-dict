package cache_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/lbpay-lab/core-dict/internal/infrastructure/cache"
)

type TestData struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

// setupRedisContainer starts a Redis container for testing
func setupRedisContainer(t *testing.T) (*cache.RedisClient, func()) {
	ctx := context.Background()

	// Start Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor: wait.ForLog("Ready to accept connections").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get container host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	// Create Redis config
	redisURL := fmt.Sprintf("redis://%s:%s/0", host, port.Port())
	config := &cache.RedisConfig{
		URL:              redisURL,
		Password:         "",
		DB:               0,
		PoolSize:         10,
		MinIdleConns:     2,
		MaxRetries:       3,
		ConnMaxIdleTime:  time.Minute * 5,
		ConnMaxLifetime:  time.Hour,
		DialTimeout:      time.Second * 10,
		ReadTimeout:      time.Second * 10,
		WriteTimeout:     time.Second * 10,
		TLSEnabled:       false,
		TLSSkipVerify:    false,
	}

	// Create Redis client with retry logic
	var client *cache.RedisClient
	maxRetries := 10
	retryDelay := 500 * time.Millisecond

	for i := 0; i < maxRetries; i++ {
		client, err = cache.NewRedisClient(config)
		if err == nil {
			// Test connection
			err = client.Ping(ctx)
			if err == nil {
				break
			}
			client.Close()
		}

		if i < maxRetries-1 {
			t.Logf("Redis connection attempt %d/%d failed, retrying in %v...", i+1, maxRetries, retryDelay)
			time.Sleep(retryDelay)
		}
	}
	require.NoError(t, err, "Failed to connect to Redis after %d retries", maxRetries)

	cleanup := func() {
		client.Close()
		container.Terminate(ctx)
	}

	return client, cleanup
}

func TestCache_Get_Hit(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.CacheAside, "test:")

	// Set a value
	data := TestData{Name: "test", Value: 123}
	err := c.Set(context.Background(), "key1", data, 5*time.Minute)
	require.NoError(t, err)

	// Get the value
	var result TestData
	err = c.Get(context.Background(), "key1", &result)
	assert.NoError(t, err)
	assert.Equal(t, data.Name, result.Name)
	assert.Equal(t, data.Value, result.Value)
}

func TestCache_Get_Miss(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.CacheAside, "test:")

	var result TestData
	err := c.Get(context.Background(), "nonexistent", &result)
	assert.Error(t, err)
	assert.Equal(t, cache.ErrCacheMiss, err)
}

func TestCache_Set_Success(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.CacheAside, "test:")

	data := TestData{Name: "test", Value: 456}
	err := c.Set(context.Background(), "key2", data, 5*time.Minute)
	assert.NoError(t, err)

	// Verify it was set
	var result TestData
	err = c.Get(context.Background(), "key2", &result)
	assert.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestCache_Delete_Success(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.CacheAside, "test:")

	// Set a value
	data := TestData{Name: "test", Value: 789}
	err := c.Set(context.Background(), "key3", data, 5*time.Minute)
	require.NoError(t, err)

	// Delete it
	err = c.Delete(context.Background(), "key3")
	assert.NoError(t, err)

	// Verify it's gone
	var result TestData
	err = c.Get(context.Background(), "key3", &result)
	assert.Error(t, err)
	assert.Equal(t, cache.ErrCacheMiss, err)
}

func TestCache_CacheAside_Pattern(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.CacheAside, "test:")
	handler := cache.NewCacheAsideHandler(c)

	loadCalled := false
	loader := func(ctx context.Context) (interface{}, error) {
		loadCalled = true
		return TestData{Name: "loaded", Value: 999}, nil
	}

	// First call - should load from DB
	var result TestData
	err := handler.GetOrLoad(context.Background(), "aside1", &result, 5*time.Minute, loader)
	assert.NoError(t, err)
	assert.True(t, loadCalled)
	assert.Equal(t, "loaded", result.Name)

	// Second call - should get from cache
	loadCalled = false
	var result2 TestData
	err = handler.GetOrLoad(context.Background(), "aside1", &result2, 5*time.Minute, loader)
	assert.NoError(t, err)
	assert.False(t, loadCalled) // Should not call loader
}

func TestCache_WriteThrough_Pattern(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.WriteThrough, "test:")
	handler := cache.NewWriteThroughHandler(c)

	writerCalled := false
	writer := func(ctx context.Context, value interface{}) error {
		writerCalled = true
		return nil
	}

	data := TestData{Name: "writethrough", Value: 111}
	err := handler.Write(context.Background(), "wt1", data, 5*time.Minute, writer)
	assert.NoError(t, err)
	assert.True(t, writerCalled)

	// Verify it's in cache
	var result TestData
	err = c.Get(context.Background(), "wt1", &result)
	assert.NoError(t, err)
	assert.Equal(t, data, result)
}

func TestCache_WriteBehind_Pattern(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.WriteBehind, "test:")
	handler := cache.NewWriteBehindHandler(c, 2)
	defer handler.Stop()

	writerCalled := false
	writer := func(ctx context.Context, value interface{}) error {
		writerCalled = true
		return nil
	}

	data := TestData{Name: "writebehind", Value: 222}
	err := handler.Write(context.Background(), "wb1", data, 5*time.Minute, writer)
	assert.NoError(t, err)

	// Value should be in cache immediately
	var result TestData
	err = c.Get(context.Background(), "wb1", &result)
	assert.NoError(t, err)
	assert.Equal(t, data, result)

	// Wait for async write
	time.Sleep(100 * time.Millisecond)
	assert.True(t, writerCalled)
}

func TestCache_ReadThrough_Pattern(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.ReadThrough, "test:")
	handler := cache.NewReadThroughHandler(c)

	loadCalled := false
	loader := func(ctx context.Context) (interface{}, error) {
		loadCalled = true
		return TestData{Name: "readthrough", Value: 333}, nil
	}

	// First read - should load from DB
	var result TestData
	err := handler.Read(context.Background(), "rt1", &result, 5*time.Minute, loader)
	assert.NoError(t, err)
	assert.True(t, loadCalled)

	// Second read - should get from cache
	loadCalled = false
	var result2 TestData
	err = handler.Read(context.Background(), "rt1", &result2, 5*time.Minute, loader)
	assert.NoError(t, err)
	assert.False(t, loadCalled)
}

func TestCache_WriteAround_Pattern(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.WriteAround, "test:")
	handler := cache.NewWriteAroundHandler(c)

	// Set initial value in cache
	initialData := TestData{Name: "initial", Value: 100}
	err := c.Set(context.Background(), "wa1", initialData, 5*time.Minute)
	require.NoError(t, err)

	writerCalled := false
	writer := func(ctx context.Context, value interface{}) error {
		writerCalled = true
		return nil
	}

	newData := TestData{Name: "updated", Value: 200}
	err = handler.Write(context.Background(), "wa1", newData, writer)
	assert.NoError(t, err)
	assert.True(t, writerCalled)

	// Cache should be invalidated (key deleted)
	var result TestData
	err = c.Get(context.Background(), "wa1", &result)
	assert.Error(t, err)
	assert.Equal(t, cache.ErrCacheMiss, err)
}

func TestCache_Invalidate_ByPattern(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	c := cache.NewCache(client, cache.CacheAside, "test:")

	// Set multiple keys with same prefix
	for i := 0; i < 3; i++ {
		data := TestData{Name: "pattern", Value: i}
		err := c.Set(context.Background(), fmt.Sprintf("pattern:%d", i), data, 5*time.Minute)
		require.NoError(t, err)
	}

	// Clear keys matching pattern
	err := c.Clear(context.Background(), "pattern:*")
	assert.NoError(t, err)

	// Verify all keys are gone
	for i := 0; i < 3; i++ {
		var result TestData
		err = c.Get(context.Background(), fmt.Sprintf("pattern:%d", i), &result)
		assert.Error(t, err)
	}
}
