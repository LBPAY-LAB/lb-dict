package interfaces

import (
	"context"

	"github.com/lbpay-lab/conn-dict/internal/domain/aggregates"
)

// TemporalClient defines the interface for Temporal workflow operations
type TemporalClient interface {
	// StartClaimWorkflow starts a DICT claim workflow
	StartClaimWorkflow(ctx context.Context, workflowID string, claim *aggregates.Claim) error

	// StartVsyncWorkflow starts a DICT vsync workflow
	StartVsyncWorkflow(ctx context.Context, workflowID string, entry *aggregates.VsyncEntry) error

	// CancelWorkflow cancels a running workflow
	CancelWorkflow(ctx context.Context, workflowID string) error

	// GetWorkflowStatus retrieves the status of a workflow
	GetWorkflowStatus(ctx context.Context, workflowID string) (string, error)

	// SignalWorkflow sends a signal to a running workflow
	SignalWorkflow(ctx context.Context, workflowID, signalName string, data interface{}) error

	// QueryWorkflow queries a running workflow
	QueryWorkflow(ctx context.Context, workflowID, queryType string) (interface{}, error)
}
