package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// EventType representa o tipo de evento de auditoria
type EventType string

const (
	EventTypeEntryCreated     EventType = "ENTRY_CREATED"
	EventTypeEntryUpdated     EventType = "ENTRY_UPDATED"
	EventTypeEntryDeleted     EventType = "ENTRY_DELETED"
	EventTypeEntryActivated   EventType = "ENTRY_ACTIVATED"
	EventTypeEntryBlocked     EventType = "ENTRY_BLOCKED"
	EventTypeEntryUnblocked   EventType = "ENTRY_UNBLOCKED"
	EventTypeClaimCreated     EventType = "CLAIM_CREATED"
	EventTypeClaimConfirmed   EventType = "CLAIM_CONFIRMED"
	EventTypeClaimCancelled   EventType = "CLAIM_CANCELLED"
	EventTypeClaimCompleted   EventType = "CLAIM_COMPLETED"
	EventTypeClaimExpired     EventType = "CLAIM_EXPIRED"
	EventTypeInfractionReported EventType = "INFRACTION_REPORTED"
	EventTypeInfractionResolved EventType = "INFRACTION_RESOLVED"
	EventTypeSyncStarted      EventType = "SYNC_STARTED"
	EventTypeSyncCompleted    EventType = "SYNC_COMPLETED"
	EventTypeSyncFailed       EventType = "SYNC_FAILED"
)

// EntityType representa o tipo de entidade auditada
type EntityType string

const (
	EntityTypeEntry       EntityType = "ENTRY"
	EntityTypeAccount     EntityType = "ACCOUNT"
	EntityTypeClaim       EntityType = "CLAIM"
	EntityTypePortability EntityType = "PORTABILITY"
	EntityTypeInfraction  EntityType = "INFRACTION"
)

// AuditEvent representa um evento de auditoria no sistema
type AuditEvent struct {
	ID           uuid.UUID
	EventID      uuid.UUID
	EventType    EventType
	EntityType   EntityType
	EntityID     uuid.UUID
	OldValues    map[string]interface{}
	NewValues    map[string]interface{}
	Diff         map[string]interface{}
	UserID       *uuid.UUID
	IPAddress    string
	UserAgent    string
	OccurredAt   time.Time
	Metadata     map[string]interface{}
}

// NewAuditEvent cria um novo evento de auditoria
func NewAuditEvent(
	eventType EventType,
	entityType EntityType,
	entityID uuid.UUID,
	oldValues map[string]interface{},
	newValues map[string]interface{},
	userID *uuid.UUID,
) (*AuditEvent, error) {
	if err := validateEventType(eventType); err != nil {
		return nil, err
	}
	if err := validateEntityType(entityType); err != nil {
		return nil, err
	}
	if entityID == uuid.Nil {
		return nil, errors.New("entity ID cannot be nil")
	}

	diff := computeDiff(oldValues, newValues)

	return &AuditEvent{
		ID:         uuid.New(),
		EventID:    uuid.New(),
		EventType:  eventType,
		EntityType: entityType,
		EntityID:   entityID,
		OldValues:  oldValues,
		NewValues:  newValues,
		Diff:       diff,
		UserID:     userID,
		OccurredAt: time.Now(),
		Metadata:   make(map[string]interface{}),
	}, nil
}

// Validate valida o evento de auditoria
func (a *AuditEvent) Validate() error {
	if a.ID == uuid.Nil || a.EventID == uuid.Nil {
		return errors.New("IDs cannot be nil")
	}
	if err := validateEventType(a.EventType); err != nil {
		return err
	}
	if err := validateEntityType(a.EntityType); err != nil {
		return err
	}
	if a.EntityID == uuid.Nil {
		return errors.New("entity ID cannot be nil")
	}
	return nil
}

// SetRequestContext define o contexto da requisição
func (a *AuditEvent) SetRequestContext(ipAddress, userAgent string) {
	a.IPAddress = ipAddress
	a.UserAgent = userAgent
}

// AddMetadata adiciona metadados ao evento
func (a *AuditEvent) AddMetadata(key string, value interface{}) {
	a.Metadata[key] = value
}

func validateEventType(et EventType) error {
	validEventTypes := map[EventType]bool{
		EventTypeEntryCreated:       true,
		EventTypeEntryUpdated:       true,
		EventTypeEntryDeleted:       true,
		EventTypeEntryActivated:     true,
		EventTypeEntryBlocked:       true,
		EventTypeEntryUnblocked:     true,
		EventTypeClaimCreated:       true,
		EventTypeClaimConfirmed:     true,
		EventTypeClaimCancelled:     true,
		EventTypeClaimCompleted:     true,
		EventTypeClaimExpired:       true,
		EventTypeInfractionReported: true,
		EventTypeInfractionResolved: true,
		EventTypeSyncStarted:        true,
		EventTypeSyncCompleted:      true,
		EventTypeSyncFailed:         true,
	}
	if !validEventTypes[et] {
		return errors.New("invalid event type")
	}
	return nil
}

func validateEntityType(et EntityType) error {
	validEntityTypes := map[EntityType]bool{
		EntityTypeEntry:       true,
		EntityTypeAccount:     true,
		EntityTypeClaim:       true,
		EntityTypePortability: true,
		EntityTypeInfraction:  true,
	}
	if !validEntityTypes[et] {
		return errors.New("invalid entity type")
	}
	return nil
}

// computeDiff calcula a diferença entre valores antigos e novos
func computeDiff(oldValues, newValues map[string]interface{}) map[string]interface{} {
	diff := make(map[string]interface{})

	// Campos modificados
	for key, newVal := range newValues {
		oldVal, exists := oldValues[key]
		if !exists || oldVal != newVal {
			diff[key] = map[string]interface{}{
				"old": oldVal,
				"new": newVal,
			}
		}
	}

	// Campos removidos
	for key, oldVal := range oldValues {
		if _, exists := newValues[key]; !exists {
			diff[key] = map[string]interface{}{
				"old":     oldVal,
				"removed": true,
			}
		}
	}

	return diff
}
