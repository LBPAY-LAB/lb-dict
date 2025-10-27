package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// InvestigateInfractionCommand represents the command to start investigating an infraction
type InvestigateInfractionCommand struct {
	InfractionID string
}

// InvestigateInfractionHandler handles the InvestigateInfractionCommand
type InvestigateInfractionHandler struct {
	infractionRepo *repositories.InfractionRepository
	logger         *logrus.Logger
}

// NewInvestigateInfractionHandler creates a new InvestigateInfractionHandler
func NewInvestigateInfractionHandler(infractionRepo *repositories.InfractionRepository, logger *logrus.Logger) *InvestigateInfractionHandler {
	return &InvestigateInfractionHandler{
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// Handle executes the InvestigateInfractionCommand
func (h *InvestigateInfractionHandler) Handle(ctx context.Context, cmd InvestigateInfractionCommand) error {
	if cmd.InfractionID == "" {
		return fmt.Errorf("infraction_id is required")
	}

	infraction, err := h.infractionRepo.GetByInfractionID(ctx, cmd.InfractionID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	if err := infraction.Investigate(); err != nil {
		h.logger.WithError(err).Errorf("Failed to investigate infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to investigate infraction: %w", err)
	}

	if err := h.infractionRepo.Update(ctx, infraction); err != nil {
		h.logger.WithError(err).Errorf("Failed to update infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	h.logger.WithField("infraction_id", cmd.InfractionID).Info("Infraction marked as under investigation")
	return nil
}
