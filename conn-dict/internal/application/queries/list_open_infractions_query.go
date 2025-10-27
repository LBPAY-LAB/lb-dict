package queries

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// ListOpenInfractionsQuery represents a query to list open infractions needing investigation
type ListOpenInfractionsQuery struct {
	Limit int
}

// ListOpenInfractionsHandler handles the ListOpenInfractionsQuery
type ListOpenInfractionsHandler struct {
	infractionRepo *repositories.InfractionRepository
	logger         *logrus.Logger
}

// NewListOpenInfractionsHandler creates a new ListOpenInfractionsHandler
func NewListOpenInfractionsHandler(infractionRepo *repositories.InfractionRepository, logger *logrus.Logger) *ListOpenInfractionsHandler {
	return &ListOpenInfractionsHandler{
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// Handle executes the ListOpenInfractionsQuery
func (h *ListOpenInfractionsHandler) Handle(ctx context.Context, query ListOpenInfractionsQuery) ([]*entities.Infraction, error) {
	if query.Limit <= 0 {
		query.Limit = 50 // Default limit for investigation queue
	}

	infractions, err := h.infractionRepo.ListOpen(ctx, query.Limit)
	if err != nil {
		h.logger.WithError(err).Error("Failed to list open infractions")
		return nil, fmt.Errorf("failed to list open infractions: %w", err)
	}

	return infractions, nil
}
