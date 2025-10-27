// Package messaging provides Pulsar event producers for DICT Entry events
// This producer publishes Entry lifecycle events (Created, Updated, Deleted) to conn-dict service
package messaging

import (
	"context"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	connectv1 "github.com/lbpay-lab/dict-contracts/proto/conn_dict/v1"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// EntryEventProducer publishes DICT Entry events to Pulsar
// Topics:
//   - dict.entries.created: Entry creation events (Core → Connect)
//   - dict.entries.updated: Entry update events (Core → Connect)
//   - dict.entries.deleted.immediate: Entry deletion events (Core → Connect)
type EntryEventProducer struct {
	client          pulsar.Client
	createdProducer pulsar.Producer
	updatedProducer pulsar.Producer
	deletedProducer pulsar.Producer
	config          *ProducerConfig
}

// NewEntryEventProducer creates a new Entry event producer
// pulsarURL: Pulsar broker URL (e.g., "pulsar://localhost:6650")
func NewEntryEventProducer(pulsarURL string) (*EntryEventProducer, error) {
	config := DefaultEntryProducerConfig()
	config.PulsarURL = pulsarURL

	// Create Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     config.PulsarURL,
		ConnectionTimeout:       5 * time.Second,
		OperationTimeout:        30 * time.Second,
		MaxConnectionsPerBroker: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	// Create producer for "dict.entries.created" topic
	createdProducer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:                   "persistent://public/default/dict.entries.created",
		Name:                    "core-dict-entry-created",
		CompressionType:         config.CompressionType,
		BatchingMaxMessages:     config.BatchingMaxMessages,
		BatchingMaxPublishDelay: config.BatchingMaxPublishDelay,
		MaxPendingMessages:      config.MaxPendingMessages,
		SendTimeout:             config.SendTimeout,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create 'created' producer: %w", err)
	}

	// Create producer for "dict.entries.updated" topic
	updatedProducer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:                   "persistent://public/default/dict.entries.updated",
		Name:                    "core-dict-entry-updated",
		CompressionType:         config.CompressionType,
		BatchingMaxMessages:     config.BatchingMaxMessages,
		BatchingMaxPublishDelay: config.BatchingMaxPublishDelay,
		MaxPendingMessages:      config.MaxPendingMessages,
		SendTimeout:             config.SendTimeout,
	})
	if err != nil {
		createdProducer.Close()
		client.Close()
		return nil, fmt.Errorf("failed to create 'updated' producer: %w", err)
	}

	// Create producer for "dict.entries.deleted.immediate" topic
	deletedProducer, err := client.CreateProducer(pulsar.ProducerOptions{
		Topic:                   "persistent://public/default/dict.entries.deleted.immediate",
		Name:                    "core-dict-entry-deleted",
		CompressionType:         config.CompressionType,
		BatchingMaxMessages:     config.BatchingMaxMessages,
		BatchingMaxPublishDelay: config.BatchingMaxPublishDelay,
		MaxPendingMessages:      config.MaxPendingMessages,
		SendTimeout:             config.SendTimeout,
	})
	if err != nil {
		updatedProducer.Close()
		createdProducer.Close()
		client.Close()
		return nil, fmt.Errorf("failed to create 'deleted' producer: %w", err)
	}

	return &EntryEventProducer{
		client:          client,
		createdProducer: createdProducer,
		updatedProducer: updatedProducer,
		deletedProducer: deletedProducer,
		config:          config,
	}, nil
}

