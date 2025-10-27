// Package cache provides rate limiting functionality using Redis
package cache

import (
	"context"
	"fmt"
	"time"
)

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	// RequestsPerSecond is the maximum number of requests allowed per second
	RequestsPerSecond int

	// BurstSize is the maximum burst size
	BurstSize int

	// KeyPrefix is the Redis key prefix for rate limiting
	KeyPrefix string
}

// DefaultRateLimitConfig returns default rate limiting configuration
func DefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		RequestsPerSecond: 100,
		BurstSize:         20,
		KeyPrefix:         "core-dict:ratelimit:",
	}
}

// RateLimiter implements token bucket rate limiting using Redis
type RateLimiter struct {
	client *RedisClient
	config *RateLimitConfig
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(client *RedisClient, config *RateLimitConfig) *RateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}
	return &RateLimiter{
		client: client,
		config: config,
	}
}

// Allow checks if a request is allowed based on the identifier (IP or Account ID)
// Returns true if allowed, false if rate limit exceeded
func (rl *RateLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	key := rl.makeKey(identifier)
	now := time.Now().Unix()

	// Use Redis Lua script for atomic token bucket implementation
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])

		-- Get current count and window start time
		local current = redis.call('GET', key)

		if current == false then
			-- First request, initialize
			redis.call('SET', key, 1)
			redis.call('EXPIRE', key, window)
			return 1
		end

		current = tonumber(current)

		if current < limit then
			-- Within limit, increment
			redis.call('INCR', key)
			return 1
		else
			-- Rate limit exceeded
			return 0
		end
	`

	result, err := rl.client.client.Eval(ctx, script, []string{key},
		rl.config.RequestsPerSecond,
		1, // 1 second window
		now,
	).Result()

	if err != nil {
		return false, fmt.Errorf("rate limit check failed: %w", err)
	}

	allowed := result.(int64) == 1
	return allowed, nil
}

// AllowN checks if N requests are allowed
func (rl *RateLimiter) AllowN(ctx context.Context, identifier string, n int) (bool, error) {
	key := rl.makeKey(identifier)
	now := time.Now().Unix()

	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local n = tonumber(ARGV[4])

		local current = redis.call('GET', key)

		if current == false then
			if n <= limit then
				redis.call('SET', key, n)
				redis.call('EXPIRE', key, window)
				return 1
			else
				return 0
			end
		end

		current = tonumber(current)

		if current + n <= limit then
			redis.call('INCRBY', key, n)
			return 1
		else
			return 0
		end
	`

	result, err := rl.client.client.Eval(ctx, script, []string{key},
		rl.config.RequestsPerSecond,
		1,
		now,
		n,
	).Result()

	if err != nil {
		return false, fmt.Errorf("rate limit check failed: %w", err)
	}

	allowed := result.(int64) == 1
	return allowed, nil
}

// Reset resets the rate limit for an identifier
func (rl *RateLimiter) Reset(ctx context.Context, identifier string) error {
	key := rl.makeKey(identifier)
	return rl.client.Del(ctx, key)
}

// GetRemaining returns the remaining requests available for an identifier
func (rl *RateLimiter) GetRemaining(ctx context.Context, identifier string) (int, error) {
	key := rl.makeKey(identifier)

	current, err := rl.client.Get(ctx, key)
	if err == ErrCacheMiss {
		return rl.config.RequestsPerSecond, nil
	}
	if err != nil {
		return 0, fmt.Errorf("failed to get current count: %w", err)
	}

	var count int
	if _, err := fmt.Sscanf(current, "%d", &count); err != nil {
		return 0, fmt.Errorf("failed to parse count: %w", err)
	}

	remaining := rl.config.RequestsPerSecond - count
	if remaining < 0 {
		remaining = 0
	}

	return remaining, nil
}

// GetTTL returns the remaining time until the rate limit resets
func (rl *RateLimiter) GetTTL(ctx context.Context, identifier string) (time.Duration, error) {
	key := rl.makeKey(identifier)
	return rl.client.TTL(ctx, key)
}

// makeKey creates a rate limit key
func (rl *RateLimiter) makeKey(identifier string) string {
	return rl.config.KeyPrefix + identifier
}

// SlidingWindowRateLimiter implements sliding window rate limiting
type SlidingWindowRateLimiter struct {
	client *RedisClient
	config *RateLimitConfig
}

// NewSlidingWindowRateLimiter creates a new sliding window rate limiter
func NewSlidingWindowRateLimiter(client *RedisClient, config *RateLimitConfig) *SlidingWindowRateLimiter {
	if config == nil {
		config = DefaultRateLimitConfig()
	}
	return &SlidingWindowRateLimiter{
		client: client,
		config: config,
	}
}

