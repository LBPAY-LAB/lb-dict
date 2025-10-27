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

// GetStatisticsQuery representa a query para obter estatísticas agregadas
type GetStatisticsQuery struct {
	// Pode adicionar filtros no futuro (ex: período, ISPB específico)
}

// GetStatisticsQueryHandler lida com a query GetStatistics
type GetStatisticsQueryHandler struct {
	statsRepo repositories.StatisticsRepository
	cache     services.CacheService
}

// NewGetStatisticsQueryHandler cria um novo handler para GetStatistics
func NewGetStatisticsQueryHandler(
	statsRepo repositories.StatisticsRepository,
	cache services.CacheService,
) *GetStatisticsQueryHandler {
	return &GetStatisticsQueryHandler{
		statsRepo: statsRepo,
		cache:     cache,
	}
}

// Handle executa a query GetStatistics
// Estatísticas são SEMPRE cacheadas por 5 minutos devido ao custo computacional
func (h *GetStatisticsQueryHandler) Handle(ctx context.Context, query GetStatisticsQuery) (*entities.Statistics, error) {
	// 1. Try cache first (estatísticas são caras de calcular)
	cacheKey := "statistics:global"
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if stats, ok := cachedData.(*entities.Statistics); ok {
			return stats, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var stats entities.Statistics
			if err := json.Unmarshal([]byte(jsonData), &stats); err == nil {
				return &stats, nil
			}
		}
	}

	// 2. Cache miss - calculate statistics (expensive operation)
	stats, err := h.statsRepo.GetStatistics(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get statistics: %w", err)
	}

	// 3. Store in cache (TTL: 5 minutes)
	// Estatísticas não precisam ser real-time
	if err := h.cache.Set(ctx, cacheKey, stats, 5*time.Minute); err != nil {
		_ = err
	}

	return stats, nil
}

// InvalidateCache invalida o cache de estatísticas
// Deve ser chamado após operações que mudem contadores (create, delete, etc)
func (h *GetStatisticsQueryHandler) InvalidateCache(ctx context.Context) error {
	cacheKey := "statistics:global"
	return h.cache.Delete(ctx, cacheKey)
}

// RefreshCache força um refresh do cache de estatísticas
func (h *GetStatisticsQueryHandler) RefreshCache(ctx context.Context) error {
	// Invalida cache existente
	if err := h.InvalidateCache(ctx); err != nil {
		return err
	}

	// Recalcula estatísticas
	_, err := h.Handle(ctx, GetStatisticsQuery{})
	return err
}
