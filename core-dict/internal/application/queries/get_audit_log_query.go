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

// GetAuditLogQuery representa a query para buscar audit logs
type GetAuditLogQuery struct {
	EntityType string    // Tipo de entidade (Entry, Claim, Account, etc)
	EntityID   uuid.UUID // ID da entidade
	Page       int       // 1-indexed
	PageSize   int       // default: 100, max: 1000
}

// GetAuditLogByActorQuery representa a query para buscar audit logs por ator
type GetAuditLogByActorQuery struct {
	ActorID  uuid.UUID // ID do ator (usuário/sistema)
	Page     int       // 1-indexed
	PageSize int       // default: 100, max: 1000
}

// GetAuditLogResult representa o resultado paginado
type GetAuditLogResult struct {
	AuditLogs  []*entities.AuditLog `json:"audit_logs"`
	TotalCount int64                `json:"total_count"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"page_size"`
	TotalPages int                  `json:"total_pages"`
}

// GetAuditLogQueryHandler lida com queries de audit logs
type GetAuditLogQueryHandler struct {
	auditRepo repositories.AuditRepository
	cache     services.CacheService
}

// NewGetAuditLogQueryHandler cria um novo handler para GetAuditLog
func NewGetAuditLogQueryHandler(
	auditRepo repositories.AuditRepository,
	cache services.CacheService,
) *GetAuditLogQueryHandler {
	return &GetAuditLogQueryHandler{
		auditRepo: auditRepo,
		cache:     cache,
	}
}

