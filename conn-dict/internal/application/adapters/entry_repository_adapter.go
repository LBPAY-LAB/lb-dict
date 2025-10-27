package adapters

import (
	"context"

	"github.com/google/uuid"
	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/lbpay-lab/conn-dict/internal/application/usecases"
	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
)

// EntryRepositoryAdapter adapts the infrastructure EntryRepository to the use case interface
type EntryRepositoryAdapter struct {
	repo *repositories.EntryRepository
}

// NewEntryRepositoryAdapter creates a new adapter
func NewEntryRepositoryAdapter(repo *repositories.EntryRepository) *EntryRepositoryAdapter {
	return &EntryRepositoryAdapter{repo: repo}
}

// Create adapts the Create method
func (a *EntryRepositoryAdapter) Create(ctx context.Context, entry *usecases.Entry) error {
	domainEntry := &entities.Entry{
		ID:                   uuid.New(),
		EntryID:              entry.EntryID,
		Key:                  entry.KeyValue,
		KeyType:              a.mapKeyType(entry.KeyType),
		Participant:          entry.AccountISPB,
		AccountBranch:        strPtr(entry.BranchCode),
		AccountNumber:        strPtr(entry.AccountNum),
		AccountType:          a.mapAccountType(entry.AccountType),
		OwnerType:            entities.OwnerTypeNaturalPerson, // Default to natural person
		OwnerName:            strPtr(entry.HolderName),
		OwnerTaxID:           strPtr(entry.HolderDoc),
		Status:               a.mapEntryStatus(entry.Status),
		BacenEntryID:         strPtr(entry.ExternalID),
		CreatedAt:            entry.CreatedAt,
		UpdatedAt:            entry.UpdatedAt,
	}

	return a.repo.Create(ctx, domainEntry)
}

// GetByID adapts the GetByID method
func (a *EntryRepositoryAdapter) GetByID(ctx context.Context, entryID string) (*usecases.Entry, error) {
	domainEntry, err := a.repo.GetByID(ctx, entryID)
	if err != nil {
		return nil, err
	}

	return a.toUseCaseEntry(domainEntry), nil
}

// GetByKey adapts the GetByKey method
func (a *EntryRepositoryAdapter) GetByKey(ctx context.Context, keyType commonv1.KeyType, keyValue string) (*usecases.Entry, error) {
	// Note: Repository GetByKey only takes key value, not key type
	domainEntry, err := a.repo.GetByKey(ctx, keyValue)
	if err != nil {
		return nil, err
	}

	return a.toUseCaseEntry(domainEntry), nil
}

// Update adapts the Update method
func (a *EntryRepositoryAdapter) Update(ctx context.Context, entry *usecases.Entry) error {
	domainEntry := &entities.Entry{
		ID:                   uuid.MustParse(entry.EntryID),
		EntryID:              entry.EntryID,
		Key:                  entry.KeyValue,
		KeyType:              a.mapKeyType(entry.KeyType),
		Participant:          entry.AccountISPB,
		AccountBranch:        strPtr(entry.BranchCode),
		AccountNumber:        strPtr(entry.AccountNum),
		AccountType:          a.mapAccountType(entry.AccountType),
		OwnerType:            entities.OwnerTypeNaturalPerson,
		OwnerName:            strPtr(entry.HolderName),
		OwnerTaxID:           strPtr(entry.HolderDoc),
		Status:               a.mapEntryStatus(entry.Status),
		BacenEntryID:         strPtr(entry.ExternalID),
		UpdatedAt:            entry.UpdatedAt,
	}

	return a.repo.Update(ctx, domainEntry)
}

// SoftDelete adapts the Delete method (repository uses hard delete)
func (a *EntryRepositoryAdapter) SoftDelete(ctx context.Context, entryID string) error {
	return a.repo.Delete(ctx, entryID)
}

// Helper methods to convert between domain and proto types

func (a *EntryRepositoryAdapter) toUseCaseEntry(e *entities.Entry) *usecases.Entry {
	if e == nil {
		return nil
	}

	return &usecases.Entry{
		EntryID:     e.EntryID,
		ExternalID:  derefStr(e.BacenEntryID),
		KeyType:     a.reverseMapKeyType(string(e.KeyType)),
		KeyValue:    e.Key,
		AccountISPB: e.Participant,
		AccountType: a.reverseMapAccountType(string(e.AccountType)),
		AccountNum:  derefStr(e.AccountNumber),
		BranchCode:  derefStr(e.AccountBranch),
		HolderName:  derefStr(e.OwnerName),
		HolderDoc:   derefStr(e.OwnerTaxID),
		Status:      a.reverseMapEntryStatus(string(e.Status)),
		CreatedAt:   e.CreatedAt,
		UpdatedAt:   e.UpdatedAt,
		DeletedAt:   e.DeletedAt,
	}
}

