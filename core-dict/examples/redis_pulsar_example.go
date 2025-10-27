// Package examples demonstrates Redis and Pulsar integration
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/lbpay-lab/core-dict/internal/infrastructure/cache"
	"github.com/lbpay-lab/core-dict/internal/infrastructure/messaging"
)

// Example demonstrates Redis caching and Pulsar event publishing
func main() {
	ctx := context.Background()

	// ====================================
	// Redis Example
	// ====================================
	fmt.Println("=== Redis Cache Example ===")

	// Create Redis client
	redisConfig := cache.DefaultRedisConfig()
	redisConfig.URL = "redis://localhost:6379/0"

	redisClient, err := cache.NewRedisClient(redisConfig)
	if err != nil {
		log.Fatalf("Failed to create Redis client: %v", err)
	}
	defer redisClient.Close()

	// Test connection
	if err := redisClient.Ping(ctx); err != nil {
		log.Fatalf("Redis ping failed: %v", err)
	}
	fmt.Println("✓ Redis connected")

	// Create cache with Cache-Aside strategy
	dictCache := cache.NewCache(redisClient, cache.CacheAside, "core-dict:")
	cacheHandler := cache.NewCacheAsideHandler(dictCache)

	// Simulate getting a DICT entry (Cache-Aside pattern)
	fmt.Println("\n--- Cache-Aside Pattern ---")
	key := "12345678901" // CPF
	cacheKey := "entry:" + key

	// Mock loader (simulates DB query)
	loader := func(ctx context.Context) (interface{}, error) {
		fmt.Println("  ⚡ Cache MISS - Loading from database...")
		time.Sleep(100 * time.Millisecond) // Simulate DB query
		return map[string]interface{}{
			"key":       key,
			"key_type":  "CPF",
			"account":   "00001",
			"ispb":      "12345678",
			"status":    "ACTIVE",
		}, nil
	}

	// First call - cache miss, loads from DB
	var entry1 interface{}
	err = cacheHandler.GetOrLoad(ctx, cacheKey, &entry1, 5*time.Minute, loader)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Printf("  Entry: %+v\n", entry1)
	}

	// Second call - cache hit
	fmt.Println("\n  Second call (should hit cache):")
	var entry2 interface{}
	err = cacheHandler.GetOrLoad(ctx, cacheKey, &entry2, 5*time.Minute, loader)
	if err != nil {
		log.Printf("Error: %v", err)
	} else {
		fmt.Println("  ✓ Cache HIT - No DB query needed")
	}

	// ====================================
	// Rate Limiting Example
	// ====================================
	fmt.Println("\n\n=== Rate Limiting Example ===")

	rateLimiter := cache.NewIPRateLimiter(redisClient, 5) // 5 req/s

	ip := "192.168.1.100"
	fmt.Printf("Testing rate limit for IP %s (limit: 5 req/s)\n", ip)

	for i := 1; i <= 7; i++ {
		allowed, err := rateLimiter.Allow(ctx, ip)
		if err != nil {
			log.Printf("Error: %v", err)
			continue
		}

		if allowed {
			fmt.Printf("  Request %d: ✓ ALLOWED\n", i)
		} else {
			fmt.Printf("  Request %d: ✗ RATE LIMITED\n", i)
		}
	}

	// ====================================
	// Pulsar Example
	// ====================================
	fmt.Println("\n\n=== Pulsar Event Publishing Example ===")

	// Create Pulsar producer
	producerConfig := messaging.DefaultProducerConfig()
	producerConfig.BrokerURL = "pulsar://localhost:6650"
	producerConfig.Topic = "persistent://dict/events/key-events"

	producer, err := messaging.NewEventProducer(producerConfig)
	if err != nil {
		log.Fatalf("Failed to create Pulsar producer: %v", err)
	}
	defer producer.Close()

	fmt.Println("✓ Pulsar producer connected")

	// Publish KeyCreated event
	event := &messaging.DomainEvent{
		EventID:       fmt.Sprintf("key_created_%s_%d", key, time.Now().UnixNano()),
		EventType:     "KeyCreated",
		AggregateID:   key,
		AggregateType: "DictEntry",
		Timestamp:     time.Now(),
		Version:       1,
		Data: map[string]interface{}{
			"key":          key,
			"key_type":     "CPF",
			"account_ispb": "12345678",
			"created_at":   time.Now().Format(time.RFC3339),
		},
		Metadata: map[string]string{
			"source":      "core-dict",
			"environment": "development",
		},
	}

	err = producer.PublishEvent(ctx, event)
	if err != nil {
		log.Printf("Failed to publish event: %v", err)
	} else {
		fmt.Printf("✓ Event published: %s (Type: %s)\n", event.EventID, event.EventType)
	}

	// Publish asynchronously
	fmt.Println("\n--- Async Publishing ---")
	asyncEvent := &messaging.DomainEvent{
		EventID:       fmt.Sprintf("key_updated_%s_%d", key, time.Now().UnixNano()),
		EventType:     "KeyUpdated",
		AggregateID:   key,
		AggregateType: "DictEntry",
		Timestamp:     time.Now(),
		Version:       2,
		Data: map[string]interface{}{
			"key":        key,
			"old_status": "ACTIVE",
			"new_status": "BLOCKED",
		},
	}

	producer.PublishEventAsync(ctx, asyncEvent, func(msgID interface{}, err error) {
		if err != nil {
			fmt.Printf("  ✗ Async publish failed: %v\n", err)
		} else {
			fmt.Printf("  ✓ Async event published: %s\n", asyncEvent.EventID)
		}
	})

	// Wait for async publish to complete
	time.Sleep(100 * time.Millisecond)

	// Flush pending messages
	if err := producer.Flush(); err != nil {
		log.Printf("Failed to flush: %v", err)
	}

	// ====================================
	// Specialized Producers Example
	// ====================================
	fmt.Println("\n\n=== Specialized Key Event Producer ===")

	keyProducer, err := messaging.NewKeyEventProducer("pulsar://localhost:6650")
	if err != nil {
		log.Printf("Failed to create key producer: %v", err)
	} else {
		defer keyProducer.Close()

		// Publish KeyCreated
		err = keyProducer.PublishKeyCreated(ctx, "98765432100", "CPF", "12345678")
		if err != nil {
			log.Printf("Failed: %v", err)
		} else {
			fmt.Println("✓ KeyCreated event published via specialized producer")
		}

		// Publish KeyDeleted
		err = keyProducer.PublishKeyDeleted(ctx, key, "user_requested")
		if err != nil {
			log.Printf("Failed: %v", err)
		} else {
			fmt.Println("✓ KeyDeleted event published")
		}
	}

	// ====================================
	// Integration Example: Cache + Events
	// ====================================
	fmt.Println("\n\n=== Integration: Cache + Events ===")

	// Simulate creating a new key
	newKey := "11122233344"
	fmt.Printf("Creating new key: %s\n", newKey)

	// 1. Store in cache (Write-Through pattern)
	writeThroughHandler := cache.NewWriteThroughHandler(dictCache)
	newEntry := map[string]interface{}{
		"key":      newKey,
		"key_type": "CPF",
		"account":  "00002",
		"ispb":     "87654321",
		"status":   "ACTIVE",
	}

	err = writeThroughHandler.Write(ctx, "entry:"+newKey, newEntry, 5*time.Minute, func(ctx context.Context, value interface{}) error {
		// Simulate DB write
		fmt.Println("  ✓ Saved to database")
		return nil
	})
	if err != nil {
		log.Printf("Failed to write: %v", err)
	} else {
		fmt.Println("  ✓ Saved to cache")
	}

	// 2. Publish event
	if producer != nil {
		createEvent := &messaging.DomainEvent{
			EventID:       fmt.Sprintf("key_created_%s_%d", newKey, time.Now().UnixNano()),
			EventType:     "KeyCreated",
			AggregateID:   newKey,
			AggregateType: "DictEntry",
			Data:          newEntry,
		}

		producer.PublishEvent(ctx, createEvent)
		fmt.Println("  ✓ Event published to Pulsar")
	}

	fmt.Println("\n=== Example Complete ===")
}
