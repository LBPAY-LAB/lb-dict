package queries

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// ListInfractionsQuery represents a query to list infractions
type ListInfractionsQuery struct {
	Key                 string // Optional
	ReporterParticipant string // Optional (one of Key or ReporterParticipant must be provided)
	Limit               int
	Offset              int
}

// ListInfractionsHandler handles the ListInfractionsQuery
type ListInfractionsHandler struct {
	infractionRepo *repositories.InfractionRepository
	logger         *logrus.Logger
}

// NewListInfractionsHandler creates a new ListInfractionsHandler
func NewListInfractionsHandler(infractionRepo *repositories.InfractionRepository, logger *logrus.Logger) *ListInfractionsHandler {
	return &ListInfractionsHandler{
		infractionRepo: infractionRepo,
		logger:         logger,
	}
}

// Handle executes the ListInfractionsQuery
func (h *ListInfractionsHandler) Handle(ctx context.Context, query ListInfractionsQuery) ([]*entities.Infraction, error) {
	if query.Key == "" && query.ReporterParticipant == "" {
		return nil, fmt.Errorf("either key or reporter_participant must be provided")
	}

	if query.Limit <= 0 {
		query.Limit = 100
	}

	if query.Offset < 0 {
		query.Offset = 0
	}

	var infractions []*entities.Infraction
	var err error

	if query.Key != "" {
		infractions, err = h.infractionRepo.ListByKey(ctx, query.Key, query.Limit, query.Offset)
	} else {
		infractions, err = h.infractionRepo.ListByReporter(ctx, query.ReporterParticipant, query.Limit, query.Offset)
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to list infractions")
		return nil, fmt.Errorf("failed to list infractions: %w", err)
	}

	return infractions, nil
}
