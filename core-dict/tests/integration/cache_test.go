package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/tests/testhelpers"
)

func TestIntegration_Redis_CacheAside_Pattern(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create entry in database
	entry := testhelpers.NewValidEntry()
	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, "ACTIVE", entry.UserID)
	require.NoError(t, err)

	cacheKey := fmt.Sprintf("entry:%s", entry.ID)

	// Act - Step 1: Cache miss, load from DB
	_, err = env.Redis.Get(env.Ctx, cacheKey).Result()
	assert.Error(t, err, "Should be cache miss")
	assert.Equal(t, redis.Nil, err, "Should be Redis Nil error")

	// Load from database (simulate)
	var dbEntry testhelpers.ValidEntry
	err = env.PG.QueryRow(env.Ctx, `
		SELECT id, key_type, key_value, status FROM entries WHERE id = $1
	`, entry.ID).Scan(&dbEntry.ID, &dbEntry.KeyType, &dbEntry.KeyValue, &dbEntry.Status)
	require.NoError(t, err)

	// Step 2: Store in cache (Cache-Aside pattern)
	entryJSON := fmt.Sprintf(`{"id":"%s","key_type":"%s","key_value":"%s","status":"%s"}`,
		dbEntry.ID, dbEntry.KeyType, dbEntry.KeyValue, dbEntry.Status)
	err = env.Redis.Set(env.Ctx, cacheKey, entryJSON, 5*time.Minute).Err()
	require.NoError(t, err)

	// Step 3: Cache hit
	cachedData, err := env.Redis.Get(env.Ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Contains(t, cachedData, entry.ID, "Should get cached data")

	// Assert - Verify cache hit doesn't query DB
	// In real scenario, we'd track DB queries
	t.Logf("Cache hit successful: %s", cachedData)
}

