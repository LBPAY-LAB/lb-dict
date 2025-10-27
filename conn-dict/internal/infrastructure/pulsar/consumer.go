package pulsar

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	bridgepb "github.com/lbpay-lab/dict-contracts/gen/proto/bridge/v1"
	commonpb "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
)

// Consumer wraps Pulsar consumer for processing Entry events asynchronously
// This consumer handles fast operations (< 2s) that don't require Temporal workflows
type Consumer struct {
	client        pulsar.Client
	consumers     []pulsar.Consumer
	entryRepo     *repositories.EntryRepository
	bridgeClient  bridgepb.BridgeServiceClient
	logger        *logrus.Logger
	wg            sync.WaitGroup
	stopChan      chan struct{}
	config        ConsumerConfig
}

// ConsumerConfig holds configuration for Pulsar consumer
type ConsumerConfig struct {
	URL                     string
	Subscription            string
	TopicEntryCreated       string
	TopicEntryUpdated       string
	TopicEntryDeletedImmed  string
	MaxReconnectToBroker    uint
	ConnectionTimeout       time.Duration
	OperationTimeout        time.Duration
	AckTimeout              time.Duration
	NackRedeliveryDelay     time.Duration
	MaxDeliveryAttempts     int
}

// EntryCreatedEvent represents a Pulsar event for entry creation
type EntryCreatedEvent struct {
	EntryID           string    `json:"entry_id"`
	Key               string    `json:"key"`
	KeyType           string    `json:"key_type"`
	Participant       string    `json:"participant"`
	AccountBranch     *string   `json:"account_branch,omitempty"`
	AccountNumber     *string   `json:"account_number,omitempty"`
	AccountType       string    `json:"account_type"`
	AccountOpenedDate *string   `json:"account_opened_date,omitempty"`
	OwnerType         string    `json:"owner_type"`
	OwnerName         *string   `json:"owner_name,omitempty"`
	OwnerTaxID        *string   `json:"owner_tax_id,omitempty"`
	IdempotencyKey    string    `json:"idempotency_key"`
	RequestID         string    `json:"request_id"`
	Timestamp         time.Time `json:"timestamp"`
}

// EntryUpdatedEvent represents a Pulsar event for entry update
type EntryUpdatedEvent struct {
	EntryID           string    `json:"entry_id"`
	Key               string    `json:"key"`
	KeyType           string    `json:"key_type"`
	Participant       string    `json:"participant"`
	AccountBranch     *string   `json:"account_branch,omitempty"`
	AccountNumber     *string   `json:"account_number,omitempty"`
	AccountType       string    `json:"account_type"`
	AccountOpenedDate *string   `json:"account_opened_date,omitempty"`
	OwnerType         string    `json:"owner_type"`
	OwnerName         *string   `json:"owner_name,omitempty"`
	OwnerTaxID        *string   `json:"owner_tax_id,omitempty"`
	IdempotencyKey    string    `json:"idempotency_key"`
	RequestID         string    `json:"request_id"`
	Timestamp         time.Time `json:"timestamp"`
}

// EntryDeletedEvent represents a Pulsar event for immediate entry deletion
type EntryDeletedEvent struct {
	EntryID        string    `json:"entry_id"`
	Key            string    `json:"key"`
	KeyType        string    `json:"key_type"`
	Reason         string    `json:"reason"`
	IdempotencyKey string    `json:"idempotency_key"`
	RequestID      string    `json:"request_id"`
	Timestamp      time.Time `json:"timestamp"`
}

// NewConsumer creates a new Pulsar consumer for Entry events
func NewConsumer(
	config ConsumerConfig,
	entryRepo *repositories.EntryRepository,
	bridgeClient bridgepb.BridgeServiceClient,
	logger *logrus.Logger,
) (*Consumer, error) {
	if entryRepo == nil {
		return nil, fmt.Errorf("entryRepo cannot be nil")
	}
	if bridgeClient == nil {
		return nil, fmt.Errorf("bridgeClient cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Create Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     config.URL,
		ConnectionTimeout:       config.ConnectionTimeout,
		OperationTimeout:        config.OperationTimeout,
		MaxConnectionsPerBroker: 5,
		Logger:                  nil, // Use custom logger
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"url":          config.URL,
		"subscription": config.Subscription,
	}).Info("Pulsar consumer client created")

	return &Consumer{
		client:       client,
		consumers:    make([]pulsar.Consumer, 0, 3),
		entryRepo:    entryRepo,
		bridgeClient: bridgeClient,
		logger:       logger,
		stopChan:     make(chan struct{}),
		config:       config,
	}, nil
}

