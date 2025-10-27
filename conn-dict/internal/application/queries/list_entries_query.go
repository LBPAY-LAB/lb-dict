package queries

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// ListEntriesQuery represents a query to list entries
type ListEntriesQuery struct {
	ParticipantISPB string
	Limit           int
	Offset          int
}

// ListEntriesHandler handles the ListEntriesQuery
type ListEntriesHandler struct {
	entryRepo *repositories.EntryRepository
	logger    *logrus.Logger
}

// NewListEntriesHandler creates a new ListEntriesHandler
func NewListEntriesHandler(entryRepo *repositories.EntryRepository, logger *logrus.Logger) *ListEntriesHandler {
	return &ListEntriesHandler{
		entryRepo: entryRepo,
		logger:    logger,
	}
}

// Handle executes the ListEntriesQuery
func (h *ListEntriesHandler) Handle(ctx context.Context, query ListEntriesQuery) ([]*entities.Entry, error) {
	if query.ParticipantISPB == "" {
		return nil, fmt.Errorf("participant_ispb is required")
	}

	if query.Limit <= 0 {
		query.Limit = 100 // Default limit
	}

	if query.Offset < 0 {
		query.Offset = 0
	}

	entries, err := h.entryRepo.ListByParticipant(ctx, query.ParticipantISPB, query.Limit, query.Offset)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to list entries for participant: %s", query.ParticipantISPB)
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}

	return entries, nil
}
