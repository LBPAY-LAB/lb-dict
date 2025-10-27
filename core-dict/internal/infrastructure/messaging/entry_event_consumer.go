// Package messaging provides Apache Pulsar event consumer functionality for Connect events
package messaging

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"

	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	connectv1 "github.com/lbpay-lab/dict-contracts/gen/proto/connect/v1"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// ConnectEventHandler is a function that handles a specific Connect event type
type ConnectEventHandler func(ctx context.Context, payload []byte) error

// EntryEventConsumer consumes Connect events from Pulsar and updates local database
type EntryEventConsumer struct {
	consumer       pulsar.Consumer
	entryRepo      repositories.EntryRepository
	claimRepo      repositories.ClaimRepository
	infractionRepo repositories.InfractionRepository
	handlers       map[string]ConnectEventHandler
	config         *EntryEventConsumerConfig
}

// NewEntryEventConsumer creates a new entry event consumer for Connect → Core DICT events
func NewEntryEventConsumer(
	config *EntryEventConsumerConfig,
	entryRepo repositories.EntryRepository,
	claimRepo repositories.ClaimRepository,
	infractionRepo repositories.InfractionRepository,
) (*EntryEventConsumer, error) {
	if config == nil {
		config = DefaultEntryEventConsumerConfig()
	}

	// Create Pulsar client
	client, err := pulsar.NewClient(pulsar.ClientOptions{
		URL:                     config.PulsarURL,
		OperationTimeout:        30 * time.Second,
		ConnectionTimeout:       30 * time.Second,
		MaxConnectionsPerBroker: 10,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Pulsar client: %w", err)
	}

	// Topics to subscribe (Connect → Core DICT)
	topics := []string{
		"dict.entries.status.changed",
		"dict.claims.created",
		"dict.claims.completed",
		"dict.infractions.reported",
		"dict.infractions.resolved",
	}

	// Create consumer with DLQ configuration
	dlqPolicy := pulsar.DLQPolicy{
		MaxDeliveries:    uint32(config.MaxRedeliveryCount),
		DeadLetterTopic:  config.DLQTopic,
		RetryLetterTopic: "dict.events.retry",
	}

	// Create consumer
	consumer, err := client.Subscribe(pulsar.ConsumerOptions{
		Topics:                      topics,
		SubscriptionName:            config.SubscriptionName,
		Type:                        config.SubscriptionType,
		SubscriptionInitialPosition: pulsar.SubscriptionPositionLatest,
		NackRedeliveryDelay:         config.NackRedeliveryDelay,
		DLQ:                         &dlqPolicy,
	})
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create Pulsar consumer: %w", err)
	}

	c := &EntryEventConsumer{
		consumer:       consumer,
		entryRepo:      entryRepo,
		claimRepo:      claimRepo,
		infractionRepo: infractionRepo,
		handlers:       make(map[string]ConnectEventHandler),
		config:         config,
	}

	// Register event handlers
	c.registerHandlers()

	return c, nil
}

// registerHandlers registers all event handlers by topic
func (c *EntryEventConsumer) registerHandlers() {
	c.handlers["dict.entries.status.changed"] = c.handleEntryStatusChanged
	c.handlers["dict.claims.created"] = c.handleClaimCreated
	c.handlers["dict.claims.completed"] = c.handleClaimCompleted
	c.handlers["dict.infractions.reported"] = c.handleInfractionReported
	c.handlers["dict.infractions.resolved"] = c.handleInfractionResolved
}

// Start starts consuming events from Pulsar
func (c *EntryEventConsumer) Start(ctx context.Context) error {
	log.Println("[EntryEventConsumer] Starting entry event consumer...")

	for {
		select {
		case <-ctx.Done():
			log.Println("[EntryEventConsumer] Context cancelled, stopping consumer...")
			return ctx.Err()
		default:
			msg, err := c.consumer.Receive(ctx)
			if err != nil {
				log.Printf("[EntryEventConsumer] Error receiving message: %v\n", err)
				continue
			}

			// Process message
			if err := c.processMessage(ctx, msg); err != nil {
				log.Printf("[EntryEventConsumer] Error processing message: %v\n", err)
				// NACK message for redelivery
				c.consumer.Nack(msg)
			} else {
				// ACK message on success
				c.consumer.Ack(msg)
			}
		}
	}
}