// Start subscribes to all Entry topics and starts consuming messages
func (c *Consumer) Start(ctx context.Context) error {
	// Subscribe to dict.entries.created
	consumerCreated, err := c.client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       c.config.TopicEntryCreated,
		SubscriptionName:            c.config.Subscription + "-created",
		Type:                        pulsar.Shared, // Multiple workers can share subscription
		SubscriptionInitialPosition: pulsar.SubscriptionPositionLatest,
		NackRedeliveryDelay:         c.config.NackRedeliveryDelay,
		RetryEnable:                 true,
		MaxReconnectToBroker:        &c.config.MaxReconnectToBroker,
		ReceiverQueueSize:           1000,
	})
	if err != nil {
		c.client.Close()
		return fmt.Errorf("failed to subscribe to %s: %w", c.config.TopicEntryCreated, err)
	}
	c.consumers = append(c.consumers, consumerCreated)
	c.logger.Infof("Subscribed to topic: %s", c.config.TopicEntryCreated)

	// Subscribe to dict.entries.updated
	consumerUpdated, err := c.client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       c.config.TopicEntryUpdated,
		SubscriptionName:            c.config.Subscription + "-updated",
		Type:                        pulsar.Shared,
		SubscriptionInitialPosition: pulsar.SubscriptionPositionLatest,
		NackRedeliveryDelay:         c.config.NackRedeliveryDelay,
		RetryEnable:                 true,
		MaxReconnectToBroker:        &c.config.MaxReconnectToBroker,
		ReceiverQueueSize:           1000,
	})
	if err != nil {
		c.Stop()
		return fmt.Errorf("failed to subscribe to %s: %w", c.config.TopicEntryUpdated, err)
	}
	c.consumers = append(c.consumers, consumerUpdated)
	c.logger.Infof("Subscribed to topic: %s", c.config.TopicEntryUpdated)

	// Subscribe to dict.entries.deleted.immediate
	consumerDeleted, err := c.client.Subscribe(pulsar.ConsumerOptions{
		Topic:                       c.config.TopicEntryDeletedImmed,
		SubscriptionName:            c.config.Subscription + "-deleted-immediate",
		Type:                        pulsar.Shared,
		SubscriptionInitialPosition: pulsar.SubscriptionPositionLatest,
		NackRedeliveryDelay:         c.config.NackRedeliveryDelay,
		RetryEnable:                 true,
		MaxReconnectToBroker:        &c.config.MaxReconnectToBroker,
		ReceiverQueueSize:           1000,
	})
	if err != nil {
		c.Stop()
		return fmt.Errorf("failed to subscribe to %s: %w", c.config.TopicEntryDeletedImmed, err)
	}
	c.consumers = append(c.consumers, consumerDeleted)
	c.logger.Infof("Subscribed to topic: %s", c.config.TopicEntryDeletedImmed)

	// Start goroutines to consume messages
	c.wg.Add(3)
	go c.consumeCreatedEvents(ctx, consumerCreated)
	go c.consumeUpdatedEvents(ctx, consumerUpdated)
	go c.consumeDeletedEvents(ctx, consumerDeleted)

	c.logger.Info("Pulsar consumer started successfully - processing Entry events")
	return nil
}

// consumeCreatedEvents consumes messages from dict.entries.created topic
func (c *Consumer) consumeCreatedEvents(ctx context.Context, consumer pulsar.Consumer) {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopChan:
			c.logger.Info("Stopping dict.entries.created consumer")
			return
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping dict.entries.created consumer")
			return
		default:
			msg, err := consumer.Receive(ctx)
			if err != nil {
				c.logger.Errorf("Error receiving message from dict.entries.created: %v", err)
				continue
			}

			// Process message
			if err := c.handleEntryCreated(ctx, msg); err != nil {
				c.logger.WithError(err).WithFields(logrus.Fields{
					"message_id": msg.ID(),
					"topic":      msg.Topic(),
				}).Error("Failed to process EntryCreated event")

				// Nack message for redelivery
				consumer.Nack(msg)
			} else {
				// Ack message on success
				consumer.Ack(msg)
			}
		}
	}
}

