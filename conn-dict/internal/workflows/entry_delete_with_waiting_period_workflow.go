package workflows

import (
	"fmt"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/activities"
	"go.temporal.io/sdk/workflow"
)

// DeleteEntryWithWaitingPeriodWorkflowInput represents the input parameters for the DeleteEntry workflow with 30-day waiting period
type DeleteEntryWithWaitingPeriodWorkflowInput struct {
	EntryID         string `json:"entry_id"`
	DeletionReason  string `json:"deletion_reason"`
	RequestedBy     string `json:"requested_by"`
}

// DeleteEntryWithWaitingPeriodWorkflowResult represents the result of the DeleteEntry workflow with 30-day waiting period
type DeleteEntryWithWaitingPeriodWorkflowResult struct {
	EntryID        string    `json:"entry_id"`
	Status         string    `json:"status"` // "DEACTIVATED", "DELETED", "CANCELLED", "FAILED"
	DeactivatedAt  time.Time `json:"deactivated_at,omitempty"`
	DeletedAt      time.Time `json:"deleted_at,omitempty"`
	Message        string    `json:"message"`
	ErrorReason    string    `json:"error_reason,omitempty"`
}

const (
	// DeleteEntryTimeout is the maximum duration for entry deletion (31 days to account for 30-day wait)
	DeleteEntryTimeout = 31 * 24 * time.Hour

	// EntryDeletionWaitPeriod is the mandatory 30-day waiting period before permanent deletion
	// This follows Bacen DICT regulations
	EntryDeletionWaitPeriod = 30 * 24 * time.Hour

	// EntryDeletionReason constants
	DeletionReasonUserRequest  = "USER_REQUEST"
	DeletionReasonCompliance   = "COMPLIANCE"
	DeletionReasonFraud        = "FRAUD"
	DeletionReasonDuplicate    = "DUPLICATE"
	DeletionReasonAdminAction  = "ADMIN_ACTION"
)

