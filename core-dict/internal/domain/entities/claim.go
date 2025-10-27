package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// Claim representa uma reivindicação de chave PIX
type Claim struct {
	ID                    uuid.UUID
	EntryKey              string                    // Chave PIX reivindicada
	ClaimType             valueobjects.ClaimType    // Tipo de reivindicação
	Status                valueobjects.ClaimStatus  // Status da reivindicação
	ClaimerParticipant    valueobjects.Participant  // Participante reivindicante
	DonorParticipant      valueobjects.Participant  // Participante doador
	ClaimerAccountID      uuid.UUID                 // Conta do reivindicante
	DonorAccountID        uuid.UUID                 // Conta do doador
	BacenClaimID          string                    // ID do claim no Bacen
	WorkflowID            string                    // ID do Temporal workflow
	CompletionPeriodDays  int                       // Período de conclusão (30 dias)
	ExpiresAt             time.Time                 // Data de expiração
	ResolutionType        string                    // Tipo de resolução (APPROVED, REJECTED, TIMEOUT)
	ResolutionReason      string                    // Razão da resolução
	ResolutionDate        *time.Time                // Data da resolução
	Metadata              map[string]interface{}
	CreatedAt             time.Time
	UpdatedAt             time.Time
	DeletedAt             *time.Time
}

// NewClaim cria uma nova reivindicação de chave PIX
func NewClaim(
	entryKey string,
	claimType valueobjects.ClaimType,
	claimer valueobjects.Participant,
	donor valueobjects.Participant,
	claimerAccountID uuid.UUID,
	donorAccountID uuid.UUID,
) (*Claim, error) {
	if entryKey == "" {
		return nil, errors.New("entry key cannot be empty")
	}
	if !claimType.IsValid() {
		return nil, errors.New("invalid claim type")
	}
	if claimer.ISPB == donor.ISPB {
		return nil, errors.New("claimer and donor cannot be the same participant")
	}
	if claimerAccountID == uuid.Nil || donorAccountID == uuid.Nil {
		return nil, errors.New("account IDs cannot be nil")
	}

	now := time.Now()
	completionPeriod := 30 // 30 dias conforme regulação
	expiresAt := now.AddDate(0, 0, completionPeriod)

	return &Claim{
		ID:                   uuid.New(),
		EntryKey:             entryKey,
		ClaimType:            claimType,
		Status:               valueobjects.ClaimStatusOpen,
		ClaimerParticipant:   claimer,
		DonorParticipant:     donor,
		ClaimerAccountID:     claimerAccountID,
		DonorAccountID:       donorAccountID,
		CompletionPeriodDays: completionPeriod,
		ExpiresAt:            expiresAt,
		Metadata:             make(map[string]interface{}),
		CreatedAt:            now,
		UpdatedAt:            now,
	}, nil
}

// Validate valida as regras de negócio da reivindicação
func (c *Claim) Validate() error {
	if c.ID == uuid.Nil {
		return errors.New("claim ID cannot be nil")
	}
	if c.EntryKey == "" {
		return errors.New("entry key cannot be empty")
	}
	if !c.ClaimType.IsValid() {
		return errors.New("invalid claim type")
	}
	if !c.Status.IsValid() {
		return errors.New("invalid status")
	}
	if c.ClaimerParticipant.ISPB == c.DonorParticipant.ISPB {
		return errors.New("claimer and donor must be different participants")
	}
	if c.ExpiresAt.Before(c.CreatedAt) {
		return errors.New("expiration date must be after creation date")
	}
	return nil
}

// Confirm confirma a reivindicação (doador aceita)
func (c *Claim) Confirm(reason string) error {
	if !c.Status.CanTransitionTo(valueobjects.ClaimStatusConfirmed) {
		return errors.New("cannot confirm claim in current status")
	}
	now := time.Now()
	c.Status = valueobjects.ClaimStatusConfirmed
	c.ResolutionType = "APPROVED"
	c.ResolutionReason = reason
	c.ResolutionDate = &now
	c.UpdatedAt = now
	return nil
}

// Cancel cancela a reivindicação (doador rejeita ou reivindicante desiste)
func (c *Claim) Cancel(reason string) error {
	if c.Status.IsFinal() {
		return errors.New("cannot cancel claim in final status")
	}
	now := time.Now()
	c.Status = valueobjects.ClaimStatusCancelled
	c.ResolutionType = "CANCELLED"
	c.ResolutionReason = reason
	c.ResolutionDate = &now
	c.UpdatedAt = now
	return nil
}

// Complete completa a reivindicação com sucesso
func (c *Claim) Complete() error {
	if c.Status != valueobjects.ClaimStatusConfirmed {
		return errors.New("claim must be confirmed before completion")
	}
	now := time.Now()
	c.Status = valueobjects.ClaimStatusCompleted
	c.ResolutionType = "APPROVED"
	c.ResolutionDate = &now
	c.UpdatedAt = now
	return nil
}

// Expire expira a reivindicação por timeout (30 dias)
func (c *Claim) Expire() error {
	if c.Status.IsFinal() {
		return errors.New("claim is already in final status")
	}
	if time.Now().Before(c.ExpiresAt) {
		return errors.New("claim has not expired yet")
	}
	now := time.Now()
	c.Status = valueobjects.ClaimStatusExpired
	c.ResolutionType = "TIMEOUT"
	c.ResolutionReason = "No response within completion period"
	c.ResolutionDate = &now
	c.UpdatedAt = now
	return nil
}

// AutoConfirm auto-confirma a reivindicação por timeout
func (c *Claim) AutoConfirm() error {
	if c.Status != valueobjects.ClaimStatusWaitingResolution {
		return errors.New("claim must be waiting resolution for auto-confirmation")
	}
	if time.Now().Before(c.ExpiresAt) {
		return errors.New("auto-confirmation deadline not reached")
	}
	now := time.Now()
	c.Status = valueobjects.ClaimStatusAutoConfirmed
	c.ResolutionType = "TIMEOUT"
	c.ResolutionReason = "Auto-confirmed after completion period"
	c.ResolutionDate = &now
	c.UpdatedAt = now
	return nil
}

// SetWaitingResolution marca a reivindicação como aguardando resolução
func (c *Claim) SetWaitingResolution() error {
	if !c.Status.CanTransitionTo(valueobjects.ClaimStatusWaitingResolution) {
		return errors.New("cannot set waiting resolution in current status")
	}
	c.Status = valueobjects.ClaimStatusWaitingResolution
	c.UpdatedAt = time.Now()
	return nil
}

// IsExpired verifica se a reivindicação expirou
func (c *Claim) IsExpired() bool {
	return time.Now().After(c.ExpiresAt)
}

// IsFinal verifica se o status da reivindicação é final
func (c *Claim) IsFinal() bool {
	return c.Status.IsFinal()
}

// SetWorkflowID define o ID do workflow Temporal
func (c *Claim) SetWorkflowID(workflowID string) {
	c.WorkflowID = workflowID
	c.UpdatedAt = time.Now()
}

// SetBacenClaimID define o ID do claim no Bacen
func (c *Claim) SetBacenClaimID(bacenClaimID string) {
	c.BacenClaimID = bacenClaimID
	c.UpdatedAt = time.Now()
}
