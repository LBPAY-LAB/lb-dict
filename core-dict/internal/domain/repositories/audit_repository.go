package repositories

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// AuditRepository define as operações de persistência para AuditEvent
type AuditRepository interface {
	// Create cria um novo evento de auditoria
	Create(ctx context.Context, event *entities.AuditEvent) error

	// FindByID busca evento por ID
	FindByID(ctx context.Context, eventID uuid.UUID) (*entities.AuditEvent, error)

	// FindByEntityID lista eventos por entidade
	FindByEntityID(ctx context.Context, entityType entities.EntityType, entityID uuid.UUID, limit, offset int) ([]*entities.AuditEvent, error)

	// FindByEventType lista eventos por tipo
	FindByEventType(ctx context.Context, eventType entities.EventType, limit, offset int) ([]*entities.AuditEvent, error)

	// FindByUserID lista eventos por usuário
	FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.AuditEvent, error)

	// FindByDateRange lista eventos em um período
	FindByDateRange(ctx context.Context, from, to time.Time, limit, offset int) ([]*entities.AuditEvent, error)

	// List lista eventos com paginação e filtros
	List(ctx context.Context, filters AuditFilters) ([]*entities.AuditEvent, error)

	// Count conta total de eventos
	Count(ctx context.Context, filters AuditFilters) (int64, error)
}

// AuditFilters define filtros para listagem de eventos de auditoria
type AuditFilters struct {
	EventType    *entities.EventType
	EntityType   *entities.EntityType
	EntityID     *uuid.UUID
	UserID       *uuid.UUID
	IPAddress    *string
	OccurredAfter  *time.Time
	OccurredBefore *time.Time
	Limit        int
	Offset       int
}
