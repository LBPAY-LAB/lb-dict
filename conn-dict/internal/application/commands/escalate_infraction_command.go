package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// EscalateInfractionCommand represents the command to escalate an infraction to Bacen
type EscalateInfractionCommand struct {
	InfractionID    string
	EscalationNotes string
}

// EscalateInfractionHandler handles the EscalateInfractionCommand
type EscalateInfractionHandler struct {
	infractionRepo *repositories.InfractionRepository
	logger         *logrus.Logger
}

// NewEscalateInfractionHandler creates a new EscalateInfractionHandler
func NewEscalateInfractionHandler(infractionRepo *repositories.InfractionRepository, logger *logrus.Logger) *EscalateInfractionHandler {
	return &EscalateInfractionHandler{
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// Handle executes the EscalateInfractionCommand
func (h *EscalateInfractionHandler) Handle(ctx context.Context, cmd EscalateInfractionCommand) error {
	if cmd.InfractionID == "" {
		return fmt.Errorf("infraction_id is required")
	}

	if cmd.EscalationNotes == "" {
		return fmt.Errorf("escalation_notes are required")
	}

	infraction, err := h.infractionRepo.GetByInfractionID(ctx, cmd.InfractionID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to get infraction: %w", err)
	}

	if err := infraction.EscalateToBacen(cmd.EscalationNotes); err != nil {
		h.logger.WithError(err).Errorf("Failed to escalate infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to escalate infraction: %w", err)
	}

	if err := h.infractionRepo.Update(ctx, infraction); err != nil {
		h.logger.WithError(err).Errorf("Failed to update infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	h.logger.WithField("infraction_id", cmd.InfractionID).Info("Infraction escalated to Bacen successfully")
	return nil
}
