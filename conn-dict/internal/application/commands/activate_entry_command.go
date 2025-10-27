package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// ActivateEntryCommand represents the command to activate an entry
type ActivateEntryCommand struct {
	EntryID string
}

// ActivateEntryHandler handles the ActivateEntryCommand
type ActivateEntryHandler struct {
	entryRepo *repositories.EntryRepository
	logger    *logrus.Logger
}

// NewActivateEntryHandler creates a new ActivateEntryHandler
func NewActivateEntryHandler(entryRepo *repositories.EntryRepository, logger *logrus.Logger) *ActivateEntryHandler {
	return &ActivateEntryHandler{
		entryRepo: entryRepo,
		logger:    logger,
	}
}

// Handle executes the ActivateEntryCommand
func (h *ActivateEntryHandler) Handle(ctx context.Context, cmd ActivateEntryCommand) error {
	// Validate input
	if cmd.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}

	// Get entry
	entry, err := h.entryRepo.GetByEntryID(ctx, cmd.EntryID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get entry: %s", cmd.EntryID)
		return fmt.Errorf("failed to get entry: %w", err)
	}

	// Activate entry (business logic)
	if err := entry.Activate(); err != nil {
		h.logger.WithError(err).Errorf("Failed to activate entry: %s", cmd.EntryID)
		return fmt.Errorf("failed to activate entry: %w", err)
	}

	// Update entry in repository
	if err := h.entryRepo.Update(ctx, entry); err != nil {
		h.logger.WithError(err).Errorf("Failed to update entry after activation: %s", cmd.EntryID)
		return fmt.Errorf("failed to update entry: %w", err)
	}

	h.logger.WithField("entry_id", cmd.EntryID).Info("Entry activated successfully")
	return nil
}
