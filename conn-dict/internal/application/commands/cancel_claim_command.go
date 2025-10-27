package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// CancelClaimCommand represents the command to cancel a claim
type CancelClaimCommand struct {
	ClaimID string
	Reason  string
}

// CancelClaimHandler handles the CancelClaimCommand
type CancelClaimHandler struct {
	claimRepo *repositories.ClaimRepository
	logger    *logrus.Logger
}

// NewCancelClaimHandler creates a new CancelClaimHandler
func NewCancelClaimHandler(claimRepo *repositories.ClaimRepository, logger *logrus.Logger) *CancelClaimHandler {
	return &CancelClaimHandler{
		claimRepo: claimRepo,
		logger:    logger,
	}
}

// Handle executes the CancelClaimCommand
func (h *CancelClaimHandler) Handle(ctx context.Context, cmd CancelClaimCommand) error {
	// Validate input
	if cmd.ClaimID == "" {
		return fmt.Errorf("claim_id is required")
	}

	if cmd.Reason == "" {
		return fmt.Errorf("reason is required for cancellation")
	}

	// Get claim
	claim, err := h.claimRepo.GetByClaimID(ctx, cmd.ClaimID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get claim: %s", cmd.ClaimID)
		return fmt.Errorf("failed to get claim: %w", err)
	}

	// Cancel claim (business logic)
	if err := claim.Cancel(cmd.Reason); err != nil {
		h.logger.WithError(err).Errorf("Failed to cancel claim: %s", cmd.ClaimID)
		return fmt.Errorf("failed to cancel claim: %w", err)
	}

	// Update claim in repository
	if err := h.claimRepo.Update(ctx, claim); err != nil {
		h.logger.WithError(err).Errorf("Failed to update claim after cancellation: %s", cmd.ClaimID)
		return fmt.Errorf("failed to update claim: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id": cmd.ClaimID,
		"reason":   cmd.Reason,
	}).Info("Claim cancelled successfully")

	return nil
}
