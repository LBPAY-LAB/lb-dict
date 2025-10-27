package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// PortabilityStatus representa o status de uma portabilidade
type PortabilityStatus string

const (
	PortabilityStatusInitiated       PortabilityStatus = "INITIATED"
	PortabilityStatusPendingApproval PortabilityStatus = "PENDING_APPROVAL"
	PortabilityStatusApproved        PortabilityStatus = "APPROVED"
	PortabilityStatusRejected        PortabilityStatus = "REJECTED"
	PortabilityStatusCompleted       PortabilityStatus = "COMPLETED"
	PortabilityStatusCancelled       PortabilityStatus = "CANCELLED"
	PortabilityStatusFailed          PortabilityStatus = "FAILED"
)

// Portability representa uma portabilidade de chave PIX entre contas
type Portability struct {
	ID                    uuid.UUID
	EntryKey              string
	OriginParticipant     valueobjects.Participant
	DestinationParticipant valueobjects.Participant
	OriginAccountID       uuid.UUID
	DestinationAccountID  uuid.UUID
	Status                PortabilityStatus
	WorkflowID            string
	BacenPortabilityID    string
	RequiresOTP           bool
	OTPValidatedAt        *time.Time
	InitiatedAt           time.Time
	CompletedAt           *time.Time
	RejectionReason       string
	Metadata              map[string]interface{}
	CreatedAt             time.Time
	UpdatedAt             time.Time
}

// NewPortability cria uma nova portabilidade
func NewPortability(
	entryKey string,
	origin valueobjects.Participant,
	destination valueobjects.Participant,
	originAccountID uuid.UUID,
	destinationAccountID uuid.UUID,
	requiresOTP bool,
) (*Portability, error) {
	if entryKey == "" {
		return nil, errors.New("entry key cannot be empty")
	}
	if origin.ISPB == destination.ISPB {
		return nil, errors.New("origin and destination must be different participants")
	}
	if originAccountID == uuid.Nil || destinationAccountID == uuid.Nil {
		return nil, errors.New("account IDs cannot be nil")
	}

	now := time.Now()
	return &Portability{
		ID:                     uuid.New(),
		EntryKey:               entryKey,
		OriginParticipant:      origin,
		DestinationParticipant: destination,
		OriginAccountID:        originAccountID,
		DestinationAccountID:   destinationAccountID,
		Status:                 PortabilityStatusInitiated,
		RequiresOTP:            requiresOTP,
		InitiatedAt:            now,
		Metadata:               make(map[string]interface{}),
		CreatedAt:              now,
		UpdatedAt:              now,
	}, nil
}

// Validate valida as regras de negócio da portabilidade
func (p *Portability) Validate() error {
	if p.ID == uuid.Nil {
		return errors.New("portability ID cannot be nil")
	}
	if p.EntryKey == "" {
		return errors.New("entry key cannot be empty")
	}
	if p.OriginParticipant.ISPB == p.DestinationParticipant.ISPB {
		return errors.New("origin and destination must be different")
	}
	if p.OriginAccountID == uuid.Nil || p.DestinationAccountID == uuid.Nil {
		return errors.New("account IDs cannot be nil")
	}
	return nil
}

// ValidateOTP valida o OTP e marca como validado
func (p *Portability) ValidateOTP() error {
	if !p.RequiresOTP {
		return errors.New("OTP validation not required for this portability")
	}
	if p.OTPValidatedAt != nil {
		return errors.New("OTP already validated")
	}
	now := time.Now()
	p.OTPValidatedAt = &now
	p.UpdatedAt = now
	return nil
}

// Approve aprova a portabilidade
func (p *Portability) Approve() error {
	if p.Status != PortabilityStatusPendingApproval && p.Status != PortabilityStatusInitiated {
		return errors.New("portability must be pending approval or initiated")
	}
	if p.RequiresOTP && p.OTPValidatedAt == nil {
		return errors.New("OTP must be validated before approval")
	}
	p.Status = PortabilityStatusApproved
	p.UpdatedAt = time.Now()
	return nil
}

// Reject rejeita a portabilidade
func (p *Portability) Reject(reason string) error {
	if p.IsFinal() {
		return errors.New("cannot reject portability in final status")
	}
	p.Status = PortabilityStatusRejected
	p.RejectionReason = reason
	p.UpdatedAt = time.Now()
	return nil
}

// Complete completa a portabilidade com sucesso
func (p *Portability) Complete() error {
	if p.Status != PortabilityStatusApproved {
		return errors.New("portability must be approved before completion")
	}
	now := time.Now()
	p.Status = PortabilityStatusCompleted
	p.CompletedAt = &now
	p.UpdatedAt = now
	return nil
}

// Fail marca a portabilidade como falha
func (p *Portability) Fail(reason string) error {
	if p.IsFinal() {
		return errors.New("portability is already in final status")
	}
	p.Status = PortabilityStatusFailed
	p.RejectionReason = reason
	p.UpdatedAt = time.Now()
	return nil
}

// Cancel cancela a portabilidade
func (p *Portability) Cancel(reason string) error {
	if p.IsFinal() {
		return errors.New("cannot cancel portability in final status")
	}
	p.Status = PortabilityStatusCancelled
	p.RejectionReason = reason
	p.UpdatedAt = time.Now()
	return nil
}

// SetPendingApproval marca como aguardando aprovação
func (p *Portability) SetPendingApproval() error {
	if p.Status != PortabilityStatusInitiated {
		return errors.New("can only set pending approval from initiated status")
	}
	p.Status = PortabilityStatusPendingApproval
	p.UpdatedAt = time.Now()
	return nil
}

// IsFinal verifica se o status é final
func (p *Portability) IsFinal() bool {
	finalStatuses := map[PortabilityStatus]bool{
		PortabilityStatusCompleted: true,
		PortabilityStatusRejected:  true,
		PortabilityStatusCancelled: true,
		PortabilityStatusFailed:    true,
	}
	return finalStatuses[p.Status]
}

// SetWorkflowID define o ID do workflow Temporal
func (p *Portability) SetWorkflowID(workflowID string) {
	p.WorkflowID = workflowID
	p.UpdatedAt = time.Now()
}

// SetBacenPortabilityID define o ID da portabilidade no Bacen
func (p *Portability) SetBacenPortabilityID(bacenID string) {
	p.BacenPortabilityID = bacenID
	p.UpdatedAt = time.Now()
}
