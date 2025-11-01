package ports

import (
	"context"
	"time"

	"github.com/lb-conn/connector-dict/domain/ratelimit"
)

// PolicyRepository defines the interface for rate limit policy storage operations
type PolicyRepository interface {
	// GetAll retrieves all rate limit policies
	GetAll(ctx context.Context) ([]*ratelimit.Policy, error)

	// GetByID retrieves a specific policy by endpoint ID
	GetByID(ctx context.Context, endpointID string) (*ratelimit.Policy, error)

	// GetByCategory retrieves policies for a specific PSP category
	GetByCategory(ctx context.Context, category string) ([]*ratelimit.Policy, error)

	// Upsert inserts or updates a policy
	Upsert(ctx context.Context, policy *ratelimit.Policy) error

	// UpsertBatch inserts or updates multiple policies in a single transaction
	UpsertBatch(ctx context.Context, policies []*ratelimit.Policy) error
}

// StateRepository defines the interface for rate limit state storage operations
// States are time-series data with 5-minute snapshots
type StateRepository interface {
	// Save inserts a new state snapshot
	Save(ctx context.Context, state *ratelimit.PolicyState) error

	// SaveBatch inserts multiple state snapshots in a single transaction
	SaveBatch(ctx context.Context, states []*ratelimit.PolicyState) error

	// GetLatest retrieves the most recent state for an endpoint
	GetLatest(ctx context.Context, endpointID string) (*ratelimit.PolicyState, error)

	// GetLatestAll retrieves the most recent state for all endpoints
	GetLatestAll(ctx context.Context) ([]*ratelimit.PolicyState, error)

	// GetHistory retrieves historical states for an endpoint in a time range
	GetHistory(ctx context.Context, endpointID string, since, until time.Time) ([]*ratelimit.PolicyState, error)

	// GetByCategory retrieves latest states for endpoints of a specific category
	GetByCategory(ctx context.Context, category string, limit int) ([]*ratelimit.PolicyState, error)

	// GetPreviousState retrieves the state immediately before the given timestamp
	// Used for calculating consumption rates
	GetPreviousState(ctx context.Context, endpointID string, before time.Time) (*ratelimit.PolicyState, error)

	// DeleteOlderThan deletes states older than the specified timestamp
	// Returns the number of records deleted (for maintenance/cleanup)
	DeleteOlderThan(ctx context.Context, timestamp time.Time) (int64, error)
}

// AlertRepository defines the interface for rate limit alert storage operations
type AlertRepository interface {
	// Save inserts a new alert
	Save(ctx context.Context, alert *ratelimit.Alert) error

	// GetUnresolved retrieves all unresolved alerts
	GetUnresolved(ctx context.Context) ([]*ratelimit.Alert, error)

	// GetUnresolvedByEndpoint retrieves unresolved alerts for a specific endpoint
	GetUnresolvedByEndpoint(ctx context.Context, endpointID string) ([]*ratelimit.Alert, error)

	// GetUnresolvedBySeverity retrieves unresolved alerts by severity
	GetUnresolvedBySeverity(ctx context.Context, severity ratelimit.AlertSeverity) ([]*ratelimit.Alert, error)

	// Resolve marks an alert as resolved
	Resolve(ctx context.Context, alertID int64, notes string) error

	// ResolveBulk marks multiple alerts as resolved
	ResolveBulk(ctx context.Context, alertIDs []int64, notes string) error

	// AutoResolve automatically resolves alerts based on current state
	// Returns the number of alerts resolved
	AutoResolve(ctx context.Context, endpointID string, availableTokens, capacity int) (int, error)

	// GetHistory retrieves alert history in a time range
	GetHistory(ctx context.Context, since, until time.Time) ([]*ratelimit.Alert, error)

	// GetHistoryByEndpoint retrieves alert history for a specific endpoint
	GetHistoryByEndpoint(ctx context.Context, endpointID string, since, until time.Time) ([]*ratelimit.Alert, error)
}
