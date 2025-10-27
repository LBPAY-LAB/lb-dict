package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// CompleteClaimCommand represents the command to complete a claim
type CompleteClaimCommand struct {
	ClaimID string
}

// CompleteClaimHandler handles the CompleteClaimCommand
type CompleteClaimHandler struct {
	claimRepo *repositories.ClaimRepository
	logger    *logrus.Logger
}

// NewCompleteClaimHandler creates a new CompleteClaimHandler
func NewCompleteClaimHandler(claimRepo *repositories.ClaimRepository, logger *logrus.Logger) *CompleteClaimHandler {
	return &CompleteClaimHandler{
		claimRepo: claimRepo,
		logger:    logger,
	}
}

// Handle executes the CompleteClaimCommand
func (h *CompleteClaimHandler) Handle(ctx context.Context, cmd CompleteClaimCommand) error {
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

	// Complete claim (business logic)
	if err := claim.Complete(); err != nil {
		h.logger.WithError(err).Errorf("Failed to complete claim: %s", cmd.ClaimID)
		return fmt.Errorf("failed to complete claim: %w", err)
	}

	// Update claim in repository
	if err := h.claimRepo.Update(ctx, claim); err != nil {
		h.logger.WithError(err).Errorf("Failed to update claim after completion: %s", cmd.ClaimID)
		return fmt.Errorf("failed to update claim: %w", err)
	}

	h.logger.WithField("claim_id", cmd.ClaimID).Info("Claim completed successfully")
	return nil
}
