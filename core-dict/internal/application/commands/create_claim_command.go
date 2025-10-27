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

// CreateClaimCommand comando para criar claim (reivindicação)
type CreateClaimCommand struct {
	KeyValue      string
	ClaimType     valueobjects.ClaimType // OWNERSHIP ou PORTABILITY
	ClaimerISPB   string                 // ISPB do PSP reivindicador
	ClaimedISPB   string                 // ISPB do PSP reivindicado (nós)
	AccountID     uuid.UUID              // Nova conta (destino do claim)
	OwnerTaxID    string
	RequestedBy   uuid.UUID
	BacenClaimID  string // ID retornado pelo Bacen
}

// CreateClaimResult resultado do comando
type CreateClaimResult struct {
	ClaimID    uuid.UUID
	Status     string
	DeadlineAt time.Time // 7 dias para responder
}

// CreateClaimCommandHandler handler para criação de claim
type CreateClaimCommandHandler struct {
	entryRepo      repositories.EntryRepository
	claimRepo      repositories.ClaimRepository
	eventPublisher EventPublisher
}

// NewCreateClaimCommandHandler cria nova instância
func NewCreateClaimCommandHandler(
	entryRepo repositories.EntryRepository,
	claimRepo repositories.ClaimRepository,
	eventPublisher EventPublisher,
) *CreateClaimCommandHandler {
	return &CreateClaimCommandHandler{
		entryRepo:      entryRepo,
		claimRepo:      claimRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle executa o comando
func (h *CreateClaimCommandHandler) Handle(ctx context.Context, cmd CreateClaimCommand) (*CreateClaimResult, error) {
	// 1. Validar entrada
	if cmd.KeyValue == "" {
		return nil, errors.New("key_value is required")
	}
	if cmd.BacenClaimID == "" {
		return nil, errors.New("bacen_claim_id is required")
	}

	// 2. Buscar entry existente
	entry, err := h.entryRepo.FindByKey(ctx, cmd.KeyValue)
	if err != nil {
		return nil, errors.New("entry not found for key: " + cmd.KeyValue)
	}

	// 3. Validar que entry pertence a este ISPB
	if entry.ISPB != cmd.ClaimedISPB {
		return nil, errors.New("entry does not belong to claimed ISPB")
	}

	// 4. Validar que não existe claim ativo para esta chave
	existingClaim, err := h.claimRepo.FindActiveByEntryID(ctx, entry.ID)
	if err == nil && existingClaim != nil {
		return nil, errors.New("active claim already exists for this key")
	}

	// 5. Criar entidade Claim usando domain factory
	claimerParticipant := valueobjects.Participant{
		ISPB: cmd.ClaimerISPB,
		Name: "", // TODO: buscar nome do participante
	}
	donorParticipant := valueobjects.Participant{
		ISPB: cmd.ClaimedISPB,
		Name: "", // TODO: buscar nome do participante
	}

	claim, err := entities.NewClaim(
		cmd.KeyValue,
		cmd.ClaimType,
		claimerParticipant,
		donorParticipant,
		cmd.AccountID,
		entry.AccountID,
	)
	if err != nil {
		return nil, errors.New("failed to create claim: " + err.Error())
	}

	// Adicionar dados do Bacen
	claim.SetBacenClaimID(cmd.BacenClaimID)

	// 6. Persistir claim
	if err := h.claimRepo.Create(ctx, claim); err != nil {
		return nil, errors.New("failed to create claim: " + err.Error())
	}

	// 7. Publicar evento (para notificar usuário via app/email)
	now := time.Now()
	event := DomainEvent{
		EventType:     "ClaimReceived",
		AggregateID:   claim.ID.String(),
		AggregateType: "Claim",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"claim_id":       claim.ID.String(),
			"entry_id":       entry.ID.String(),
			"key_value":      entry.KeyValue,
			"claim_type":     string(cmd.ClaimType),
			"claimer_ispb":   cmd.ClaimerISPB,
			"deadline_at":    claim.ExpiresAt,
			"bacen_claim_id": cmd.BacenClaimID,
		},
	}
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		return nil, errors.New("failed to publish event: " + err.Error())
	}

	return &CreateClaimResult{
		ClaimID:    claim.ID,
		Status:     string(claim.Status),
		DeadlineAt: claim.ExpiresAt,
	}, nil
}
