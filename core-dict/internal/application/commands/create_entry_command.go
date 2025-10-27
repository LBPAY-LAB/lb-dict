package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
	"github.com/lbpay-lab/core-dict/internal/application/services"
)

// CreateEntryCommand representa o comando para criar uma nova chave PIX
type CreateEntryCommand struct {
	KeyType       entities.KeyType
	KeyValue      string
	AccountID     uuid.UUID
	AccountISPB   string
	AccountBranch string
	AccountNumber string
	AccountType   string
	OwnerType     string
	OwnerTaxID    string
	OwnerName     string
	RequestedBy   uuid.UUID
	OTP           string // Para validação Email/Phone
}

// CreateEntryResult resultado do comando
type CreateEntryResult struct {
	EntryID   uuid.UUID
	Status    string
	CreatedAt time.Time
}

// CreateEntryCommandHandler handler para criação de chave PIX
type CreateEntryCommandHandler struct {
	entryRepo        repositories.EntryRepository
	eventPublisher   EventPublisher
	keyValidator     KeyValidatorService
	ownershipChecker OwnershipService
	duplicateChecker DuplicateCheckerService
	cacheService     services.CacheService
	connectClient    services.ConnectClient // NEW: gRPC client for RSFN operations
	entryProducer    EntryEventProducer     // NEW: Pulsar event producer
}

// NewCreateEntryCommandHandler cria nova instância do handler
func NewCreateEntryCommandHandler(
	entryRepo repositories.EntryRepository,
	eventPublisher EventPublisher,
	keyValidator KeyValidatorService,
	ownershipChecker OwnershipService,
	duplicateChecker DuplicateCheckerService,
	cacheService services.CacheService,
	connectClient services.ConnectClient,
	entryProducer EntryEventProducer,
) *CreateEntryCommandHandler {
	return &CreateEntryCommandHandler{
		entryRepo:        entryRepo,
		eventPublisher:   eventPublisher,
		keyValidator:     keyValidator,
		ownershipChecker: ownershipChecker,
		duplicateChecker: duplicateChecker,
		cacheService:     cacheService,
		connectClient:    connectClient,
		entryProducer:    entryProducer,
	}
}

// Handle executa o comando de criação de chave PIX
func (h *CreateEntryCommandHandler) Handle(ctx context.Context, cmd CreateEntryCommand) (*CreateEntryResult, error) {
	// 1. Validar formato da chave
	if err := h.keyValidator.ValidateFormat(cmd.KeyType, cmd.KeyValue); err != nil {
		return nil, errors.New("invalid key format: " + err.Error())
	}

	// 2. Validar ownership (CPF/CNPJ deve pertencer ao titular da conta)
	if err := h.ownershipChecker.ValidateOwnership(ctx, cmd.KeyType, cmd.KeyValue, cmd.OwnerTaxID); err != nil {
		return nil, errors.New("ownership validation failed: " + err.Error())
	}

	// 3. Verificar duplicação local (chave já existe neste PSP?)
	isDuplicate, err := h.duplicateChecker.IsDuplicate(ctx, cmd.KeyValue)
	if err != nil {
		return nil, errors.New("duplicate check failed: " + err.Error())
	}
	if isDuplicate {
		return nil, errors.New("key already registered in this PSP")
	}

	// 3a. Verificar duplicação GLOBAL via Connect (consulta RSFN)
	if h.connectClient != nil {
		existingEntry, err := h.connectClient.GetEntryByKey(ctx, cmd.KeyValue)
		if err == nil && existingEntry != nil {
			return nil, errors.New("key already registered in RSFN DICT")
		}
		// If error is "not found", continue (key is available)
	}

	// 4. Validar limites (max 5 CPF, 20 CNPJ, etc.)
	if err := h.keyValidator.ValidateLimits(ctx, cmd.KeyType, cmd.OwnerTaxID); err != nil {
		return nil, errors.New("key limit exceeded: " + err.Error())
	}

	// 5. Criar entidade Entry (Domain Layer)
	now := time.Now()
	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       cmd.KeyType,
		KeyValue:      cmd.KeyValue,
		Status:        entities.KeyStatusPending,
		AccountID:     cmd.AccountID,
		ISPB:          cmd.AccountISPB,
		Branch:        cmd.AccountBranch,
		AccountNumber: cmd.AccountNumber,
		AccountType:   cmd.AccountType,
		OwnerName:     cmd.OwnerName,
		OwnerTaxID:    cmd.OwnerTaxID,
		OwnerType:     cmd.OwnerType,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// 6. Persistir no repositório
	if err := h.entryRepo.Create(ctx, entry); err != nil {
		return nil, errors.New("failed to create entry: " + err.Error())
	}

	// 7. Publicar evento de domínio via Pulsar (EntryCreated)
	// This event triggers: Connect → Bridge → Bacen DICT registration
	// Non-blocking: uses goroutine to avoid blocking user response
	if h.entryProducer != nil {
		go func() {
			bgCtx := context.Background()
			// TODO: Convert Entry to proper domain entity format
			// For now, publish using legacy event publisher
			if err := h.eventPublisher.Publish(bgCtx, DomainEvent{
				EventType:     "EntryCreated",
				AggregateID:   entry.ID.String(),
				AggregateType: "Entry",
				OccurredAt:    time.Now(),
				Payload: map[string]interface{}{
					"entry_id":       entry.ID.String(),
					"key_type":       string(entry.KeyType),
					"key_value":      entry.KeyValue,
					"ispb":           entry.ISPB,
					"account_branch": entry.Branch,
					"account_number": entry.AccountNumber,
					"owner_name":     entry.OwnerName,
					"owner_tax_id":   entry.OwnerTaxID,
				},
			}); err != nil {
				// Log error but don't fail the request
				// TODO: Use proper logger
				// Background worker will retry failed events
			}
		}()
	}

	// 8. Invalidar cache
	h.cacheService.Delete(ctx, "entry:"+cmd.KeyValue)

	return &CreateEntryResult{
		EntryID:   entry.ID,
		Status:    string(entry.Status),
		CreatedAt: entry.CreatedAt,
	}, nil
}

// DomainEvent represents an event in the domain
type DomainEvent struct {
	EventType     string
	AggregateID   string
	AggregateType string
	OccurredAt    time.Time
	Payload       map[string]interface{}
}

// Service interfaces
type EventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
}

type KeyValidatorService interface {
	ValidateFormat(keyType entities.KeyType, keyValue string) error
	ValidateLimits(ctx context.Context, keyType entities.KeyType, ownerTaxID string) error
}

type OwnershipService interface {
	ValidateOwnership(ctx context.Context, keyType entities.KeyType, keyValue, ownerTaxID string) error
}

type DuplicateCheckerService interface {
	IsDuplicate(ctx context.Context, keyValue string) (bool, error)
}

// EntryEventProducer interface for publishing events to Pulsar
type EntryEventProducer interface {
	PublishCreated(ctx context.Context, entry interface{}, userID string) error
	PublishUpdated(ctx context.Context, entry interface{}, userID string) error
	PublishDeleted(ctx context.Context, entryID, keyValue, reason, userID string) error
}
