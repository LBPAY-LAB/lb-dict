package events

import (
	"time"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
)

// EventType represents the type of DICT event
type EventType string

const (
	EventTypeEntryCreated EventType = "entry.created"
	EventTypeEntryUpdated EventType = "entry.updated"
	EventTypeEntryDeleted EventType = "entry.deleted"
	EventTypeError        EventType = "error"
)

// DictEvent represents a base DICT event
type DictEvent struct {
	EventType EventType `json:"event_type"`
	Source    string    `json:"source"`
	Timestamp string    `json:"timestamp"`
	TraceID   string    `json:"trace_id"`
}

// EntryCreatedEvent represents an entry created event
type EntryCreatedEvent struct {
	DictEvent
	Entry *entities.DictEntry `json:"entry"`
}

// EntryUpdatedEvent represents an entry updated event
type EntryUpdatedEvent struct {
	DictEvent
	Entry    *entities.DictEntry `json:"entry"`
	OldEntry *entities.DictEntry `json:"old_entry,omitempty"`
}

// EntryDeletedEvent represents an entry deleted event
type EntryDeletedEvent struct {
	DictEvent
	KeyID       string `json:"key_id"`
	KeyType     string `json:"key_type"`
	Participant string `json:"participant"`
}

// ErrorEvent represents an error event
type ErrorEvent struct {
	DictEvent
	ErrorCode    string `json:"error_code"`
	ErrorMessage string `json:"error_message"`
	Context      string `json:"context,omitempty"`
}

// NewEntryCreatedEvent creates a new entry created event
func NewEntryCreatedEvent(entry *entities.DictEntry, traceID string) *EntryCreatedEvent {
	return &EntryCreatedEvent{
		DictEvent: DictEvent{
			EventType: EventTypeEntryCreated,
			Source:    "rsfn-bridge",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			TraceID:   traceID,
		},
		Entry: entry,
	}
}

// NewEntryUpdatedEvent creates a new entry updated event
func NewEntryUpdatedEvent(entry *entities.DictEntry, oldEntry *entities.DictEntry, traceID string) *EntryUpdatedEvent {
	return &EntryUpdatedEvent{
		DictEvent: DictEvent{
			EventType: EventTypeEntryUpdated,
			Source:    "rsfn-bridge",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			TraceID:   traceID,
		},
		Entry:    entry,
		OldEntry: oldEntry,
	}
}

// NewEntryDeletedEvent creates a new entry deleted event
func NewEntryDeletedEvent(keyID, keyType, participant, traceID string) *EntryDeletedEvent {
	return &EntryDeletedEvent{
		DictEvent: DictEvent{
			EventType: EventTypeEntryDeleted,
			Source:    "rsfn-bridge",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			TraceID:   traceID,
		},
		KeyID:       keyID,
		KeyType:     keyType,
		Participant: participant,
	}
}

// NewErrorEvent creates a new error event
func NewErrorEvent(errorCode, errorMessage, context, traceID string) *ErrorEvent {
	return &ErrorEvent{
		DictEvent: DictEvent{
			EventType: EventTypeError,
			Source:    "rsfn-bridge",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			TraceID:   traceID,
		},
		ErrorCode:    errorCode,
		ErrorMessage: errorMessage,
		Context:      context,
	}
}
