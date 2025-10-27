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

// UpdateEntryCommand comando para atualizar chave PIX
type UpdateEntryCommand struct {
	EntryID       uuid.UUID
	AccountID     uuid.UUID // Nova conta (opcional)
	AccountISPB   string
	AccountBranch string
	AccountNumber string
	RequestedBy   uuid.UUID
	TwoFactorCode string // 2FA obrigatório
}

// UpdateEntryResult resultado do comando
type UpdateEntryResult struct {
	EntryID   uuid.UUID
	UpdatedAt time.Time
}

// UpdateEntryCommandHandler handler para atualização de chave PIX
type UpdateEntryCommandHandler struct {
	entryRepo      repositories.EntryRepository
	eventPublisher EventPublisher
	cacheService   services.CacheService
	connectClient  services.ConnectClient      // NEW: gRPC client for RSFN
	entryProducer  EntryEventProducer // NEW: Pulsar event producer
}

// NewUpdateEntryCommandHandler cria nova instância
func NewUpdateEntryCommandHandler(
	entryRepo repositories.EntryRepository,
	eventPublisher EventPublisher,
	cacheService services.CacheService,
	connectClient services.ConnectClient,
	entryProducer EntryEventProducer,
) *UpdateEntryCommandHandler {
	return &UpdateEntryCommandHandler{
		entryRepo:      entryRepo,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
		connectClient:  connectClient,
		entryProducer:  entryProducer,
	}
}

// Handle executa o comando
func (h *UpdateEntryCommandHandler) Handle(ctx context.Context, cmd UpdateEntryCommand) (*UpdateEntryResult, error) {
	// 1. Validar 2FA
	if cmd.TwoFactorCode == "" {
		return nil, errors.New("2FA code required")
	}
	// TODO: Validar código 2FA com serviço de autenticação

	// 2. Buscar entry existente
	entry, err := h.entryRepo.FindByID(ctx, cmd.EntryID)
	if err != nil {
		return nil, errors.New("entry not found")
	}

	// 3. Validar status (apenas ACTIVE pode ser atualizado)
	if entry.Status != entities.KeyStatusActive {
		return nil, errors.New("only active entries can be updated")
	}

	// 4. Atualizar campos (flat structure)
	if cmd.AccountID != uuid.Nil {
		entry.AccountID = cmd.AccountID
	}
	if cmd.AccountISPB != "" {
		entry.ISPB = cmd.AccountISPB
	}
	if cmd.AccountBranch != "" {
		entry.Branch = cmd.AccountBranch
	}
	if cmd.AccountNumber != "" {
		entry.AccountNumber = cmd.AccountNumber
	}
	entry.UpdatedAt = time.Now()

	// 5. Persistir mudanças
	if err := h.entryRepo.Update(ctx, entry); err != nil {
		return nil, errors.New("failed to update entry: " + err.Error())
	}

	// 6. Publicar evento via Pulsar (EntryUpdated)
	// Non-blocking: triggers Connect → Bridge → Bacen DICT update
	if h.entryProducer != nil {
		go func() {
			bgCtx := context.Background()
			if err := h.eventPublisher.Publish(bgCtx, DomainEvent{
				EventType:     "EntryUpdated",
				AggregateID:   entry.ID.String(),
				AggregateType: "Entry",
				OccurredAt:    time.Now(),
				Payload: map[string]interface{}{
					"entry_id":       entry.ID.String(),
					"key_value":      entry.KeyValue,
					"new_account_id": cmd.AccountID.String(),
					"new_ispb":       cmd.AccountISPB,
					"new_branch":     cmd.AccountBranch,
					"new_number":     cmd.AccountNumber,
					"updated_by":     cmd.RequestedBy.String(),
				},
			}); err != nil {
				// Log error but don't fail request
			}
		}()
	}

	// 7. Invalidar cache
	h.cacheService.Delete(ctx, "entry:"+entry.KeyValue)

	return &UpdateEntryResult{
		EntryID:   entry.ID,
		UpdatedAt: entry.UpdatedAt,
	}, nil
}
