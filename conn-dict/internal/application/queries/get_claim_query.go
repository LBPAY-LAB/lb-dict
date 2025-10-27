package queries

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// GetClaimQuery represents a query to get a claim by ID
type GetClaimQuery struct {
	ClaimID string
}

// GetClaimHandler handles the GetClaimQuery
type GetClaimHandler struct {
	claimRepo *repositories.ClaimRepository
	logger    *logrus.Logger
}

// NewGetClaimHandler creates a new GetClaimHandler
func NewGetClaimHandler(claimRepo *repositories.ClaimRepository, logger *logrus.Logger) *GetClaimHandler {
	return &GetClaimHandler{
		claimRepo: claimRepo,
		logger:    logger,
	}
}

// Handle executes the GetClaimQuery
func (h *GetClaimHandler) Handle(ctx context.Context, query GetClaimQuery) (*entities.Claim, error) {
	if query.ClaimID == "" {
		return nil, fmt.Errorf("claim_id is required")
	}

	claim, err := h.claimRepo.GetByClaimID(ctx, query.ClaimID)
	if err != nil {
		h.logger.WithError(err).Errorf("Failed to get claim: %s", query.ClaimID)
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}

	return claim, nil
}