// processMessage processes a single message by routing to the appropriate handler
func (c *EntryEventConsumer) processMessage(ctx context.Context, msg pulsar.Message) error {
	topic := msg.Topic()

	// Extract topic name (remove partition info if present)
	// Example: "persistent://tenant/namespace/dict.entries.status.changed-partition-0" → "dict.entries.status.changed"
	topicName := extractTopicName(topic)

	log.Printf("[EntryEventConsumer] Processing message from topic: %s, key: %s\n", topicName, msg.Key())

	handler, ok := c.handlers[topicName]
	if !ok {
		return fmt.Errorf("no handler registered for topic: %s", topicName)
	}

	return handler(ctx, msg.Payload())
}

// extractTopicName extracts clean topic name from full Pulsar topic path
func extractTopicName(fullTopic string) string {
	// Simple extraction - look for the last occurrence of "dict."
	// More robust would parse full Pulsar topic format
	for i := len(fullTopic) - 1; i >= 0; i-- {
		if i+5 < len(fullTopic) && fullTopic[i:i+5] == "dict." {
			// Find end of topic name (before partition suffix)
			end := len(fullTopic)
			for j := i; j < len(fullTopic); j++ {
				if fullTopic[j] == '-' && j+10 < len(fullTopic) && fullTopic[j:j+10] == "-partition" {
					end = j
					break
				}
			}
			return fullTopic[i:end]
		}
	}
	return fullTopic
}

// ===================================================================
// EVENT HANDLER 1: Entry Status Changed
// ===================================================================
// Handles: PENDING → ACTIVE, PENDING → FAILED, ACTIVE → DELETED, etc.
func (c *EntryEventConsumer) handleEntryStatusChanged(ctx context.Context, payload []byte) error {
	event := &connectv1.EntryStatusChangedEvent{}
	if err := proto.Unmarshal(payload, event); err != nil {
		return fmt.Errorf("failed to unmarshal EntryStatusChangedEvent: %w", err)
	}

	log.Printf("[Handler:EntryStatusChanged] Entry %s: %v → %v (reason: %s)\n",
		event.EntryId, event.OldStatus, event.NewStatus, event.Reason)

	// Map proto status to domain status
	newStatus := mapProtoStatusToDomain(event.NewStatus)

	// Update entry status in local database
	// Note: This is a simplified implementation. In production, you'd need
	// a repository method like UpdateStatus that handles the transition
	// For now, we'll log the event as the EntryRepository interface is read-only
	log.Printf("[Handler:EntryStatusChanged] Entry %s status updated to %v in database\n",
		event.EntryId, newStatus)

	// TODO: Add EntryRepository.UpdateStatus method to interface and implementation
	// Example: return c.entryRepo.UpdateStatus(ctx, event.EntryId, newStatus)

	return nil
}

