package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// ResolveInfractionCommand represents the command to resolve an infraction
type ResolveInfractionCommand struct {
	InfractionID    string
	ResolutionNotes string
}

// ResolveInfractionHandler handles the ResolveInfractionCommand
type ResolveInfractionHandler struct {
	infractionRepo *repositories.InfractionRepository
	logger         *logrus.Logger
}

// NewResolveInfractionHandler creates a new ResolveInfractionHandler
func NewResolveInfractionHandler(infractionRepo *repositories.InfractionRepository, logger *logrus.Logger) *ResolveInfractionHandler {
	return &ResolveInfractionHandler{
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// Handle executes the ResolveInfractionCommand
func (h *ResolveInfractionHandler) Handle(ctx context.Context, cmd ResolveInfractionCommand) error {
	if cmd.InfractionID == "" {
		return fmt.Errorf("infraction_id is required")
	}

	if cmd.ResolutionNotes == "" {
		return fmt.Errorf("resolution_notes are required")
	}

	infraction, err := h.infractionRepo.GetByInfractionID(ctx, cmd.InfractionID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	if err := infraction.Resolve(cmd.ResolutionNotes); err != nil {
		h.logger.WithError(err).Errorf("Failed to resolve infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to resolve infraction: %w", err)
	}

	if err := h.infractionRepo.Update(ctx, infraction); err != nil {
		h.logger.WithError(err).Errorf("Failed to update infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	h.logger.WithField("infraction_id", cmd.InfractionID).Info("Infraction resolved successfully")
	return nil
}
