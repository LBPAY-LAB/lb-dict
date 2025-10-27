package commands

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

// CreateEntryCommand representa o comando para criar uma nova chave PIX
type CreateEntryCommand struct {
	KeyType       KeyType
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
	entryRepo        EntryRepository
	eventPublisher   EventPublisher
	keyValidator     KeyValidatorService
	ownershipChecker OwnershipService
	duplicateChecker DuplicateCheckerService
	cacheService     CacheService
	connectClient    ConnectClient      // NEW: gRPC client for RSFN operations
	entryProducer    EntryEventProducer // NEW: Pulsar event producer
}

// NewCreateEntryCommandHandler cria nova instância do handler
func NewCreateEntryCommandHandler(
	entryRepo EntryRepository,
	eventPublisher EventPublisher,
	keyValidator KeyValidatorService,
	ownershipChecker OwnershipService,
	duplicateChecker DuplicateCheckerService,
	cacheService CacheService,
	connectClient ConnectClient,
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
	entry := &Entry{
		ID:        uuid.New(),
		KeyType:   cmd.KeyType,
		KeyValue:  cmd.KeyValue,
		Status:    "PENDING",
		AccountID: cmd.AccountID,
		Account: Account{
			ISPB:          cmd.AccountISPB,
			Branch:        cmd.AccountBranch,
			AccountNumber: cmd.AccountNumber,
			AccountType:   cmd.AccountType,
		},
		Owner: Owner{
			Type:  cmd.OwnerType,
			TaxID: cmd.OwnerTaxID,
			Name:  cmd.OwnerName,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
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
					"ispb":           entry.Account.ISPB,
					"account_branch": entry.Account.Branch,
					"account_number": entry.Account.AccountNumber,
					"owner_name":     entry.Owner.Name,
					"owner_tax_id":   entry.Owner.TaxID,
				},
			}); err != nil {
				// Log error but don't fail the request
				// TODO: Use proper logger
				// Background worker will retry failed events
			}
		}()
	}

	// 8. Invalidar cache
	h.cacheService.InvalidateKey(ctx, "entry:"+cmd.KeyValue)

	return &CreateEntryResult{
		EntryID:   entry.ID,
		Status:    entry.Status,
		CreatedAt: entry.CreatedAt,
	}, nil
}

// Temporary interfaces (Domain Layer será implementado por outro agente)
type KeyType string

const (
	KeyTypeCPF   KeyType = "CPF"
	KeyTypeCNPJ  KeyType = "CNPJ"
	KeyTypeEmail KeyType = "EMAIL"
	KeyTypePhone KeyType = "PHONE"
	KeyTypeEVP   KeyType = "EVP"
)

type Entry struct {
	ID        uuid.UUID
	KeyType   KeyType
	KeyValue  string
	Status    string
	AccountID uuid.UUID
	Account   Account
	Owner     Owner
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Account struct {
	ISPB          string
	Branch        string
	AccountNumber string
	AccountType   string
}

type Owner struct {
	Type  string
	TaxID string
	Name  string
}

type DomainEvent struct {
	EventType     string
	AggregateID   string
	AggregateType string
	OccurredAt    time.Time
	Payload       map[string]interface{}
}

// Repository interfaces (temporárias)
type EntryRepository interface {
	Create(ctx context.Context, entry *Entry) error
	FindByID(ctx context.Context, id uuid.UUID) (*Entry, error)
	FindByKeyValue(ctx context.Context, keyValue string) (*Entry, error)
	Update(ctx context.Context, entry *Entry) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByOwnerAndType(ctx context.Context, ownerTaxID string, keyType KeyType) (int, error)
}

// Service interfaces (temporárias)
type EventPublisher interface {
	Publish(ctx context.Context, event DomainEvent) error
}

type KeyValidatorService interface {
	ValidateFormat(keyType KeyType, keyValue string) error
	ValidateLimits(ctx context.Context, keyType KeyType, ownerTaxID string) error
}

type OwnershipService interface {
	ValidateOwnership(ctx context.Context, keyType KeyType, keyValue, ownerTaxID string) error
}

type DuplicateCheckerService interface {
	IsDuplicate(ctx context.Context, keyValue string) (bool, error)
}

type CacheService interface {
	Get(ctx context.Context, key string) (interface{}, error)
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	InvalidateKey(ctx context.Context, key string) error
	InvalidatePattern(ctx context.Context, pattern string) error
}

// ConnectClient interface for gRPC communication with conn-dict service
type ConnectClient interface {
	GetEntryByKey(ctx context.Context, keyValue string) (interface{}, error)
	CreateEntry(ctx context.Context, keyType, keyValue, accountISPB string) (string, error)
	UpdateEntry(ctx context.Context, entryID, newAccountISPB string) error
	DeleteEntry(ctx context.Context, entryID, reason string) error
}

// EntryEventProducer interface for publishing events to Pulsar
type EntryEventProducer interface {
	PublishCreated(ctx context.Context, entry interface{}, userID string) error
	PublishUpdated(ctx context.Context, entry interface{}, userID string) error
	PublishDeleted(ctx context.Context, entryID, keyValue, reason, userID string) error
}
