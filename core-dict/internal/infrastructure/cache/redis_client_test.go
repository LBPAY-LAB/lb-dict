package cache

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisClient_Connection(t *testing.T) {
	// This test requires a running Redis instance
	// Skip in CI if Redis is not available
	config := DefaultRedisConfig()
	config.URL = "redis://localhost:6379/15" // Use DB 15 for tests

	client, err := NewRedisClient(config)
	if err != nil {
		t.Skip("Redis not available, skipping test")
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Test Ping
	err = client.Ping(ctx)
	require.NoError(t, err)
}

func TestRedisClient_SetGet(t *testing.T) {
	config := DefaultRedisConfig()
	config.URL = "redis://localhost:6379/15"

	client, err := NewRedisClient(config)
	if err != nil {
		t.Skip("Redis not available, skipping test")
		return
	}
	defer client.Close()

	ctx := context.Background()

	// Test Set and Get
	key := "test-key"
	value := "test-value"

	err = client.Set(ctx, key, value, 5*time.Minute)
	require.NoError(t, err)

	result, err := client.Get(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, value, result)

	// Cleanup
	client.Del(ctx, key)
}

func TestRedisClient_Exists(t *testing.T) {
	config := DefaultRedisConfig()
	config.URL = "redis://localhost:6379/15"

	client, err := NewRedisClient(config)
	if err != nil {
		t.Skip("Redis not available, skipping test")
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-exists-key"

	// Key should not exist
	exists, err := client.Exists(ctx, key)
	require.NoError(t, err)
	assert.False(t, exists)

	// Set key
	client.Set(ctx, key, "value", time.Minute)

	// Key should exist
	exists, err = client.Exists(ctx, key)
	require.NoError(t, err)
	assert.True(t, exists)

	// Cleanup
	client.Del(ctx, key)
}

func TestRedisClient_Incr(t *testing.T) {
	config := DefaultRedisConfig()
	config.URL = "redis://localhost:6379/15"

	client, err := NewRedisClient(config)
	if err != nil {
		t.Skip("Redis not available, skipping test")
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-counter"

	// First increment should return 1
	val, err := client.Incr(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(1), val)

	// Second increment should return 2
	val, err = client.Incr(ctx, key)
	require.NoError(t, err)
	assert.Equal(t, int64(2), val)

	// IncrBy 5 should return 7
	val, err = client.IncrBy(ctx, key, 5)
	require.NoError(t, err)
	assert.Equal(t, int64(7), val)

	// Cleanup
	client.Del(ctx, key)
}

func TestRedisClient_SetNX(t *testing.T) {
	config := DefaultRedisConfig()
	config.URL = "redis://localhost:6379/15"

	client, err := NewRedisClient(config)
	if err != nil {
		t.Skip("Redis not available, skipping test")
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-lock"

	// First SetNX should succeed
	acquired, err := client.SetNX(ctx, key, "holder1", time.Minute)
	require.NoError(t, err)
	assert.True(t, acquired)

	// Second SetNX should fail (key already exists)
	acquired, err = client.SetNX(ctx, key, "holder2", time.Minute)
	require.NoError(t, err)
	assert.False(t, acquired)

	// Cleanup
	client.Del(ctx, key)
}

func TestRedisClient_TTL(t *testing.T) {
	config := DefaultRedisConfig()
	config.URL = "redis://localhost:6379/15"

	client, err := NewRedisClient(config)
	if err != nil {
		t.Skip("Redis not available, skipping test")
		return
	}
	defer client.Close()

	ctx := context.Background()
	key := "test-ttl"

	// Set key with 10 second TTL
	client.Set(ctx, key, "value", 10*time.Second)

	// Check TTL
	ttl, err := client.TTL(ctx, key)
	require.NoError(t, err)
	assert.Greater(t, ttl, 5*time.Second)
	assert.LessOrEqual(t, ttl, 10*time.Second)

	// Cleanup
	client.Del(ctx, key)
}
