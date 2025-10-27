package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// CreateClaimCommand represents the command to create a new claim
type CreateClaimCommand struct {
	ClaimID                string
	Type                   entities.ClaimType
	Key                    string
	KeyType                string
	DonorParticipant       string // ISPB (8 digits)
	ClaimerParticipant     string // ISPB (8 digits)
	ClaimerAccountBranch   string
	ClaimerAccountNumber   string
	ClaimerAccountType     string
}

// Validate validates the CreateClaimCommand input
func (c *CreateClaimCommand) Validate() error {
	if c.ClaimID == "" {
		return errors.New("claim_id is required")
	}
	if c.Key == "" {
		return errors.New("key is required")
	}
	if c.KeyType == "" {
		return errors.New("key_type is required")
	}
	if len(c.DonorParticipant) != 8 {
		return errors.New("donor_participant must be 8 digits (ISPB)")
	}
	if len(c.ClaimerParticipant) != 8 {
		return errors.New("claimer_participant must be 8 digits (ISPB)")
	}
	if c.DonorParticipant == c.ClaimerParticipant {
		return errors.New("donor and claimer must be different institutions")
	}
	if c.Type != entities.ClaimTypePortability && c.Type != entities.ClaimTypeOwnership {
		return errors.New("claim type must be PORTABILITY or OWNERSHIP")
	}

	// Validate account information for claimer
	if c.ClaimerAccountNumber == "" {
		return errors.New("claimer_account_number is required")
	}
	if c.ClaimerAccountType == "" {
		return errors.New("claimer_account_type is required")
	}

	return nil
}

// CreateClaimHandler handles the CreateClaimCommand
type CreateClaimHandler struct {
	repo   *repositories.ClaimRepository
	logger *logrus.Logger
}

// NewCreateClaimHandler creates a new CreateClaimHandler
func NewCreateClaimHandler(
	repo *repositories.ClaimRepository,
	logger *logrus.Logger,
) *CreateClaimHandler {
	return &CreateClaimHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the CreateClaimCommand
func (h *CreateClaimHandler) Handle(ctx context.Context, cmd *CreateClaimCommand) (*entities.Claim, error) {
	h.logger.WithFields(logrus.Fields{
		"claim_id":            cmd.ClaimID,
		"key":                 cmd.Key,
		"key_type":            cmd.KeyType,
		"type":                cmd.Type,
		"donor_participant":   cmd.DonorParticipant,
		"claimer_participant": cmd.ClaimerParticipant,
	}).Info("Handling CreateClaimCommand")

	// Validate command
	if err := cmd.Validate(); err != nil {
		h.logger.WithError(err).Error("CreateClaimCommand validation failed")
		return nil, fmt.Errorf("invalid command: %w", err)
	}

	// Check if there's already an active claim for this key
	hasActive, err := h.repo.HasActiveClaim(ctx, cmd.Key)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check for active claims")
		return nil, fmt.Errorf("failed to check active claims: %w", err)
	}
	if hasActive {
		h.logger.WithField("key", cmd.Key).Warn("Key already has an active claim")
		return nil, errors.New("key already has an active claim")
	}

	// Create claim entity using domain factory
	claim, err := entities.NewClaim(
		cmd.ClaimID,
		cmd.Type,
		cmd.Key,
		cmd.KeyType,
		cmd.DonorParticipant,
		cmd.ClaimerParticipant,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create claim entity")
		return nil, fmt.Errorf("failed to create claim entity: %w", err)
	}

	// Set account information
	claim.ClaimerAccountBranch = cmd.ClaimerAccountBranch
	claim.ClaimerAccountNumber = cmd.ClaimerAccountNumber
	claim.ClaimerAccountType = cmd.ClaimerAccountType

	// Persist claim to repository
	if err := h.repo.Create(ctx, claim); err != nil {
		h.logger.WithError(err).Error("Failed to persist claim")
		return nil, fmt.Errorf("failed to persist claim: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"claim_id": claim.ClaimID,
		"id":       claim.ID,
		"status":   claim.Status,
	}).Info("Claim created successfully")

	return claim, nil
}