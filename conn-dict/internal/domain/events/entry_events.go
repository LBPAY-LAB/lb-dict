package events

import (
	"time"

	"github.com/google/uuid"
)

// EntryCreatedEvent represents an event when a new DICT entry is created
type EntryCreatedEvent struct {
	EventHeader
	EntryID       string `json:"entry_id"`
	KeyValue      string `json:"key_value"`
	KeyType       string `json:"key_type"`
	ISPB          string `json:"ispb"`
	AccountNumber string `json:"account_number"`
}

// NewEntryCreatedEvent creates a new EntryCreatedEvent
func NewEntryCreatedEvent(entryID, keyValue, keyType, ispb, accountNumber string) *EntryCreatedEvent {
	return &EntryCreatedEvent{
		EventHeader: EventHeader{
			ID:        uuid.New().String(),
			Type:      "EntryCreated",
			Timestamp: time.Now(),
			Version:   "1.0",
		},
		EntryID:       entryID,
		KeyValue:      keyValue,
		KeyType:       keyType,
		ISPB:          ispb,
		AccountNumber: accountNumber,
	}
}

// EventID returns the event ID
func (e *EntryCreatedEvent) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e *EntryCreatedEvent) EventType() string {
	return e.Type
}

// AggregateID returns the aggregate ID (entry ID)
func (e *EntryCreatedEvent) AggregateID() string {
	return e.EntryID
}

// OccurredAt returns when the event occurred
func (e *EntryCreatedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// Payload returns the event payload
func (e *EntryCreatedEvent) Payload() interface{} {
	return e
}

// EntryUpdatedEvent represents an event when a DICT entry is updated
type EntryUpdatedEvent struct {
	EventHeader
	EntryID       string `json:"entry_id"`
	ISPB          string `json:"ispb"`
	AccountNumber string `json:"account_number"`
}

// NewEntryUpdatedEvent creates a new EntryUpdatedEvent
func NewEntryUpdatedEvent(entryID, ispb, accountNumber string) *EntryUpdatedEvent {
	return &EntryUpdatedEvent{
		EventHeader: EventHeader{
			ID:        uuid.New().String(),
			Type:      "EntryUpdated",
			Timestamp: time.Now(),
			Version:   "1.0",
		},
		EntryID:       entryID,
		ISPB:          ispb,
		AccountNumber: accountNumber,
	}
}

// EventID returns the event ID
func (e *EntryUpdatedEvent) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e *EntryUpdatedEvent) EventType() string {
	return e.Type
}

// AggregateID returns the aggregate ID (entry ID)
func (e *EntryUpdatedEvent) AggregateID() string {
	return e.EntryID
}

// OccurredAt returns when the event occurred
func (e *EntryUpdatedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// Payload returns the event payload
func (e *EntryUpdatedEvent) Payload() interface{} {
	return e
}

// EntryDeletedEvent represents an event when a DICT entry is deleted
type EntryDeletedEvent struct {
	EventHeader
	EntryID string `json:"entry_id"`
}

// NewEntryDeletedEvent creates a new EntryDeletedEvent
func NewEntryDeletedEvent(entryID string) *EntryDeletedEvent {
	return &EntryDeletedEvent{
		EventHeader: EventHeader{
			ID:        uuid.New().String(),
			Type:      "EntryDeleted",
			Timestamp: time.Now(),
			Version:   "1.0",
		},
		EntryID: entryID,
	}
}

// EventID returns the event ID
func (e *EntryDeletedEvent) EventID() string {
	return e.ID
}

// EventType returns the event type
func (e *EntryDeletedEvent) EventType() string {
	return e.Type
}

// AggregateID returns the aggregate ID (entry ID)
func (e *EntryDeletedEvent) AggregateID() string {
	return e.EntryID
}

// OccurredAt returns when the event occurred
func (e *EntryDeletedEvent) OccurredAt() time.Time {
	return e.Timestamp
}

// Payload returns the event payload
func (e *EntryDeletedEvent) Payload() interface{} {
	return e
}

// EventHeader represents common event metadata
type EventHeader struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}
