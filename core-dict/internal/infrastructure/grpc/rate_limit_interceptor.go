package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimitInterceptor handles rate limiting for gRPC requests
// In production, this should use Redis for distributed rate limiting
type RateLimitInterceptor struct {
	// Redis client would be injected here in production
	// redisClient *redis.Client

	// In-memory rate limiter for development (not suitable for multi-instance production)
	mu              sync.RWMutex
	userLimits      map[string]*UserRateLimit
	globalLimit     *TokenBucket
	enableGlobal    bool
	enablePerUser   bool
	globalRPS       int
	perUserRPS      int
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	EnableGlobalLimit bool // Enable global rate limiting (e.g., 100 req/s for entire service)
	EnablePerUserLimit bool // Enable per-user rate limiting (e.g., 10 req/s per user)
	GlobalRPS         int  // Global requests per second
	PerUserRPS        int  // Per-user requests per second
}

// NewRateLimitInterceptor creates a new rate limit interceptor
func NewRateLimitInterceptor(config *RateLimitConfig) *RateLimitInterceptor {
	if config == nil {
		config = &RateLimitConfig{
			EnableGlobalLimit:  true,
			EnablePerUserLimit: true,
			GlobalRPS:          100,  // 100 req/s globally
			PerUserRPS:         10,   // 10 req/s per user
		}
	}

	interceptor := &RateLimitInterceptor{
		userLimits:     make(map[string]*UserRateLimit),
		enableGlobal:   config.EnableGlobalLimit,
		enablePerUser:  config.EnablePerUserLimit,
		globalRPS:      config.GlobalRPS,
		perUserRPS:     config.PerUserRPS,
	}

	// Initialize global rate limiter
	if config.EnableGlobalLimit {
		interceptor.globalLimit = NewTokenBucket(config.GlobalRPS, config.GlobalRPS)
	}

	return interceptor
}

// Unary returns a unary server interceptor for rate limiting
func (i *RateLimitInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check global rate limit first
		if i.enableGlobal {
			if !i.globalLimit.Allow() {
				return nil, status.Error(
					codes.ResourceExhausted,
					"global rate limit exceeded, please try again later",
				)
			}
		}

		// Check per-user rate limit
		if i.enablePerUser {
			userID, _ := ctx.Value("user_id").(string)
			if userID != "" {
				allowed, retryAfter := i.checkUserRateLimit(userID)
				if !allowed {
					return nil, status.Errorf(
						codes.ResourceExhausted,
						"user rate limit exceeded, retry after %d seconds",
						retryAfter,
					)
				}
			}
		}

		// Call the handler
		return handler(ctx, req)
	}
}

// checkUserRateLimit checks if user has exceeded their rate limit
func (i *RateLimitInterceptor) checkUserRateLimit(userID string) (allowed bool, retryAfterSeconds int) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Get or create user rate limit
	userLimit, exists := i.userLimits[userID]
	if !exists {
		userLimit = &UserRateLimit{
			bucket: NewTokenBucket(i.perUserRPS, i.perUserRPS),
		}
		i.userLimits[userID] = userLimit
	}

	// Check if user can make request
	if userLimit.bucket.Allow() {
		return true, 0
	}

	// Calculate retry after (1 second for token bucket)
	return false, 1
}

// UserRateLimit holds per-user rate limiting state
type UserRateLimit struct {
	bucket *TokenBucket
}

// TokenBucket implements a simple token bucket rate limiter
type TokenBucket struct {
	mu           sync.Mutex
	capacity     int       // Maximum tokens
	tokens       int       // Current tokens
	refillRate   int       // Tokens added per second
	lastRefill   time.Time
}

