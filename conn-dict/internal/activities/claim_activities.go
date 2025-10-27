package activities

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/pulsar"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/lbpay-lab/conn-dict/workflows"
	"github.com/sirupsen/logrus"
)

// ClaimActivities contains all Temporal activities for claim workflow
type ClaimActivities struct {
	logger         *logrus.Logger
	claimRepo      *repositories.ClaimRepository
	pulsarProducer *pulsar.Producer
	// TODO Sprint 2: Add Bridge gRPC client
	// bridgeClient   *grpc.BridgeClient
}

// NewClaimActivities creates a new instance of ClaimActivities
func NewClaimActivities(
	logger *logrus.Logger,
	claimRepo *repositories.ClaimRepository,
	pulsarProducer *pulsar.Producer,
) *ClaimActivities {
	return &ClaimActivities{
		logger:         logger,
		claimRepo:      claimRepo,
		pulsarProducer: pulsarProducer,
	}
}

// CreateClaimActivity creates a new claim in the database
//
// Sprint 1: This activity persists the claim to PostgreSQL and returns the claim UUID
// This activity is idempotent: if a claim with the same ClaimID already exists,
// it returns the existing claim UUID instead of creating a duplicate.
//
// Input: workflows.ClaimWorkflowInput
// Output: workflows.CreateClaimResult
func (a *ClaimActivities) CreateClaimActivity(ctx context.Context, input workflows.ClaimWorkflowInput) (*workflows.CreateClaimResult, error) {
	logger := activity.GetLogger(ctx)
	activityInfo := activity.GetInfo(ctx)

	logger.Info("CreateClaimActivity started",
		"claim_id", input.ClaimID,
		"key", input.Key,
		"activity_id", activityInfo.ActivityID,
		"attempt", activityInfo.Attempt,
	)

	// Check if claim already exists (idempotency)
	existingClaim, err := a.claimRepo.GetByClaimID(ctx, input.ClaimID)
	if err == nil && existingClaim != nil {
		logger.Info("Claim already exists (idempotent operation)",
			"claim_id", input.ClaimID,
			"claim_uuid", existingClaim.ID,
		)
		return &workflows.CreateClaimResult{
			ClaimUUID: existingClaim.ID.String(),
			ClaimID:   existingClaim.ClaimID,
			Success:   true,
		}, nil
	}

	// Parse claim type
	var claimType entities.ClaimType
	if input.ClaimType == "PORTABILITY" {
		claimType = entities.ClaimTypePortability
	} else if input.ClaimType == "OWNERSHIP" {
		claimType = entities.ClaimTypeOwnership
	} else {
		return nil, fmt.Errorf("invalid claim type: %s", input.ClaimType)
	}

	// Create new claim entity
	claim, err := entities.NewClaim(
		input.ClaimID,
		claimType,
		input.Key,
		input.KeyType,
		input.DonorISPB,
		input.ClaimerISPB,
	)
	if err != nil {
		logger.Error("Failed to create claim entity", "error", err)
		return nil, fmt.Errorf("failed to create claim entity: %w", err)
	}

	// Set account information
	claim.ClaimerAccountBranch = input.ClaimerAccountBranch
	claim.ClaimerAccountNumber = input.ClaimerAccountNumber
	claim.ClaimerAccountType = input.ClaimerAccountType

	// Persist claim to database
	err = a.claimRepo.Create(ctx, claim)
	if err != nil {
		logger.Error("Failed to persist claim to database", "error", err)
		return nil, fmt.Errorf("failed to persist claim: %w", err)
	}

	logger.Info("Claim created successfully in database",
		"claim_id", claim.ClaimID,
		"claim_uuid", claim.ID,
		"key", claim.Key,
		"status", claim.Status,
	)

	// Record heartbeat for long-running activities
	activity.RecordHeartbeat(ctx, "Claim created in database")

	return &workflows.CreateClaimResult{
		ClaimUUID: claim.ID.String(),
		ClaimID:   claim.ClaimID,
		Success:   true,
	}, nil
}