// Allow checks if a request is allowed using sliding window algorithm
func (sw *SlidingWindowRateLimiter) Allow(ctx context.Context, identifier string) (bool, error) {
	key := sw.makeKey(identifier)
	now := time.Now().UnixMilli()
	windowStart := now - 1000 // 1 second window

	// Use Redis sorted set (ZSET) for sliding window
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local now = tonumber(ARGV[2])
		local window_start = tonumber(ARGV[3])

		-- Remove old entries outside the window
		redis.call('ZREMRANGEBYSCORE', key, 0, window_start)

		-- Count current entries in window
		local current = redis.call('ZCARD', key)

		if current < limit then
			-- Add current request
			redis.call('ZADD', key, now, now)
			redis.call('EXPIRE', key, 2)
			return 1
		else
			return 0
		end
	`

	result, err := sw.client.client.Eval(ctx, script, []string{key},
		sw.config.RequestsPerSecond,
		now,
		windowStart,
	).Result()

	if err != nil {
		return false, fmt.Errorf("sliding window rate limit check failed: %w", err)
	}

	allowed := result.(int64) == 1
	return allowed, nil
}

// GetCount returns the current count in the sliding window
func (sw *SlidingWindowRateLimiter) GetCount(ctx context.Context, identifier string) (int64, error) {
	key := sw.makeKey(identifier)
	now := time.Now().UnixMilli()
	windowStart := now - 1000

	// Remove old entries
	if err := sw.client.client.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart)).Err(); err != nil {
		return 0, fmt.Errorf("failed to remove old entries: %w", err)
	}

	// Count current entries
	count, err := sw.client.client.ZCard(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get count: %w", err)
	}

	return count, nil
}

// Reset resets the rate limit for an identifier
func (sw *SlidingWindowRateLimiter) Reset(ctx context.Context, identifier string) error {
	key := sw.makeKey(identifier)
	return sw.client.Del(ctx, key)
}

// makeKey creates a rate limit key
func (sw *SlidingWindowRateLimiter) makeKey(identifier string) string {
	return sw.config.KeyPrefix + "sw:" + identifier
}

// IPRateLimiter is a convenience wrapper for rate limiting by IP address
type IPRateLimiter struct {
	limiter *RateLimiter
}

// NewIPRateLimiter creates a new IP-based rate limiter
func NewIPRateLimiter(client *RedisClient, requestsPerSecond int) *IPRateLimiter {
	config := &RateLimitConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         requestsPerSecond / 5, // 20% burst
		KeyPrefix:         "core-dict:ratelimit:ip:",
	}
	return &IPRateLimiter{
		limiter: NewRateLimiter(client, config),
	}
}

// Allow checks if a request from the given IP is allowed
func (ip *IPRateLimiter) Allow(ctx context.Context, ipAddress string) (bool, error) {
	return ip.limiter.Allow(ctx, ipAddress)
}

// AccountRateLimiter is a convenience wrapper for rate limiting by account
type AccountRateLimiter struct {
	limiter *RateLimiter
}

// NewAccountRateLimiter creates a new account-based rate limiter
func NewAccountRateLimiter(client *RedisClient, requestsPerSecond int) *AccountRateLimiter {
	config := &RateLimitConfig{
		RequestsPerSecond: requestsPerSecond,
		BurstSize:         requestsPerSecond / 5,
		KeyPrefix:         "core-dict:ratelimit:account:",
	}
	return &AccountRateLimiter{
		limiter: NewRateLimiter(client, config),
	}
}

// Allow checks if a request from the given account is allowed
func (a *AccountRateLimiter) Allow(ctx context.Context, accountID string) (bool, error) {
	return a.limiter.Allow(ctx, accountID)
}

// MultiKeyRateLimiter allows rate limiting by multiple keys (e.g., IP + Account)
type MultiKeyRateLimiter struct {
	limiters []*RateLimiter
}

// NewMultiKeyRateLimiter creates a rate limiter that checks multiple keys
func NewMultiKeyRateLimiter(limiters ...*RateLimiter) *MultiKeyRateLimiter {
	return &MultiKeyRateLimiter{
		limiters: limiters,
	}
}

// Allow checks if request is allowed for all keys
func (m *MultiKeyRateLimiter) Allow(ctx context.Context, identifiers ...string) (bool, error) {
	if len(identifiers) != len(m.limiters) {
		return false, fmt.Errorf("number of identifiers (%d) doesn't match number of limiters (%d)",
			len(identifiers), len(m.limiters))
	}

	for i, limiter := range m.limiters {
		allowed, err := limiter.Allow(ctx, identifiers[i])
		if err != nil {
			return false, err
		}
		if !allowed {
			return false, nil
		}
	}

	return true, nil
}