// consumeUpdatedEvents consumes messages from dict.entries.updated topic
func (c *Consumer) consumeUpdatedEvents(ctx context.Context, consumer pulsar.Consumer) {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopChan:
			c.logger.Info("Stopping dict.entries.updated consumer")
			return
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping dict.entries.updated consumer")
			return
		default:
			msg, err := consumer.Receive(ctx)
			if err != nil {
				c.logger.Errorf("Error receiving message from dict.entries.updated: %v", err)
				continue
			}

			// Process message
			if err := c.handleEntryUpdated(ctx, msg); err != nil {
				c.logger.WithError(err).WithFields(logrus.Fields{
					"message_id": msg.ID(),
					"topic":      msg.Topic(),
				}).Error("Failed to process EntryUpdated event")

				// Nack message for redelivery
				consumer.Nack(msg)
			} else {
				// Ack message on success
				consumer.Ack(msg)
			}
		}
	}
}

// consumeDeletedEvents consumes messages from dict.entries.deleted.immediate topic
func (c *Consumer) consumeDeletedEvents(ctx context.Context, consumer pulsar.Consumer) {
	defer c.wg.Done()

	for {
		select {
		case <-c.stopChan:
			c.logger.Info("Stopping dict.entries.deleted.immediate consumer")
			return
		case <-ctx.Done():
			c.logger.Info("Context cancelled, stopping dict.entries.deleted.immediate consumer")
			return
		default:
			msg, err := consumer.Receive(ctx)
			if err != nil {
				c.logger.Errorf("Error receiving message from dict.entries.deleted.immediate: %v", err)
				continue
			}

			// Process message
			if err := c.handleEntryDeleteImmediate(ctx, msg); err != nil {
				c.logger.WithError(err).WithFields(logrus.Fields{
					"message_id": msg.ID(),
					"topic":      msg.Topic(),
				}).Error("Failed to process EntryDeleted event")

				// Nack message for redelivery
				consumer.Nack(msg)
			} else {
				// Ack message on success
				consumer.Ack(msg)
			}
		}
	}
}

// handleEntryCreated processes EntryCreated event by calling Bridge gRPC directly
// NO TEMPORAL WORKFLOW - This is a fast operation (< 1.5s)
func (c *Consumer) handleEntryCreated(ctx context.Context, msg pulsar.Message) error {
	startTime := time.Now()

	// Parse event
	var event EntryCreatedEvent
	if err := json.Unmarshal(msg.Payload(), &event); err != nil {
		return fmt.Errorf("failed to unmarshal EntryCreated event: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"entry_id":   event.EntryID,
		"key":        event.Key,
		"key_type":   event.KeyType,
		"request_id": event.RequestID,
	}).Info("Processing EntryCreated event")

	// 1. Fetch Entry from database
	entry, err := c.entryRepo.GetByEntryID(ctx, event.EntryID)
	if err != nil {
		return fmt.Errorf("failed to get entry from DB: %w", err)
	}

	// 2. Build Bridge gRPC request
	req := &bridgepb.CreateEntryRequest{
		Key: &commonpb.DictKey{
			KeyType:  c.mapKeyType(event.KeyType),
			KeyValue: event.Key,
		},
		Account: &commonpb.Account{
			Ispb:                  event.Participant,
			BranchCode:            ptrToString(event.AccountBranch),
			AccountNumber:         ptrToString(event.AccountNumber),
			AccountType:           c.mapAccountType(event.AccountType),
			AccountHolderName:     ptrToString(event.OwnerName),
			AccountHolderDocument: ptrToString(event.OwnerTaxID),
			DocumentType:          c.mapDocumentType(event.OwnerType),
		},
		IdempotencyKey: event.IdempotencyKey,
		RequestId:      event.RequestID,
	}

	// 3. Call Bridge gRPC CreateEntry directly (NO TEMPORAL!)
	resp, err := c.bridgeClient.CreateEntry(ctx, req)
	if err != nil {
		// Update status to FAILED
		c.logger.WithError(err).WithFields(logrus.Fields{
			"entry_id":   event.EntryID,
			"request_id": event.RequestID,
		}).Error("Bridge CreateEntry call failed")

		if updateErr := c.entryRepo.UpdateStatus(ctx, entry.EntryID, entities.EntryStatusInactive); updateErr != nil {
			c.logger.WithError(updateErr).Error("Failed to update entry status to FAILED")
		}

		return fmt.Errorf("bridge CreateEntry failed: %w", err)
	}

	// 4. Update entry status to ACTIVE + store Bacen Entry ID
	entry.Status = entities.EntryStatusActive
	entry.BacenEntryID = &resp.ExternalId
	now := time.Now()
	entry.ActivatedAt = &now
	entry.UpdatedAt = now

	if err := c.entryRepo.Update(ctx, entry); err != nil {
		c.logger.WithError(err).Error("Failed to update entry after Bridge call")
		return fmt.Errorf("failed to update entry: %w", err)
	}

	duration := time.Since(startTime)
	c.logger.WithFields(logrus.Fields{
		"entry_id":        event.EntryID,
		"bacen_entry_id":  resp.ExternalId,
		"status":          "ACTIVE",
		"duration_ms":     duration.Milliseconds(),
		"request_id":      event.RequestID,
	}).Info("EntryCreated processed successfully")

	return nil
}

