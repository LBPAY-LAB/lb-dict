package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// ClaimType represents the type of claim
type ClaimType string

const (
	ClaimTypePortability ClaimType = "PORTABILITY"
	ClaimTypeOwnership   ClaimType = "OWNERSHIP"
)

// ClaimStatus represents the current status of a claim
type ClaimStatus string

const (
	ClaimStatusOpen              ClaimStatus = "OPEN"
	ClaimStatusWaitingResolution ClaimStatus = "WAITING_RESOLUTION"
	ClaimStatusConfirmed         ClaimStatus = "CONFIRMED"
	ClaimStatusCancelled         ClaimStatus = "CANCELLED"
	ClaimStatusCompleted         ClaimStatus = "COMPLETED"
	ClaimStatusExpired           ClaimStatus = "EXPIRED"
)

// Claim represents a DICT portability or ownership claim (aggregate root)
type Claim struct {
	// Identity
	ID      uuid.UUID
	ClaimID string

	// Type and Status
	Type   ClaimType
	Status ClaimStatus

	// Key information
	Key     string
	KeyType string

	// Participants
	DonorParticipant   string // ISPB (8 digits)
	ClaimerParticipant string // ISPB (8 digits)

	// Account information (claimer)
	ClaimerAccountBranch string
	ClaimerAccountNumber string
	ClaimerAccountType   string

	// Timestamps
	CompletionPeriodEnd time.Time // 7 days for donor response
	ClaimExpiryDate     time.Time // 30 days from creation
	ConfirmedAt         *time.Time
	CompletedAt         *time.Time
	CancelledAt         *time.Time
	ExpiredAt           *time.Time

	// Metadata
	CancellationReason string
	Notes              string

	// Audit
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewClaim creates a new claim with validation
func NewClaim(
	claimID string,
	claimType ClaimType,
	key string,
	keyType string,
	donorISPB string,
	claimerISPB string,
) (*Claim, error) {
	// Validate inputs
	if claimID == "" {
		return nil, errors.New("claim_id cannot be empty")
	}
	if key == "" {
		return nil, errors.New("key cannot be empty")
	}
	if len(donorISPB) != 8 {
		return nil, errors.New("donor ISPB must be 8 digits")
	}
	if len(claimerISPB) != 8 {
		return nil, errors.New("claimer ISPB must be 8 digits")
	}
	if donorISPB == claimerISPB {
		return nil, errors.New("donor and claimer must be different")
	}

	now := time.Now()

	claim := &Claim{
		ID:                   uuid.New(),
		ClaimID:              claimID,
		Type:                 claimType,
		Status:               ClaimStatusOpen,
		Key:                  key,
		KeyType:              keyType,
		DonorParticipant:     donorISPB,
		ClaimerParticipant:   claimerISPB,
		CompletionPeriodEnd:  now.Add(7 * 24 * time.Hour),  // 7 days
		ClaimExpiryDate:      now.Add(30 * 24 * time.Hour), // 30 days
		CreatedAt:            now,
		UpdatedAt:            now,
	}

	return claim, nil
}

// Confirm marks the claim as confirmed by the donor
func (c *Claim) Confirm() error {
	if c.Status != ClaimStatusOpen && c.Status != ClaimStatusWaitingResolution {
		return errors.New("claim can only be confirmed from OPEN or WAITING_RESOLUTION status")
	}

	now := time.Now()
	c.Status = ClaimStatusConfirmed
	c.ConfirmedAt = &now
	c.UpdatedAt = now

	return nil
}

// Complete marks the claim as completed (portability done)
func (c *Claim) Complete() error {
	if c.Status != ClaimStatusConfirmed {
		return errors.New("claim must be confirmed before completion")
	}

	now := time.Now()
	c.Status = ClaimStatusCompleted
	c.CompletedAt = &now
	c.UpdatedAt = now

	return nil
}

// Cancel cancels the claim with a reason
func (c *Claim) Cancel(reason string) error {
	if c.Status == ClaimStatusCompleted || c.Status == ClaimStatusExpired {
		return errors.New("cannot cancel a completed or expired claim")
	}

	now := time.Now()
	c.Status = ClaimStatusCancelled
	c.CancellationReason = reason
	c.CancelledAt = &now
	c.UpdatedAt = now

	return nil
}

// Expire marks the claim as expired (after 30 days)
func (c *Claim) Expire() error {
	if c.Status == ClaimStatusCompleted {
		return errors.New("cannot expire a completed claim")
	}

	now := time.Now()
	c.Status = ClaimStatusExpired
	c.ExpiredAt = &now
	c.UpdatedAt = now

	return nil
}

// MoveToWaitingResolution moves claim to waiting resolution status
func (c *Claim) MoveToWaitingResolution() error {
	if c.Status != ClaimStatusOpen {
		return errors.New("claim must be in OPEN status")
	}

	c.Status = ClaimStatusWaitingResolution
	c.UpdatedAt = time.Now()

	return nil
}

// IsExpired checks if the claim has passed its expiry date
func (c *Claim) IsExpired() bool {
	return time.Now().After(c.ClaimExpiryDate)
}

// IsCompletionPeriodExpired checks if the 7-day period has passed
func (c *Claim) IsCompletionPeriodExpired() bool {
	return time.Now().After(c.CompletionPeriodEnd)
}

// IsActive checks if the claim is still active (not terminal status)
func (c *Claim) IsActive() bool {
	return c.Status != ClaimStatusCompleted &&
		c.Status != ClaimStatusCancelled &&
		c.Status != ClaimStatusExpired
}

// CanBeCancelled checks if the claim can be cancelled
func (c *Claim) CanBeCancelled() bool {
	return c.Status != ClaimStatusCompleted && c.Status != ClaimStatusExpired
}

// ValidateStatusTransition validates if a status transition is allowed
func (c *Claim) ValidateStatusTransition(newStatus ClaimStatus) error {
	validTransitions := map[ClaimStatus][]ClaimStatus{
		ClaimStatusOpen: {
			ClaimStatusWaitingResolution,
			ClaimStatusCancelled,
		},
		ClaimStatusWaitingResolution: {
			ClaimStatusConfirmed,
			ClaimStatusCancelled,
			ClaimStatusExpired,
		},
		ClaimStatusConfirmed: {
			ClaimStatusCompleted,
			ClaimStatusCancelled,
		},
		ClaimStatusCancelled: {}, // Terminal
		ClaimStatusCompleted: {}, // Terminal
		ClaimStatusExpired:   {}, // Terminal
	}

	allowedStatuses, exists := validTransitions[c.Status]
	if !exists {
		return errors.New("invalid current status")
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return nil
		}
	}

	return errors.New("invalid status transition")
}