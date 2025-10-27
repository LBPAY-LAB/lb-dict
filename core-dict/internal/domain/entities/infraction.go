package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// InfractionType representa o tipo de infração
type InfractionType string

const (
	InfractionTypeFraud             InfractionType = "FRAUD"
	InfractionTypeDataMismatch      InfractionType = "DATA_MISMATCH"
	InfractionTypeUnauthorizedKey   InfractionType = "UNAUTHORIZED_KEY"
	InfractionTypeAccountClosed     InfractionType = "ACCOUNT_CLOSED"
	InfractionTypeInvalidAccount    InfractionType = "INVALID_ACCOUNT"
	InfractionTypeKeyOwnershipIssue InfractionType = "KEY_OWNERSHIP_ISSUE"
)

// InfractionStatus representa o status de uma infração
type InfractionStatus string

const (
	InfractionStatusReported     InfractionStatus = "REPORTED"
	InfractionStatusUnderReview  InfractionStatus = "UNDER_REVIEW"
	InfractionStatusConfirmed    InfractionStatus = "CONFIRMED"
	InfractionStatusRejected     InfractionStatus = "REJECTED"
	InfractionStatusResolved     InfractionStatus = "RESOLVED"
	InfractionStatusEscalated    InfractionStatus = "ESCALATED"
)

// Infraction representa uma infração reportada no DICT
type Infraction struct {
	ID                   uuid.UUID
	EntryKey             string
	Type                 InfractionType
	Status               InfractionStatus
	ReporterParticipant  valueobjects.Participant
	ReportedParticipant  valueobjects.Participant
	BacenInfractionID    string
	Description          string
	Evidence             map[string]interface{}
	Resolution           string
	ResolvedAt           *time.Time
	Metadata             map[string]interface{}
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// NewInfraction cria uma nova infração
func NewInfraction(
	entryKey string,
	infractionType InfractionType,
	reporter valueobjects.Participant,
	reported valueobjects.Participant,
	description string,
) (*Infraction, error) {
	if entryKey == "" {
		return nil, errors.New("entry key cannot be empty")
	}
	if err := validateInfractionType(infractionType); err != nil {
		return nil, err
	}
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}
	if reporter.ISPB == reported.ISPB {
		return nil, errors.New("reporter and reported must be different participants")
	}

	now := time.Now()
	return &Infraction{
		ID:                  uuid.New(),
		EntryKey:            entryKey,
		Type:                infractionType,
		Status:              InfractionStatusReported,
		ReporterParticipant: reporter,
		ReportedParticipant: reported,
		Description:         description,
		Evidence:            make(map[string]interface{}),
		Metadata:            make(map[string]interface{}),
		CreatedAt:           now,
		UpdatedAt:           now,
	}, nil
}

// Validate valida as regras de negócio da infração
func (i *Infraction) Validate() error {
	if i.ID == uuid.Nil {
		return errors.New("infraction ID cannot be nil")
	}
	if i.EntryKey == "" {
		return errors.New("entry key cannot be empty")
	}
	if err := validateInfractionType(i.Type); err != nil {
		return err
	}
	if err := validateInfractionStatus(i.Status); err != nil {
		return err
	}
	if i.Description == "" {
		return errors.New("description cannot be empty")
	}
	if i.ReporterParticipant.ISPB == i.ReportedParticipant.ISPB {
		return errors.New("reporter and reported must be different")
	}
	return nil
}

// StartReview inicia a revisão da infração
func (i *Infraction) StartReview() error {
	if i.Status != InfractionStatusReported {
		return errors.New("can only start review on reported infractions")
	}
	i.Status = InfractionStatusUnderReview
	i.UpdatedAt = time.Now()
	return nil
}

// Confirm confirma a infração
func (i *Infraction) Confirm() error {
	if i.Status != InfractionStatusUnderReview {
		return errors.New("infraction must be under review to confirm")
	}
	i.Status = InfractionStatusConfirmed
	i.UpdatedAt = time.Now()
	return nil
}

// Reject rejeita a infração
func (i *Infraction) Reject(reason string) error {
	if i.IsFinal() {
		return errors.New("cannot reject infraction in final status")
	}
	now := time.Now()
	i.Status = InfractionStatusRejected
	i.Resolution = reason
	i.ResolvedAt = &now
	i.UpdatedAt = now
	return nil
}

// Resolve resolve a infração
func (i *Infraction) Resolve(resolution string) error {
	if i.Status != InfractionStatusConfirmed && i.Status != InfractionStatusEscalated {
		return errors.New("can only resolve confirmed or escalated infractions")
	}
	now := time.Now()
	i.Status = InfractionStatusResolved
	i.Resolution = resolution
	i.ResolvedAt = &now
	i.UpdatedAt = now
	return nil
}

// Escalate escala a infração para níveis superiores
func (i *Infraction) Escalate() error {
	if i.IsFinal() {
		return errors.New("cannot escalate infraction in final status")
	}
	i.Status = InfractionStatusEscalated
	i.UpdatedAt = time.Now()
	return nil
}

// AddEvidence adiciona evidência à infração
func (i *Infraction) AddEvidence(key string, value interface{}) {
	i.Evidence[key] = value
	i.UpdatedAt = time.Now()
}

// IsFinal verifica se o status é final
func (i *Infraction) IsFinal() bool {
	finalStatuses := map[InfractionStatus]bool{
		InfractionStatusResolved: true,
		InfractionStatusRejected: true,
	}
	return finalStatuses[i.Status]
}

// SetBacenInfractionID define o ID da infração no Bacen
func (i *Infraction) SetBacenInfractionID(bacenID string) {
	i.BacenInfractionID = bacenID
	i.UpdatedAt = time.Now()
}

func validateInfractionType(t InfractionType) error {
	validTypes := map[InfractionType]bool{
		InfractionTypeFraud:             true,
		InfractionTypeDataMismatch:      true,
		InfractionTypeUnauthorizedKey:   true,
		InfractionTypeAccountClosed:     true,
		InfractionTypeInvalidAccount:    true,
		InfractionTypeKeyOwnershipIssue: true,
	}
	if !validTypes[t] {
		return errors.New("invalid infraction type")
	}
	return nil
}

func validateInfractionStatus(s InfractionStatus) error {
	validStatuses := map[InfractionStatus]bool{
		InfractionStatusReported:    true,
		InfractionStatusUnderReview: true,
		InfractionStatusConfirmed:   true,
		InfractionStatusRejected:    true,
		InfractionStatusResolved:    true,
		InfractionStatusEscalated:   true,
	}
	if !validStatuses[s] {
		return errors.New("invalid infraction status")
	}
	return nil
}
