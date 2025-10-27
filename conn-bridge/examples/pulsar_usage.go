package examples

import (
	"context"
	"log"
	"time"

	pulsarClient "github.com/apache/pulsar-client-go/pulsar"
	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	"github.com/lbpay-lab/conn-bridge/internal/infrastructure/pulsar"
)

// ExampleBasicUsage demonstrates basic usage of DictPublisher
func ExampleBasicUsage() {
	// Create publisher with default config
	config := pulsar.DefaultDictPublisherConfig("pulsar://localhost:6650")
	publisher, err := pulsar.NewDictPublisher(config)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// Create a DICT entry
	entry := &entities.DictEntry{
		Key:  "user@example.com",
		Type: entities.KeyTypeEmail,
		Account: entities.Account{
			ISPB:   "12345678",
			Branch: "0001",
			Number: "123456",
			Type:   entities.AccountTypeChecking,
		},
		Owner: entities.Owner{
			Type:     entities.OwnerTypePerson,
			Document: "12345678900",
			Name:     "John Doe",
		},
		Status:    entities.StatusActive,
		CreatedAt: time.Now(),
	}

	// Publish entry created event
	ctx := context.Background()
	traceID := "trace-id-from-otel"

	err = publisher.PublishEntryCreated(ctx, entry, traceID)
	if err != nil {
		log.Printf("Warning: Failed to publish event: %v", err)
		// Note: This is async and non-blocking, so error handling
		// should not fail the main operation
	}

	log.Println("Event published successfully")
}

// ExampleUseCaseIntegration demonstrates integration with use cases
func ExampleUseCaseIntegration() {
	// Create publisher
	config := pulsar.DefaultDictPublisherConfig("pulsar://localhost:6650")
	pub, err := pulsar.NewDictPublisher(config)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer pub.Close()

	// In a use case, you would inject the publisher
	type CreateEntryUseCase struct {
		publisher *pulsar.DictPublisher
	}

	// Example method
	execute := func(ctx context.Context, entry *entities.DictEntry, publisher *pulsar.DictPublisher) error {
		// ... do main operation (e.g., call Bacen API)

		// Publish event (non-blocking, won't fail the operation)
		traceID := "trace-from-context"
		if err := publisher.PublishEntryCreated(ctx, entry, traceID); err != nil {
			// Log but don't return error - publishing is best effort
			log.Printf("Failed to publish entry created event: %v", err)
		}

		return nil
	}

	// Use the function
	_ = execute // prevent unused error
}

// ExampleCustomConfiguration shows custom configuration
func ExampleCustomConfiguration() {
	config := &pulsar.DictPublisherConfig{
		BrokerURL:         "pulsar://pulsar.prod:6650",
		Topic:             "rsfn-dict-res-out",
		ProducerName:      "rsfn-bridge-producer",
		BatchingEnabled:   true,
		MaxMessages:       200,                  // Larger batch
		BatchingMaxDelay:  5 * time.Millisecond, // Faster flush
		CompressionType:   pulsarClient.LZ4,
		OperationTimeout:  60 * time.Second, // Longer timeout
		ConnectionTimeout: 15 * time.Second,
	}

	publisher, err := pulsar.NewDictPublisher(config)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	log.Println("Publisher created with custom config")
}

// ExampleErrorPublishing shows how to publish error events
func ExampleErrorPublishing() {
	config := pulsar.DefaultDictPublisherConfig("pulsar://localhost:6650")
	publisher, err := pulsar.NewDictPublisher(config)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}
	defer publisher.Close()

	// Simulate an error in business logic
	businessErr := &BacenError{
		Code:    "ENTRY_NOT_FOUND",
		Message: "Entry does not exist in Bacen",
	}

	// Publish error event
	ctx := context.Background()
	traceID := "trace-id-123"
	context := "QueryEntry operation"

	err = publisher.PublishError(ctx, businessErr, context, traceID)
	if err != nil {
		log.Printf("Failed to publish error event: %v", err)
	}

	log.Println("Error event published")
}

// BacenError is an example error type
type BacenError struct {
	Code    string
	Message string
}

func (e *BacenError) Error() string {
	return e.Message
}

// ExampleGracefulShutdown shows proper cleanup
func ExampleGracefulShutdown() {
	config := pulsar.DefaultDictPublisherConfig("pulsar://localhost:6650")
	publisher, err := pulsar.NewDictPublisher(config)
	if err != nil {
		log.Fatalf("Failed to create publisher: %v", err)
	}

	// Publish some events
	ctx := context.Background()
	entry := &entities.DictEntry{
		Key:  "test@example.com",
		Type: entities.KeyTypeEmail,
	}

	_ = publisher.PublishEntryCreated(ctx, entry, "trace-1")
	_ = publisher.PublishEntryCreated(ctx, entry, "trace-2")
	_ = publisher.PublishEntryCreated(ctx, entry, "trace-3")

	// Give some time for async publishing
	time.Sleep(100 * time.Millisecond)

	// Graceful shutdown - flushes pending messages
	log.Println("Closing publisher...")
	if err := publisher.Close(); err != nil {
		log.Printf("Error closing publisher: %v", err)
	}

	log.Println("Publisher closed successfully")
}
