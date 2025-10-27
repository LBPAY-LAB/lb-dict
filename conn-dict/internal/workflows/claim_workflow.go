package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ClaimWorkflowInput represents the input parameters for the Claim workflow
type ClaimWorkflowInput struct {
	ClaimID       string `json:"claim_id"`
	EntryID       string `json:"entry_id"`
	ClaimType     string `json:"claim_type"`      // "OWNERSHIP" or "PORTABILITY"
	ClaimerISPB   string `json:"claimer_ispb"`
	DonorISPB     string `json:"donor_ispb"`
	ClaimerAccount string `json:"claimer_account"`
	RequestedBy   string `json:"requested_by"`
}

// ClaimWorkflowResult represents the result of the Claim workflow
type ClaimWorkflowResult struct {
	ClaimID       string    `json:"claim_id"`
	Status        string    `json:"status"` // "COMPLETED", "CANCELLED", "EXPIRED"
	CompletedAt   time.Time `json:"completed_at,omitempty"`
	CancelledAt   time.Time `json:"cancelled_at,omitempty"`
	ExpiredAt     time.Time `json:"expired_at,omitempty"`
	Reason        string    `json:"reason,omitempty"`
	Message       string    `json:"message"`
}

const (
	// ClaimTimeout is the maximum duration for a claim (30 days)
	ClaimTimeout = 30 * 24 * time.Hour

	// ClaimStatusPending indicates the claim is waiting for confirmation
	ClaimStatusPending = "PENDING"

	// ClaimStatusConfirmed indicates the donor confirmed the claim
	ClaimStatusConfirmed = "CONFIRMED"

	// ClaimStatusCancelled indicates the claim was cancelled
	ClaimStatusCancelled = "CANCELLED"

	// ClaimStatusCompleted indicates the claim was completed successfully
	ClaimStatusCompleted = "COMPLETED"

	// ClaimStatusExpired indicates the claim expired after 30 days
	ClaimStatusExpired = "EXPIRED"
)

// ClaimWorkflow is the main Temporal workflow for handling PIX key portability claims
//
// This workflow implements the 30-day claim process defined by Bacen:
// 1. Claim is created (PENDING)
// 2. Wait for donor confirmation or 30 days timeout
// 3. Three possible outcomes:
//    a) Donor confirms → Claim COMPLETED
//    b) Claimer/Donor cancels → Claim CANCELLED
//    c) 30 days timeout → Claim EXPIRED
//
// Signals:
// - "confirm" → Confirms the claim (donor accepts)
// - "cancel"  → Cancels the claim (claimer or donor rejects)
func ClaimWorkflow(ctx workflow.Context, input ClaimWorkflowInput) (*ClaimWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ClaimWorkflow started",
		"claim_id", input.ClaimID,
		"entry_id", input.EntryID,
		"claim_type", input.ClaimType,
		"claimer_ispb", input.ClaimerISPB,
		"donor_ispb", input.DonorISPB,
	)

	// Validate input
	if err := validateClaimInput(input); err != nil {
		return nil, fmt.Errorf("invalid claim input: %w", err)
	}

	// Activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts: 3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 1: Create claim in database
	logger.Info("Creating claim in database", "claim_id", input.ClaimID)
	err := workflow.ExecuteActivity(ctx, "CreateClaimActivity", input).Get(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create claim: %w", err)
	}

	// Step 2: Notify donor about the claim
	logger.Info("Notifying donor ISPB", "donor_ispb", input.DonorISPB)
	err = workflow.ExecuteActivity(ctx, "NotifyDonorActivity", input).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to notify donor", "error", err)
		// Non-critical error, continue workflow
	}

	// Step 3: Wait for confirmation, cancellation, or timeout (30 days)
	result := &ClaimWorkflowResult{
		ClaimID: input.ClaimID,
	}

	// Create channels for signals
	confirmChannel := workflow.GetSignalChannel(ctx, "confirm")
	cancelChannel := workflow.GetSignalChannel(ctx, "cancel")

	// Selector to wait for signals or timeout
	selector := workflow.NewSelector(ctx)

	// Handle "confirm" signal
	selector.AddReceive(confirmChannel, func(c workflow.ReceiveChannel, more bool) {
		var confirmData map[string]interface{}
		c.Receive(ctx, &confirmData)

		logger.Info("Claim confirmed by donor", "claim_id", input.ClaimID)

		// Execute completion activity
		err := workflow.ExecuteActivity(ctx, "CompleteClaimActivity", input.ClaimID).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to complete claim", "error", err)
			result.Status = ClaimStatusPending // Retry later
			return
		}

		result.Status = ClaimStatusCompleted
		result.CompletedAt = workflow.Now(ctx)
		result.Message = "Claim completed successfully - donor confirmed"
	})

	// Handle "cancel" signal
	selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
		var cancelData struct {
			Reason    string `json:"reason"`
			CancelledBy string `json:"cancelled_by"`
		}
		c.Receive(ctx, &cancelData)

		logger.Info("Claim cancelled",
			"claim_id", input.ClaimID,
			"reason", cancelData.Reason,
			"cancelled_by", cancelData.CancelledBy,
		)

		// Execute cancellation activity
		err := workflow.ExecuteActivity(ctx, "CancelClaimActivity", input.ClaimID, cancelData.Reason).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to cancel claim", "error", err)
			result.Status = ClaimStatusPending // Retry later
			return
		}

		result.Status = ClaimStatusCancelled
		result.CancelledAt = workflow.Now(ctx)
		result.Reason = cancelData.Reason
		result.Message = fmt.Sprintf("Claim cancelled by %s", cancelData.CancelledBy)
	})

	// Wait for signal or timeout (30 days)
	logger.Info("Waiting for confirmation or cancellation (30 days timeout)...")

	// Use a timer for 30 days
	timer := workflow.NewTimer(ctx, ClaimTimeout)
	selector.AddFuture(timer, func(f workflow.Future) {
		logger.Info("Claim expired after 30 days", "claim_id", input.ClaimID)

		// Execute expiration activity
		err := workflow.ExecuteActivity(ctx, "ExpireClaimActivity", input.ClaimID).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to expire claim", "error", err)
			result.Status = ClaimStatusPending // Retry later
			return
		}

		result.Status = ClaimStatusExpired
		result.ExpiredAt = workflow.Now(ctx)
		result.Message = "Claim expired after 30 days without confirmation"
	})

	// Block until one of the conditions is met
	selector.Select(ctx)

	logger.Info("ClaimWorkflow completed",
		"claim_id", input.ClaimID,
		"status", result.Status,
	)

	return result, nil
}

// validateClaimInput validates the claim workflow input
func validateClaimInput(input ClaimWorkflowInput) error {
	if input.ClaimID == "" {
		return fmt.Errorf("claim_id is required")
	}
	if input.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}
	if input.ClaimType != "OWNERSHIP" && input.ClaimType != "PORTABILITY" {
		return fmt.Errorf("claim_type must be OWNERSHIP or PORTABILITY")
	}
	if input.ClaimerISPB == "" {
		return fmt.Errorf("claimer_ispb is required")
	}
	if input.DonorISPB == "" {
		return fmt.Errorf("donor_ispb is required")
	}

	return nil
}