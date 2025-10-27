package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// RedisClient wraps Redis client with application-specific caching logic
type RedisClient struct {
	client *redis.Client
	logger *logrus.Logger
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Addr         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	MaxRetries   int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config RedisConfig, logger *logrus.Logger) (*RedisClient, error) {
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

	logger.Infof("Redis client connected: addr=%s, db=%d", config.Addr, config.DB)

	return &RedisClient{
		client: client,
		logger: logger,
	}, nil
}

// --- STRATEGY 1: Cache-Aside (Lazy Loading) ---

// Get retrieves a value from cache
func (r *RedisClient) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, key).Result()
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

	r.logger.Debugf("Cache hit: key=%s", key)
	return nil
}

// Set stores a value in cache with TTL
func (r *RedisClient) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Marshal to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	if err := r.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	r.logger.Debugf("Cache set: key=%s, ttl=%v", key, ttl)
	return nil
}

// --- STRATEGY 2: Write-Through Cache ---

// SetWithWriteThrough sets cache and also writes to database (callback)
func (r *RedisClient) SetWithWriteThrough(ctx context.Context, key string, value interface{}, ttl time.Duration, dbWrite func() error) error {
	// Write to database first
	if err := dbWrite(); err != nil {
		return fmt.Errorf("database write failed: %w", err)
	}

	// Then write to cache
	return r.Set(ctx, key, value, ttl)
}

// --- STRATEGY 3: Write-Behind Cache (Async) ---

// SetAsync sets cache and queues database write (non-blocking)
func (r *RedisClient) SetAsync(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	// Only set cache (database write happens asynchronously via queue/worker)
	return r.Set(ctx, key, value, ttl)
}

// --- STRATEGY 4: Refresh-Ahead Cache ---

// GetWithRefresh retrieves from cache and triggers background refresh if TTL is low
func (r *RedisClient) GetWithRefresh(ctx context.Context, key string, dest interface{}, ttl time.Duration, refreshThreshold time.Duration, refreshFunc func() (interface{}, error)) error {
	// Get current value
	err := r.Get(ctx, key, dest)
	if err != nil && err != ErrCacheMiss {
		return err
	}

	// If cache miss, load from source
	if err == ErrCacheMiss {
		newValue, err := refreshFunc()
		if err != nil {
			return fmt.Errorf("refresh failed: %w", err)
		}
		if err := r.Set(ctx, key, newValue, ttl); err != nil {
			return err
		}
		// Copy to dest
		return r.Get(ctx, key, dest)
	}

	// Check TTL
	remainingTTL, err := r.client.TTL(ctx, key).Result()
	if err != nil {
		return err
	}

	// If TTL is below threshold, trigger background refresh
	if remainingTTL < refreshThreshold {
		go func() {
			ctx := context.Background()
			newValue, err := refreshFunc()
			if err != nil {
				r.logger.Errorf("Background refresh failed: key=%s, error=%v", key, err)
				return
			}
			if err := r.Set(ctx, key, newValue, ttl); err != nil {
				r.logger.Errorf("Failed to set refreshed cache: key=%s, error=%v", key, err)
			} else {
				r.logger.Debugf("Cache refreshed in background: key=%s", key)
			}
		}()
	}

	return nil
}

// --- STRATEGY 5: Cache Invalidation Patterns ---

// Delete removes a key from cache
func (r *RedisClient) Delete(ctx context.Context, key string) error {
	if err := r.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}
	r.logger.Debugf("Cache deleted: key=%s", key)
	return nil
}

// DeletePattern removes all keys matching a pattern
func (r *RedisClient) DeletePattern(ctx context.Context, pattern string) error {
	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	pipe := r.client.Pipeline()

	count := 0
	for iter.Next(ctx) {
		pipe.Del(ctx, iter.Val())
		count++
	}
	if err := iter.Err(); err != nil {
		return err
	}

	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("redis pipeline exec error: %w", err)
	}

	r.logger.Debugf("Cache invalidated: pattern=%s, keys=%d", pattern, count)
	return nil
}

// --- Helper Functions ---

// Exists checks if a key exists
func (r *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Expire sets expiration on a key
func (r *RedisClient) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return r.client.Expire(ctx, key, ttl).Err()
}

// Close closes the Redis client
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// Note: ErrCacheMiss is defined in redis_cache.go