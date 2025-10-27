package cache_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/infrastructure/cache"
)

func TestRateLimiter_Allow_UnderLimit(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	limiter := cache.NewRateLimiter(client, 5, time.Second)

	// Make 5 requests (under limit)
	for i := 0; i < 5; i++ {
		allowed, err := limiter.Allow(context.Background(), "user:123")
		assert.NoError(t, err)
		assert.True(t, allowed)
	}
}

func TestRateLimiter_Deny_OverLimit(t *testing.T) {
	client, cleanup := setupRedisContainer(t)
	defer cleanup()

	limiter := cache.NewRateLimiter(client, 3, time.Second)

	// Make 3 requests (at limit)
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(context.Background(), "user:456")
		require.NoError(t, err)
		assert.True(t, allowed)
	}

	// 4th request should be denied
	allowed, err := limiter.Allow(context.Background(), "user:456")
	assert.NoError(t, err)
	assert.False(t, allowed)

	// Wait for window to reset
	time.Sleep(1100 * time.Millisecond)

	// Should be allowed again
	allowed, err = limiter.Allow(context.Background(), "user:456")
	assert.NoError(t, err)
	assert.True(t, allowed)
}
