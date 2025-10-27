package interfaces

import (
	"context"
	"time"
)

// CacheRepository defines the interface for cache operations
type CacheRepository interface {
	// Get retrieves a value from cache
	Get(ctx context.Context, key string) ([]byte, error)

	// Set stores a value in cache with expiration
	Set(ctx context.Context, key string, value []byte, expiration time.Duration) error

	// Delete removes a value from cache
	Delete(ctx context.Context, key string) error

	// Exists checks if a key exists in cache
	Exists(ctx context.Context, key string) (bool, error)

	// SetNX sets a value only if it doesn't exist (atomic)
	SetNX(ctx context.Context, key string, value []byte, expiration time.Duration) (bool, error)

	// Increment increments a counter
	Increment(ctx context.Context, key string) (int64, error)

	// Expire sets expiration on an existing key
	Expire(ctx context.Context, key string, expiration time.Duration) error
}
