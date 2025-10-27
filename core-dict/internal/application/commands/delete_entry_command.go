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

// DeleteEntryCommand comando para deletar chave PIX
type DeleteEntryCommand struct {
	EntryID       uuid.UUID
	RequestedBy   uuid.UUID
	TwoFactorCode string // 2FA obrigatório
	Reason        string // Motivo da deleção (opcional)
}

// DeleteEntryResult resultado do comando
type DeleteEntryResult struct {
	Success   bool
	DeletedAt time.Time
}

// DeleteEntryCommandHandler handler para deleção de chave PIX
type DeleteEntryCommandHandler struct {
	entryRepo      repositories.EntryRepository
	eventPublisher EventPublisher
	cacheService   services.CacheService
	connectClient  services.ConnectClient // NEW: gRPC client for RSFN
	entryProducer  EntryEventProducer     // NEW: Pulsar event producer
}

// NewDeleteEntryCommandHandler cria nova instância
func NewDeleteEntryCommandHandler(
	entryRepo repositories.EntryRepository,
	eventPublisher EventPublisher,
	cacheService services.CacheService,
	connectClient services.ConnectClient,
	entryProducer EntryEventProducer,
) *DeleteEntryCommandHandler {
	return &DeleteEntryCommandHandler{
		entryRepo:      entryRepo,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
		connectClient:  connectClient,
		entryProducer:  entryProducer,
	}
}

// Handle executa o comando
func (h *DeleteEntryCommandHandler) Handle(ctx context.Context, cmd DeleteEntryCommand) (*DeleteEntryResult, error) {
	// 1. Validar 2FA obrigatório
	if cmd.TwoFactorCode == "" {
		return nil, errors.New("2FA code required for key deletion")
	}
	// TODO: Validar código 2FA

	// 2. Buscar entry
	entry, err := h.entryRepo.FindByID(ctx, cmd.EntryID)
	if err != nil {
		return nil, errors.New("entry not found")
	}

	// 3. Validar status (apenas ACTIVE pode ser deletado)
	if entry.Status != entities.KeyStatusActive {
		return nil, errors.New("only active entries can be deleted")
	}

	// 4. Atualizar status para DELETED (soft delete)
	entry.Status = entities.KeyStatusDeleted
	now := time.Now()
	entry.UpdatedAt = now

	// 5. Persistir mudança
	if err := h.entryRepo.Update(ctx, entry); err != nil {
		return nil, errors.New("failed to delete entry: " + err.Error())
	}

	// 6. Publicar evento de deleção via Pulsar (EntryDeleted)
	// Non-blocking: triggers Connect → Bridge → Bacen DICT deletion
	if h.entryProducer != nil {
		go func() {
			bgCtx := context.Background()
			if err := h.eventPublisher.Publish(bgCtx, DomainEvent{
				EventType:     "EntryDeleted",
				AggregateID:   entry.ID.String(),
				AggregateType: "Entry",
				OccurredAt:    now,
				Payload: map[string]interface{}{
					"entry_id":   entry.ID.String(),
					"key_value":  entry.KeyValue,
					"key_type":   string(entry.KeyType),
					"deleted_by": cmd.RequestedBy.String(),
					"reason":     cmd.Reason,
				},
			}); err != nil {
				// Log error but don't fail request
			}
		}()
	}

	// 7. Invalidar cache
	h.cacheService.Delete(ctx, "entry:"+entry.KeyValue)
	h.cacheService.Invalidate(ctx, "entries:account:"+entry.AccountID.String())

	return &DeleteEntryResult{
		Success:   true,
		DeletedAt: now,
	}, nil
}
