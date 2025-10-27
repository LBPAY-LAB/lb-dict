package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// ConfirmClaimCommand comando para confirmar claim (usuário aceita reivindicação)
type ConfirmClaimCommand struct {
	ClaimID       uuid.UUID
	RequestedBy   uuid.UUID
	TwoFactorCode string // 2FA obrigatório
	ConfirmedBy   string // Nome do usuário que confirmou
}

// ConfirmClaimResult resultado do comando
type ConfirmClaimResult struct {
	ClaimID     uuid.UUID
	Status      valueobjects.ClaimStatus
	ConfirmedAt time.Time
}

// ConfirmClaimCommandHandler handler para confirmar claim
type ConfirmClaimCommandHandler struct {
	claimRepo      repositories.ClaimRepository
	entryRepo      repositories.EntryRepository
	eventPublisher EventPublisher
}

// NewConfirmClaimCommandHandler cria nova instância
func NewConfirmClaimCommandHandler(
	claimRepo repositories.ClaimRepository,
	entryRepo repositories.EntryRepository,
	eventPublisher EventPublisher,
) *ConfirmClaimCommandHandler {
	return &ConfirmClaimCommandHandler{
		claimRepo:      claimRepo,
		entryRepo:      entryRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle executa o comando
func (h *ConfirmClaimCommandHandler) Handle(ctx context.Context, cmd ConfirmClaimCommand) (*ConfirmClaimResult, error) {
	// 1. Validar 2FA
	if cmd.TwoFactorCode == "" {
		return nil, errors.New("2FA code required for claim confirmation")
	}
	// TODO: Validar 2FA

	// 2. Buscar claim
	claim, err := h.claimRepo.FindByID(ctx, cmd.ClaimID)
	if err != nil {
		return nil, errors.New("claim not found")
	}

	// 3. Validar se pode confirmar (status deve permitir transição)
	if !claim.Status.CanTransitionTo(valueobjects.ClaimStatusConfirmed) {
		return nil, errors.New("claim cannot be confirmed in current status")
	}

	// 4. Validar deadline (não pode confirmar após expiração)
	if claim.IsExpired() {
		return nil, errors.New("claim deadline exceeded")
	}

	// 5. Confirmar claim usando domain method
	reason := "Confirmed by user: " + cmd.ConfirmedBy
	if err := claim.Confirm(reason); err != nil {
		return nil, errors.New("failed to confirm claim: " + err.Error())
	}

	// 6. Persistir mudança
	if err := h.claimRepo.Update(ctx, claim); err != nil {
		return nil, errors.New("failed to update claim: " + err.Error())
	}

	// 7. Atualizar entry (marcar como em transferência)
	entry, err := h.entryRepo.FindByKey(ctx, claim.EntryKey)
	if err != nil {
		return nil, errors.New("entry not found")
	}

	// Marcar entry como tendo claim em andamento (usar string literal pois KeyStatus está duplicado)
	entry.Status = "CLAIM_PENDING"
	entry.UpdatedAt = time.Now()
	if err := h.entryRepo.Update(ctx, entry); err != nil {
		return nil, errors.New("failed to update entry status: " + err.Error())
	}

	// 8. Publicar evento (para iniciar workflow no Temporal)
	now := time.Now()
	event := DomainEvent{
		EventType:     "ClaimConfirmed",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"claim_id":       claim.ID.String(),
			"entry_key":      claim.EntryKey,
			"claimer_ispb":   claim.ClaimerParticipant.ISPB,
			"donor_ispb":     claim.DonorParticipant.ISPB,
			"bacen_claim_id": claim.BacenClaimID,
			"confirmed_by":   cmd.ConfirmedBy,
		},
	}
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		return nil, errors.New("failed to publish event: " + err.Error())
	}

	return &ConfirmClaimResult{
		ClaimID:     claim.ID,
		Status:      claim.Status,
		ConfirmedAt: now,
	}, nil
}
