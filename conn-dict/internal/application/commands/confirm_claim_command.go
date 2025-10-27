package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// ConfirmClaimCommand represents the command to confirm a claim
type ConfirmClaimCommand struct {
	ClaimID string
}

// ConfirmClaimHandler handles the ConfirmClaimCommand
type ConfirmClaimHandler struct {
	claimRepo *repositories.ClaimRepository
	logger    *logrus.Logger
}

// NewConfirmClaimHandler creates a new ConfirmClaimHandler
func NewConfirmClaimHandler(claimRepo *repositories.ClaimRepository, logger *logrus.Logger) *ConfirmClaimHandler {
	return &ConfirmClaimHandler{
		claimRepo: claimRepo,
		logger:    logger,
	}
}

// Handle executes the ConfirmClaimCommand
func (h *ConfirmClaimHandler) Handle(ctx context.Context, cmd ConfirmClaimCommand) error {
	// Validate input
	if cmd.ClaimID == "" {
		return fmt.Errorf("claim_id is required")
	}

	// Get claim
	claim, err := h.claimRepo.GetByClaimID(ctx, cmd.ClaimID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get claim: %s", cmd.ClaimID)
		return fmt.Errorf("failed to get claim: %w", err)
	}

	// Confirm claim (business logic)
	if err := claim.Confirm(); err != nil {
		h.logger.WithError(err).Errorf("Failed to confirm claim: %s", cmd.ClaimID)
		return fmt.Errorf("failed to confirm claim: %w", err)
	}

	// Update claim in repository
	if err := h.claimRepo.Update(ctx, claim); err != nil {
		h.logger.WithError(err).Errorf("Failed to update claim after confirmation: %s", cmd.ClaimID)
		return fmt.Errorf("failed to update claim: %w", err)
	}

	h.logger.WithField("claim_id", cmd.ClaimID).Info("Claim confirmed successfully")
	return nil
}