// NewTokenBucket creates a new token bucket
func NewTokenBucket(capacity, refillRate int) *TokenBucket {
	return &TokenBucket{
		capacity:   capacity,
		tokens:     capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// Allow checks if a request is allowed under the rate limit
func (tb *TokenBucket) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(elapsed.Seconds() * float64(tb.refillRate))

	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.capacity {
			tb.tokens = tb.capacity
		}
		tb.lastRefill = now
	}

	// Check if we have tokens available
	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

// GetTokens returns current number of tokens (for testing/monitoring)
func (tb *TokenBucket) GetTokens() int {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	return tb.tokens
}

// RedisRateLimiter is a Redis-based rate limiter (for production use)
// TODO: Implement this using Redis for distributed rate limiting
type RedisRateLimiter struct {
	// redisClient *redis.Client
	keyPrefix string
	ttl       time.Duration
}

// NewRedisRateLimiter creates a Redis-based rate limiter
// func NewRedisRateLimiter(redisClient *redis.Client, keyPrefix string, ttl time.Duration) *RedisRateLimiter {
// 	return &RedisRateLimiter{
// 		redisClient: redisClient,
// 		keyPrefix:   keyPrefix,
// 		ttl:         ttl,
// 	}
// }

// CheckRateLimit checks if a request is within rate limits using Redis
// func (r *RedisRateLimiter) CheckRateLimit(ctx context.Context, key string, limit int, window time.Duration) (allowed bool, remaining int, resetAt time.Time, err error) {
// 	// Redis Lua script for atomic rate limiting
// 	script := `
// 		local key = KEYS[1]
// 		local limit = tonumber(ARGV[1])
// 		local window = tonumber(ARGV[2])
// 		local now = tonumber(ARGV[3])
//
// 		local current = redis.call('GET', key)
// 		if current and tonumber(current) >= limit then
// 			local ttl = redis.call('TTL', key)
// 			return {0, 0, now + ttl}
// 		end
//
// 		local count = redis.call('INCR', key)
// 		if count == 1 then
// 			redis.call('EXPIRE', key, window)
// 		end
//
// 		local ttl = redis.call('TTL', key)
// 		return {1, limit - count, now + ttl}
// 	`
//
// 	fullKey := fmt.Sprintf("%s:%s", r.keyPrefix, key)
// 	now := time.Now().Unix()
//
// 	result, err := r.redisClient.Eval(ctx, script, []string{fullKey}, limit, int(window.Seconds()), now).Result()
// 	if err != nil {
// 		return false, 0, time.Time{}, err
// 	}
//
// 	values := result.([]interface{})
// 	allowed = values[0].(int64) == 1
// 	remaining = int(values[1].(int64))
// 	resetAt = time.Unix(values[2].(int64), 0)
//
// 	return allowed, remaining, resetAt, nil
// }

// RateLimitStats returns current rate limiting statistics
func (i *RateLimitInterceptor) RateLimitStats() map[string]interface{} {
	i.mu.RLock()
	defer i.mu.RUnlock()

	stats := make(map[string]interface{})

	// Global stats
	if i.enableGlobal {
		stats["global_enabled"] = true
		stats["global_rps"] = i.globalRPS
		stats["global_tokens"] = i.globalLimit.GetTokens()
	}

	// Per-user stats
	if i.enablePerUser {
		stats["per_user_enabled"] = true
		stats["per_user_rps"] = i.perUserRPS
		stats["total_users"] = len(i.userLimits)

		userStats := make(map[string]int)
		for userID, limit := range i.userLimits {
			userStats[userID] = limit.bucket.GetTokens()
		}
		stats["user_tokens"] = userStats
	}

	return stats
}

// CleanupStaleUsers removes inactive users from the rate limiter
// Should be called periodically to prevent memory leaks
func (i *RateLimitInterceptor) CleanupStaleUsers(maxIdleTime time.Duration) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// In a real implementation, track last access time per user
	// For now, just clear all users if map is too large
	if len(i.userLimits) > 10000 {
		i.userLimits = make(map[string]*UserRateLimit)
		fmt.Println("[RATE_LIMIT] Cleared user rate limit cache (exceeded 10k users)")
	}
}