// SubmitClaimToBacenActivity submits the claim to Bacen via Bridge gRPC
//
// Sprint 1: This activity simulates submission to Bacen. In Sprint 2, it will call
// the actual Bridge gRPC service to submit the claim to Bacen DICT.
//
// This activity handles retries automatically via Temporal's retry policy.
// It is idempotent and can be safely retried.
//
// Input: workflows.SubmitClaimToBacenInput
// Output: workflows.SubmitClaimToBacenResult
func (a *ClaimActivities) SubmitClaimToBacenActivity(ctx context.Context, input workflows.SubmitClaimToBacenInput) (*workflows.SubmitClaimToBacenResult, error) {
	logger := activity.GetLogger(ctx)
	activityInfo := activity.GetInfo(ctx)

	logger.Info("SubmitClaimToBacenActivity started",
		"claim_id", input.ClaimID,
		"key", input.Key,
		"donor_ispb", input.DonorISPB,
		"claimer_ispb", input.ClaimerISPB,
		"activity_id", activityInfo.ActivityID,
		"attempt", activityInfo.Attempt,
	)

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Preparing to submit claim to Bacen")

	// TODO Sprint 2: Implement actual Bridge gRPC call
	// For now, we simulate a successful submission
	//
	// In production, this should:
	// 1. Create gRPC request with claim data
	// 2. Call Bridge service: bridgeClient.CreateClaim(ctx, request)
	// 3. Handle gRPC errors (connection, timeout, etc.)
	// 4. Parse Bridge response
	// 5. Return appropriate result
	//
	// Example implementation:
	/*
		bridgeReq := &bridgepb.CreateClaimRequest{
			ClaimId:              input.ClaimID,
			Key:                  input.Key,
			KeyType:              input.KeyType,
			DonorIspb:            input.DonorISPB,
			ClaimerIspb:          input.ClaimerISPB,
			ClaimerAccountBranch: input.ClaimerAccountBranch,
			ClaimerAccountNumber: input.ClaimerAccountNumber,
			ClaimerAccountType:   input.ClaimerAccountType,
			ClaimType:            input.ClaimType,
			CorrelationId:        input.CorrelationID,
		}

		// Call Bridge with context timeout
		ctxTimeout, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		bridgeResp, err := a.bridgeClient.CreateClaim(ctxTimeout, bridgeReq)
		if err != nil {
			logger.Error("Bridge gRPC call failed", "error", err)
			return nil, fmt.Errorf("bridge call failed: %w", err)
		}

		if !bridgeResp.Success {
			logger.Warn("Bacen rejected claim",
				"error_code", bridgeResp.ErrorCode,
				"error_message", bridgeResp.ErrorMessage,
			)
			return &workflows.SubmitClaimToBacenResult{
				Success:            false,
				BacenCorrelationID: bridgeResp.CorrelationId,
				ErrorCode:          bridgeResp.ErrorCode,
				ErrorMessage:       bridgeResp.ErrorMessage,
			}, nil
		}

		logger.Info("Claim submitted to Bacen successfully",
			"bacen_correlation_id", bridgeResp.CorrelationId,
		)

		return &workflows.SubmitClaimToBacenResult{
			Success:            true,
			BacenCorrelationID: bridgeResp.CorrelationId,
		}, nil
	*/

	// Simulate processing time
	time.Sleep(500 * time.Millisecond)

	// Simulate successful submission
	bacenCorrelationID := fmt.Sprintf("BACEN-%s-%d", input.ClaimID, time.Now().Unix())

	logger.Info("Claim submitted to Bacen successfully (simulated)",
		"claim_id", input.ClaimID,
		"bacen_correlation_id", bacenCorrelationID,
	)

	// Record heartbeat
	activity.RecordHeartbeat(ctx, "Claim submitted to Bacen")

	return &workflows.SubmitClaimToBacenResult{
		Success:            true,
		BacenCorrelationID: bacenCorrelationID,
		ErrorCode:          "",
		ErrorMessage:       "",
	}, nil
}

