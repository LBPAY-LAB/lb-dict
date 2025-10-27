package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
	"github.com/lbpay-lab/core-dict/internal/application/services"
)

// BlockEntryCommand comando para bloquear chave PIX (infração, fraude, etc.)
type BlockEntryCommand struct {
	EntryID     uuid.UUID
	Reason      string // Motivo do bloqueio (obrigatório)
	BlockedBy   string // Admin/Sistema que bloqueou
	InfractionID *uuid.UUID // Opcional: ID da infração relacionada
}

// BlockEntryResult resultado do comando
type BlockEntryResult struct {
	EntryID   uuid.UUID
	Status    entities.KeyStatus
	BlockedAt time.Time
}

// BlockEntryCommandHandler handler para bloquear chave PIX
type BlockEntryCommandHandler struct {
	entryRepo      repositories.EntryRepository
	eventPublisher EventPublisher
	cacheService   services.CacheService
}

// NewBlockEntryCommandHandler cria nova instância
func NewBlockEntryCommandHandler(
	entryRepo repositories.EntryRepository,
	eventPublisher EventPublisher,
	cacheService services.CacheService,
) *BlockEntryCommandHandler {
	return &BlockEntryCommandHandler{
		entryRepo:      entryRepo,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
	}
}

// Handle executa o comando
func (h *BlockEntryCommandHandler) Handle(ctx context.Context, cmd BlockEntryCommand) (*BlockEntryResult, error) {
	// 1. Validar entrada
	if cmd.Reason == "" {
		return nil, errors.New("reason is required for blocking")
	}

	// 2. Buscar entry
	entry, err := h.entryRepo.FindByID(ctx, cmd.EntryID)
	if err != nil {
		return nil, errors.New("entry not found")
	}

	// 3. Validar status (apenas ACTIVE pode ser bloqueado)
	if entry.Status != entities.KeyStatusActive {
		return nil, errors.New("only active entries can be blocked")
	}

	// 4. Atualizar status para BLOCKED
	now := time.Now()
	entry.Status = entities.KeyStatusBlocked
	entry.UpdatedAt = now

	// 5. Persistir mudança
	if err := h.entryRepo.Update(ctx, entry); err != nil {
		return nil, errors.New("failed to block entry: " + err.Error())
	}

	// 6. Publicar evento (para notificar Bacen e usuário)
	event := DomainEvent{
		EventType:     "EntryBlocked",
		AggregateID:   entry.ID.String(),
		AggregateType: "Entry",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"entry_id":      entry.ID.String(),
			"key_value":     entry.KeyValue,
			"reason":        cmd.Reason,
			"blocked_by":    cmd.BlockedBy,
			"infraction_id": cmd.InfractionID,
		},
	}
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		return nil, errors.New("failed to publish event: " + err.Error())
	}

	// 7. Invalidar cache
	h.cacheService.Delete(ctx, "entry:"+entry.KeyValue)

	return &BlockEntryResult{
		EntryID:   entry.ID,
		Status:    entry.Status,
		BlockedAt: now,
	}, nil
}
