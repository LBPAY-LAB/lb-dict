// Example demonstrating how to use EntryEventProducer
// To run: go run examples/producer_example.go
//
// Prerequisites:
//   - Pulsar running on localhost:6650
//   - docker run -d -p 6650:6650 -p 8080:8080 apachepulsar/pulsar:latest bin/pulsar standalone

package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/infrastructure/messaging"
)

func main() {
	log.Println("========================================")
	log.Println("Entry Event Producer Example")
	log.Println("========================================")

	// Step 1: Create producer
	log.Println("\n[1] Creating EntryEventProducer...")
	producer, err := messaging.NewEntryEventProducer("pulsar://localhost:6650")
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer func() {
		log.Println("\n[CLEANUP] Flushing and closing producer...")
		if err := producer.Flush(); err != nil {
			log.Printf("Failed to flush: %v", err)
		}
		if err := producer.Close(); err != nil {
			log.Printf("Failed to close: %v", err)
		}
		log.Println("Producer closed successfully")
	}()
	log.Println("✅ Producer created successfully")

	// Step 2: Create a sample entry
	log.Println("\n[2] Creating sample Entry...")
	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		AccountType:   "CACC",
		OwnerName:     "João Silva",
		OwnerTaxID:    "12345678901",
		OwnerType:     "NATURAL_PERSON",
		Status:        entities.KeyStatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
	log.Printf("✅ Entry created: ID=%s, KeyType=%s, KeyValue=%s", entry.ID, entry.KeyType, entry.KeyValue)

	ctx := context.Background()
	userID := "user-123"

	// Step 3: Publish EntryCreatedEvent
	log.Println("\n[3] Publishing EntryCreatedEvent...")
	if err := producer.PublishCreated(ctx, entry, userID); err != nil {
		log.Fatalf("Failed to publish EntryCreatedEvent: %v", err)
	}
	log.Println("✅ EntryCreatedEvent published to topic: dict.entries.created")

	// Step 4: Update entry and publish update event
	log.Println("\n[4] Updating entry and publishing EntryUpdatedEvent...")
	time.Sleep(100 * time.Millisecond)
	entry.AccountNumber = "654321"
	entry.Branch = "0002"
	entry.UpdatedAt = time.Now()

	if err := producer.PublishUpdated(ctx, entry, userID); err != nil {
		log.Fatalf("Failed to publish EntryUpdatedEvent: %v", err)
	}
	log.Println("✅ EntryUpdatedEvent published to topic: dict.entries.updated")

	// Step 5: Delete entry and publish deletion event
	log.Println("\n[5] Deleting entry and publishing EntryDeletedEvent...")
	time.Sleep(100 * time.Millisecond)

	if err := producer.PublishDeletedImmediate(ctx, entry, userID); err != nil {
		log.Fatalf("Failed to publish EntryDeletedEvent: %v", err)
	}
	log.Println("✅ EntryDeletedEvent published to topic: dict.entries.deleted.immediate")

	// Step 6: Summary
	log.Println("\n========================================")
	log.Println("Summary")
	log.Println("========================================")
	log.Println("Events published:")
	log.Println("  1. EntryCreatedEvent  → dict.entries.created")
	log.Println("  2. EntryUpdatedEvent  → dict.entries.updated")
	log.Println("  3. EntryDeletedEvent  → dict.entries.deleted.immediate")
	log.Println()
	log.Println("Next steps:")
	log.Println("  - Verify messages in Pulsar admin UI: http://localhost:8080")
	log.Println("  - Implement consumer in conn-dict to process these events")
	log.Println("  - Add monitoring metrics (Prometheus)")
	log.Println("========================================")
}
