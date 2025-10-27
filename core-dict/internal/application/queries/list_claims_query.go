package queries

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/lbpay-lab/core-dict/internal/application/services"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
)

// ListClaimsQuery representa a query para listar claims com paginação
type ListClaimsQuery struct {
	ISPB     string // ISPB do participante (donor ou claimer)
	Page     int    // 1-indexed
	PageSize int    // default: 100, max: 1000
}

// ListClaimsResult representa o resultado paginado
type ListClaimsResult struct {
	Claims     []*entities.Claim `json:"claims"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// ListClaimsQueryHandler lida com a query ListClaims
type ListClaimsQueryHandler struct {
	claimRepo repositories.ClaimRepository
	cache     services.CacheService
}

// NewListClaimsQueryHandler cria um novo handler para ListClaims
func NewListClaimsQueryHandler(
	claimRepo repositories.ClaimRepository,
	cache services.CacheService,
) *ListClaimsQueryHandler {
	return &ListClaimsQueryHandler{
		claimRepo: claimRepo,
		cache:     cache,
	}
}

// Handle executa a query ListClaims com paginação
func (h *ListClaimsQueryHandler) Handle(ctx context.Context, query ListClaimsQuery) (*ListClaimsResult, error) {
	// Validação
	if query.ISPB == "" {
		return nil, fmt.Errorf("ispb is required")
	}

	// Default e limites de paginação
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 100
	}
	if query.PageSize > 1000 {
		query.PageSize = 1000
	}

	// Calcular offset
	offset := (query.Page - 1) * query.PageSize

	// 1. Try cache first
	cacheKey := fmt.Sprintf("claims:ispb:%s:page:%d:size:%d", query.ISPB, query.Page, query.PageSize)
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if result, ok := cachedData.(*ListClaimsResult); ok {
			return result, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var result ListClaimsResult
			if err := json.Unmarshal([]byte(jsonData), &result); err == nil {
				return &result, nil
			}
		}
	}

	// 2. Cache miss - query database
	filters := repositories.ClaimFilters{
		DonorISPB: &query.ISPB,
		Limit:     query.PageSize,
		Offset:    offset,
	}
	claims, err := h.claimRepo.List(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to list claims: %w", err)
	}

	// 3. Get total count for pagination metadata
	totalCount, err := h.claimRepo.Count(ctx, filters)
	if err != nil {
		return nil, fmt.Errorf("failed to count claims: %w", err)
	}

	// 4. Calculate total pages
	totalPages := int(totalCount) / query.PageSize
	if int(totalCount)%query.PageSize > 0 {
		totalPages++
	}

	result := &ListClaimsResult{
		Claims:     claims,
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}

	// 5. Store in cache (TTL: 1 minute - claims são frequentemente atualizados)
	if err := h.cache.Set(ctx, cacheKey, result, 1*time.Minute); err != nil {
		_ = err
	}

	return result, nil
}

// InvalidateCache invalida o cache de listagem de claims de um participante
func (h *ListClaimsQueryHandler) InvalidateCache(ctx context.Context, ispb string) error {
	pattern := fmt.Sprintf("claims:ispb:%s:*", ispb)
	return h.cache.Invalidate(ctx, pattern)
}