// Handle executa a query GetAuditLog por entidade
func (h *GetAuditLogQueryHandler) Handle(ctx context.Context, query GetAuditLogQuery) (*GetAuditLogResult, error) {
	// Validação
	if query.EntityType == "" || query.EntityID == uuid.Nil {
		return nil, fmt.Errorf("entity_type and entity_id are required")
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
	cacheKey := fmt.Sprintf("audit:entity:%s:%s:page:%d:size:%d",
		query.EntityType, query.EntityID.String(), query.Page, query.PageSize)
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if result, ok := cachedData.(*GetAuditLogResult); ok {
			return result, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var result GetAuditLogResult
			if err := json.Unmarshal([]byte(jsonData), &result); err == nil {
				return &result, nil
			}
		}
	}

	// 2. Cache miss - query database
	// Convert string to EntityType
	entityType := entities.EntityType(query.EntityType)
	auditEvents, err := h.auditRepo.FindByEntityID(ctx, entityType, query.EntityID, query.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	// Convert AuditEvent to AuditLog
	auditLogs := make([]*entities.AuditLog, len(auditEvents))
	for i, event := range auditEvents {
		var actorID uuid.UUID
		if event.UserID != nil {
			actorID = *event.UserID
		}

		// Merge OldValues, NewValues, and Diff into Changes
		changes := make(map[string]interface{})
		if event.OldValues != nil {
			changes["old"] = event.OldValues
		}
		if event.NewValues != nil {
			changes["new"] = event.NewValues
		}
		if event.Diff != nil {
			changes["diff"] = event.Diff
		}

		auditLogs[i] = &entities.AuditLog{
			ID:         event.ID,
			EntityType: string(event.EntityType),
			EntityID:   event.EntityID,
			Action:     string(event.EventType),
			ActorID:    actorID,
			ActorType:  "user",
			Changes:    changes,
			Metadata:   event.Metadata,
			Timestamp:  event.OccurredAt,
		}
	}

	// 3. Calcular total count usando Count method
	filters := repositories.AuditFilters{
		EntityType: &entityType,
		EntityID:   &query.EntityID,
		Limit:      query.PageSize,
		Offset:     offset,
	}
	totalCount, err := h.auditRepo.Count(ctx, filters)
	if err != nil {
		// Fallback to estimation if Count fails
		totalCount = int64(len(auditLogs))
		if len(auditLogs) == query.PageSize {
			totalCount = int64(query.Page * query.PageSize)
		}
	}

	totalPages := int(totalCount) / query.PageSize
	if int(totalCount)%query.PageSize > 0 {
		totalPages++
	}

	result := &GetAuditLogResult{
		AuditLogs:  auditLogs,
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}

	// 4. Store in cache (TTL: 15 minutes - audit logs não mudam)
	if err := h.cache.Set(ctx, cacheKey, result, 15*time.Minute); err != nil {
		_ = err
	}

	return result, nil
}

// HandleByActor executa a query GetAuditLog por ator
func (h *GetAuditLogQueryHandler) HandleByActor(ctx context.Context, query GetAuditLogByActorQuery) (*GetAuditLogResult, error) {
	// Validação
	if query.ActorID == uuid.Nil {
		return nil, fmt.Errorf("actor_id is required")
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
	cacheKey := fmt.Sprintf("audit:actor:%s:page:%d:size:%d",
		query.ActorID.String(), query.Page, query.PageSize)
	if cachedData, err := h.cache.Get(ctx, cacheKey); err == nil && cachedData != nil {
		// Cache hit
		if result, ok := cachedData.(*GetAuditLogResult); ok {
			return result, nil
		}
		// Try to unmarshal
		if jsonData, ok := cachedData.(string); ok {
			var result GetAuditLogResult
			if err := json.Unmarshal([]byte(jsonData), &result); err == nil {
				return &result, nil
			}
		}
	}

	// 2. Cache miss - query database
	auditEvents, err := h.auditRepo.FindByUserID(ctx, query.ActorID, query.PageSize, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}

	// Convert AuditEvent to AuditLog
	auditLogs := make([]*entities.AuditLog, len(auditEvents))
	for i, event := range auditEvents {
		var actorID uuid.UUID
		if event.UserID != nil {
			actorID = *event.UserID
		}

		// Merge OldValues, NewValues, and Diff into Changes
		changes := make(map[string]interface{})
		if event.OldValues != nil {
			changes["old"] = event.OldValues
		}
		if event.NewValues != nil {
			changes["new"] = event.NewValues
		}
		if event.Diff != nil {
			changes["diff"] = event.Diff
		}

		auditLogs[i] = &entities.AuditLog{
			ID:         event.ID,
			EntityType: string(event.EntityType),
			EntityID:   event.EntityID,
			Action:     string(event.EventType),
			ActorID:    actorID,
			ActorType:  "user",
			Changes:    changes,
			Metadata:   event.Metadata,
			Timestamp:  event.OccurredAt,
		}
	}

	// 3. Calcular total count
	filters := repositories.AuditFilters{
		UserID: &query.ActorID,
		Limit:  query.PageSize,
		Offset: offset,
	}
	totalCount, err := h.auditRepo.Count(ctx, filters)
	if err != nil {
		totalCount = int64(len(auditLogs))
		if len(auditLogs) == query.PageSize {
			totalCount = int64(query.Page * query.PageSize)
		}
	}

	totalPages := int(totalCount) / query.PageSize
	if int(totalCount)%query.PageSize > 0 {
		totalPages++
	}

	result := &GetAuditLogResult{
		AuditLogs:  auditLogs,
		TotalCount: totalCount,
		Page:       query.Page,
		PageSize:   query.PageSize,
		TotalPages: totalPages,
	}

	// 4. Store in cache (TTL: 15 minutes)
	if err := h.cache.Set(ctx, cacheKey, result, 15*time.Minute); err != nil {
		_ = err
	}

	return result, nil
}

// InvalidateCache invalida o cache de audit logs de uma entidade
func (h *GetAuditLogQueryHandler) InvalidateCache(ctx context.Context, entityType string, entityID uuid.UUID) error {
	pattern := fmt.Sprintf("audit:entity:%s:%s:*", entityType, entityID.String())
	return h.cache.Invalidate(ctx, pattern)
}