// handleEntryUpdated processes EntryUpdated event by calling Bridge gRPC directly
// NO TEMPORAL WORKFLOW - This is a fast operation (< 1s)
func (c *Consumer) handleEntryUpdated(ctx context.Context, msg pulsar.Message) error {
	startTime := time.Now()

	// Parse event
	var event EntryUpdatedEvent
	if err := json.Unmarshal(msg.Payload(), &event); err != nil {
		return fmt.Errorf("failed to unmarshal EntryUpdated event: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"entry_id":   event.EntryID,
		"key":        event.Key,
		"request_id": event.RequestID,
	}).Info("Processing EntryUpdated event")

	// 1. Fetch Entry from database
	entry, err := c.entryRepo.GetByEntryID(ctx, event.EntryID)
	if err != nil {
		return fmt.Errorf("failed to get entry from DB: %w", err)
	}

	// 2. Build Bridge gRPC request
	req := &bridgepb.UpdateEntryRequest{
		EntryId: event.EntryID,
		NewAccount: &commonpb.Account{
			Ispb:                  event.Participant,
			BranchCode:            ptrToString(event.AccountBranch),
			AccountNumber:         ptrToString(event.AccountNumber),
			AccountType:           c.mapAccountType(event.AccountType),
			AccountHolderName:     ptrToString(event.OwnerName),
			AccountHolderDocument: ptrToString(event.OwnerTaxID),
			DocumentType:          c.mapDocumentType(event.OwnerType),
		},
		IdempotencyKey: event.IdempotencyKey,
		RequestId:      event.RequestID,
	}

	// 3. Call Bridge gRPC UpdateEntry directly (NO TEMPORAL!)
	resp, err := c.bridgeClient.UpdateEntry(ctx, req)
	if err != nil {
		// Log error but don't change status
		c.logger.WithError(err).WithFields(logrus.Fields{
			"entry_id":   event.EntryID,
			"request_id": event.RequestID,
		}).Error("Bridge UpdateEntry call failed")

		return fmt.Errorf("bridge UpdateEntry failed: %w", err)
	}

	// 4. Update entry in database with new account info
	if event.AccountBranch != nil {
		entry.AccountBranch = event.AccountBranch
	}
	if event.AccountNumber != nil {
		entry.AccountNumber = event.AccountNumber
	}
	if event.OwnerName != nil {
		entry.OwnerName = event.OwnerName
	}
	if event.OwnerTaxID != nil {
		entry.OwnerTaxID = event.OwnerTaxID
	}
	entry.UpdatedAt = time.Now()

	if err := c.entryRepo.Update(ctx, entry); err != nil {
		c.logger.WithError(err).Error("Failed to update entry after Bridge call")
		return fmt.Errorf("failed to update entry: %w", err)
	}

	duration := time.Since(startTime)
	c.logger.WithFields(logrus.Fields{
		"entry_id":    event.EntryID,
		"duration_ms": duration.Milliseconds(),
		"request_id":  event.RequestID,
		"bacen_tx_id": resp.BacenTransactionId,
	}).Info("EntryUpdated processed successfully")

	return nil
}

