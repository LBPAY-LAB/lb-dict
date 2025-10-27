package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
	"github.com/lbpay-lab/core-dict/internal/application/services"
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
	Status      valueobjects.ClaimStatus
	CompletedAt time.Time
}

// CompleteClaimCommandHandler handler para completar claim
type CompleteClaimCommandHandler struct {
	claimRepo      repositories.ClaimRepository
	entryRepo      repositories.EntryRepository
	eventPublisher EventPublisher
	cacheService   services.CacheService
}

// NewCompleteClaimCommandHandler cria nova instância
func NewCompleteClaimCommandHandler(
	claimRepo repositories.ClaimRepository,
	entryRepo repositories.EntryRepository,
	eventPublisher EventPublisher,
	cacheService services.CacheService,
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
	if claim.Status != valueobjects.ClaimStatusConfirmed {
		return nil, errors.New("claim must be confirmed to be completed")
	}

	// 3. Completar claim usando domain method
	if err := claim.Complete(); err != nil {
		return nil, errors.New("failed to complete claim: " + err.Error())
	}

	// Atualizar resolution reason com detalhes do Bacen
	claim.ResolutionReason = "Completed by " + cmd.CompletedBy + ". Bacen response: " + cmd.BacenResponseID

	// 4. Persistir mudança
	if err := h.claimRepo.Update(ctx, claim); err != nil {
		return nil, errors.New("failed to update claim: " + err.Error())
	}

	// 5. Atualizar entry (transferir para outro PSP - deletar localmente)
	entry, err := h.entryRepo.FindByKey(ctx, claim.EntryKey)
	if err != nil {
		return nil, errors.New("entry not found")
	}

	// Marcar entry como transferido (soft delete)
	now := time.Now()
	entry.Status = entities.KeyStatusDeleted // Chave foi transferida
	entry.UpdatedAt = now
	entry.DeletedAt = &now // Soft delete

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
			"entry_key":          claim.EntryKey,
			"key_value":          entry.KeyValue,
			"claimer_ispb":       claim.ClaimerParticipant.ISPB,
			"donor_ispb":         claim.DonorParticipant.ISPB,
			"bacen_response_id":  cmd.BacenResponseID,
		},
	}
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		return nil, errors.New("failed to publish event: " + err.Error())
	}

	// 7. Invalidar cache
	h.cacheService.Delete(ctx, "entry:"+entry.KeyValue)
	h.cacheService.Invalidate(ctx, "claims:entry:"+entry.ID.String())

	return &CompleteClaimResult{
		ClaimID:     claim.ID,
		Status:      claim.Status,
		CompletedAt: now,
	}, nil
}