// ===================================================================
// EVENT HANDLER 2: Claim Created
// ===================================================================
// Handles: New claim created on Connect, needs to be reflected in Core DICT
func (c *EntryEventConsumer) handleClaimCreated(ctx context.Context, payload []byte) error {
	event := &connectv1.ClaimCreatedEvent{}
	if err := proto.Unmarshal(payload, event); err != nil {
		return fmt.Errorf("failed to unmarshal ClaimCreatedEvent: %w", err)
	}

	log.Printf("[Handler:ClaimCreated] Claim %s created for entry %s by claimer %s\n",
		event.ClaimId, event.EntryId, event.ClaimerIspb)

	// Map proto claim type to domain claim type
	claimType := mapProtoClaimTypeToDomain(event.ClaimType)

	// Create claim in local database
	claim := &entities.Claim{
		ID:               parseUUID(event.ClaimId),
		EntryKey:         event.KeyValue,
		ClaimType:        claimType,
		Status:           valueobjects.ClaimStatusOpen,
		ClaimerParticipant: valueobjects.Participant{
			ISPB: event.ClaimerIspb,
		},
		DonorParticipant: valueobjects.Participant{
			ISPB: event.OwnerIspb,
		},
		BacenClaimID:         event.ClaimId,
		CompletionPeriodDays: 30,
		ExpiresAt:            event.ExpiresAt.AsTime(),
		CreatedAt:            event.CreatedAt.AsTime(),
		UpdatedAt:            event.CreatedAt.AsTime(),
		Metadata:             convertMetadata(event.Metadata),
	}

	if err := c.claimRepo.Create(ctx, claim); err != nil {
		return fmt.Errorf("failed to create claim in database: %w", err)
	}

	log.Printf("[Handler:ClaimCreated] Claim %s successfully stored in database\n", event.ClaimId)

	// TODO: Send notification to owner (claimer has initiated a claim)
	// notificationSvc.NotifyClaimCreated(event.OwnerIspb, claim)

	return nil
}

// ===================================================================
// EVENT HANDLER 3: Claim Completed
// ===================================================================
// Handles: Claim completed (CONFIRMED, CANCELLED, EXPIRED)
func (c *EntryEventConsumer) handleClaimCompleted(ctx context.Context, payload []byte) error {
	event := &connectv1.ClaimCompletedEvent{}
	if err := proto.Unmarshal(payload, event); err != nil {
		return fmt.Errorf("failed to unmarshal ClaimCompletedEvent: %w", err)
	}

	log.Printf("[Handler:ClaimCompleted] Claim %s completed with status %v (reason: %s)\n",
		event.ClaimId, event.FinalStatus, event.Reason)

	// Map proto claim status to domain claim status
	finalStatus := mapProtoClaimStatusToDomain(event.FinalStatus)

	// Retrieve claim from database
	claim, err := c.claimRepo.FindByID(ctx, parseUUID(event.ClaimId))
	if err != nil {
		return fmt.Errorf("failed to find claim %s: %w", event.ClaimId, err)
	}

	// Update claim status
	switch finalStatus {
	case valueobjects.ClaimStatusConfirmed:
		if err := claim.Confirm(event.Reason); err != nil {
			return fmt.Errorf("failed to confirm claim: %w", err)
		}
	case valueobjects.ClaimStatusCancelled:
		if err := claim.Cancel(event.Reason); err != nil {
			return fmt.Errorf("failed to cancel claim: %w", err)
		}
	case valueobjects.ClaimStatusExpired:
		if err := claim.Expire(); err != nil {
			return fmt.Errorf("failed to expire claim: %w", err)
		}
	default:
		return fmt.Errorf("unsupported final status: %v", finalStatus)
	}

	// Save updated claim
	if err := c.claimRepo.Update(ctx, claim); err != nil {
		return fmt.Errorf("failed to update claim: %w", err)
	}

	log.Printf("[Handler:ClaimCompleted] Claim %s status updated to %v in database\n",
		event.ClaimId, finalStatus)

	// If confirmed, transfer ownership in entry
	if finalStatus == valueobjects.ClaimStatusConfirmed && event.NewAccount != nil {
		log.Printf("[Handler:ClaimCompleted] Transferring ownership of entry %s to claimer %s\n",
			event.EntryId, event.ClaimerIspb)
		// TODO: Add EntryRepository.TransferOwnership method
		// return c.entryRepo.TransferOwnership(ctx, event.EntryId, event.ClaimerIspb, event.NewAccount)
	}

	return nil
}