// UpdateClaimStatusActivity updates the claim status in the database and publishes an event
//
// Sprint 1: This activity updates the claim status and can publish events to Pulsar.
// This activity is idempotent: it can be safely retried without side effects.
//
// Input: workflows.UpdateClaimStatusInput
// Output: error (nil on success)
func (a *ClaimActivities) UpdateClaimStatusActivity(ctx context.Context, input workflows.UpdateClaimStatusInput) error {
	logger := activity.GetLogger(ctx)
	activityInfo := activity.GetInfo(ctx)

	logger.Info("UpdateClaimStatusActivity started",
		"claim_id", input.ClaimID,
		"new_status", input.Status,
		"reason", input.Reason,
		"activity_id", activityInfo.ActivityID,
		"attempt", activityInfo.Attempt,
	)

	// Get existing claim
	claim, err := a.claimRepo.GetByClaimID(ctx, input.ClaimID)
	if err != nil {
		logger.Error("Failed to get claim", "error", err)
		return fmt.Errorf("failed to get claim: %w", err)
	}

	// Check if status is already updated (idempotency)
	if string(claim.Status) == input.Status {
		logger.Info("Claim status already updated (idempotent operation)",
			"claim_id", input.ClaimID,
			"status", input.Status,
		)
		return nil
	}

	// Parse and validate new status
	newStatus := entities.ClaimStatus(input.Status)
	if err := claim.ValidateStatusTransition(newStatus); err != nil {
		logger.Warn("Invalid status transition",
			"current_status", claim.Status,
			"new_status", newStatus,
			"error", err,
		)
		// Don't fail - this might be a legitimate state (e.g., claim already processed)
		return nil
	}

	oldStatus := claim.Status

	// Update status based on new state
	switch newStatus {
	case entities.ClaimStatusConfirmed:
		err = claim.Confirm()
	case entities.ClaimStatusCancelled:
		err = claim.Cancel(input.Reason)
	case entities.ClaimStatusExpired:
		err = claim.Expire()
	case entities.ClaimStatusWaitingResolution:
		err = claim.MoveToWaitingResolution()
	default:
		// Direct status update
		claim.Status = newStatus
		claim.UpdatedAt = time.Now()
	}

	if err != nil {
		logger.Error("Failed to update claim status", "error", err)
		return fmt.Errorf("failed to update status: %w", err)
	}

	// Persist to database
	err = a.claimRepo.Update(ctx, claim)
	if err != nil {
		logger.Error("Failed to persist claim status update", "error", err)
		return fmt.Errorf("failed to persist status update: %w", err)
	}

	logger.Info("Claim status updated successfully",
		"claim_id", claim.ClaimID,
		"old_status", oldStatus,
		"new_status", newStatus,
	)

	// Publish status change event to Pulsar
	event := map[string]interface{}{
		"event_type":  "claim_status_changed",
		"claim_id":    claim.ClaimID,
		"key":         claim.Key,
		"old_status":  string(oldStatus),
		"new_status":  string(newStatus),
		"reason":      input.Reason,
		"updated_at":  claim.UpdatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, claim.ClaimID); err != nil {
		logger.Warn("Failed to publish status change event (non-critical)", "error", err)
		// Non-critical - event publishing can be retried
	}

	// Record heartbeat
	activity.RecordHeartbeat(ctx, fmt.Sprintf("Status updated to %s", newStatus))

	return nil
}

// GetClaimStatusActivity retrieves the current status of a claim
func (a *ClaimActivities) GetClaimStatusActivity(ctx context.Context, claimID string) (string, error) {
	a.logger.Infof("Getting claim status: %s", claimID)

	claim, err := a.claimRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		a.logger.WithError(err).Errorf("Failed to get claim: %s", claimID)
		return "", fmt.Errorf("claim not found: %w", err)
	}

	return string(claim.Status), nil
}

// CompleteClaimActivity completes the claim after donor confirmation
func (a *ClaimActivities) CompleteClaimActivity(ctx context.Context, claimID string) error {
	a.logger.WithField("claim_id", claimID).Info("Completing claim")

	// Get claim from database
	claim, err := a.claimRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}

	// Update claim status to completed
	if err := claim.Complete(); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update in database
	if err := a.claimRepo.Update(ctx, claim); err != nil {
		return fmt.Errorf("failed to update claim: %w", err)
	}

	// Publish completion event
	event := map[string]interface{}{
		"event_type":   "claim_completed",
		"claim_id":     claim.ClaimID,
		"key":          claim.Key,
		"completed_at": claim.CompletedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, claim.ClaimID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish claim completed event")
	}

	a.logger.WithField("claim_id", claimID).Info("Claim completed successfully")

	return nil
}

// CancelClaimActivity cancels the claim
func (a *ClaimActivities) CancelClaimActivity(ctx context.Context, claimID, reason string) error {
	a.logger.WithFields(logrus.Fields{
		"claim_id": claimID,
		"reason":   reason,
	}).Info("Cancelling claim")

	// Get claim from database
	claim, err := a.claimRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}

	// Update claim status to cancelled
	if err := claim.Cancel(reason); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update in database
	if err := a.claimRepo.Update(ctx, claim); err != nil {
		return fmt.Errorf("failed to update claim: %w", err)
	}

	// Publish cancellation event
	event := map[string]interface{}{
		"event_type":   "claim_cancelled",
		"claim_id":     claim.ClaimID,
		"key":          claim.Key,
		"reason":       reason,
		"cancelled_at": claim.CancelledAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, claim.ClaimID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish claim cancelled event")
	}

	a.logger.WithField("claim_id", claimID).Info("Claim cancelled successfully")

	return nil
}

