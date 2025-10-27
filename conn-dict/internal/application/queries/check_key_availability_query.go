package queries

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// CheckKeyAvailabilityQuery represents a query to check if a PIX key is available
type CheckKeyAvailabilityQuery struct {
	Key string
}

// CheckKeyAvailabilityHandler handles the CheckKeyAvailabilityQuery
type CheckKeyAvailabilityHandler struct {
	entryRepo *repositories.EntryRepository
	logger    *logrus.Logger
}

// NewCheckKeyAvailabilityHandler creates a new CheckKeyAvailabilityHandler
func NewCheckKeyAvailabilityHandler(entryRepo *repositories.EntryRepository, logger *logrus.Logger) *CheckKeyAvailabilityHandler {
	return &CheckKeyAvailabilityHandler{
		entryRepo: entryRepo,
		logger:    logger,
	}
}

// Handle executes the CheckKeyAvailabilityQuery
// Returns true if key is available (not in use), false if already taken
func (h *CheckKeyAvailabilityHandler) Handle(ctx context.Context, query CheckKeyAvailabilityQuery) (bool, error) {
	if query.Key == "" {
		return false, fmt.Errorf("key is required")
	}

	hasActiveKey, err := h.entryRepo.HasActiveKey(ctx, query.Key)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to check key availability: %s", query.Key)
		return false, fmt.Errorf("failed to check key availability: %w", err)
	}

	// If hasActiveKey is true, key is NOT available
	// If hasActiveKey is false, key IS available
	return !hasActiveKey, nil
}