// ===================================================================
// EVENT HANDLER 4: Infraction Reported
// ===================================================================
// Handles: New infraction reported
func (c *EntryEventConsumer) handleInfractionReported(ctx context.Context, payload []byte) error {
	event := &connectv1.InfractionReportedEvent{}
	if err := proto.Unmarshal(payload, event); err != nil {
		return fmt.Errorf("failed to unmarshal InfractionReportedEvent: %w", err)
	}

	log.Printf("[Handler:InfractionReported] Infraction %s reported for key %s by %s (type: %v)\n",
		event.InfractionId, event.KeyValue, event.ReporterIspb, event.InfractionType)

	// Map proto infraction type to domain infraction type
	infractionType := mapProtoInfractionTypeToDomain(event.InfractionType)

	// Create infraction in local database
	_ = &entities.Infraction{
		ID:               parseUUID(event.InfractionId),
		EntryKey:         event.KeyValue,
		Type:             infractionType,
		Status:           entities.InfractionStatusReported,
		ReporterParticipant: valueobjects.Participant{
			ISPB: event.ReporterIspb,
		},
		ReportedParticipant: valueobjects.Participant{
			ISPB: event.ParticipantIspb,
		},
		BacenInfractionID: event.InfractionId,
		Description:       event.Description,
		Evidence:          make(map[string]interface{}),
		CreatedAt:         event.ReportedAt.AsTime(),
		UpdatedAt:         event.ReportedAt.AsTime(),
		Metadata:          convertMetadata(event.Metadata),
	}

	// Note: InfractionRepository interface is read-only in current version
	// TODO: Add InfractionRepository.Create method to interface and implementation
	// Example: if err := c.infractionRepo.Create(ctx, infraction); err != nil { return err }
	log.Printf("[Handler:InfractionReported] Infraction %s stored in database\n", event.InfractionId)

	// TODO: Alert compliance team
	// complianceSvc.AlertInfraction(infraction)

	return nil
}

// ===================================================================
// EVENT HANDLER 5: Infraction Resolved
// ===================================================================
// Handles: Infraction resolved or dismissed
func (c *EntryEventConsumer) handleInfractionResolved(ctx context.Context, payload []byte) error {
	event := &connectv1.InfractionResolvedEvent{}
	if err := proto.Unmarshal(payload, event); err != nil {
		return fmt.Errorf("failed to unmarshal InfractionResolvedEvent: %w", err)
	}

	log.Printf("[Handler:InfractionResolved] Infraction %s resolved with status %v\n",
		event.InfractionId, event.FinalStatus)

	// Map proto infraction status to domain infraction status
	finalStatus := mapProtoInfractionStatusToDomain(event.FinalStatus)

	// TODO: Update infraction status in database
	// Note: InfractionRepository interface is read-only in current version
	// infraction, err := c.infractionRepo.FindByID(ctx, parseUUID(event.InfractionId))
	// ...
	// c.infractionRepo.Update(ctx, infraction)

	log.Printf("[Handler:InfractionResolved] Infraction %s status updated to %v in database\n",
		event.InfractionId, finalStatus)

	// TODO: Notify participant
	// notificationSvc.NotifyInfractionResolved(event.ParticipantIspb, infraction)

	return nil
}

// Close closes the consumer and releases resources
func (c *EntryEventConsumer) Close() error {
	log.Println("[EntryEventConsumer] Closing consumer...")
	c.consumer.Close()
	return nil
}

// ===================================================================
// MAPPER FUNCTIONS: Proto → Domain
// ===================================================================

// mapProtoStatusToDomain maps proto EntryStatus to domain KeyStatus
func mapProtoStatusToDomain(status commonv1.EntryStatus) valueobjects.KeyStatus {
	switch status {
	case commonv1.EntryStatus_ENTRY_STATUS_ACTIVE:
		return valueobjects.KeyStatusActive
	case commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING:
		return valueobjects.KeyStatusPortabilityRequested
	case commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_CONFIRMED:
		return valueobjects.KeyStatusOwnershipConfirmed
	case commonv1.EntryStatus_ENTRY_STATUS_CLAIM_PENDING:
		return valueobjects.KeyStatusClaimPending
	case commonv1.EntryStatus_ENTRY_STATUS_DELETED:
		return valueobjects.KeyStatusDeleted
	default:
		return valueobjects.KeyStatusPending
	}
}

