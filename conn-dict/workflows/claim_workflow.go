package workflows

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

// ClaimWorkflowInput represents the input for the claim workflow
type ClaimWorkflowInput struct {
	// Claim information
	ClaimID            string
	Key                string
	KeyType            string
	DonorISPB          string
	ClaimerISPB        string
	ClaimerAccountBranch string
	ClaimerAccountNumber string
	ClaimerAccountType   string
	ClaimType          string // "PORTABILITY" or "OWNERSHIP"

	// Request metadata
	RequestedBy    string
	CorrelationID  string
}

// ClaimWorkflowResult represents the result of the claim workflow
type ClaimWorkflowResult struct {
	WorkflowID    string
	ClaimID       string
	Status        string // "PENDING", "CONFIRMED", "CANCELLED", "FAILED"
	Message       string
	CreatedAt     time.Time
	ErrorReason   string
}

// ClaimWorkflow orchestrates the DICT claim process
//
// Sprint 1: Creates claim skeleton without 30-day timer
// Sprint 2: Will add 30-day timer and signal handling for confirmation/cancellation
//
// Flow:
// 1. Validate input
// 2. Create claim in database (via CreateClaimActivity)
// 3. Submit claim to Bacen via Bridge (via SubmitClaimToBacenActivity)
// 4. Update claim status to PENDING (via UpdateClaimStatusActivity)
// 5. Return result with workflow ID and claim ID
//
// TODO Sprint 2: Add 30-day timer and signal handling for:
// - confirm-claim signal
// - cancel-claim signal
func ClaimWorkflow(ctx workflow.Context, input ClaimWorkflowInput) (*ClaimWorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ClaimWorkflow started",
		"claim_id", input.ClaimID,
		"key", input.Key,
		"key_type", input.KeyType,
		"claimer_ispb", input.ClaimerISPB,
		"donor_ispb", input.DonorISPB,
	)

	// Get workflow info
	workflowInfo := workflow.GetInfo(ctx)

	// Initialize result
	result := &ClaimWorkflowResult{
		WorkflowID: workflowInfo.WorkflowExecution.ID,
		ClaimID:    input.ClaimID,
		CreatedAt:  workflow.Now(ctx),
	}

	// Step 1: Validate input
	logger.Info("Step 1: Validating claim workflow input")
	if err := validateClaimWorkflowInput(input); err != nil {
		logger.Error("Invalid claim workflow input", "error", err)
		result.Status = "FAILED"
		result.Message = "Validation failed"
		result.ErrorReason = err.Error()
		return result, fmt.Errorf("invalid claim input: %w", err)
	}

	// Configure activity options with retry policy
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    2 * time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    1 * time.Minute,
			MaximumAttempts:    3,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Step 2: Create claim in database
	logger.Info("Step 2: Creating claim in database")
	var createResult CreateClaimResult
	err := workflow.ExecuteActivity(ctx, "CreateClaimActivity", input).Get(ctx, &createResult)
	if err != nil {
		logger.Error("Failed to create claim in database", "error", err)
		result.Status = "FAILED"
		result.Message = "Failed to create claim"
		result.ErrorReason = err.Error()
		return result, fmt.Errorf("create claim failed: %w", err)
	}

	logger.Info("Claim created in database", "claim_uuid", createResult.ClaimUUID)

	// Step 3: Submit claim to Bacen via Bridge
	logger.Info("Step 3: Submitting claim to Bacen")
	var submitResult SubmitClaimToBacenResult
	err = workflow.ExecuteActivity(ctx, "SubmitClaimToBacenActivity", SubmitClaimToBacenInput{
		ClaimID:              input.ClaimID,
		Key:                  input.Key,
		KeyType:              input.KeyType,
		DonorISPB:            input.DonorISPB,
		ClaimerISPB:          input.ClaimerISPB,
		ClaimerAccountBranch: input.ClaimerAccountBranch,
		ClaimerAccountNumber: input.ClaimerAccountNumber,
		ClaimerAccountType:   input.ClaimerAccountType,
		ClaimType:            input.ClaimType,
		CorrelationID:        input.CorrelationID,
	}).Get(ctx, &submitResult)

	if err != nil {
		logger.Error("Failed to submit claim to Bacen", "error", err)
		// Update claim status to FAILED
		_ = workflow.ExecuteActivity(ctx, "UpdateClaimStatusActivity", UpdateClaimStatusInput{
			ClaimID: input.ClaimID,
			Status:  "CANCELLED",
			Reason:  fmt.Sprintf("Bacen submission failed: %v", err),
		}).Get(ctx, nil)

		result.Status = "FAILED"
		result.Message = "Failed to submit claim to Bacen"
		result.ErrorReason = err.Error()
		return result, fmt.Errorf("submit to Bacen failed: %w", err)
	}

	logger.Info("Claim submitted to Bacen successfully",
		"bacen_correlation_id", submitResult.BacenCorrelationID,
		"success", submitResult.Success,
	)

	// Step 4: Update claim status to PENDING
	logger.Info("Step 4: Updating claim status to PENDING")
	err = workflow.ExecuteActivity(ctx, "UpdateClaimStatusActivity", UpdateClaimStatusInput{
		ClaimID: input.ClaimID,
		Status:  "WAITING_RESOLUTION",
		Reason:  "Claim submitted to Bacen, waiting for resolution",
	}).Get(ctx, nil)

	if err != nil {
		logger.Warn("Failed to update claim status", "error", err)
		// Non-critical - claim is submitted, status update can be retried
	}

	// Step 5: Return success result
	result.Status = "PENDING"
	result.Message = "Claim created and submitted to Bacen successfully"

	logger.Info("ClaimWorkflow completed successfully",
		"workflow_id", result.WorkflowID,
		"claim_id", result.ClaimID,
		"status", result.Status,
	)

	// TODO Sprint 2: Add 30-day timer here
	// TODO Sprint 2: Add signal handlers for confirm-claim and cancel-claim

	return result, nil
}

