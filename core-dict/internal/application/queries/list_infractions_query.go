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

// ListInfractionsQuery representa a query para listar infrações com paginação
type ListInfractionsQuery struct {
	ISPB     string // ISPB do participante infrator
	Page     int    // 1-indexed
	PageSize int    // default: 100, max: 1000
}

// ListInfractionsResult representa o resultado paginado
type ListInfractionsResult struct {
	Infractions []*entities.Infraction `json:"infractions"`
	TotalCount  int64                  `json:"total_count"`
	Page        int                    `json:"page"`
	PageSize    int                    `json:"page_size"`
	TotalPages  int                    `json:"total_pages"`
}

// ListInfractionsQueryHandler lida com a query ListInfractions
type ListInfractionsQueryHandler struct {
	infractionRepo repositories.InfractionRepository
	cache          services.CacheService
}

// NewListInfractionsQueryHandler cria um novo handler para ListInfractions
func NewListInfractionsQueryHandler(
	infractionRepo repositories.InfractionRepository,
	cache services.CacheService,
) *ListInfractionsQueryHandler {
	return &ListInfractionsQueryHandler{
		infractionRepo: infractionRepo,
		cache:          cache,
	}
}

// Handle executa a query ListInfractions com paginação
func (h *ListInfractionsQueryHandler) Handle(ctx context.Context, query ListInfractionsQuery) (*ListInfractionsResult, error) {
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
	cacheKey := fmt.Sprintf("infractions:ispb:%s:page:%d:size:%d", query.ISPB, query.Page, query.PageSize)
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if result, ok := cachedData.(*ListInfractionsResult); ok {
			return result, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var result ListInfractionsResult
			if err := json.Unmarshal([]byte(jsonData), &result); err == nil {
				return &result, nil
			}
		}
	}

	// 2. Cache miss - query database
	infractions, err := h.infractionRepo.List(ctx, query.ISPB, query.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list infractions: %w", err)
	}

	// 3. Get total count for pagination metadata
	totalCount, err := h.infractionRepo.CountByISPB(ctx, query.ISPB)
	if err != nil {
		return nil, fmt.Errorf("failed to count infractions: %w", err)
	}

	// 4. Calculate total pages
	totalPages := int(totalCount) / query.PageSize
	if int(totalCount)%query.PageSize > 0 {
		totalPages++
	}

	result := &ListInfractionsResult{
		Infractions: infractions,
		TotalCount:  totalCount,
		Page:        query.Page,
		PageSize:    query.PageSize,
		TotalPages:  totalPages,
	}

	// 5. Store in cache (TTL: 10 minutes - infrações raramente mudam)
	if err := h.cache.Set(ctx, cacheKey, result, 10*time.Minute); err != nil {
		_ = err
	}

	return result, nil
}

// InvalidateCache invalida o cache de listagem de infrações de um participante
func (h *ListInfractionsQueryHandler) InvalidateCache(ctx context.Context, ispb string) error {
	pattern := fmt.Sprintf("infractions:ispb:%s:*", ispb)
	return h.cache.Invalidate(ctx, pattern)
}
