package adapters

import (
	"context"
	"encoding/json"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/cache"
)

// CacheAdapter adapts the Redis client to the use case cache interface
type CacheAdapter struct {
	client *cache.RedisClient
}

// NewCacheAdapter creates a new cache adapter
func NewCacheAdapter(client *cache.RedisClient) *CacheAdapter {
	return &CacheAdapter{client: client}
}

// Get retrieves a value from cache
func (a *CacheAdapter) Get(ctx context.Context, key string) ([]byte, error) {
	var result interface{}
	err := a.client.Get(ctx, key, &result)
	if err != nil {
		if err == cache.ErrCacheMiss {
			return nil, nil // Return nil, nil for cache miss
		}
		return nil, err
	}

	// Marshal the result to JSON bytes
	data, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return data, nil
}

// Set stores a value in cache with TTL
func (a *CacheAdapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	// Unmarshal to interface{} first
	var data interface{}
	if err := json.Unmarshal(value, &data); err != nil {
		return err
	}

	return a.client.Set(ctx, key, data, ttl)
}

// Delete removes a value from cache
func (a *CacheAdapter) Delete(ctx context.Context, key string) error {
	return a.client.Delete(ctx, key)
}
