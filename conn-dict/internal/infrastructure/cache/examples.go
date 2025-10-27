package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// Example demonstrates how to use each caching strategy

// ExampleCacheAside demonstrates Cache-Aside pattern for entry lookups
func ExampleCacheAside(cache Cache, logger *logrus.Logger) {
	ctx := context.Background()
	keyBuilder := NewCacheKeyBuilder()

	// Build cache key for entry lookup
	entryID := "550e8400-e29b-41d4-a716-446655440000"
	cacheKey := keyBuilder.BuildEntryKey(entryID)

	// Loader function simulates database fetch
	loader := func() (interface{}, error) {
		logger.Info("Fetching entry from database...")
		// Simulate DB query
		entry := map[string]interface{}{
			"id":        entryID,
			"key":       "+5511999887766",
			"key_type":  "PHONE",
			"ispb":      "12345678",
			"status":    "ACTIVE",
			"created_at": time.Now(),
		}
		return entry, nil
	}

	// Use Cache-Aside: check cache, load from DB if miss, cache result
	result, err := cache.GetOrLoad(ctx, cacheKey, loader, EntryKeyPattern.TTL)
	if err != nil {
		logger.Errorf("Cache-Aside failed: %v", err)
		return
	}

	logger.Infof("Entry retrieved: %+v", result)
}

// ExampleWriteThrough demonstrates Write-Through pattern for entry creation
func ExampleWriteThrough(cache Cache, logger *logrus.Logger) {
	ctx := context.Background()
	keyBuilder := NewCacheKeyBuilder()

	// New entry to create
	newEntry := map[string]interface{}{
		"id":        "new-entry-123",
		"key":       "email@example.com",
		"key_type":  "EMAIL",
		"ispb":      "87654321",
		"status":    "ACTIVE",
		"created_at": time.Now(),
	}

	cacheKey := keyBuilder.BuildEntryKey(newEntry["id"].(string))

	// DB writer function simulates database insert
	dbWriter := func() error {
		logger.Info("Writing entry to database...")
		// Simulate DB insert
		// db.Insert(newEntry)
		return nil
	}

	// Use Write-Through: write to DB first, then cache
	err := cache.SetWithDB(ctx, cacheKey, newEntry, EntryCreatePattern.TTL, dbWriter)
	if err != nil {
		logger.Errorf("Write-Through failed: %v", err)
		return
	}

	logger.Info("Entry created and cached successfully")
}

// ExampleWriteBehind demonstrates Write-Behind pattern for high-frequency metrics
func ExampleWriteBehind(cache Cache, logger *logrus.Logger) {
	ctx := context.Background()
	keyBuilder := NewCacheKeyBuilder()

	participantISPB := "12345678"
	cacheKey := keyBuilder.BuildMetricsKey(participantISPB)

	// Metrics to update
	metrics := map[string]interface{}{
		"participant_ispb": participantISPB,
		"total_entries":    1000,
		"active_entries":   950,
		"last_updated":     time.Now(),
	}

	// DB writer for async update
	dbWriter := func() error {
		logger.Info("Async writing metrics to database...")
		// Simulate DB update
		// db.UpdateMetrics(metrics)
		return nil
	}

	// Use Write-Behind: cache immediately, queue DB write
	err := cache.SetAsync(ctx, cacheKey, metrics, MetricsPattern.TTL, dbWriter)
	if err != nil {
		logger.Errorf("Write-Behind failed: %v", err)
		return
	}

	logger.Info("Metrics cached immediately, DB write queued")
}

