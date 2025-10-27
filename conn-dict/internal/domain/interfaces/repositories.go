package interfaces

import (
	"context"

	"github.com/lbpay-lab/conn-dict/internal/domain/aggregates"
)

// ClaimRepository defines the interface for claim persistence
type ClaimRepository interface {
	// Save persists a claim aggregate
	Save(ctx context.Context, claim *aggregates.Claim) error

	// FindByID retrieves a claim by ID
	FindByID(ctx context.Context, id string) (*aggregates.Claim, error)

	// FindByKey retrieves a claim by key and key type
	FindByKey(ctx context.Context, key, keyType string) (*aggregates.Claim, error)

	// FindPendingClaims retrieves all pending claims
	FindPendingClaims(ctx context.Context) ([]*aggregates.Claim, error)

	// FindExpiredClaims retrieves all expired claims
	FindExpiredClaims(ctx context.Context) ([]*aggregates.Claim, error)

	// Delete removes a claim
	Delete(ctx context.Context, id string) error
}

// VsyncRepository defines the interface for vsync entry persistence
type VsyncRepository interface {
	// Save persists a vsync entry
	Save(ctx context.Context, entry *aggregates.VsyncEntry) error

	// FindByKey retrieves a vsync entry by key and key type
	FindByKey(ctx context.Context, key, keyType string) (*aggregates.VsyncEntry, error)

	// FindPendingEntries retrieves all pending vsync entries
	FindPendingEntries(ctx context.Context) ([]*aggregates.VsyncEntry, error)

	// FindFailedEntries retrieves all failed vsync entries that can be retried
	FindFailedEntries(ctx context.Context) ([]*aggregates.VsyncEntry, error)

	// Delete removes a vsync entry
	Delete(ctx context.Context, key, keyType string) error
}
