package aggregates

import (
	"time"

	"github.com/lbpay-lab/conn-dict/internal/domain/events"
)

// VsyncStatus represents the synchronization status
type VsyncStatus string

const (
	VsyncStatusPending    VsyncStatus = "PENDING"
	VsyncStatusInProgress VsyncStatus = "IN_PROGRESS"
	VsyncStatusCompleted  VsyncStatus = "COMPLETED"
	VsyncStatusFailed     VsyncStatus = "FAILED"
)

// VsyncEntry represents a DICT vsync entry aggregate root
type VsyncEntry struct {
	Key           string
	KeyType       string
	ISPB          string
	Branch        string
	Account       string
	AccountType   string
	OwnerName     string
	OwnerDocument string
	Status        VsyncStatus
	SyncedAt      time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Attempts      int
	LastError     string
	Version       int

	// Domain events
	events []events.DomainEvent
}

// NewVsyncEntry creates a new VsyncEntry aggregate
func NewVsyncEntry(
	key, keyType, ispb, branch, account, accountType, ownerName, ownerDocument string,
) *VsyncEntry {
	now := time.Now()
	entry := &VsyncEntry{
		Key:           key,
		KeyType:       keyType,
		ISPB:          ispb,
		Branch:        branch,
		Account:       account,
		AccountType:   accountType,
		OwnerName:     ownerName,
		OwnerDocument: ownerDocument,
		Status:        VsyncStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
		Attempts:      0,
		Version:       1,
		events:        make([]events.DomainEvent, 0),
	}

	// Add domain event
	entry.addEvent(events.NewVsyncEntryCreatedEvent(entry))

	return entry
}

// StartSync marks the entry as in progress
func (v *VsyncEntry) StartSync() {
	v.Status = VsyncStatusInProgress
	v.UpdatedAt = time.Now()
	v.Attempts++
	v.Version++

	// Add domain event
	v.addEvent(events.NewVsyncStartedEvent(v))
}

// CompleteSync marks the entry as completed
func (v *VsyncEntry) CompleteSync() {
	v.Status = VsyncStatusCompleted
	v.SyncedAt = time.Now()
	v.UpdatedAt = time.Now()
	v.Version++

	// Add domain event
	v.addEvent(events.NewVsyncCompletedEvent(v))
}

// FailSync marks the entry as failed
func (v *VsyncEntry) FailSync(errorMsg string) {
	v.Status = VsyncStatusFailed
	v.LastError = errorMsg
	v.UpdatedAt = time.Now()
	v.Version++

	// Add domain event
	v.addEvent(events.NewVsyncFailedEvent(v, errorMsg))
}

// CanRetry checks if the entry can be retried
func (v *VsyncEntry) CanRetry() bool {
	return v.Attempts < 3 && v.Status == VsyncStatusFailed
}

// addEvent adds a domain event to the aggregate
func (v *VsyncEntry) addEvent(event events.DomainEvent) {
	v.events = append(v.events, event)
}

// GetEvents returns all domain events and clears the event list
func (v *VsyncEntry) GetEvents() []events.DomainEvent {
	result := v.events
	v.events = []events.DomainEvent{}
	return result
}
