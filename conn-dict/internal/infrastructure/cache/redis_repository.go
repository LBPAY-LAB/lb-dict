package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/lbpay-lab/conn-dict/internal/domain/interfaces"
)

// RedisRepository implements CacheRepository using Redis
type RedisRepository struct {
	client *redis.Client
}

// NewRedisRepository creates a new RedisRepository
func NewRedisRepository(addr, password string, db int) (*RedisRepository, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisRepository{
		client: client,
	}, nil
}

// Get retrieves a value from cache
func (r *RedisRepository) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("key not found: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get value: %w", err)
	}
	return val, nil
}

// Set stores a value in cache with expiration
func (r *RedisRepository) Set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	err := r.client.Set(ctx, key, value, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set value: %w", err)
	}
	return nil
}

// Delete removes a value from cache
func (r *RedisRepository) Delete(ctx context.Context, key string) error {
	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (r *RedisRepository) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check existence: %w", err)
	}
	return n > 0, nil
}

// SetNX sets a value only if it doesn't exist (atomic)
func (r *RedisRepository) SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error) {
	ok, err := r.client.SetNX(ctx, key, value, expiration).Result()
	if err != nil {
		return false, fmt.Errorf("failed to set NX: %w", err)
	}
	return ok, nil
}

// Increment increments a counter
func (r *RedisRepository) Increment(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment: %w", err)
	}
	return val, nil
}

// Expire sets expiration on an existing key
func (r *RedisRepository) Expire(ctx context.Context, key string, expiration time.Duration) error {
	err := r.client.Expire(ctx, key, expiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set expiration: %w", err)
	}
	return nil
}

// Close closes the Redis connection
func (r *RedisRepository) Close() error {
	return r.client.Close()
}

// Ensure RedisRepository implements CacheRepository
var _ interfaces.CacheRepository = (*RedisRepository)(nil)
