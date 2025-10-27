package activities

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ActivityOptions provides standardized activity configuration for different operation types
type ActivityOptions struct {
	Database      workflow.ActivityOptions
	ExternalAPI   workflow.ActivityOptions
	Messaging     workflow.ActivityOptions
	Validation    workflow.ActivityOptions
	LongRunning   workflow.ActivityOptions
}

// NewActivityOptions creates activity options with production-ready defaults
func NewActivityOptions() *ActivityOptions {
	return &ActivityOptions{
		// Database operations: fast, retriable, short timeout
		Database: workflow.ActivityOptions{
			StartToCloseTimeout: 10 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    1 * time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    30 * time.Second,
				MaximumAttempts:    5,
			},
			HeartbeatTimeout: 5 * time.Second,
		},

		// External API calls: longer timeout, more retries
		ExternalAPI: workflow.ActivityOptions{
			StartToCloseTimeout:    30 * time.Second,
			ScheduleToCloseTimeout: 2 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    2 * time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    1 * time.Minute,
				MaximumAttempts:    10,
			},
			HeartbeatTimeout: 10 * time.Second,
		},

		// Messaging (Pulsar): fast, retriable
		Messaging: workflow.ActivityOptions{
			StartToCloseTimeout: 15 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    500 * time.Millisecond,
				BackoffCoefficient: 2.0,
				MaximumInterval:    20 * time.Second,
				MaximumAttempts:    7,
			},
			HeartbeatTimeout: 5 * time.Second,
		},

		// Validation: fast, no retries on validation errors
		Validation: workflow.ActivityOptions{
			StartToCloseTimeout: 5 * time.Second,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    1 * time.Second,
				BackoffCoefficient: 1.5,
				MaximumInterval:    10 * time.Second,
				MaximumAttempts:    3,
			},
		},

		// Long-running operations: extended timeouts, periodic heartbeats
		LongRunning: workflow.ActivityOptions{
			StartToCloseTimeout:    5 * time.Minute,
			ScheduleToCloseTimeout: 10 * time.Minute,
			RetryPolicy: &temporal.RetryPolicy{
				InitialInterval:    5 * time.Second,
				BackoffCoefficient: 2.0,
				MaximumInterval:    2 * time.Minute,
				MaximumAttempts:    15,
			},
			HeartbeatTimeout: 30 * time.Second,
		},
	}
}

// ClaimActivityTimeouts provides specific timeouts for claim workflow activities
type ClaimActivityTimeouts struct {
	CreateClaim         time.Duration
	ValidateEligibility time.Duration
	NotifyDonor         time.Duration
	SendConfirmation    time.Duration
	CompleteClaim       time.Duration
	CancelClaim         time.Duration
	ExpireClaim         time.Duration
	GetStatus           time.Duration
	UpdateOwnership     time.Duration
	PublishEvent        time.Duration
}

// NewClaimActivityTimeouts returns production timeouts for claim activities
func NewClaimActivityTimeouts() *ClaimActivityTimeouts {
	return &ClaimActivityTimeouts{
		CreateClaim:         10 * time.Second, // Database insert
		ValidateEligibility: 5 * time.Second,  // Quick validation
		NotifyDonor:         30 * time.Second, // External API call
		SendConfirmation:    30 * time.Second, // External API call
		CompleteClaim:       15 * time.Second, // Database update + event
		CancelClaim:         15 * time.Second, // Database update + event
		ExpireClaim:         15 * time.Second, // Database update + event
		GetStatus:           3 * time.Second,  // Simple query
		UpdateOwnership:     10 * time.Second, // Database update
		PublishEvent:        5 * time.Second,  // Pulsar publish
	}
}

// GetActivityOptionsByType returns the appropriate activity options for an operation type
func (ao *ActivityOptions) GetActivityOptionsByType(activityType string) workflow.ActivityOptions {
	switch activityType {
	case "database", "db":
		return ao.Database
	case "api", "external":
		return ao.ExternalAPI
	case "messaging", "pulsar", "event":
		return ao.Messaging
	case "validation", "validate":
		return ao.Validation
	case "long-running", "long":
		return ao.LongRunning
	default:
		// Default to database options (most common)
		return ao.Database
	}
}