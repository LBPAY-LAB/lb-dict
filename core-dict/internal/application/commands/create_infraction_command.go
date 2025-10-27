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

// CreateInfractionCommand comando para criar infração (violação de regras PIX)
type CreateInfractionCommand struct {
	EntryID          uuid.UUID
	InfractionType   InfractionType
	Description      string
	ReportedBy       string // ISPB do PSP que reportou
	Severity         string // LOW, MEDIUM, HIGH, CRITICAL
	BacenInfractionID string // ID retornado pelo Bacen
}

type InfractionType string

const (
	InfractionTypeFraud          InfractionType = "FRAUD"
	InfractionTypeSpam           InfractionType = "SPAM"
	InfractionTypeUnauthorizedUse InfractionType = "UNAUTHORIZED_USE"
	InfractionTypeOther          InfractionType = "OTHER"
)

// CreateInfractionResult resultado do comando
type CreateInfractionResult struct {
	InfractionID uuid.UUID
	Status       string
	CreatedAt    time.Time
}

// CreateInfractionCommandHandler handler para criar infração
type CreateInfractionCommandHandler struct {
	entryRepo       repositories.EntryRepository
	infractionRepo  repositories.InfractionRepository
	eventPublisher  EventPublisher
}

// NewCreateInfractionCommandHandler cria nova instância
func NewCreateInfractionCommandHandler(
	entryRepo repositories.EntryRepository,
	infractionRepo repositories.InfractionRepository,
	eventPublisher EventPublisher,
) *CreateInfractionCommandHandler {
	return &CreateInfractionCommandHandler{
		entryRepo:      entryRepo,
		infractionRepo: infractionRepo,
		eventPublisher: eventPublisher,
	}
}

// Handle executa o comando
func (h *CreateInfractionCommandHandler) Handle(ctx context.Context, cmd CreateInfractionCommand) (*CreateInfractionResult, error) {
	// 1. Validar entrada
	if cmd.Description == "" {
		return nil, errors.New("description is required")
	}
	if cmd.BacenInfractionID == "" {
		return nil, errors.New("bacen_infraction_id is required")
	}

	// 2. Buscar entry
	entry, err := h.entryRepo.FindByID(ctx, cmd.EntryID)
	if err != nil {
		return nil, errors.New("entry not found")
	}

	// 3. Criar entidade Infraction usando domain factory
	reporter := valueobjects.Participant{
		ISPB: cmd.ReportedBy,
		Name: "", // TODO: buscar nome do participante
	}
	reported := valueobjects.Participant{
		ISPB: entry.ISPB,
		Name: "", // Nome do PSP dono da chave
	}

	// Converter tipo de infração
	infractionType := entities.InfractionType(cmd.InfractionType)

	infraction, err := entities.NewInfraction(
		entry.KeyValue,
		infractionType,
		reporter,
		reported,
		cmd.Description,
	)
	if err != nil {
		return nil, errors.New("failed to create infraction: " + err.Error())
	}

	// Adicionar dados adicionais
	infraction.BacenInfractionID = cmd.BacenInfractionID
	if cmd.Severity == "CRITICAL" {
		infraction.Status = entities.InfractionStatusEscalated
	}

	// 4. Persistir infraction
	if err := h.infractionRepo.Create(ctx, infraction); err != nil {
		return nil, errors.New("failed to create infraction: " + err.Error())
	}

	// 5. Se severidade CRITICAL, bloquear chave automaticamente
	now := time.Now()
	if cmd.Severity == "CRITICAL" {
		entry.Status = entities.KeyStatusBlocked
		entry.UpdatedAt = now
		if err := h.entryRepo.Update(ctx, entry); err != nil {
			return nil, errors.New("failed to block entry: " + err.Error())
		}
	}

	// 6. Publicar evento (para notificar compliance e usuário)
	event := DomainEvent{
		EventType:     "InfractionCreated",
		AggregateID:   infraction.ID.String(),
		AggregateType: "Infraction",
		OccurredAt:    now,
		Payload: map[string]interface{}{
			"infraction_id":       infraction.ID.String(),
			"entry_id":            entry.ID.String(),
			"key_value":           entry.KeyValue,
			"infraction_type":     string(cmd.InfractionType),
			"severity":            cmd.Severity,
			"reported_by":         cmd.ReportedBy,
			"bacen_infraction_id": cmd.BacenInfractionID,
		},
	}
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		return nil, errors.New("failed to publish event: " + err.Error())
	}

	return &CreateInfractionResult{
		InfractionID: infraction.ID,
		Status:       string(infraction.Status),
		CreatedAt:    infraction.CreatedAt,
	}, nil
}
