package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
	Status      string
	ConfirmedAt time.Time
}

// ConfirmClaimCommandHandler handler para confirmar claim
type ConfirmClaimCommandHandler struct {
	claimRepo      ClaimRepository
	entryRepo      EntryRepository
	eventPublisher EventPublisher
}

// NewConfirmClaimCommandHandler cria nova instância
func NewConfirmClaimCommandHandler(
	claimRepo ClaimRepository,
	entryRepo EntryRepository,
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

	// 3. Validar status (deve estar PENDING)
	if claim.Status != "PENDING" {
		return nil, errors.New("claim must be PENDING to be confirmed")
	}

	// 4. Validar deadline (não pode confirmar após 7 dias)
	if time.Now().After(claim.DeadlineAt) {
		return nil, errors.New("claim deadline exceeded")
	}

	// 5. Atualizar claim para CONFIRMED
	now := time.Now()
	claim.Status = "CONFIRMED"
	claim.ResolvedAt = &now
	claim.ResolutionNote = "Confirmed by user: " + cmd.ConfirmedBy
	claim.UpdatedAt = now

	// 6. Persistir mudança
	if err := h.claimRepo.Update(ctx, claim); err != nil {
		return nil, errors.New("failed to confirm claim: " + err.Error())
	}

	// 7. Atualizar entry (marcar como em transferência)
	entry, err := h.entryRepo.FindByID(ctx, claim.EntryID)
	if err != nil {
		return nil, errors.New("entry not found")
	}
	entry.Status = "TRANSFERRING"
	entry.UpdatedAt = now
	if err := h.entryRepo.Update(ctx, entry); err != nil {
		return nil, errors.New("failed to update entry status: " + err.Error())
	}

	// 8. Publicar evento (para iniciar workflow no Temporal)
	event := DomainEvent{
		EventType:     "ClaimConfirmed",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"claim_id":       claim.ID.String(),
			"entry_id":       claim.EntryID.String(),
			"claimer_ispb":   claim.ClaimerISPB,
			"claimed_ispb":   claim.ClaimedISPB,
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
