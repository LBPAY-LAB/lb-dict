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

// GetEntryQuery representa a query para buscar uma chave PIX
type GetEntryQuery struct {
	KeyValue string
}

// GetEntryQueryHandler lida com a query GetEntry
type GetEntryQueryHandler struct {
	entryRepo     repositories.EntryRepository
	cache         services.CacheService
	connectClient services.ConnectClient // NEW: Optional fallback to RSFN
}

// NewGetEntryQueryHandler cria um novo handler para GetEntry
func NewGetEntryQueryHandler(
	entryRepo repositories.EntryRepository,
	cache services.CacheService,
	connectClient services.ConnectClient, // Optional: can be nil
) *GetEntryQueryHandler {
	return &GetEntryQueryHandler{
		entryRepo:     entryRepo,
		cache:         cache,
		connectClient: connectClient,
	}
}

// Handle executa a query GetEntry seguindo o padrão Cache-Aside
func (h *GetEntryQueryHandler) Handle(ctx context.Context, query GetEntryQuery) (*entities.Entry, error) {
	// Validação
	if query.KeyValue == "" {
		return nil, fmt.Errorf("key_value is required")
	}

	// 1. Try cache first (Cache-Aside pattern)
	cacheKey := fmt.Sprintf("entry:%s", query.KeyValue)
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit - deserialize
		if entry, ok := cachedData.(*entities.Entry); ok {
			return entry, nil
		}
		// Try to unmarshal if it's a string/bytes
		if jsonData, ok := cachedData.(string); ok {
			var entry entities.Entry
			if err := json.Unmarshal([]byte(jsonData), &entry); err == nil {
				return &entry, nil
			}
		}
	}

	// 2. Cache miss - query database
	entry, err := h.entryRepo.FindByKey(ctx, query.KeyValue)
	if err == nil {
		// Found in database - cache and return
		if err := h.cache.Set(ctx, cacheKey, entry, 5*time.Minute); err != nil {
			_ = err // Log error but don't fail
		}
		return entry, nil
	}

	// 3. Database miss - try Connect service as fallback (optional)
	// This allows querying RSFN DICT directly for keys owned by other ISPBs
	if h.connectClient != nil {
		// TODO: Connect client would need a method to query by key value
		// For now, just return the database error
		// Future: rsfnEntry, err := h.connectClient.GetEntryByKey(ctx, query.KeyValue)
		// if err == nil { return mapRSFNEntryToDomain(rsfnEntry), nil }
	}

	// 4. Not found anywhere - return error
	return nil, fmt.Errorf("entry not found: %w", err)
}

// InvalidateCache invalida o cache de uma chave específica
func (h *GetEntryQueryHandler) InvalidateCache(ctx context.Context, keyValue string) error {
	cacheKey := fmt.Sprintf("entry:%s", keyValue)
	return h.cache.Delete(ctx, cacheKey)
}
