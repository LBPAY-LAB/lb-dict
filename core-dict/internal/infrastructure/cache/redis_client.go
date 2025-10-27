// Package cache provides Redis client and caching functionality
package cache

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisConfig holds Redis connection configuration
type RedisConfig struct {
	URL              string
	Password         string
	DB               int
	PoolSize         int
	MinIdleConns     int
	MaxRetries       int
	ConnMaxIdleTime  time.Duration
	ConnMaxLifetime  time.Duration
	DialTimeout      time.Duration
	ReadTimeout      time.Duration
	WriteTimeout     time.Duration
	TLSEnabled       bool
	TLSSkipVerify    bool
}

// DefaultRedisConfig returns default Redis configuration
func DefaultRedisConfig() *RedisConfig {
	return &RedisConfig{
		URL:              "redis://localhost:6379",
		Password:         "",
		DB:               0,
		PoolSize:         10,
		MinIdleConns:     5,
		MaxRetries:       3,
		ConnMaxIdleTime:  5 * time.Minute,
		ConnMaxLifetime:  30 * time.Minute,
		DialTimeout:      5 * time.Second,
		ReadTimeout:      3 * time.Second,
		WriteTimeout:     3 * time.Second,
		TLSEnabled:       false,
		TLSSkipVerify:    false,
	}
}

// RedisClient wraps the redis.Client with additional functionality
type RedisClient struct {
	client *redis.Client
	config *RedisConfig
}

// NewRedisClient creates a new Redis client with the provided configuration
func NewRedisClient(config *RedisConfig) (*RedisClient, error) {
	if config == nil {
		config = DefaultRedisConfig()
	}

	opts, err := redis.ParseURL(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Override with config values
	opts.Password = config.Password
	opts.DB = config.DB
	opts.PoolSize = config.PoolSize
	opts.MinIdleConns = config.MinIdleConns
	opts.MaxRetries = config.MaxRetries
	opts.ConnMaxIdleTime = config.ConnMaxIdleTime
	opts.ConnMaxLifetime = config.ConnMaxLifetime
	opts.DialTimeout = config.DialTimeout
	opts.ReadTimeout = config.ReadTimeout
	opts.WriteTimeout = config.WriteTimeout

	// TLS configuration
	if config.TLSEnabled {
		opts.TLSConfig = &tls.Config{
			InsecureSkipVerify: config.TLSSkipVerify,
		}
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisClient{
		client: client,
		config: config,
	}, nil
}

// Client returns the underlying redis.Client
func (rc *RedisClient) Client() *redis.Client {
	return rc.client
}

// Close closes the Redis connection
func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

// Ping checks if Redis is alive
func (rc *RedisClient) Ping(ctx context.Context) error {
	return rc.client.Ping(ctx).Err()
}

// Get retrieves a value by key
func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", ErrCacheMiss
	}
	if err != nil {
		return "", fmt.Errorf("redis get failed: %w", err)
	}
	return val, nil
}

// Set sets a key-value pair with expiration
func (rc *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	err := rc.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis set failed: %w", err)
	}
	return nil
}

// SetNX sets a key-value pair only if the key does not exist (used for distributed locks)
func (rc *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	result, err := rc.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("redis setnx failed: %w", err)
	}
	return result, nil
}

// Del deletes one or more keys
func (rc *RedisClient) Del(ctx context.Context, keys ...string) error {
	err := rc.client.Del(ctx, keys...).Err()
	if err != nil {
		return fmt.Errorf("redis del failed: %w", err)
	}
	return nil
}

// Exists checks if a key exists
func (rc *RedisClient) Exists(ctx context.Context, key string) (bool, error) {
	result, err := rc.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists failed: %w", err)
	}
	return result > 0, nil
}

// Expire sets expiration on a key
func (rc *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := rc.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("redis expire failed: %w", err)
	}
	return nil
}

// Incr increments the integer value of a key by one
func (rc *RedisClient) Incr(ctx context.Context, key string) (int64, error) {
	val, err := rc.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis incr failed: %w", err)
	}
	return val, nil
}

// IncrBy increments the integer value of a key by the given amount
func (rc *RedisClient) IncrBy(ctx context.Context, key string, value int64) (int64, error) {
	val, err := rc.client.IncrBy(ctx, key, value).Result()
	if err != nil {
		return 0, fmt.Errorf("redis incrby failed: %w", err)
	}
	return val, nil
}

// Decr decrements the integer value of a key by one
func (rc *RedisClient) Decr(ctx context.Context, key string) (int64, error) {
	val, err := rc.client.Decr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis decr failed: %w", err)
	}
	return val, nil
}

// TTL returns the remaining time to live of a key
func (rc *RedisClient) TTL(ctx context.Context, key string) (time.Duration, error) {
	val, err := rc.client.TTL(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl failed: %w", err)
	}
	return val, nil
}

// Keys returns all keys matching pattern (use with caution in production)
func (rc *RedisClient) Keys(ctx context.Context, pattern string) ([]string, error) {
	keys, err := rc.client.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("redis keys failed: %w", err)
	}
	return keys, nil
}

// FlushDB deletes all keys in the current database (use with caution)
func (rc *RedisClient) FlushDB(ctx context.Context) error {
	err := rc.client.FlushDB(ctx).Err()
	if err != nil {
		return fmt.Errorf("redis flushdb failed: %w", err)
	}
	return nil
}

// Info returns Redis server information
func (rc *RedisClient) Info(ctx context.Context) (string, error) {
	info, err := rc.client.Info(ctx).Result()
	if err != nil {
		return "", fmt.Errorf("redis info failed: %w", err)
	}
	return info, nil
}

// DBSize returns the number of keys in the current database
func (rc *RedisClient) DBSize(ctx context.Context) (int64, error) {
	size, err := rc.client.DBSize(ctx).Result()
	if err != nil {
		return 0, fmt.Errorf("redis dbsize failed: %w", err)
	}
	return size, nil
}

// Pipeline creates a new pipeline for batch operations
func (rc *RedisClient) Pipeline() redis.Pipeliner {
	return rc.client.Pipeline()
}

// TxPipeline creates a new transaction pipeline
func (rc *RedisClient) TxPipeline() redis.Pipeliner {
	return rc.client.TxPipeline()
}

// ErrCacheMiss is returned when a key is not found in cache
var ErrCacheMiss = fmt.Errorf("cache miss")
