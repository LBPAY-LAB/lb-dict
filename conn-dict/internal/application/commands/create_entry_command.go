package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// CreateEntryCommand represents a command to create a new DICT entry
type CreateEntryCommand struct {
	EntryID           string
	Key               string
	KeyType           entities.KeyType
	Participant       string // ISPB (8 digits)
	AccountBranch     *string
	AccountNumber     *string
	AccountType       entities.AccountType
	AccountOpenedDate *time.Time
	OwnerType         entities.OwnerType
	OwnerName         *string
	OwnerTaxID        *string // CPF (11) or CNPJ (14)
	BacenEntryID      *string
}

// CreateEntryHandler handles the CreateEntryCommand
type CreateEntryHandler struct {
	repo   *repositories.EntryRepository
	logger *logrus.Logger
}

// NewCreateEntryHandler creates a new CreateEntryHandler
func NewCreateEntryHandler(repo *repositories.EntryRepository, logger *logrus.Logger) *CreateEntryHandler {
	return &CreateEntryHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the CreateEntryCommand
func (h *CreateEntryHandler) Handle(ctx context.Context, cmd *CreateEntryCommand) (string, error) {
	h.logger.WithFields(logrus.Fields{
		"entry_id":    cmd.EntryID,
		"key":         cmd.Key,
		"key_type":    cmd.KeyType,
		"participant": cmd.Participant,
	}).Info("Handling CreateEntryCommand")

	// Validate command inputs
	if err := h.validateCommand(cmd); err != nil {
		h.logger.WithError(err).Error("CreateEntryCommand validation failed")
		return "", fmt.Errorf("invalid command: %w", err)
	}

	// Check if key already exists and is active
	hasActiveKey, err := h.repo.HasActiveKey(ctx, cmd.Key)
	if err != nil {
		h.logger.WithError(err).Error("Failed to check if key is active")
		return "", fmt.Errorf("failed to check key availability: %w", err)
	}

	if hasActiveKey {
		h.logger.WithField("key", cmd.Key).Warn("Key already exists and is active")
		return "", fmt.Errorf("key already registered and active: %s", cmd.Key)
	}

	// Check if entryID is already in use
	existingEntry, err := h.repo.GetByEntryID(ctx, cmd.EntryID)
	if err == nil && existingEntry != nil {
		h.logger.WithField("entry_id", cmd.EntryID).Warn("EntryID already exists")
		return "", fmt.Errorf("entry_id already exists: %s", cmd.EntryID)
	}

	// Create new entry entity
	entry, err := entities.NewEntry(
		cmd.EntryID,
		cmd.Key,
		cmd.KeyType,
		cmd.Participant,
		cmd.AccountType,
		cmd.OwnerType,
	)
	if err != nil {
		h.logger.WithError(err).Error("Failed to create entry entity")
		return "", fmt.Errorf("failed to create entry entity: %w", err)
	}

	// Set optional fields
	entry.AccountBranch = cmd.AccountBranch
	entry.AccountNumber = cmd.AccountNumber
	entry.AccountOpenedDate = cmd.AccountOpenedDate
	entry.OwnerName = cmd.OwnerName
	entry.OwnerTaxID = cmd.OwnerTaxID
	entry.BacenEntryID = cmd.BacenEntryID

	// Persist entry
	if err := h.repo.Create(ctx, entry); err != nil {
		h.logger.WithError(err).Error("Failed to persist entry")
		return "", fmt.Errorf("failed to create entry: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"entry_id":  entry.EntryID,
		"uuid":      entry.ID.String(),
		"key":       entry.Key,
		"key_type":  entry.KeyType,
		"status":    entry.Status,
	}).Info("Entry created successfully")

	return entry.ID.String(), nil
}

// validateCommand validates the CreateEntryCommand inputs
func (h *CreateEntryHandler) validateCommand(cmd *CreateEntryCommand) error {
	if cmd.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}

	if cmd.Key == "" {
		return fmt.Errorf("key is required")
	}

	if cmd.KeyType == "" {
		return fmt.Errorf("key_type is required")
	}

	if cmd.Participant == "" {
		return fmt.Errorf("participant (ISPB) is required")
	}

	// Validate ISPB format (8 digits)
	if len(cmd.Participant) != 8 {
		return fmt.Errorf("participant must be 8 digits, got %d", len(cmd.Participant))
	}

	// Validate key type is valid
	validKeyTypes := map[entities.KeyType]bool{
		entities.KeyTypeCPF:   true,
		entities.KeyTypeCNPJ:  true,
		entities.KeyTypeEMAIL: true,
		entities.KeyTypePHONE: true,
		entities.KeyTypeEVP:   true,
	}
	if !validKeyTypes[cmd.KeyType] {
		return fmt.Errorf("invalid key_type: %s", cmd.KeyType)
	}

	// Validate account type is valid
	validAccountTypes := map[entities.AccountType]bool{
		entities.AccountTypeCACC: true,
		entities.AccountTypeSLRY: true,
		entities.AccountTypeSVGS: true,
		entities.AccountTypeTRAN: true,
	}
	if !validAccountTypes[cmd.AccountType] {
		return fmt.Errorf("invalid account_type: %s", cmd.AccountType)
	}

	// Validate owner type is valid
	validOwnerTypes := map[entities.OwnerType]bool{
		entities.OwnerTypeNaturalPerson: true,
		entities.OwnerTypeLegalPerson:   true,
	}
	if !validOwnerTypes[cmd.OwnerType] {
		return fmt.Errorf("invalid owner_type: %s", cmd.OwnerType)
	}

	// Validate owner tax ID format if provided
	if cmd.OwnerTaxID != nil {
		taxIDLen := len(*cmd.OwnerTaxID)
		if taxIDLen != 11 && taxIDLen != 14 {
			return fmt.Errorf("owner_tax_id must be 11 (CPF) or 14 (CNPJ) digits, got %d", taxIDLen)
		}
	}

	return nil
}