// validateClaimWorkflowInput validates the claim workflow input
func validateClaimWorkflowInput(input ClaimWorkflowInput) error {
	if input.ClaimID == "" {
		return fmt.Errorf("claim_id is required")
	}
	if input.Key == "" {
		return fmt.Errorf("key is required")
	}
	if input.KeyType == "" {
		return fmt.Errorf("key_type is required")
	}
	if input.DonorISPB == "" {
		return fmt.Errorf("donor_ispb is required")
	}
	if len(input.DonorISPB) != 8 {
		return fmt.Errorf("donor_ispb must be 8 digits")
	}
	if input.ClaimerISPB == "" {
		return fmt.Errorf("claimer_ispb is required")
	}
	if len(input.ClaimerISPB) != 8 {
		return fmt.Errorf("claimer_ispb must be 8 digits")
	}
	if input.DonorISPB == input.ClaimerISPB {
		return fmt.Errorf("donor and claimer must be different")
	}
	if input.ClaimType != "PORTABILITY" && input.ClaimType != "OWNERSHIP" {
		return fmt.Errorf("claim_type must be PORTABILITY or OWNERSHIP")
	}
	return nil
}

// Activity input/output structs

// CreateClaimResult is the result of CreateClaimActivity
type CreateClaimResult struct {
	ClaimUUID string
	ClaimID   string
	Success   bool
}

// SubmitClaimToBacenInput is the input for SubmitClaimToBacenActivity
type SubmitClaimToBacenInput struct {
	ClaimID              string
	Key                  string
	KeyType              string
	DonorISPB            string
	ClaimerISPB          string
	ClaimerAccountBranch string
	ClaimerAccountNumber string
	ClaimerAccountType   string
	ClaimType            string
	CorrelationID        string
}

// SubmitClaimToBacenResult is the result of SubmitClaimToBacenActivity
type SubmitClaimToBacenResult struct {
	Success            bool
	BacenCorrelationID string
	ErrorCode          string
	ErrorMessage       string
}

// UpdateClaimStatusInput is the input for UpdateClaimStatusActivity
type UpdateClaimStatusInput struct {
	ClaimID string
	Status  string
	Reason  string
}