// DeleteEntryWithWaitingPeriodWorkflow is the main Temporal workflow for deleting PIX key entries with a 30-day waiting period
//
// This workflow implements the entry deletion process following Bacen regulations:
// 1. Deactivate entry immediately (set status to INACTIVE)
// 2. Wait for mandatory 30 days period
// 3. Perform soft delete after waiting period
// 4. Notify Bacen DICT about the deletion
//
// The mandatory 30-day waiting period allows for:
// - User reconsideration (can reactivate during this period)
// - Pending transactions to complete
// - Compliance and audit requirements
// - Fraud investigation if needed
//
// Signals supported:
// - "cancel_deletion" → Cancels the deletion and reactivates the entry
//
// Error handling:
// - Deactivation errors: retry with exponential backoff
// - Delete errors: retry, log for manual intervention
// - Bacen notification errors: non-critical, retry separately
//
// Note: For immediate deletion (admin/compliance), use Pulsar Consumer directly instead of this workflow
func DeleteEntryWithWaitingPeriodWorkflow(ctx workflow.Context, input DeleteEntryWithWaitingPeriodWorkflowInput) (*DeleteEntryWithWaitingPeriodWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("DeleteEntryWithWaitingPeriodWorkflow started",
		"entry_id", input.EntryID,
		"reason", input.DeletionReason,
	)

	// Validate input
	if err := validateDeleteEntryInput(input); err != nil {
		logger.Error("Invalid delete input", "error", err)
		return &DeleteEntryWithWaitingPeriodWorkflowResult{
			EntryID:     input.EntryID,
			Status:      "FAILED",
			Message:     "Validation failed",
			ErrorReason: err.Error(),
		}, fmt.Errorf("invalid delete input: %w", err)
	}

	// Get activity options
	activityOpts := activities.NewActivityOptions()
	result := &DeleteEntryWithWaitingPeriodWorkflowResult{
		EntryID: input.EntryID,
	}

	// Step 1: Deactivate entry (set to INACTIVE status)
	logger.Info("Step 1: Deactivating entry", "entry_id", input.EntryID)
	ctx1 := workflow.WithActivityOptions(ctx, activityOpts.Database)

	err := workflow.ExecuteActivity(ctx1, "DeactivateEntryActivity", input.EntryID, input.DeletionReason).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to deactivate entry", "error", err)
		result.Status = "FAILED"
		result.Message = "Entry deactivation failed"
		result.ErrorReason = err.Error()
		return result, fmt.Errorf("deactivate entry failed: %w", err)
	}

	result.Status = "DEACTIVATED"
	result.DeactivatedAt = workflow.Now(ctx)
	logger.Info("Entry deactivated successfully", "entry_id", input.EntryID)

	// Step 2: Wait for mandatory 30-day period
	logger.Info("Step 2: Waiting for mandatory 30-day period before deletion",
		"entry_id", input.EntryID,
		"wait_until", workflow.Now(ctx).Add(EntryDeletionWaitPeriod),
	)

	// Create channel for cancel signal
	cancelChannel := workflow.GetSignalChannel(ctx, "cancel_deletion")

	// Selector to wait for signals or timeout
	selector := workflow.NewSelector(ctx)

	// Handle "cancel_deletion" signal - reactivate entry
	selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
		var cancelData struct {
			Reason      string `json:"reason"`
			CancelledBy string `json:"cancelled_by"`
		}
		c.Receive(ctx, &cancelData)

		logger.Info("Deletion cancelled - reactivating entry",
			"entry_id", input.EntryID,
			"reason", cancelData.Reason,
			"cancelled_by", cancelData.CancelledBy,
		)

		// Reactivate entry
		ctxReactivate := workflow.WithActivityOptions(ctx, activityOpts.Database)
		err := workflow.ExecuteActivity(ctxReactivate, "ActivateEntryActivity", input.EntryID).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to reactivate entry", "error", err)
			result.Status = "FAILED"
			result.Message = "Failed to cancel deletion"
			result.ErrorReason = err.Error()
			return
		}

		result.Status = "CANCELLED"
		result.Message = fmt.Sprintf("Deletion cancelled by %s: %s", cancelData.CancelledBy, cancelData.Reason)
	})

	// Wait for 30 days or until signal received
	timer := workflow.NewTimer(ctx, EntryDeletionWaitPeriod)
	selector.AddFuture(timer, func(f workflow.Future) {
		logger.Info("30-day waiting period completed", "entry_id", input.EntryID)
	})

	// Block until one of the conditions is met
	selector.Select(ctx)

	// If deletion was cancelled, return early
	if result.Status == "CANCELLED" {
		logger.Info("DeleteEntryWithWaitingPeriodWorkflow cancelled",
			"entry_id", input.EntryID,
			"message", result.Message,
		)
		return result, nil
	}

	// Step 3: Perform soft delete
	logger.Info("Step 3: Performing soft delete", "entry_id", input.EntryID)
	ctx3 := workflow.WithActivityOptions(ctx, activityOpts.Database)

	err = workflow.ExecuteActivity(ctx3, "DeleteEntryActivity", input.EntryID).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to delete entry", "error", err)
		result.Status = "FAILED"
		result.Message = "Entry deletion failed"
		result.ErrorReason = err.Error()
		return result, fmt.Errorf("delete entry failed: %w", err)
	}

	result.Status = "DELETED"
	result.DeletedAt = workflow.Now(ctx)

	// Step 4: Notify Bacen DICT about deletion
	logger.Info("Step 4: Notifying Bacen about entry deletion", "entry_id", input.EntryID)
	ctx4 := workflow.WithActivityOptions(ctx, activityOpts.ExternalAPI)

	err = workflow.ExecuteActivity(ctx4, "NotifyBacenActivity", input.EntryID).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to notify Bacen (non-critical)", "error", err)
		// Non-critical - entry is deleted, but Bacen notification failed
		result.Message = "Entry deleted but Bacen notification pending - will retry"
		result.ErrorReason = fmt.Sprintf("Bacen notification failed: %s", err.Error())
		return result, nil
	}

	// Success
	if result.Message == "" {
		result.Message = "Entry deleted successfully after 30-day waiting period and Bacen notified"
	}

	logger.Info("DeleteEntryWithWaitingPeriodWorkflow completed successfully",
		"entry_id", input.EntryID,
		"status", result.Status,
		"deactivated_at", result.DeactivatedAt,
		"deleted_at", result.DeletedAt,
	)

	return result, nil
}

// validateDeleteEntryInput validates the DeleteEntryWithWaitingPeriod workflow input
func validateDeleteEntryInput(input DeleteEntryWithWaitingPeriodWorkflowInput) error {
	if input.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}
	if input.DeletionReason == "" {
		return fmt.Errorf("deletion_reason is required")
	}

	// Validate deletion reason
	validReasons := map[string]bool{
		DeletionReasonUserRequest:  true,
		DeletionReasonCompliance:   true,
		DeletionReasonFraud:        true,
		DeletionReasonDuplicate:    true,
		DeletionReasonAdminAction:  true,
	}

	if !validReasons[input.DeletionReason] {
		return fmt.Errorf("invalid deletion_reason: must be one of USER_REQUEST, COMPLIANCE, FRAUD, DUPLICATE, ADMIN_ACTION")
	}

	return nil
}