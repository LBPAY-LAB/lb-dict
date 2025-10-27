package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// DeleteEntryCommand represents the command to delete an entry
type DeleteEntryCommand struct {
	EntryID string
}

// DeleteEntryHandler handles the DeleteEntryCommand
type DeleteEntryHandler struct {
	entryRepo *repositories.EntryRepository
	logger    *logrus.Logger
}

// NewDeleteEntryHandler creates a new DeleteEntryHandler
func NewDeleteEntryHandler(entryRepo *repositories.EntryRepository, logger *logrus.Logger) *DeleteEntryHandler {
	return &DeleteEntryHandler{
		entryRepo: entryRepo,
		logger:    logger,
	}
}

// Handle executes the DeleteEntryCommand (soft delete)
func (h *DeleteEntryHandler) Handle(ctx context.Context, cmd DeleteEntryCommand) error {
	// Validate input
	if cmd.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}

	// Soft delete entry
	if err := h.entryRepo.Delete(ctx, cmd.EntryID); err != nil {
		h.logger.WithError(err).Errorf("Failed to delete entry: %s", cmd.EntryID)
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	h.logger.WithField("entry_id", cmd.EntryID).Info("Entry deleted successfully")
	return nil
}
