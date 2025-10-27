package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// DeactivateEntryCommand represents the command to deactivate an entry
type DeactivateEntryCommand struct {
	EntryID string
	Reason  string
}

// DeactivateEntryHandler handles the DeactivateEntryCommand
type DeactivateEntryHandler struct {
	entryRepo *repositories.EntryRepository
	logger    *logrus.Logger
}

// NewDeactivateEntryHandler creates a new DeactivateEntryHandler
func NewDeactivateEntryHandler(entryRepo *repositories.EntryRepository, logger *logrus.Logger) *DeactivateEntryHandler {
	return &DeactivateEntryHandler{
		entryRepo: entryRepo,
		logger:    logger,
	}
}

// Handle executes the DeactivateEntryCommand
func (h *DeactivateEntryHandler) Handle(ctx context.Context, cmd DeactivateEntryCommand) error {
	// Validate input
	if cmd.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}

	if cmd.Reason == "" {
		return fmt.Errorf("reason is required for deactivation")
	}

	// Get entry
	entry, err := h.entryRepo.GetByEntryID(ctx, cmd.EntryID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get entry: %s", cmd.EntryID)
		return fmt.Errorf("failed to get entry: %w", err)
	}

	// Deactivate entry (business logic)
	if err := entry.Deactivate(cmd.Reason); err != nil {
		h.logger.WithError(err).Errorf("Failed to deactivate entry: %s", cmd.EntryID)
		return fmt.Errorf("failed to deactivate entry: %w", err)
	}

	// Update entry in repository
	if err := h.entryRepo.Update(ctx, entry); err != nil {
		h.logger.WithError(err).Errorf("Failed to update entry after deactivation: %s", cmd.EntryID)
		return fmt.Errorf("failed to update entry: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"entry_id": cmd.EntryID,
		"reason":   cmd.Reason,
	}).Info("Entry deactivated successfully")

	return nil
}
