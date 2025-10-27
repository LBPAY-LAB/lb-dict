package events

import (
	"time"

	"github.com/google/uuid"
)

// DomainEvent represents a domain event interface
type DomainEvent interface {
	EventID() string
	EventType() string
	AggregateID() string
	OccurredAt() time.Time
	Payload() interface{}
}

// BaseEvent provides common event fields
type BaseEvent struct {
	ID          string
	Type        string
	Aggregate   string
	OccurredOn  time.Time
	EventData   interface{}
}

func (e *BaseEvent) EventID() string       { return e.ID }
func (e *BaseEvent) EventType() string     { return e.Type }
func (e *BaseEvent) AggregateID() string   { return e.Aggregate }
func (e *BaseEvent) OccurredAt() time.Time { return e.OccurredOn }
func (e *BaseEvent) Payload() interface{}  { return e.EventData }

// ClaimCreatedEvent represents a claim creation event
type ClaimCreatedEvent struct {
	BaseEvent
	ClaimID       string
	Key           string
	KeyType       string
	ISPB          string
	Branch        string
	Account       string
	AccountType   string
	OwnerName     string
	OwnerDocument string
}

// NewClaimCreatedEvent creates a new ClaimCreatedEvent
func NewClaimCreatedEvent(claim interface{}) *ClaimCreatedEvent {
	return &ClaimCreatedEvent{
		BaseEvent: BaseEvent{
			ID:         uuid.New().String(),
			Type:       "claim.created",
			OccurredOn: time.Now(),
		},
		// TODO: Map claim fields properly
	}
}

// ClaimConfirmedEvent represents a claim confirmation event
type ClaimConfirmedEvent struct {
	BaseEvent
	ClaimID      string
	ConfirmedBy  string
	ConfirmedAt  time.Time
}

// NewClaimConfirmedEvent creates a new ClaimConfirmedEvent
func NewClaimConfirmedEvent(claim interface{}, confirmedBy string) *ClaimConfirmedEvent {
	return &ClaimConfirmedEvent{
		BaseEvent: BaseEvent{
			ID:         uuid.New().String(),
			Type:       "claim.confirmed",
			OccurredOn: time.Now(),
		},
		ConfirmedBy: confirmedBy,
	}
}

// ClaimCancelledEvent represents a claim cancellation event
type ClaimCancelledEvent struct {
	BaseEvent
	ClaimID     string
	Reason      string
	CancelledBy string
}

// NewClaimCancelledEvent creates a new ClaimCancelledEvent
func NewClaimCancelledEvent(claim interface{}, reason, cancelledBy string) *ClaimCancelledEvent {
	return &ClaimCancelledEvent{
		BaseEvent: BaseEvent{
			ID:         uuid.New().String(),
			Type:       "claim.cancelled",
			OccurredOn: time.Now(),
		},
		Reason:      reason,
		CancelledBy: cancelledBy,
	}
}

// ClaimExpiredEvent represents a claim expiration event
type ClaimExpiredEvent struct {
	BaseEvent
	ClaimID   string
	ExpiredAt time.Time
}

// NewClaimExpiredEvent creates a new ClaimExpiredEvent
func NewClaimExpiredEvent(claim interface{}) *ClaimExpiredEvent {
	return &ClaimExpiredEvent{
		BaseEvent: BaseEvent{
			ID:         uuid.New().String(),
			Type:       "claim.expired",
			OccurredOn: time.Now(),
		},
	}
}