// ExampleReadThrough demonstrates Read-Through pattern for configuration
func ExampleReadThrough(cache Cache, logger *logrus.Logger) {
	ctx := context.Background()
	keyBuilder := NewCacheKeyBuilder()

	configName := "max_retries"
	cacheKey := keyBuilder.BuildConfigKey(configName)

	// Loader function fetches config from database
	loader := func() (interface{}, error) {
		logger.Info("Loading configuration from database...")
		// Simulate config fetch
		config := map[string]interface{}{
			"name":  configName,
			"value": "5",
			"type":  "int",
		}
		return config, nil
	}

	// Use Read-Through: check cache, auto-load and cache if miss
	result, err := cache.GetWithLoader(ctx, cacheKey, loader, ConfigPattern.TTL)
	if err != nil {
		logger.Errorf("Read-Through failed: %v", err)
		return
	}

	logger.Infof("Configuration retrieved: %+v", result)
}

// ExampleWriteAround demonstrates Write-Around pattern for bulk operations
func ExampleWriteAround(cache Cache, logger *logrus.Logger) {
	ctx := context.Background()
	keyBuilder := NewCacheKeyBuilder()

	syncID := "vsync-batch-2024-01-15"
	cacheKey := keyBuilder.BuildVSyncKey(syncID)

	// DB writer for bulk operation
	dbWriter := func() error {
		logger.Info("Performing bulk write to database...")
		// Simulate bulk insert/update
		// db.BulkInsert(entries)
		return nil
	}

	// Use Write-Around: write to DB, invalidate cache (don't populate)
	err := cache.InvalidateAndWrite(ctx, cacheKey, dbWriter)
	if err != nil {
		logger.Errorf("Write-Around failed: %v", err)
		return
	}

	logger.Info("Bulk operation completed, cache invalidated")
}

// ExampleInvalidateParticipantCache demonstrates cache invalidation patterns
func ExampleInvalidateParticipantCache(cache Cache, logger *logrus.Logger) {
	ctx := context.Background()
	patterns := GetInvalidationPatterns()

	participantISPB := "12345678"

	// Invalidate all entries for a participant
	pattern := patterns.BuildParticipantEntriesPattern(participantISPB)
	err := cache.DeletePattern(ctx, pattern)
	if err != nil {
		logger.Errorf("Cache invalidation failed: %v", err)
		return
	}

	logger.Infof("Invalidated all entries for participant: %s", participantISPB)
}

// ExampleCompleteWorkflow demonstrates a complete workflow using all strategies
func ExampleCompleteWorkflow(cache Cache, logger *logrus.Logger) {
	ctx := context.Background()
	keyBuilder := NewCacheKeyBuilder()

	logger.Info("=== Starting Complete Cache Workflow ===")

	// 1. User looks up an entry (Cache-Aside)
	logger.Info("\n1. Entry Lookup (Cache-Aside)")
	entryKey := keyBuilder.BuildEntryKey("entry-001")
	entry, err := cache.GetOrLoad(ctx, entryKey, func() (interface{}, error) {
		return map[string]string{"id": "entry-001", "status": "ACTIVE"}, nil
	}, 5*time.Minute)
	if err != nil {
		logger.Errorf("Lookup failed: %v", err)
		return
	}
	logger.Infof("Entry found: %+v", entry)

	// 2. User creates a new entry (Write-Through)
	logger.Info("\n2. Entry Creation (Write-Through)")
	newEntryKey := keyBuilder.BuildEntryKey("entry-002")
	err = cache.SetWithDB(ctx, newEntryKey, map[string]string{
		"id": "entry-002", "status": "PENDING",
	}, 5*time.Minute, func() error {
		logger.Info("   -> Writing to database...")
		return nil
	})
	if err != nil {
		logger.Errorf("Creation failed: %v", err)
		return
	}
	logger.Info("Entry created successfully")

	// 3. System updates metrics (Write-Behind)
	logger.Info("\n3. Metrics Update (Write-Behind)")
	metricsKey := keyBuilder.BuildMetricsKey("12345678")
	err = cache.SetAsync(ctx, metricsKey, map[string]int{
		"total": 100, "active": 95,
	}, 10*time.Minute, func() error {
		logger.Info("   -> DB write will happen in next batch...")
		return nil
	})
	if err != nil {
		logger.Errorf("Metrics update failed: %v", err)
		return
	}
	logger.Info("Metrics cached, DB write queued")

	// 4. System loads configuration (Read-Through)
	logger.Info("\n4. Config Load (Read-Through)")
	configKey := keyBuilder.BuildConfigKey("timeout")
	config, err := cache.GetWithLoader(ctx, configKey, func() (interface{}, error) {
		return map[string]string{"timeout": "30s"}, nil
	}, 1*time.Hour)
	if err != nil {
		logger.Errorf("Config load failed: %v", err)
		return
	}
	logger.Infof("Config loaded: %+v", config)

	// 5. Bulk sync operation (Write-Around)
	logger.Info("\n5. Bulk Sync (Write-Around)")
	syncKey := keyBuilder.BuildVSyncKey("vsync-001")
	err = cache.InvalidateAndWrite(ctx, syncKey, func() error {
		logger.Info("   -> Performing bulk database operation...")
		return nil
	})
	if err != nil {
		logger.Errorf("Bulk sync failed: %v", err)
		return
	}
	logger.Info("Bulk sync completed")

	logger.Info("\n=== Workflow Complete ===")
}