// PublishCreated publishes an EntryCreatedEvent
// Flow: Core DICT → Pulsar → Connect → Bridge → Bacen DICT
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - entry: The Entry entity that was created
//   - userID: ID of the user who created the entry
//
// Returns error if serialization or send fails
func (p *EntryEventProducer) PublishCreated(
	ctx context.Context,
	entry *entities.Entry,
	userID string,
) error {
	startTime := time.Now()

	// Build protobuf event
	event := &connectv1.EntryCreatedEvent{
		EntryId:         entry.ID.String(),
		ParticipantIspb: entry.ISPB,
		KeyType:         mapKeyType(entry.KeyType),
		KeyValue:        entry.KeyValue,
		Account:         mapAccountFromEntry(entry),
		IdempotencyKey:  uuid.NewString(),
		RequestId:       uuid.NewString(),
		UserId:          userID,
		CreatedAt:       timestamppb.New(entry.CreatedAt),
		Metadata: map[string]string{
			"source":      "core-dict",
			"entry_id":    entry.ID.String(),
			"key_type":    string(entry.KeyType),
			"participant": entry.ISPB,
		},
	}

	// Serialize protobuf to bytes
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal EntryCreatedEvent: %w", err)
	}

	// Send to Pulsar
	messageID, err := p.createdProducer.Send(ctx, &pulsar.ProducerMessage{
		Payload:   data,
		Key:       entry.ID.String(), // Partition key = EntryID (ensures FIFO ordering)
		EventTime: time.Now(),
		Properties: map[string]string{
			"event_type":      "entry.created",
			"version":         "v1",
			"entry_id":        entry.ID.String(),
			"key_type":        string(entry.KeyType),
			"key_value":       entry.KeyValue,
			"idempotency_key": event.IdempotencyKey,
			"request_id":      event.RequestId,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send EntryCreatedEvent to Pulsar: %w", err)
	}

	// Log latency (in production, use proper structured logger)
	latency := time.Since(startTime)
	fmt.Printf("[EntryEventProducer] Published EntryCreatedEvent: messageID=%s, entryID=%s, latency=%dms\n",
		messageID.String(), entry.ID.String(), latency.Milliseconds())

	return nil
}

// PublishUpdated publishes an EntryUpdatedEvent
// Flow: Core DICT → Pulsar → Connect → Bridge → Bacen DICT
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - entry: The Entry entity that was updated
//   - userID: ID of the user who updated the entry
//
// Returns error if serialization or send fails
func (p *EntryEventProducer) PublishUpdated(
	ctx context.Context,
	entry *entities.Entry,
	userID string,
) error {
	startTime := time.Now()

	// Build protobuf event
	event := &connectv1.EntryUpdatedEvent{
		EntryId:         entry.ID.String(),
		ParticipantIspb: entry.ISPB,
		KeyType:         mapKeyType(entry.KeyType),
		KeyValue:        entry.KeyValue,
		NewAccount:      mapAccountFromEntry(entry),
		IdempotencyKey:  uuid.NewString(),
		RequestId:       uuid.NewString(),
		UserId:          userID,
		UpdatedAt:       timestamppb.New(entry.UpdatedAt),
		Metadata: map[string]string{
			"source":      "core-dict",
			"entry_id":    entry.ID.String(),
			"key_type":    string(entry.KeyType),
			"participant": entry.ISPB,
		},
	}

	// Serialize protobuf to bytes
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal EntryUpdatedEvent: %w", err)
	}

	// Send to Pulsar
	messageID, err := p.updatedProducer.Send(ctx, &pulsar.ProducerMessage{
		Payload:   data,
		Key:       entry.ID.String(), // Partition key = EntryID (ensures FIFO ordering)
		EventTime: time.Now(),
		Properties: map[string]string{
			"event_type":      "entry.updated",
			"version":         "v1",
			"entry_id":        entry.ID.String(),
			"key_type":        string(entry.KeyType),
			"key_value":       entry.KeyValue,
			"idempotency_key": event.IdempotencyKey,
			"request_id":      event.RequestId,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send EntryUpdatedEvent to Pulsar: %w", err)
	}

	// Log latency
	latency := time.Since(startTime)
	fmt.Printf("[EntryEventProducer] Published EntryUpdatedEvent: messageID=%s, entryID=%s, latency=%dms\n",
		messageID.String(), entry.ID.String(), latency.Milliseconds())

	return nil
}

// PublishDeleted publishes an EntryDeletedEvent
// Flow: Core DICT → Pulsar → Connect → Bridge → Bacen DICT
//
// Parameters:
//   - ctx: Context for cancellation and timeouts
//   - entryID: UUID of the deleted entry
//   - keyValue: The key value that was deleted
//   - keyType: Type of the key (CPF, CNPJ, EMAIL, PHONE, EVP)
//   - participantISPB: ISPB of the participant
//   - deletionType: IMMEDIATE or WAITING_PERIOD
//   - userID: ID of the user who deleted the entry
//
// Returns error if serialization or send fails
func (p *EntryEventProducer) PublishDeleted(
	ctx context.Context,
	entryID uuid.UUID,
	keyValue string,
	keyType entities.KeyType,
	participantISPB string,
	deletionType connectv1.EntryDeletedEvent_DeletionType,
	userID string,
) error {
	startTime := time.Now()

	// Build protobuf event
	event := &connectv1.EntryDeletedEvent{
		EntryId:         entryID.String(),
		ParticipantIspb: participantISPB,
		KeyType:         mapKeyType(keyType),
		KeyValue:        keyValue,
		DeletionType:    deletionType,
		IdempotencyKey:  uuid.NewString(),
		RequestId:       uuid.NewString(),
		UserId:          userID,
		DeletedAt:       timestamppb.Now(),
		Metadata: map[string]string{
			"source":      "core-dict",
			"entry_id":    entryID.String(),
			"key_type":    string(keyType),
			"participant": participantISPB,
		},
	}

	// Serialize protobuf to bytes
	data, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal EntryDeletedEvent: %w", err)
	}

	// Send to Pulsar
	messageID, err := p.deletedProducer.Send(ctx, &pulsar.ProducerMessage{
		Payload:   data,
		Key:       entryID.String(), // Partition key = EntryID (ensures FIFO ordering)
		EventTime: time.Now(),
		Properties: map[string]string{
			"event_type":      "entry.deleted",
			"version":         "v1",
			"entry_id":        entryID.String(),
			"key_type":        string(keyType),
			"key_value":       keyValue,
			"deletion_type":   deletionType.String(),
			"idempotency_key": event.IdempotencyKey,
			"request_id":      event.RequestId,
		},
	})
	if err != nil {
		return fmt.Errorf("failed to send EntryDeletedEvent to Pulsar: %w", err)
	}

	// Log latency
	latency := time.Since(startTime)
	fmt.Printf("[EntryEventProducer] Published EntryDeletedEvent: messageID=%s, entryID=%s, latency=%dms\n",
		messageID.String(), entryID.String(), latency.Milliseconds())

	return nil
}

// PublishDeletedImmediate is a convenience method for immediate deletions
func (p *EntryEventProducer) PublishDeletedImmediate(
	ctx context.Context,
	entry *entities.Entry,
	userID string,
) error {
	return p.PublishDeleted(
		ctx,
		entry.ID,
		entry.KeyValue,
		entry.KeyType,
		entry.ISPB,
		connectv1.EntryDeletedEvent_DELETION_TYPE_IMMEDIATE,
		userID,
	)
}

// Flush flushes all pending messages (useful before shutdown)
func (p *EntryEventProducer) Flush() error {
	if err := p.createdProducer.Flush(); err != nil {
		return fmt.Errorf("failed to flush 'created' producer: %w", err)
	}
	if err := p.updatedProducer.Flush(); err != nil {
		return fmt.Errorf("failed to flush 'updated' producer: %w", err)
	}
	if err := p.deletedProducer.Flush(); err != nil {
		return fmt.Errorf("failed to flush 'deleted' producer: %w", err)
	}
	return nil
}

// Close closes all producers and the client
func (p *EntryEventProducer) Close() error {
	var errMsg string

	if p.createdProducer != nil {
		p.createdProducer.Close()
	}
	if p.updatedProducer != nil {
		p.updatedProducer.Close()
	}
	if p.deletedProducer != nil {
		p.deletedProducer.Close()
	}
	if p.client != nil {
		p.client.Close()
	}

	if errMsg != "" {
		return fmt.Errorf("close errors: %s", errMsg)
	}
	return nil
}

// ========================================================================
// MAPPING HELPERS - Domain → Protobuf
// ========================================================================

// mapKeyType maps domain.KeyType to proto commonv1.KeyType
func mapKeyType(kt entities.KeyType) commonv1.KeyType {
	switch kt {
	case entities.KeyTypeCPF:
		return commonv1.KeyType_KEY_TYPE_CPF
	case entities.KeyTypeCNPJ:
		return commonv1.KeyType_KEY_TYPE_CNPJ
	case entities.KeyTypeEmail:
		return commonv1.KeyType_KEY_TYPE_EMAIL
	case entities.KeyTypePhone:
		return commonv1.KeyType_KEY_TYPE_PHONE
	case entities.KeyTypeEVP:
		return commonv1.KeyType_KEY_TYPE_EVP
	default:
		return commonv1.KeyType_KEY_TYPE_UNSPECIFIED
	}
}

// mapAccountFromEntry maps Entry fields to proto Account message
func mapAccountFromEntry(entry *entities.Entry) *commonv1.Account {
	return &commonv1.Account{
		Ispb:                  entry.ISPB,
		AccountType:           mapAccountType(entry.AccountType),
		AccountNumber:         entry.AccountNumber,
		AccountCheckDigit:     "", // Not stored in Entry, will need to add if required
		BranchCode:            entry.Branch,
		AccountHolderName:     entry.OwnerName,
		AccountHolderDocument: entry.OwnerTaxID,
		DocumentType:          mapDocumentType(entry.OwnerTaxID),
	}
}

// mapAccountType maps string account type to proto AccountType
func mapAccountType(accountType string) commonv1.AccountType {
	switch accountType {
	case "CACC", "CHECKING":
		return commonv1.AccountType_ACCOUNT_TYPE_CHECKING
	case "SVGS", "SAVINGS":
		return commonv1.AccountType_ACCOUNT_TYPE_SAVINGS
	case "TRAN", "PAYMENT":
		return commonv1.AccountType_ACCOUNT_TYPE_PAYMENT
	case "SLRY", "SALARY":
		return commonv1.AccountType_ACCOUNT_TYPE_SALARY
	default:
		return commonv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

// mapDocumentType infers document type from TaxID length
func mapDocumentType(taxID string) commonv1.DocumentType {
	switch len(taxID) {
	case 11:
		return commonv1.DocumentType_DOCUMENT_TYPE_CPF
	case 14:
		return commonv1.DocumentType_DOCUMENT_TYPE_CNPJ
	default:
		return commonv1.DocumentType_DOCUMENT_TYPE_UNSPECIFIED
	}
}