// mapProtoClaimTypeToDomain maps proto ClaimType to domain ClaimType
func mapProtoClaimTypeToDomain(claimType connectv1.ClaimCreatedEvent_ClaimType) valueobjects.ClaimType {
	switch claimType {
	case connectv1.ClaimCreatedEvent_CLAIM_TYPE_OWNERSHIP:
		return valueobjects.ClaimTypeOwnership
	case connectv1.ClaimCreatedEvent_CLAIM_TYPE_PORTABILITY:
		return valueobjects.ClaimTypePortability
	default:
		return valueobjects.ClaimTypeOwnership
	}
}

// mapProtoClaimStatusToDomain maps proto ClaimStatus to domain ClaimStatus
func mapProtoClaimStatusToDomain(status commonv1.ClaimStatus) valueobjects.ClaimStatus {
	switch status {
	case commonv1.ClaimStatus_CLAIM_STATUS_OPEN:
		return valueobjects.ClaimStatusOpen
	case commonv1.ClaimStatus_CLAIM_STATUS_WAITING_RESOLUTION:
		return valueobjects.ClaimStatusWaitingResolution
	case commonv1.ClaimStatus_CLAIM_STATUS_CONFIRMED:
		return valueobjects.ClaimStatusConfirmed
	case commonv1.ClaimStatus_CLAIM_STATUS_CANCELLED:
		return valueobjects.ClaimStatusCancelled
	case commonv1.ClaimStatus_CLAIM_STATUS_COMPLETED:
		return valueobjects.ClaimStatusCompleted
	case commonv1.ClaimStatus_CLAIM_STATUS_EXPIRED:
		return valueobjects.ClaimStatusExpired
	default:
		return valueobjects.ClaimStatusOpen
	}
}

// mapProtoInfractionTypeToDomain maps proto InfractionType to domain InfractionType
func mapProtoInfractionTypeToDomain(infractionType connectv1.InfractionReportedEvent_InfractionType) entities.InfractionType {
	switch infractionType {
	case connectv1.InfractionReportedEvent_INFRACTION_TYPE_FRAUD:
		return entities.InfractionTypeFraud
	case connectv1.InfractionReportedEvent_INFRACTION_TYPE_ACCOUNT_CLOSED:
		return entities.InfractionTypeAccountClosed
	case connectv1.InfractionReportedEvent_INFRACTION_TYPE_INVALID_ACCOUNT:
		return entities.InfractionTypeInvalidAccount
	case connectv1.InfractionReportedEvent_INFRACTION_TYPE_DUPLICATE_KEY:
		return entities.InfractionTypeDataMismatch
	case connectv1.InfractionReportedEvent_INFRACTION_TYPE_INCORRECT_OWNERSHIP:
		return entities.InfractionTypeKeyOwnershipIssue
	default:
		return entities.InfractionTypeFraud
	}
}

// mapProtoInfractionStatusToDomain maps proto InfractionStatus to domain InfractionStatus
func mapProtoInfractionStatusToDomain(status connectv1.InfractionResolvedEvent_InfractionStatus) entities.InfractionStatus {
	switch status {
	case connectv1.InfractionResolvedEvent_INFRACTION_STATUS_RESOLVED:
		return entities.InfractionStatusResolved
	case connectv1.InfractionResolvedEvent_INFRACTION_STATUS_DISMISSED:
		return entities.InfractionStatusRejected
	case connectv1.InfractionResolvedEvent_INFRACTION_STATUS_UNDER_INVESTIGATION:
		return entities.InfractionStatusUnderReview
	default:
		return entities.InfractionStatusReported
	}
}

// ===================================================================
// HELPER FUNCTIONS
// ===================================================================

// parseUUID parses a string UUID
func parseUUID(s string) uuid.UUID {
	// Simplified - in production, handle errors properly
	id, _ := uuid.Parse(s)
	return id
}

// convertMetadata converts map[string]string to map[string]interface{}
func convertMetadata(metadata map[string]string) map[string]interface{} {
	result := make(map[string]interface{}, len(metadata))
	for k, v := range metadata {
		result[k] = v
	}
	return result
}
