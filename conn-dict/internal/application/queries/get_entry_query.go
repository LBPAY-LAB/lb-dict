package queries

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// GetEntryQuery represents a query to get an entry by ID or Key
type GetEntryQuery struct {
	EntryID string // Optional
	Key     string // Optional (one of EntryID or Key must be provided)
}

// GetEntryHandler handles the GetEntryQuery
type GetEntryHandler struct {
	entryRepo *repositories.EntryRepository
	logger    *logrus.Logger
}

// NewGetEntryHandler creates a new GetEntryHandler
func NewGetEntryHandler(entryRepo *repositories.EntryRepository, logger *logrus.Logger) *GetEntryHandler {
	return &GetEntryHandler{
		entryRepo: entryRepo,
		logger:    logger,
	}
}

// Handle executes the GetEntryQuery
func (h *GetEntryHandler) Handle(ctx context.Context, query GetEntryQuery) (*entities.Entry, error) {
	if query.EntryID == "" && query.Key == "" {
		return nil, fmt.Errorf("either entry_id or key must be provided")
	}

	var entry *entities.Entry
	var err error

	if query.EntryID != "" {
		entry, err = h.entryRepo.GetByEntryID(ctx, query.EntryID)
	} else {
		entry, err = h.entryRepo.GetByKey(ctx, query.Key)
	}

	if err != nil {
		h.logger.WithError(err).Error("Failed to get entry")
		return nil, fmt.Errorf("failed to get entry: %w", err)
	}

	return entry, nil
}
