package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/application/services"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
)

// GetClaimQuery representa a query para buscar um claim por ID
type GetClaimQuery struct {
	ClaimID uuid.UUID
}

// GetClaimQueryHandler lida com a query GetClaim
type GetClaimQueryHandler struct {
	claimRepo repositories.ClaimRepository
	cache     services.CacheService
}

// NewGetClaimQueryHandler cria um novo handler para GetClaim
func NewGetClaimQueryHandler(
	claimRepo repositories.ClaimRepository,
	cache services.CacheService,
) *GetClaimQueryHandler {
	return &GetClaimQueryHandler{
		claimRepo: claimRepo,
		cache:     cache,
	}
}

// Handle executa a query GetClaim
func (h *GetClaimQueryHandler) Handle(ctx context.Context, query GetClaimQuery) (*entities.Claim, error) {
	// Validação
	if query.ClaimID == uuid.Nil {
		return nil, fmt.Errorf("claim_id is required")
	}

	// 1. Try cache first (Cache-Aside pattern)
	cacheKey := fmt.Sprintf("claim:%s", query.ClaimID.String())
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if claim, ok := cachedData.(*entities.Claim); ok {
			return claim, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var claim entities.Claim
			if err := json.Unmarshal([]byte(jsonData), &claim); err == nil {
				return &claim, nil
			}
		}
	}

	// 2. Cache miss - query database
	claim, err := h.claimRepo.FindByID(ctx, query.ClaimID)
	if err != nil {
		return nil, fmt.Errorf("claim not found: %w", err)
	}

	// 3. Store in cache (TTL: 3 minutes - claims são mutáveis)
	if err := h.cache.Set(ctx, cacheKey, claim, 3*time.Minute); err != nil {
		_ = err
	}

	// 4. Return result
	return claim, nil
}

// InvalidateCache invalida o cache de um claim específico
func (h *GetClaimQueryHandler) InvalidateCache(ctx context.Context, claimID uuid.UUID) error {
	cacheKey := fmt.Sprintf("claim:%s", claimID.String())
	return h.cache.Delete(ctx, cacheKey)
}
