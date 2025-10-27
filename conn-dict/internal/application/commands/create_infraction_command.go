package commands

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// CreateInfractionCommand represents the command to create a new infraction
type CreateInfractionCommand struct {
	InfractionID        string
	EntryID             string
	ClaimID             string
	Key                 string
	Type                string
	Description         string
	EvidenceURLs        []string
	ReporterParticipant string
	ReportedParticipant string
}

// CreateInfractionHandler handles the CreateInfractionCommand
type CreateInfractionHandler struct {
	infractionRepo *repositories.InfractionRepository
	logger         *logrus.Logger
}

// NewCreateInfractionHandler creates a new CreateInfractionHandler
func NewCreateInfractionHandler(infractionRepo *repositories.InfractionRepository, logger *logrus.Logger) *CreateInfractionHandler {
	return &CreateInfractionHandler{
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// Handle executes the CreateInfractionCommand
func (h *CreateInfractionHandler) Handle(ctx context.Context, cmd CreateInfractionCommand) error {
	// Validate input
	if cmd.InfractionID == "" {
		return fmt.Errorf("infraction_id is required")
	}

	if cmd.Key == "" {
		return fmt.Errorf("key is required")
	}

	if cmd.Description == "" {
		return fmt.Errorf("description is required")
	}

	if cmd.ReporterParticipant == "" {
		return fmt.Errorf("reporter_participant is required")
	}

	// Convert type string to InfractionType
	var infractionType entities.InfractionType
	switch cmd.Type {
	case "FRAUD":
		infractionType = entities.InfractionTypeFraud
	case "ACCOUNT_CLOSED":
		infractionType = entities.InfractionTypeAccountClosed
	case "INCORRECT_DATA":
		infractionType = entities.InfractionTypeIncorrectData
	case "UNAUTHORIZED_USE":
		infractionType = entities.InfractionTypeUnauthorizedUse
	case "DUPLICATE_KEY":
		infractionType = entities.InfractionTypeDuplicateKey
	case "OTHER":
		infractionType = entities.InfractionTypeOther
	default:
		return fmt.Errorf("invalid infraction type: %s", cmd.Type)
	}

	// Create infraction entity
	infraction, err := entities.NewInfraction(
		cmd.InfractionID,
		cmd.Key,
		infractionType,
		cmd.Description,
		cmd.ReporterParticipant,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create infraction entity")
		return fmt.Errorf("invalid infraction data: %w", err)
	}

	// Set optional fields
	if cmd.EntryID != "" {
		infraction.EntryID = &cmd.EntryID
	}

	if cmd.ClaimID != "" {
		infraction.ClaimID = &cmd.ClaimID
	}

	if cmd.ReportedParticipant != "" {
		infraction.ReportedParticipant = &cmd.ReportedParticipant
	}

	if len(cmd.EvidenceURLs) > 0 {
		infraction.EvidenceURLs = cmd.EvidenceURLs
	}

	// Persist to repository
	if err := h.infractionRepo.Create(ctx, infraction); err != nil {
		h.logger.WithError(err).Errorf("Failed to create infraction: %s", cmd.InfractionID)
		return fmt.Errorf("failed to create infraction: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"infraction_id": cmd.InfractionID,
		"key":           cmd.Key,
		"type":          cmd.Type,
	}).Info("Infraction created successfully")

	return nil
}
