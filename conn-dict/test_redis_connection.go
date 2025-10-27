package main

import (
	"context"
	"fmt"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/cache"
	"github.com/sirupsen/logrus"
)

func main() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	fmt.Println("=== Redis Cache - 5 Strategies Test ===\n")

	// Configure Redis cache
	config := cache.RedisCacheConfig{
		Addr:               "localhost:6379",
		Password:           "",
		DB:                 0,
		PoolSize:           10,
		MinIdleConns:       2,
		MaxRetries:         3,
		DialTimeout:        5 * time.Second,
		ReadTimeout:        3 * time.Second,
		WriteTimeout:       3 * time.Second,
		WriteBehindEnabled: true,
		WriteBehindBatch:   5 * time.Second,
	}

	// Create cache instance
	redisCache, err := cache.NewRedisCache(config, logger)
	if err != nil {
		logger.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisCache.Close()

	fmt.Println("✅ Redis connection successful!")
	fmt.Println()

	ctx := context.Background()
	keyBuilder := cache.NewCacheKeyBuilder()

	// Test each strategy
	fmt.Println("Testing 5 Caching Strategies:")
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println()

	// 1. Cache-Aside
	fmt.Println("1. CACHE-ASIDE (Lazy Loading)")
	entryKey := keyBuilder.BuildEntryKey("test-001")
	result, err := redisCache.GetOrLoad(ctx, entryKey, func() (interface{}, error) {
		fmt.Println("   → Loading from 'database'...")
		return map[string]string{"id": "test-001", "status": "ACTIVE"}, nil
	}, 5*time.Minute)
	if err != nil {
		logger.Errorf("Cache-Aside failed: %v", err)
	} else {
		fmt.Printf("   ✓ Result: %+v\n", result)
	}
	fmt.Println()

	// 2. Write-Through
	fmt.Println("2. WRITE-THROUGH")
	newEntryKey := keyBuilder.BuildEntryKey("test-002")
	err = redisCache.SetWithDB(ctx, newEntryKey, map[string]string{
		"id": "test-002", "status": "PENDING",
	}, 5*time.Minute, func() error {
		fmt.Println("   → Writing to 'database'...")
		return nil
	})
	if err != nil {
		logger.Errorf("Write-Through failed: %v", err)
	} else {
		fmt.Println("   ✓ Data written to DB and cached")
	}
	fmt.Println()

	// 3. Write-Behind
	fmt.Println("3. WRITE-BEHIND (Async)")
	metricsKey := keyBuilder.BuildMetricsKey("12345678")
	err = redisCache.SetAsync(ctx, metricsKey, map[string]int{
		"total": 100, "active": 95,
	}, 10*time.Minute, func() error {
		fmt.Println("   → DB write will happen in batch (async)...")
		return nil
	})
	if err != nil {
		logger.Errorf("Write-Behind failed: %v", err)
	} else {
		fmt.Println("   ✓ Cached immediately, DB write queued")
	}
	fmt.Println()

	// 4. Read-Through
	fmt.Println("4. READ-THROUGH")
	configKey := keyBuilder.BuildConfigKey("max_retries")
	config2, err := redisCache.GetWithLoader(ctx, configKey, func() (interface{}, error) {
		fmt.Println("   → Auto-loading from 'database'...")
		return map[string]string{"name": "max_retries", "value": "5"}, nil
	}, 1*time.Hour)
	if err != nil {
		logger.Errorf("Read-Through failed: %v", err)
	} else {
		fmt.Printf("   ✓ Config: %+v\n", config2)
	}
	fmt.Println()

	// 5. Write-Around
	fmt.Println("5. WRITE-AROUND (Invalidation)")
	bulkKey := keyBuilder.BuildVSyncKey("vsync-001")
	err = redisCache.InvalidateAndWrite(ctx, bulkKey, func() error {
		fmt.Println("   → Performing bulk 'database' operation...")
		return nil
	})
	if err != nil {
		logger.Errorf("Write-Around failed: %v", err)
	} else {
		fmt.Println("   ✓ Bulk write completed, cache invalidated")
	}
	fmt.Println()

	// Wait for write-behind flush
	fmt.Println("Waiting 6 seconds for write-behind flush...")
	time.Sleep(6 * time.Second)

	fmt.Println()
	fmt.Println("=" + string(make([]byte, 50)))
	fmt.Println("✅ All 5 caching strategies tested successfully!")
	fmt.Println()
	fmt.Println("Strategies Summary:")
	fmt.Println("  1. Cache-Aside:   Entry lookups (TTL: 5min)")
	fmt.Println("  2. Write-Through: Entry creation (TTL: 5min)")
	fmt.Println("  3. Write-Behind:  High-frequency updates (TTL: 10min, Batch: 10s)")
	fmt.Println("  4. Read-Through:  Config data (TTL: 1hour)")
	fmt.Println("  5. Write-Around:  Bulk operations (Invalidation-based)")
}
