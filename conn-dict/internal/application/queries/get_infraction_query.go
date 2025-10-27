package queries

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// GetInfractionQuery represents a query to get an infraction by ID
type GetInfractionQuery struct {
	InfractionID string
}

// GetInfractionHandler handles the GetInfractionQuery
type GetInfractionHandler struct {
	infractionRepo *repositories.InfractionRepository
	logger         *logrus.Logger
}

// NewGetInfractionHandler creates a new GetInfractionHandler
func NewGetInfractionHandler(infractionRepo *repositories.InfractionRepository, logger *logrus.Logger) *GetInfractionHandler {
	return &GetInfractionHandler{
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// Handle executes the GetInfractionQuery
func (h *GetInfractionHandler) Handle(ctx context.Context, query GetInfractionQuery) (*entities.Infraction, error) {
	if query.InfractionID == "" {
		return nil, fmt.Errorf("infraction_id is required")
	}

	infraction, err := h.infractionRepo.GetByInfractionID(ctx, query.InfractionID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get infraction: %s", query.InfractionID)
		return nil, fmt.Errorf("failed to get infraction: %w", err)
	}

	return infraction, nil
}
