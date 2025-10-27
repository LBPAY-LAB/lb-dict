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

// ListEntriesQuery representa a query para listar chaves PIX com paginação
type ListEntriesQuery struct {
	AccountID uuid.UUID
	Page      int // 1-indexed
	PageSize  int // default: 100, max: 1000
}

// ListEntriesResult representa o resultado paginado
type ListEntriesResult struct {
	Entries    []*entities.Entry `json:"entries"`
	TotalCount int64             `json:"total_count"`
	Page       int               `json:"page"`
	PageSize   int               `json:"page_size"`
	TotalPages int               `json:"total_pages"`
}

// ListEntriesQueryHandler lida com a query ListEntries
type ListEntriesQueryHandler struct {
	entryRepo repositories.EntryRepository
	cache     services.CacheService
}

// NewListEntriesQueryHandler cria um novo handler para ListEntries
func NewListEntriesQueryHandler(
	entryRepo repositories.EntryRepository,
	cache services.CacheService,
) *ListEntriesQueryHandler {
	return &ListEntriesQueryHandler{
		entryRepo: entryRepo,
		cache:     cache,
	}
}

// Handle executa a query ListEntries com paginação cursor-based
func (h *ListEntriesQueryHandler) Handle(ctx context.Context, query ListEntriesQuery) (*ListEntriesResult, error) {
	// Validação
	if query.AccountID == uuid.Nil {
		return nil, fmt.Errorf("account_id is required")
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
	cacheKey := fmt.Sprintf("entries:account:%s:page:%d:size:%d", query.AccountID.String(), query.Page, query.PageSize)
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if result, ok := cachedData.(*ListEntriesResult); ok {
			return result, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var result ListEntriesResult
			if err := json.Unmarshal([]byte(jsonData), &result); err == nil {
				return &result, nil
			}
		}
	}

	// 2. Cache miss - query database
	entries, err := h.entryRepo.List(ctx, query.AccountID, query.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}

	// 3. Get total count for pagination metadata
	totalCount, err := h.entryRepo.CountByAccount(ctx, query.AccountID)
	if err != nil {
		return nil, fmt.Errorf("failed to count entries: %w", err)
	}

	// 4. Calculate total pages
	totalPages := int(totalCount) / query.PageSize
	if int(totalCount)%query.PageSize > 0 {
		totalPages++
	}

	result := &ListEntriesResult{
		Entries:    entries,
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}

	// 5. Store in cache (TTL: 2 minutes - shorter for lists)
	if err := h.cache.Set(ctx, cacheKey, result, 2*time.Minute); err != nil {
		// Log error but don't fail
		_ = err
	}

	return result, nil
}

// InvalidateCache invalida o cache de listagem de uma conta
func (h *ListEntriesQueryHandler) InvalidateCache(ctx context.Context, accountID uuid.UUID) error {
	pattern := fmt.Sprintf("entries:account:%s:*", accountID.String())
	return h.cache.Invalidate(ctx, pattern)
}
