package entities

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// InfractionType represents the type of infraction
type InfractionType string

const (
	InfractionTypeFraud            InfractionType = "FRAUD"
	InfractionTypeAccountClosed    InfractionType = "ACCOUNT_CLOSED"
	InfractionTypeIncorrectData    InfractionType = "INCORRECT_DATA"
	InfractionTypeUnauthorizedUse  InfractionType = "UNAUTHORIZED_USE"
	InfractionTypeDuplicateKey     InfractionType = "DUPLICATE_KEY"
	InfractionTypeOther            InfractionType = "OTHER"
)

// InfractionStatus represents the status of an infraction
type InfractionStatus string

const (
	InfractionStatusOpen               InfractionStatus = "OPEN"
	InfractionStatusUnderInvestigation InfractionStatus = "UNDER_INVESTIGATION"
	InfractionStatusResolved           InfractionStatus = "RESOLVED"
	InfractionStatusDismissed          InfractionStatus = "DISMISSED"
	InfractionStatusEscalatedToBacen   InfractionStatus = "ESCALATED_TO_BACEN"
)

// Infraction represents a fraud report or infraction in the DICT system
type Infraction struct {
	ID           uuid.UUID
	InfractionID string // External infraction ID

	// Related entities
	EntryID *string // Optional - related entry
	ClaimID *string // Optional - related claim
	Key     string  // PIX key involved

	// Infraction details
	Type        InfractionType
	Description string
	EvidenceURLs []string // Array of evidence URLs

	// Reporter information
	ReporterParticipant string  // ISPB that reported (8 digits)
	ReportedParticipant *string // ISPB being reported (optional)

	// Status and resolution
	Status          InfractionStatus
	ResolutionNotes *string

	// Timestamps
	ReportedAt     time.Time
	InvestigatedAt *time.Time
	ResolvedAt     *time.Time

	// Audit
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewInfraction creates a new Infraction with validation
func NewInfraction(
	infractionID string,
	key string,
	infractionType InfractionType,
	description string,
	reporterISPB string,
) (*Infraction, error) {
	// Validate reporter ISPB
	if !isValidISPB(reporterISPB) {
		return nil, fmt.Errorf("invalid reporter ISPB: must be 8 digits, got %s", reporterISPB)
	}

	// Validate description
	if description == "" {
		return nil, errors.New("description is required")
	}

	if key == "" {
		return nil, errors.New("key is required")
	}

	now := time.Now()

	infraction := &Infraction{
		ID:                  uuid.New(),
		InfractionID:        infractionID,
		Key:                 key,
		Type:                infractionType,
		Description:         description,
		ReporterParticipant: reporterISPB,
		Status:              InfractionStatusOpen,
		ReportedAt:          now,
		CreatedAt:           now,
		UpdatedAt:           now,
		EvidenceURLs:        []string{},
	}

	return infraction, nil
}

// Investigate marks the infraction as under investigation
func (i *Infraction) Investigate() error {
	if i.Status != InfractionStatusOpen {
		return fmt.Errorf("can only investigate open infractions, current status: %s", i.Status)
	}

	now := time.Now()
	i.Status = InfractionStatusUnderInvestigation
	i.InvestigatedAt = &now
	i.UpdatedAt = now

	return nil
}

// Resolve marks the infraction as resolved with resolution notes
func (i *Infraction) Resolve(notes string) error {
	if i.Status != InfractionStatusUnderInvestigation && i.Status != InfractionStatusOpen {
		return fmt.Errorf("can only resolve open or under investigation infractions, current status: %s", i.Status)
	}

	if notes == "" {
		return errors.New("resolution notes are required")
	}

	now := time.Now()
	i.Status = InfractionStatusResolved
	i.ResolvedAt = &now
	i.ResolutionNotes = &notes
	i.UpdatedAt = now

	return nil
}

// Dismiss dismisses the infraction with notes
func (i *Infraction) Dismiss(notes string) error {
	if i.Status == InfractionStatusResolved || i.Status == InfractionStatusDismissed {
		return fmt.Errorf("cannot dismiss resolved or already dismissed infraction, current status: %s", i.Status)
	}

	if notes == "" {
		return errors.New("dismissal notes are required")
	}

	now := time.Now()
	i.Status = InfractionStatusDismissed
	i.ResolvedAt = &now
	i.ResolutionNotes = &notes
	i.UpdatedAt = now

	return nil
}

// EscalateToBacen escalates the infraction to Bacen
func (i *Infraction) EscalateToBacen(notes string) error {
	if i.Status == InfractionStatusResolved || i.Status == InfractionStatusDismissed {
		return fmt.Errorf("cannot escalate resolved or dismissed infraction, current status: %s", i.Status)
	}

	if i.Status == InfractionStatusEscalatedToBacen {
		return errors.New("infraction already escalated to Bacen")
	}

	now := time.Now()
	i.Status = InfractionStatusEscalatedToBacen
	i.ResolutionNotes = &notes
	i.UpdatedAt = now

	return nil
}

// AddEvidence adds an evidence URL to the infraction
func (i *Infraction) AddEvidence(url string) error {
	if url == "" {
		return errors.New("evidence URL cannot be empty")
	}

	// Check for duplicates
	for _, existingURL := range i.EvidenceURLs {
		if existingURL == url {
			return errors.New("evidence URL already exists")
		}
	}

	i.EvidenceURLs = append(i.EvidenceURLs, url)
	i.UpdatedAt = time.Now()

	return nil
}

// IsOpen checks if infraction is open or under investigation
func (i *Infraction) IsOpen() bool {
	return i.Status == InfractionStatusOpen || i.Status == InfractionStatusUnderInvestigation
}

// IsClosed checks if infraction is resolved or dismissed
func (i *Infraction) IsClosed() bool {
	return i.Status == InfractionStatusResolved || i.Status == InfractionStatusDismissed
}

// IsEscalated checks if infraction is escalated to Bacen
func (i *Infraction) IsEscalated() bool {
	return i.Status == InfractionStatusEscalatedToBacen
}

// ValidateStatusTransition validates if a status transition is allowed
func (i *Infraction) ValidateStatusTransition(newStatus InfractionStatus) error {
	validTransitions := map[InfractionStatus][]InfractionStatus{
		InfractionStatusOpen: {
			InfractionStatusUnderInvestigation,
			InfractionStatusDismissed,
			InfractionStatusEscalatedToBacen,
		},
		InfractionStatusUnderInvestigation: {
			InfractionStatusResolved,
			InfractionStatusDismissed,
			InfractionStatusEscalatedToBacen,
		},
		InfractionStatusResolved: {
			// Terminal state - no transitions
		},
		InfractionStatusDismissed: {
			// Terminal state - no transitions
		},
		InfractionStatusEscalatedToBacen: {
			InfractionStatusResolved,
			InfractionStatusDismissed,
		},
	}

	allowedStatuses, exists := validTransitions[i.Status]
	if !exists {
		return fmt.Errorf("unknown current status: %s", i.Status)
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return nil // Transition is valid
		}
	}

	return fmt.Errorf(
		"invalid status transition from %s to %s",
		i.Status,
		newStatus,
	)
}
