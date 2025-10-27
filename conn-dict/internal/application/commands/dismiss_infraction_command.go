package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// DismissInfractionCommand represents the command to dismiss an infraction
type DismissInfractionCommand struct {
	InfractionID    string
	DismissalNotes  string
}

// DismissInfractionHandler handles the DismissInfractionCommand
type DismissInfractionHandler struct {
	infractionRepo *repositories.InfractionRepository
	logger         *logrus.Logger
}

// NewDismissInfractionHandler creates a new DismissInfractionHandler
func NewDismissInfractionHandler(infractionRepo *repositories.InfractionRepository, logger *logrus.Logger) *DismissInfractionHandler {
	return &DismissInfractionHandler{
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// Handle executes the DismissInfractionCommand
func (h *DismissInfractionHandler) Handle(ctx context.Context, cmd DismissInfractionCommand) error {
	if cmd.InfractionID == "" {
		return fmt.Errorf("infraction_id is required")
	}

	if cmd.DismissalNotes == "" {
		return fmt.Errorf("dismissal_notes are required")
	}

	infraction, err := h.infractionRepo.GetByInfractionID(ctx, cmd.InfractionID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	if err := infraction.Dismiss(cmd.DismissalNotes); err != nil {
		h.logger.WithError(err).Errorf("Failed to dismiss infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to dismiss infraction: %w", err)
	}

	if err := h.infractionRepo.Update(ctx, infraction); err != nil {
		h.logger.WithError(err).Errorf("Failed to update infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	h.logger.WithField("infraction_id", cmd.InfractionID).Info("Infraction dismissed successfully")
	return nil
}