// ExampleWithRealService demonstrates integration with service layer
func ExampleWithRealService(cache Cache, logger *logrus.Logger) {
	// This would be in your service layer
	type EntryService struct {
		cache      Cache
		keyBuilder *CacheKeyBuilder
		logger     *logrus.Logger
	}

	service := &EntryService{
		cache:      cache,
		keyBuilder: NewCacheKeyBuilder(),
		logger:     logger,
	}

	// Example service method using cache
	getEntry := func(ctx context.Context, entryID string) (interface{}, error) {
		cacheKey := service.keyBuilder.BuildEntryKey(entryID)

		// Use Cache-Aside pattern
		return service.cache.GetOrLoad(ctx, cacheKey, func() (interface{}, error) {
			// This would be your actual DB call
			service.logger.Infof("Loading entry %s from database", entryID)
			return map[string]string{"id": entryID, "data": "..."}, nil
		}, 5*time.Minute)
	}

	ctx := context.Background()
	entry, err := getEntry(ctx, "test-entry")
	if err != nil {
		logger.Errorf("Service call failed: %v", err)
		return
	}

	logger.Infof("Service returned: %+v", entry)
}

// RunAllExamples runs all example scenarios
func RunAllExamples() {
	logger := logrus.New()
	logger.SetLevel(logrus.InfoLevel)

	// Create cache instance
	config := RedisCacheConfig{
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
		WriteBehindBatch:   10 * time.Second,
	}

	cache, err := NewRedisCache(config, logger)
	if err != nil {
		logger.Fatalf("Failed to create cache: %v", err)
	}
	defer cache.Close()

	// Run examples
	fmt.Println("\n=== EXAMPLE 1: Cache-Aside ===")
	ExampleCacheAside(cache, logger)

	fmt.Println("\n=== EXAMPLE 2: Write-Through ===")
	ExampleWriteThrough(cache, logger)

	fmt.Println("\n=== EXAMPLE 3: Write-Behind ===")
	ExampleWriteBehind(cache, logger)

	fmt.Println("\n=== EXAMPLE 4: Read-Through ===")
	ExampleReadThrough(cache, logger)

	fmt.Println("\n=== EXAMPLE 5: Write-Around ===")
	ExampleWriteAround(cache, logger)

	fmt.Println("\n=== EXAMPLE 6: Cache Invalidation ===")
	ExampleInvalidateParticipantCache(cache, logger)

	fmt.Println("\n=== EXAMPLE 7: Complete Workflow ===")
	ExampleCompleteWorkflow(cache, logger)

	fmt.Println("\n=== EXAMPLE 8: Service Integration ===")
	ExampleWithRealService(cache, logger)

	logger.Info("\n✅ All examples completed successfully!")
}