// handleEntryDeleteImmediate processes EntryDeleted (immediate) event by calling Bridge gRPC
// NO TEMPORAL WORKFLOW - This is a fast operation (< 1s)
// For delete with 30-day waiting period, use DeleteEntryWithWaitingPeriodWorkflow instead
func (c *Consumer) handleEntryDeleteImmediate(ctx context.Context, msg pulsar.Message) error {
	startTime := time.Now()

	// Parse event
	var event EntryDeletedEvent
	if err := json.Unmarshal(msg.Payload(), &event); err != nil {
		return fmt.Errorf("failed to unmarshal EntryDeleted event: %w", err)
	}

	c.logger.WithFields(logrus.Fields{
		"entry_id":   event.EntryID,
		"key":        event.Key,
		"reason":     event.Reason,
		"request_id": event.RequestID,
	}).Info("Processing EntryDeleted (immediate) event")

	// 1. Fetch Entry from database
	entry, err := c.entryRepo.GetByEntryID(ctx, event.EntryID)
	if err != nil {
		return fmt.Errorf("failed to get entry from DB: %w", err)
	}

	// 2. Build Bridge gRPC request
	req := &bridgepb.DeleteEntryRequest{
		EntryId: event.EntryID,
		Key: &commonpb.DictKey{
			KeyType:  c.mapKeyType(event.KeyType),
			KeyValue: event.Key,
		},
		IdempotencyKey: event.IdempotencyKey,
		RequestId:      event.RequestID,
	}

	// 3. Call Bridge gRPC DeleteEntry directly (NO TEMPORAL!)
	resp, err := c.bridgeClient.DeleteEntry(ctx, req)
	if err != nil {
		c.logger.WithError(err).WithFields(logrus.Fields{
			"entry_id":   event.EntryID,
			"request_id": event.RequestID,
		}).Error("Bridge DeleteEntry call failed")

		return fmt.Errorf("bridge DeleteEntry failed: %w", err)
	}

	// 4. Soft delete entry in database
	if err := c.entryRepo.Delete(ctx, entry.EntryID); err != nil {
		c.logger.WithError(err).Error("Failed to soft delete entry after Bridge call")
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	duration := time.Since(startTime)
	c.logger.WithFields(logrus.Fields{
		"entry_id":    event.EntryID,
		"deleted":     resp.Deleted,
		"duration_ms": duration.Milliseconds(),
		"request_id":  event.RequestID,
		"bacen_tx_id": resp.BacenTransactionId,
	}).Info("EntryDeleted processed successfully")

	return nil
}

// Stop gracefully stops the consumer
func (c *Consumer) Stop() {
	c.logger.Info("Stopping Pulsar consumer...")

	// Signal goroutines to stop
	close(c.stopChan)

	// Wait for all goroutines to finish
	c.wg.Wait()

	// Close all consumers
	for _, consumer := range c.consumers {
		consumer.Close()
	}

	// Close client
	c.client.Close()

	c.logger.Info("Pulsar consumer stopped successfully")
}

// Helper functions

func (c *Consumer) mapKeyType(keyType string) commonpb.KeyType {
	switch keyType {
	case "CPF":
		return commonpb.KeyType_KEY_TYPE_CPF
	case "CNPJ":
		return commonpb.KeyType_KEY_TYPE_CNPJ
	case "EMAIL":
		return commonpb.KeyType_KEY_TYPE_EMAIL
	case "PHONE":
		return commonpb.KeyType_KEY_TYPE_PHONE
	case "EVP":
		return commonpb.KeyType_KEY_TYPE_EVP
	default:
		return commonpb.KeyType_KEY_TYPE_UNSPECIFIED
	}
}

func (c *Consumer) mapAccountType(accountType string) commonpb.AccountType {
	switch accountType {
	case "CACC":
		return commonpb.AccountType_ACCOUNT_TYPE_CHECKING
	case "SLRY":
		return commonpb.AccountType_ACCOUNT_TYPE_SALARY
	case "SVGS":
		return commonpb.AccountType_ACCOUNT_TYPE_SAVINGS
	case "TRAN":
		return commonpb.AccountType_ACCOUNT_TYPE_PAYMENT
	default:
		return commonpb.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

func (c *Consumer) mapDocumentType(ownerType string) commonpb.DocumentType {
	switch ownerType {
	case "NATURAL_PERSON":
		return commonpb.DocumentType_DOCUMENT_TYPE_CPF
	case "LEGAL_PERSON":
		return commonpb.DocumentType_DOCUMENT_TYPE_CNPJ
	default:
		return commonpb.DocumentType_DOCUMENT_TYPE_UNSPECIFIED
	}
}

func ptrToString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

// DefaultConsumerConfig returns default consumer configuration
func DefaultConsumerConfig() ConsumerConfig {
	return ConsumerConfig{
		URL:                    "pulsar://localhost:6650",
		Subscription:           "conn-dict-entry-consumer",
		TopicEntryCreated:      "dict.entries.created",
		TopicEntryUpdated:      "dict.entries.updated",
		TopicEntryDeletedImmed: "dict.entries.deleted.immediate",
		MaxReconnectToBroker:   10,
		ConnectionTimeout:      30 * time.Second,
		OperationTimeout:       30 * time.Second,
		AckTimeout:             20 * time.Second,
		NackRedeliveryDelay:    60 * time.Second,
		MaxDeliveryAttempts:    5,
	}
}