func (a *EntryRepositoryAdapter) mapKeyType(kt commonv1.KeyType) entities.KeyType {
	switch kt {
	case commonv1.KeyType_KEY_TYPE_CPF:
		return entities.KeyTypeCPF
	case commonv1.KeyType_KEY_TYPE_CNPJ:
		return entities.KeyTypeCNPJ
	case commonv1.KeyType_KEY_TYPE_EMAIL:
		return entities.KeyTypeEMAIL
	case commonv1.KeyType_KEY_TYPE_PHONE:
		return entities.KeyTypePHONE
	case commonv1.KeyType_KEY_TYPE_EVP:
		return entities.KeyTypeEVP
	default:
		return entities.KeyTypeCPF // Default
	}
}

func (a *EntryRepositoryAdapter) reverseMapKeyType(kt string) commonv1.KeyType {
	switch entities.KeyType(kt) {
	case entities.KeyTypeCPF:
		return commonv1.KeyType_KEY_TYPE_CPF
	case entities.KeyTypeCNPJ:
		return commonv1.KeyType_KEY_TYPE_CNPJ
	case entities.KeyTypeEMAIL:
		return commonv1.KeyType_KEY_TYPE_EMAIL
	case entities.KeyTypePHONE:
		return commonv1.KeyType_KEY_TYPE_PHONE
	case entities.KeyTypeEVP:
		return commonv1.KeyType_KEY_TYPE_EVP
	default:
		return commonv1.KeyType_KEY_TYPE_UNSPECIFIED
	}
}

func (a *EntryRepositoryAdapter) mapAccountType(at commonv1.AccountType) entities.AccountType {
	switch at {
	case commonv1.AccountType_ACCOUNT_TYPE_CHECKING:
		return entities.AccountTypeCACC
	case commonv1.AccountType_ACCOUNT_TYPE_SAVINGS:
		return entities.AccountTypeSVGS
	case commonv1.AccountType_ACCOUNT_TYPE_PAYMENT:
		return entities.AccountTypeTRAN
	case commonv1.AccountType_ACCOUNT_TYPE_SALARY:
		return entities.AccountTypeSLRY
	default:
		return entities.AccountTypeCACC // Default
	}
}

func (a *EntryRepositoryAdapter) reverseMapAccountType(at string) commonv1.AccountType {
	switch entities.AccountType(at) {
	case entities.AccountTypeCACC:
		return commonv1.AccountType_ACCOUNT_TYPE_CHECKING
	case entities.AccountTypeSVGS:
		return commonv1.AccountType_ACCOUNT_TYPE_SAVINGS
	case entities.AccountTypeTRAN:
		return commonv1.AccountType_ACCOUNT_TYPE_PAYMENT
	case entities.AccountTypeSLRY:
		return commonv1.AccountType_ACCOUNT_TYPE_SALARY
	default:
		return commonv1.AccountType_ACCOUNT_TYPE_UNSPECIFIED
	}
}

func (a *EntryRepositoryAdapter) mapEntryStatus(s commonv1.EntryStatus) entities.EntryStatus {
	switch s {
	case commonv1.EntryStatus_ENTRY_STATUS_ACTIVE:
		return entities.EntryStatusActive
	case commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING:
		return entities.EntryStatusPortabilityPending
	case commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_CONFIRMED:
		return entities.EntryStatusPortabilityPending // Map to closest match
	case commonv1.EntryStatus_ENTRY_STATUS_CLAIM_PENDING:
		return entities.EntryStatusOwnershipChangePending
	case commonv1.EntryStatus_ENTRY_STATUS_DELETED:
		return entities.EntryStatusInactive
	default:
		return entities.EntryStatusActive
	}
}

func (a *EntryRepositoryAdapter) reverseMapEntryStatus(s string) commonv1.EntryStatus {
	switch entities.EntryStatus(s) {
	case entities.EntryStatusActive:
		return commonv1.EntryStatus_ENTRY_STATUS_ACTIVE
	case entities.EntryStatusInactive:
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED
	case entities.EntryStatusBlocked:
		return commonv1.EntryStatus_ENTRY_STATUS_DELETED
	case entities.EntryStatusPortabilityPending:
		return commonv1.EntryStatus_ENTRY_STATUS_PORTABILITY_PENDING
	case entities.EntryStatusOwnershipChangePending:
		return commonv1.EntryStatus_ENTRY_STATUS_CLAIM_PENDING
	default:
		return commonv1.EntryStatus_ENTRY_STATUS_UNSPECIFIED
	}
}

// Helper functions for pointer conversion
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func derefStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