func TestIntegration_Redis_WriteThrough_Pattern(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange
	entry := testhelpers.NewValidEntry()
	cacheKey := fmt.Sprintf("entry:%s", entry.ID)

	// Act - Write-Through: Write to DB and cache simultaneously
	// Step 1: Write to database
	_, err := env.PG.Exec(env.Ctx, `
		INSERT INTO entries (id, key_type, key_value, account_id, ispb, status, user_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, entry.ID, entry.KeyType, entry.KeyValue, entry.AccountID, entry.ISPB, "ACTIVE", entry.UserID)
	require.NoError(t, err)

	// Step 2: Write to cache (Write-Through pattern)
	entryJSON := fmt.Sprintf(`{"id":"%s","key_type":"%s","status":"ACTIVE"}`, entry.ID, entry.KeyType)
	err = env.Redis.Set(env.Ctx, cacheKey, entryJSON, 5*time.Minute).Err()
	require.NoError(t, err)

	// Assert - Verify both DB and cache have the data
	// Check DB
	var dbID string
	err = env.PG.QueryRow(env.Ctx, `SELECT id FROM entries WHERE id = $1`, entry.ID).Scan(&dbID)
	require.NoError(t, err)
	assert.Equal(t, entry.ID, dbID)

	// Check cache
	cachedData, err := env.Redis.Get(env.Ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Contains(t, cachedData, entry.ID)

	// Step 3: Update with Write-Through
	_, err = env.PG.Exec(env.Ctx, `UPDATE entries SET status = 'BLOCKED' WHERE id = $1`, entry.ID)
	require.NoError(t, err)

	// Update cache
	updatedJSON := fmt.Sprintf(`{"id":"%s","key_type":"%s","status":"BLOCKED"}`, entry.ID, entry.KeyType)
	err = env.Redis.Set(env.Ctx, cacheKey, updatedJSON, 5*time.Minute).Err()
	require.NoError(t, err)

	// Verify update
	cachedUpdated, err := env.Redis.Get(env.Ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Contains(t, cachedUpdated, "BLOCKED")
}

func TestIntegration_Redis_RateLimiter_100RPS(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Rate limiter configuration
	userID := "test-user-123"
	rateLimitKey := fmt.Sprintf("ratelimit:%s", userID)
	maxRequests := 100
	window := 1 * time.Second

	// Act - Simulate 150 requests
	allowed := 0
	blocked := 0

	for i := 0; i < 150; i++ {
		// Increment counter
		count, err := env.Redis.Incr(env.Ctx, rateLimitKey).Result()
		require.NoError(t, err)

		// Set expiry on first request
		if count == 1 {
			err = env.Redis.Expire(env.Ctx, rateLimitKey, window).Err()
			require.NoError(t, err)
		}

		// Check if within limit
		if count <= int64(maxRequests) {
			allowed++
		} else {
			blocked++
		}
	}

	// Assert
	assert.Equal(t, 100, allowed, "Should allow 100 requests")
	assert.Equal(t, 50, blocked, "Should block 50 requests")

	// Verify Redis counter
	counter, err := env.Redis.Get(env.Ctx, rateLimitKey).Int64()
	require.NoError(t, err)
	assert.Equal(t, int64(150), counter)

	// Wait for window to expire
	time.Sleep(window + 100*time.Millisecond)

	// Counter should be expired
	_, err = env.Redis.Get(env.Ctx, rateLimitKey).Result()
	assert.Error(t, err, "Counter should be expired")
}

func TestIntegration_Redis_Invalidation_ByPattern(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Create multiple cached entries
	accountID := uuid.NewString()
	entries := []testhelpers.ValidEntry{}

	for i := 0; i < 5; i++ {
		entry := testhelpers.NewValidCPFEntry(fmt.Sprintf("1234567890%d", i))
		entry.AccountID = accountID
		entries = append(entries, entry)

		// Cache each entry
		cacheKey := fmt.Sprintf("entry:%s", entry.ID)
		entryJSON := fmt.Sprintf(`{"id":"%s","account_id":"%s"}`, entry.ID, accountID)
		err := env.Redis.Set(env.Ctx, cacheKey, entryJSON, 10*time.Minute).Err()
		require.NoError(t, err)
	}

	// Also cache account-level data
	accountCacheKey := fmt.Sprintf("account:%s:entries", accountID)
	err := env.Redis.Set(env.Ctx, accountCacheKey, "entry_list", 10*time.Minute).Err()
	require.NoError(t, err)

	// Verify all cached
	for _, entry := range entries {
		cacheKey := fmt.Sprintf("entry:%s", entry.ID)
		exists, err := env.Redis.Exists(env.Ctx, cacheKey).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	}

	// Act - Invalidate all entries for this account using pattern
	// Pattern: account:{accountID}:*
	accountPattern := fmt.Sprintf("account:%s:*", accountID)

	// Get all keys matching pattern
	keys, err := env.Redis.Keys(env.Ctx, accountPattern).Result()
	require.NoError(t, err)

	// Delete matched keys
	if len(keys) > 0 {
		deleted, err := env.Redis.Del(env.Ctx, keys...).Result()
		require.NoError(t, err)
		assert.Greater(t, deleted, int64(0))
	}

	// Also invalidate individual entry caches (in practice, would use a better pattern)
	// For this test, we manually invalidate
	for _, entry := range entries {
		cacheKey := fmt.Sprintf("entry:%s", entry.ID)
		err := env.Redis.Del(env.Ctx, cacheKey).Err()
		require.NoError(t, err)
	}

	// Assert - Verify all caches invalidated
	for _, entry := range entries {
		cacheKey := fmt.Sprintf("entry:%s", entry.ID)
		_, err := env.Redis.Get(env.Ctx, cacheKey).Result()
		assert.Error(t, err, "Cache should be invalidated")
	}

	// Verify account cache invalidated
	_, err = env.Redis.Get(env.Ctx, accountCacheKey).Result()
	assert.Error(t, err, "Account cache should be invalidated")
}

func TestIntegration_Redis_TTL_Expiration(t *testing.T) {
	env := testhelpers.SetupIntegrationTest(t)
	defer env.CleanAll()

	// Arrange - Set different TTLs for different cache types
	testCases := []struct {
		key string
		ttl time.Duration
	}{
		{"short-lived-key", 1 * time.Second},
		{"medium-lived-key", 5 * time.Second},
		{"long-lived-key", 10 * time.Second},
	}

	for _, tc := range testCases {
		err := env.Redis.Set(env.Ctx, tc.key, "value", tc.ttl).Err()
		require.NoError(t, err)
	}

	// Act - Verify TTLs
	for _, tc := range testCases {
		ttl, err := env.Redis.TTL(env.Ctx, tc.key).Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0), "TTL should be positive")
		assert.LessOrEqual(t, ttl, tc.ttl, "TTL should not exceed set value")
		t.Logf("Key %s has TTL: %v", tc.key, ttl)
	}

	// Wait for short-lived key to expire
	time.Sleep(2 * time.Second)

	// Assert - Short-lived key should be expired
	_, err := env.Redis.Get(env.Ctx, "short-lived-key").Result()
	assert.Error(t, err, "Short-lived key should be expired")

	// Medium and long-lived keys should still exist
	_, err = env.Redis.Get(env.Ctx, "medium-lived-key").Result()
	assert.NoError(t, err, "Medium-lived key should still exist")

	_, err = env.Redis.Get(env.Ctx, "long-lived-key").Result()
	assert.NoError(t, err, "Long-lived key should still exist")

	// Test PERSIST (remove expiration)
	err = env.Redis.Persist(env.Ctx, "medium-lived-key").Err()
	require.NoError(t, err)

	ttl, err := env.Redis.TTL(env.Ctx, "medium-lived-key").Result()
	require.NoError(t, err)
	assert.Equal(t, time.Duration(-1), ttl, "Key should have no expiration after PERSIST")

	// Test custom expiration patterns for DICT use cases
	entry := testhelpers.NewValidEntry()
	cacheKey := fmt.Sprintf("entry:%s", entry.ID)

	// Cache entry with standard TTL (5 minutes)
	err = env.Redis.Set(env.Ctx, cacheKey, entry.ID, 5*time.Minute).Err()
	require.NoError(t, err)

	// Extend TTL on cache hit (common pattern)
	err = env.Redis.Expire(env.Ctx, cacheKey, 5*time.Minute).Err()
	require.NoError(t, err)

	// Verify extended TTL
	ttl, err = env.Redis.TTL(env.Ctx, cacheKey).Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, 4*time.Minute, "TTL should be extended")

	// Test cache invalidation on update (set TTL to 0)
	err = env.Redis.Expire(env.Ctx, cacheKey, 0).Err()
	require.NoError(t, err)

	// Key should be immediately expired
	_, err = env.Redis.Get(env.Ctx, cacheKey).Result()
	assert.Error(t, err, "Key should be expired after setting TTL to 0")
}
