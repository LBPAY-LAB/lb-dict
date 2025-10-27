package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
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
	ClaimID    uuid.UUID
	Status     string
	CancelledAt time.Time
}

// CancelClaimCommandHandler handler para cancelar claim
type CancelClaimCommandHandler struct {
	claimRepo      ClaimRepository
	entryRepo      EntryRepository
	eventPublisher EventPublisher
}

// NewCancelClaimCommandHandler cria nova instância
func NewCancelClaimCommandHandler(
	claimRepo ClaimRepository,
	entryRepo EntryRepository,
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

	// 3. Validar status (deve estar PENDING)
	if claim.Status != "PENDING" {
		return nil, errors.New("claim must be PENDING to be cancelled")
	}

	// 4. Atualizar claim para CANCELLED
	now := time.Now()
	claim.Status = "CANCELLED"
	claim.ResolvedAt = &now
	claim.ResolutionNote = "Cancelled by user. Reason: " + cmd.Reason
	claim.UpdatedAt = now

	// 5. Persistir mudança
	if err := h.claimRepo.Update(ctx, claim); err != nil {
		return nil, errors.New("failed to cancel claim: " + err.Error())
	}

	// 6. Entry continua ACTIVE (claim foi rejeitado)

	// 7. Publicar evento (para notificar Bacen via RSFN)
	event := DomainEvent{
		EventType:     "ClaimCancelled",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"claim_id":       claim.ID.String(),
			"entry_id":       claim.EntryID.String(),
			"claimer_ispb":   claim.ClaimerISPB,
			"claimed_ispb":   claim.ClaimedISPB,
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