// ExpireClaimActivity expires the claim after 30 days timeout
func (a *ClaimActivities) ExpireClaimActivity(ctx context.Context, claimID string) error {
	a.logger.WithField("claim_id", claimID).Info("Expiring claim")

	// Get claim from database
	claim, err := a.claimRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}

	// Update claim status to expired
	if err := claim.Expire(); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update in database
	if err := a.claimRepo.Update(ctx, claim); err != nil {
		return fmt.Errorf("failed to update claim: %w", err)
	}

	// Publish expiration event
	event := map[string]interface{}{
		"event_type": "claim_expired",
		"claim_id":   claim.ClaimID,
		"key":        claim.Key,
		"expired_at": claim.ExpiredAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, claim.ClaimID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish claim expired event")
	}

	a.logger.WithField("claim_id", claimID).Info("Claim expired successfully")

	return nil
}

// ValidateClaimEligibilityActivity validates if entry is eligible for claim
func (a *ClaimActivities) ValidateClaimEligibilityActivity(ctx context.Context, key string) error {
	a.logger.WithField("key", key).Info("Validating claim eligibility")

	// Check if there's already an active claim for this key
	hasActive, err := a.claimRepo.HasActiveClaim(ctx, key)
	if err != nil {
		return fmt.Errorf("failed to check active claims: %w", err)
	}

	if hasActive {
		return fmt.Errorf("key already has an active claim: %s", key)
	}

	// TODO: Additional validation rules:
	// 1. Check if entry exists in entries table
	// 2. Check if entry is ACTIVE (not already in portability)
	// 3. Check if entry is not blocked
	// 4. Check if claimer is different from current owner

	a.logger.WithField("key", key).Info("Entry is eligible for claim")

	return nil
}

// NotifyDonorActivity sends notification to the donor ISPB about the claim
func (a *ClaimActivities) NotifyDonorActivity(ctx context.Context, claimID string) error {
	a.logger.WithField("claim_id", claimID).Info("Notifying donor about claim")

	// Get claim details
	claim, err := a.claimRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}

	// TODO: Call Bridge gRPC to send DICT message to donor
	// This will be implemented when Bridge gRPC client is ready

	// Publish notification event
	event := map[string]interface{}{
		"event_type": "donor_notified",
		"claim_id":   claim.ClaimID,
		"donor":      claim.DonorParticipant,
		"key":        claim.Key,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, claim.ClaimID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish donor notified event")
	}

	a.logger.WithField("claim_id", claimID).Info("Donor notified successfully")

	return nil
}

// SendClaimConfirmationActivity sends confirmation request to donor
func (a *ClaimActivities) SendClaimConfirmationActivity(ctx context.Context, claimID string) error {
	a.logger.WithField("claim_id", claimID).Info("Sending confirmation request")

	// Get claim details
	claim, err := a.claimRepo.GetByClaimID(ctx, claimID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}

	// TODO: Call Bridge gRPC to send DICT confirmation request

	// Publish event
	event := map[string]interface{}{
		"event_type": "confirmation_sent",
		"claim_id":   claim.ClaimID,
		"donor":      claim.DonorParticipant,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, claim.ClaimID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish confirmation sent event")
	}

	a.logger.WithField("claim_id", claimID).Info("Confirmation request sent")

	return nil
}

// UpdateEntryOwnershipActivity updates entry ownership after claim completion
func (a *ClaimActivities) UpdateEntryOwnershipActivity(ctx context.Context, key, newOwnerISPB string) error {
	a.logger.WithFields(logrus.Fields{
		"key":       key,
		"new_owner": newOwnerISPB,
	}).Info("Updating entry ownership")

	// TODO: Update entry table with new owner ISPB
	// This will be implemented when EntryRepository is created

	// Publish event
	event := map[string]interface{}{
		"event_type": "ownership_transferred",
		"key":        key,
		"new_owner":  newOwnerISPB,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, key); err != nil {
		a.logger.WithError(err).Warn("Failed to publish ownership transferred event")
	}

	a.logger.WithField("key", key).Info("Entry ownership updated")

	return nil
}

// PublishClaimEventActivity publishes claim events to Pulsar
func (a *ClaimActivities) PublishClaimEventActivity(ctx context.Context, event map[string]interface{}) error {
	eventType, ok := event["event_type"].(string)
	if !ok {
		return fmt.Errorf("event must have event_type field")
	}

	key, ok := event["key"].(string)
	if !ok {
		key = "unknown"
	}

	a.logger.WithFields(logrus.Fields{
		"event_type": eventType,
		"key":        key,
	}).Info("Publishing claim event")

	if err := a.pulsarProducer.PublishEvent(ctx, event, key); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	return nil
}