package queries

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// ListClaimsQuery represents a query to list claims by key
type ListClaimsQuery struct {
	Key    string
	Limit  int
	Offset int
}

// ListClaimsHandler handles the ListClaimsQuery
type ListClaimsHandler struct {
	claimRepo *repositories.ClaimRepository
	logger    *logrus.Logger
}

// NewListClaimsHandler creates a new ListClaimsHandler
func NewListClaimsHandler(claimRepo *repositories.ClaimRepository, logger *logrus.Logger) *ListClaimsHandler {
	return &ListClaimsHandler{
		claimRepo: claimRepo,
		logger:    logger,
	}
}

// Handle executes the ListClaimsQuery
func (h *ListClaimsHandler) Handle(ctx context.Context, query ListClaimsQuery) ([]*entities.Claim, error) {
	if query.Key == "" {
		return nil, fmt.Errorf("key is required")
	}

	if query.Limit <= 0 {
		query.Limit = 100
	}

	if query.Offset < 0 {
		query.Offset = 0
	}

	claims, err := h.claimRepo.ListByKey(ctx, query.Key)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to list claims for key: %s", query.Key)
		return nil, fmt.Errorf("failed to list claims: %w", err)
	}

	return claims, nil
}
