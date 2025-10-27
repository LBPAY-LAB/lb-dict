package entities

import (
	"time"

	"github.com/google/uuid"
)

// SyncType represents the type of VSYNC synchronization
type SyncType string

const (
	SyncTypeFull        SyncType = "FULL"        // Complete sync of all entries
	SyncTypeIncremental SyncType = "INCREMENTAL" // Only entries changed since last sync
)

// SyncStatus represents the outcome of a VSYNC execution
type SyncStatus string

const (
	SyncStatusCompleted SyncStatus = "COMPLETED" // Sync completed successfully
	SyncStatusPartial   SyncStatus = "PARTIAL"   // Sync partially completed (some errors)
	SyncStatusFailed    SyncStatus = "FAILED"    // Sync failed
)

// SyncReport represents the result of a VSYNC workflow execution
type SyncReport struct {
	ID   uuid.UUID `json:"id"`
	SyncID string  `json:"sync_id"` // Temporal workflow execution ID

	// Sync details
	SyncType      SyncType   `json:"sync_type"`
	SyncTimestamp time.Time  `json:"sync_timestamp"`

	// Participant
	ParticipantISPB string `json:"participant_ispb"`

	// Statistics
	EntriesFetched  int `json:"entries_fetched"`  // Total fetched from Bacen
	EntriesCompared int `json:"entries_compared"` // Total compared
	EntriesSynced   int `json:"entries_synced"`   // Total synced
	EntriesCreated  int `json:"entries_created"`  // Created locally
	EntriesUpdated  int `json:"entries_updated"`  // Updated locally
	EntriesDeleted  int `json:"entries_deleted"`  // Deleted locally

	// Discrepancy tracking
	DiscrepanciesFound        int `json:"discrepancies_found"`
	DiscrepanciesMissingLocal int `json:"discrepancies_missing_local"`  // Missing in local DB
	DiscrepanciesOutdatedLocal int `json:"discrepancies_outdated_local"` // Outdated in local DB
	DiscrepanciesMissingBacen int `json:"discrepancies_missing_bacen"`  // Missing in Bacen (critical)

	// Execution
	Status     SyncStatus `json:"status"`
	DurationMS int        `json:"duration_ms"` // Duration in milliseconds

	// Error tracking
	ErrorMessage *string `json:"error_message,omitempty"`
	ErrorCode    *string `json:"error_code,omitempty"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewSyncReport creates a new SyncReport
func NewSyncReport(
	syncID string,
	syncType SyncType,
	participantISPB string,
) *SyncReport {
	now := time.Now()
	return &SyncReport{
		ID:              uuid.New(),
		SyncID:          syncID,
		SyncType:        syncType,
		SyncTimestamp:   now,
		ParticipantISPB: participantISPB,
		Status:          SyncStatusCompleted, // Default to completed
		Metadata:        make(map[string]interface{}),
		CreatedAt:       now,
		UpdatedAt:       now,
	}
}

// SetError marks the sync report as failed with error details
func (r *SyncReport) SetError(errorMessage, errorCode string) {
	r.Status = SyncStatusFailed
	r.ErrorMessage = &errorMessage
	r.ErrorCode = &errorCode
}

// SetPartial marks the sync report as partially completed
func (r *SyncReport) SetPartial(errorMessage string) {
	r.Status = SyncStatusPartial
	r.ErrorMessage = &errorMessage
}

// SetDuration sets the sync duration in milliseconds
func (r *SyncReport) SetDuration(duration time.Duration) {
	r.DurationMS = int(duration.Milliseconds())
}

// UpdateStatistics updates the sync statistics from discrepancy analysis
func (r *SyncReport) UpdateStatistics(
	entriesFetched int,
	discrepancies []struct {
		Type string
	},
	created, updated, deleted int,
) {
	r.EntriesFetched = entriesFetched
	r.EntriesCompared = entriesFetched
	r.DiscrepanciesFound = len(discrepancies)
	r.EntriesCreated = created
	r.EntriesUpdated = updated
	r.EntriesDeleted = deleted
	r.EntriesSynced = created + updated + deleted

	// Count discrepancies by type
	for _, d := range discrepancies {
		switch d.Type {
		case "MISSING_LOCAL":
			r.DiscrepanciesMissingLocal++
		case "OUTDATED_LOCAL":
			r.DiscrepanciesOutdatedLocal++
		case "MISSING_BACEN":
			r.DiscrepanciesMissingBacen++
		}
	}
}

// AddMetadata adds a metadata key-value pair
func (r *SyncReport) AddMetadata(key string, value interface{}) {
	if r.Metadata == nil {
		r.Metadata = make(map[string]interface{})
	}
	r.Metadata[key] = value
}