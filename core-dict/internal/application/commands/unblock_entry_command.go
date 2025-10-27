package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// UnblockEntryCommand comando para desbloquear chave PIX
type UnblockEntryCommand struct {
	EntryID     uuid.UUID
	Reason      string // Motivo do desbloqueio (obrigatório)
	UnblockedBy string // Admin/Sistema que desbloqueou
}

// UnblockEntryResult resultado do comando
type UnblockEntryResult struct {
	EntryID     uuid.UUID
	Status      string
	UnblockedAt time.Time
}

// UnblockEntryCommandHandler handler para desbloquear chave PIX
type UnblockEntryCommandHandler struct {
	entryRepo      EntryRepository
	eventPublisher EventPublisher
	cacheService   CacheService
}

// NewUnblockEntryCommandHandler cria nova instância
func NewUnblockEntryCommandHandler(
	entryRepo EntryRepository,
	eventPublisher EventPublisher,
	cacheService CacheService,
) *UnblockEntryCommandHandler {
	return &UnblockEntryCommandHandler{
		entryRepo:      entryRepo,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
	}
}

// Handle executa o comando
func (h *UnblockEntryCommandHandler) Handle(ctx context.Context, cmd UnblockEntryCommand) (*UnblockEntryResult, error) {
	// 1. Validar entrada
	if cmd.Reason == "" {
		return nil, errors.New("reason is required for unblocking")
	}

	// 2. Buscar entry
	entry, err := h.entryRepo.FindByID(ctx, cmd.EntryID)
	if err != nil {
		return nil, errors.New("entry not found")
	}

	// 3. Validar status (apenas BLOCKED pode ser desbloqueado)
	if entry.Status != "BLOCKED" {
		return nil, errors.New("only blocked entries can be unblocked")
	}

	// 4. Atualizar status para ACTIVE
	now := time.Now()
	entry.Status = "ACTIVE"
	entry.UpdatedAt = now

	// 5. Persistir mudança
	if err := h.entryRepo.Update(ctx, entry); err != nil {
		return nil, errors.New("failed to unblock entry: " + err.Error())
	}

	// 6. Publicar evento
	event := DomainEvent{
		EventType:     "EntryUnblocked",
		AggregateID:   entry.ID.String(),
		AggregateType: "Entry",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"entry_id":     entry.ID.String(),
			"key_value":    entry.KeyValue,
			"reason":       cmd.Reason,
			"unblocked_by": cmd.UnblockedBy,
		},
	}
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		return nil, errors.New("failed to publish event: " + err.Error())
	}

	// 7. Invalidar cache
	h.cacheService.InvalidateKey(ctx, "entry:"+entry.KeyValue)

	return &UnblockEntryResult{
		EntryID:     entry.ID,
		Status:      entry.Status,
		UnblockedAt: now,
	}, nil
}
