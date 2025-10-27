package aggregates

import (
	"fmt"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/domain/events"
)

// ClaimStatus represents the status of a DICT claim
type ClaimStatus string

const (
	ClaimStatusPending   ClaimStatus = "PENDING"
	ClaimStatusConfirmed ClaimStatus = "CONFIRMED"
	ClaimStatusCancelled ClaimStatus = "CANCELLED"
	ClaimStatusExpired   ClaimStatus = "EXPIRED"
)

// Claim represents a DICT claim aggregate root
type Claim struct {
	ID            string
	Key           string
	KeyType       string
	ISPB          string
	Branch        string
	Account       string
	AccountType   string
	OwnerName     string
	OwnerDocument string
	Status        ClaimStatus
	CreatedAt     time.Time
	UpdatedAt     time.Time
	ConfirmedAt   *time.Time
	CancelledAt   *time.Time
	ExpiresAt     time.Time
	Version       int

	// Domain events
	events []events.DomainEvent
}

// NewClaim creates a new Claim aggregate
func NewClaim(
	id, key, keyType, ispb, branch, account, accountType, ownerName, ownerDocument string,
) (*Claim, error) {
	// Validation
	if key == "" || keyType == "" || ispb == "" {
		return nil, fmt.Errorf("key, keyType, and ispb are required")
	}

	now := time.Now()
	claim := &Claim{
		ID:            id,
		Key:           key,
		KeyType:       keyType,
		ISPB:          ispb,
		Branch:        branch,
		Account:       account,
		AccountType:   accountType,
		OwnerName:     ownerName,
		OwnerDocument: ownerDocument,
		Status:        ClaimStatusPending,
		CreatedAt:     now,
		UpdatedAt:     now,
		ExpiresAt:     now.Add(7 * 24 * time.Hour), // 7 days expiration
		Version:       1,
		events:        make([]events.DomainEvent, 0),
	}

	// Add domain event
	claim.addEvent(events.NewClaimCreatedEvent(claim))

	return claim, nil
}

// Confirm confirms the claim
func (c *Claim) Confirm(confirmedBy string, confirmationDate time.Time) error {
	if c.Status != ClaimStatusPending {
		return fmt.Errorf("claim can only be confirmed when pending, current status: %s", c.Status)
	}

	c.Status = ClaimStatusConfirmed
	c.ConfirmedAt = &confirmationDate
	c.UpdatedAt = time.Now()
	c.Version++

	// Add domain event
	c.addEvent(events.NewClaimConfirmedEvent(c, confirmedBy))

	return nil
}

// Cancel cancels the claim
func (c *Claim) Cancel(reason, cancelledBy string) error {
	if c.Status == ClaimStatusConfirmed {
		return fmt.Errorf("confirmed claims cannot be cancelled")
	}

	now := time.Now()
	c.Status = ClaimStatusCancelled
	c.CancelledAt = &now
	c.UpdatedAt = now
	c.Version++

	// Add domain event
	c.addEvent(events.NewClaimCancelledEvent(c, reason, cancelledBy))

	return nil
}

// MarkAsExpired marks the claim as expired
func (c *Claim) MarkAsExpired() error {
	if c.Status != ClaimStatusPending {
		return fmt.Errorf("only pending claims can be expired")
	}

	c.Status = ClaimStatusExpired
	c.UpdatedAt = time.Now()
	c.Version++

	// Add domain event
	c.addEvent(events.NewClaimExpiredEvent(c))

	return nil
}

// IsExpired checks if the claim has expired
func (c *Claim) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// addEvent adds a domain event to the aggregate
func (c *Claim) addEvent(event events.DomainEvent) {
	c.events = append(c.events, event)
}

// GetEvents returns all domain events and clears the event list
func (c *Claim) GetEvents() []events.DomainEvent {
	result := c.events
	c.events = []events.DomainEvent{}
	return result
}
