package entities

import (
	"time"
)

// Claim represents a DICT key claim
type Claim struct {
	ID               string
	Key              string
	KeyType          KeyType
	ClaimantISPB     string
	DonorISPB        string
	ClaimType        ClaimType
	Status           ClaimStatus
	CreatedAt        time.Time
	CompletedAt      *time.Time
	CancelledAt      *time.Time
	Reason           string
	ResolutionReason string
}

// ClaimType represents the type of claim
type ClaimType string

const (
	ClaimTypeOwnership ClaimType = "OWNERSHIP"
	ClaimTypePortability ClaimType = "PORTABILITY"
)

// ClaimStatus represents the status of a claim
type ClaimStatus string

const (
	ClaimStatusPending    ClaimStatus = "PENDING"
	ClaimStatusConfirmed  ClaimStatus = "CONFIRMED"
	ClaimStatusCancelled  ClaimStatus = "CANCELLED"
	ClaimStatusCompleted  ClaimStatus = "COMPLETED"
)

// Validate validates the claim
func (c *Claim) Validate() error {
	// TODO: Implement validation logic
	return nil
}

// IsPending returns true if the claim is pending
func (c *Claim) IsPending() bool {
	return c.Status == ClaimStatusPending
}

// CanBeCancelled returns true if the claim can be cancelled
func (c *Claim) CanBeCancelled() bool {
	return c.Status == ClaimStatusPending
}

// Complete completes the claim
func (c *Claim) Complete(reason string) {
	now := time.Now()
	c.Status = ClaimStatusCompleted
	c.CompletedAt = &now
	c.ResolutionReason = reason
}

// Cancel cancels the claim
func (c *Claim) Cancel(reason string) {
	now := time.Now()
	c.Status = ClaimStatusCancelled
	c.CancelledAt = &now
	c.ResolutionReason = reason
}
