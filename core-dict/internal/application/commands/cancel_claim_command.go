package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// CancelClaimCommand comando para cancelar claim (usuário rejeita reivindicação)
type CancelClaimCommand struct {
	ClaimID       uuid.UUID
	RequestedBy   uuid.UUID
	TwoFactorCode string // 2FA obrigatório
	Reason        string // Motivo do cancelamento
}

// CancelClaimResult resultado do comando
type CancelClaimResult struct {
	ClaimID     uuid.UUID
	Status      valueobjects.ClaimStatus
	CancelledAt time.Time
}

// CancelClaimCommandHandler handler para cancelar claim
type CancelClaimCommandHandler struct {
	claimRepo      repositories.ClaimRepository
	entryRepo      repositories.EntryRepository
	eventPublisher EventPublisher
}

// NewCancelClaimCommandHandler cria nova instância
func NewCancelClaimCommandHandler(
	claimRepo repositories.ClaimRepository,
	entryRepo repositories.EntryRepository,
	eventPublisher EventPublisher,
) *CancelClaimCommandHandler {
	return &CancelClaimCommandHandler{
		claimRepo:      claimRepo,
		entryRepo:      entryRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle executa o comando
func (h *CancelClaimCommandHandler) Handle(ctx context.Context, cmd CancelClaimCommand) (*CancelClaimResult, error) {
	// 1. Validar 2FA
	if cmd.TwoFactorCode == "" {
		return nil, errors.New("2FA code required for claim cancellation")
	}
	// TODO: Validar 2FA

	// 2. Buscar claim
	claim, err := h.claimRepo.FindByID(ctx, cmd.ClaimID)
	if err != nil {
		return nil, errors.New("claim not found")
	}

	// 3. Validar se pode cancelar (status deve permitir)
	if claim.Status.IsFinal() {
		return nil, errors.New("cannot cancel claim in final status")
	}

	// 4. Cancelar claim usando domain method
	reason := cmd.Reason
	if reason == "" {
		reason = "Cancelled by user"
	}
	if err := claim.Cancel(reason); err != nil {
		return nil, errors.New("failed to cancel claim: " + err.Error())
	}

	// 5. Persistir mudança
	if err := h.claimRepo.Update(ctx, claim); err != nil {
		return nil, errors.New("failed to update claim: " + err.Error())
	}

	// 6. Entry continua ACTIVE (claim foi rejeitado)
	// Nota: Entry só muda para ACTIVE novamente se estava em CLAIM_PENDING

	entry, err := h.entryRepo.FindByKey(ctx, claim.EntryKey)
	if err == nil && entry.Status == "CLAIM_PENDING" {
		entry.Status = entities.KeyStatusActive
		entry.UpdatedAt = time.Now()
		_ = h.entryRepo.Update(ctx, entry)
	}

	// 7. Publicar evento (para notificar Bacen via RSFN)
	now := time.Now()
	event := DomainEvent{
		EventType:     "ClaimCancelled",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"claim_id":       claim.ID.String(),
			"entry_key":      claim.EntryKey,
			"claimer_ispb":   claim.ClaimerParticipant.ISPB,
			"donor_ispb":     claim.DonorParticipant.ISPB,
			"bacen_claim_id": claim.BacenClaimID,
			"reason":         cmd.Reason,
		},
	}
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		return nil, errors.New("failed to publish event: " + err.Error())
	}

	return &CancelClaimResult{
		ClaimID:     claim.ID,
		Status:      claim.Status,
		CancelledAt: now,
	}, nil
}
