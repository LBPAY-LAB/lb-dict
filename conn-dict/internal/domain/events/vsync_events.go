package events

import (
	"time"

	"github.com/google/uuid"
)

// VsyncEntryCreatedEvent represents a vsync entry creation event
type VsyncEntryCreatedEvent struct {
	BaseEvent
	Key           string
	KeyType       string
	ISPB          string
	Branch        string
	Account       string
	AccountType   string
	OwnerName     string
	OwnerDocument string
}

// NewVsyncEntryCreatedEvent creates a new VsyncEntryCreatedEvent
func NewVsyncEntryCreatedEvent(entry interface{}) *VsyncEntryCreatedEvent {
	return &VsyncEntryCreatedEvent{
		BaseEvent: BaseEvent{
			ID:         uuid.New().String(),
			Type:       "vsync.entry.created",
			OccurredOn: time.Now(),
		},
		// TODO: Map entry fields properly
	}
}

// VsyncStartedEvent represents a vsync start event
type VsyncStartedEvent struct {
	BaseEvent
	Key      string
	KeyType  string
	Attempts int
}

// NewVsyncStartedEvent creates a new VsyncStartedEvent
func NewVsyncStartedEvent(entry interface{}) *VsyncStartedEvent {
	return &VsyncStartedEvent{
		BaseEvent: BaseEvent{
			ID:         uuid.New().String(),
			Type:       "vsync.started",
			OccurredOn: time.Now(),
		},
	}
}

// VsyncCompletedEvent represents a vsync completion event
type VsyncCompletedEvent struct {
	BaseEvent
	Key        string
	KeyType    string
	SyncedAt   time.Time
}

// NewVsyncCompletedEvent creates a new VsyncCompletedEvent
func NewVsyncCompletedEvent(entry interface{}) *VsyncCompletedEvent {
	return &VsyncCompletedEvent{
		BaseEvent: BaseEvent{
			ID:         uuid.New().String(),
			Type:       "vsync.completed",
			OccurredOn: time.Now(),
		},
	}
}

// VsyncFailedEvent represents a vsync failure event
type VsyncFailedEvent struct {
	BaseEvent
	Key       string
	KeyType   string
	Error     string
	Attempts  int
}

// NewVsyncFailedEvent creates a new VsyncFailedEvent
func NewVsyncFailedEvent(entry interface{}, errorMsg string) *VsyncFailedEvent {
	return &VsyncFailedEvent{
		BaseEvent: BaseEvent{
			ID:         uuid.New().String(),
			Type:       "vsync.failed",
			OccurredOn: time.Now(),
		},
		Error: errorMsg,
	}
}
