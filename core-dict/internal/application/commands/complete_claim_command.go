package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// CompleteClaimCommand comando para completar claim (após confirmação do Bacen)
type CompleteClaimCommand struct {
	ClaimID         uuid.UUID
	BacenResponseID string // ID de resposta do Bacen
	CompletedBy     string // Sistema que completou (geralmente "RSFN")
}

// CompleteClaimResult resultado do comando
type CompleteClaimResult struct {
	ClaimID     uuid.UUID
	Status      string
	CompletedAt time.Time
}

// CompleteClaimCommandHandler handler para completar claim
type CompleteClaimCommandHandler struct {
	claimRepo      ClaimRepository
	entryRepo      EntryRepository
	eventPublisher EventPublisher
	cacheService   CacheService
}

// NewCompleteClaimCommandHandler cria nova instância
func NewCompleteClaimCommandHandler(
	claimRepo ClaimRepository,
	entryRepo EntryRepository,
	eventPublisher EventPublisher,
	cacheService CacheService,
) *CompleteClaimCommandHandler {
	return &CompleteClaimCommandHandler{
		claimRepo:      claimRepo,
		entryRepo:      entryRepo,
		eventPublisher: eventPublisher,
		cacheService:   cacheService,
	}
}

// Handle executa o comando
func (h *CompleteClaimCommandHandler) Handle(ctx context.Context, cmd CompleteClaimCommand) (*CompleteClaimResult, error) {
	// 1. Buscar claim
	claim, err := h.claimRepo.FindByID(ctx, cmd.ClaimID)
	if err != nil {
		return nil, errors.New("claim not found")
	}

	// 2. Validar status (deve estar CONFIRMED)
	if claim.Status != "CONFIRMED" {
		return nil, errors.New("claim must be CONFIRMED to be completed")
	}

	// 3. Atualizar claim para COMPLETED
	now := time.Now()
	claim.Status = "COMPLETED"
	claim.ResolvedAt = &now
	claim.ResolutionNote = "Completed by " + cmd.CompletedBy + ". Bacen response: " + cmd.BacenResponseID
	claim.UpdatedAt = now

	// 4. Persistir mudança
	if err := h.claimRepo.Update(ctx, claim); err != nil {
		return nil, errors.New("failed to complete claim: " + err.Error())
	}

	// 5. Atualizar entry (transferir para outro PSP - deletar localmente)
	entry, err := h.entryRepo.FindByID(ctx, claim.EntryID)
	if err != nil {
		return nil, errors.New("entry not found")
	}
	entry.Status = "TRANSFERRED"
	entry.UpdatedAt = now
	if err := h.entryRepo.Update(ctx, entry); err != nil {
		return nil, errors.New("failed to update entry status: " + err.Error())
	}

	// 6. Publicar evento (para notificação ao usuário)
	event := DomainEvent{
		EventType:     "ClaimCompleted",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"claim_id":           claim.ID.String(),
			"entry_id":           claim.EntryID.String(),
			"key_value":          entry.KeyValue,
			"claimer_ispb":       claim.ClaimerISPB,
			"bacen_response_id":  cmd.BacenResponseID,
		},
	}
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		return nil, errors.New("failed to publish event: " + err.Error())
	}

	// 7. Invalidar cache
	h.cacheService.InvalidateKey(ctx, "entry:"+entry.KeyValue)
	h.cacheService.InvalidatePattern(ctx, "claims:entry:"+entry.ID.String())

	return &CompleteClaimResult{
		ClaimID:     claim.ID,
		Status:      claim.Status,
		CompletedAt: now,
	}, nil
}
