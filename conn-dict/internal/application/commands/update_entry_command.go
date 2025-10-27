package commands

import (
	"context"
	"fmt"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// UpdateEntryCommand represents a command to update an existing DICT entry
type UpdateEntryCommand struct {
	EntryID           string
	AccountBranch     *string
	AccountNumber     *string
	AccountType       *entities.AccountType
	AccountOpenedDate *time.Time
	OwnerName         *string
	OwnerTaxID        *string
	BacenEntryID      *string
}

// UpdateEntryHandler handles the UpdateEntryCommand
type UpdateEntryHandler struct {
	repo   *repositories.EntryRepository
	logger *logrus.Logger
}

// NewUpdateEntryHandler creates a new UpdateEntryHandler
func NewUpdateEntryHandler(repo *repositories.EntryRepository, logger *logrus.Logger) *UpdateEntryHandler {
	return &UpdateEntryHandler{
		repo:   repo,
		logger: logger,
	}
}

// Handle executes the UpdateEntryCommand
func (h *UpdateEntryHandler) Handle(ctx context.Context, cmd *UpdateEntryCommand) error {
	h.logger.WithFields(logrus.Fields{
		"entry_id": cmd.EntryID,
	}).Info("Handling UpdateEntryCommand")

	// Validate command inputs
	if err := h.validateCommand(cmd); err != nil {
		h.logger.WithError(err).Error("UpdateEntryCommand validation failed")
		return fmt.Errorf("invalid command: %w", err)
	}

	// Retrieve existing entry
	entry, err := h.repo.GetByEntryID(ctx, cmd.EntryID)
	if err != nil {
		h.logger.WithError(err).Errorf("Entry not found: %s", cmd.EntryID)
		return fmt.Errorf("entry not found: %w", err)
	}

	// Check if entry can be updated (not blocked or deleted)
	if entry.IsBlocked() {
		h.logger.WithField("entry_id", cmd.EntryID).Warn("Cannot update blocked entry")
		return fmt.Errorf("cannot update blocked entry: %s", cmd.EntryID)
	}

	if entry.DeletedAt != nil {
		h.logger.WithField("entry_id", cmd.EntryID).Warn("Cannot update deleted entry")
		return fmt.Errorf("cannot update deleted entry: %s", cmd.EntryID)
	}

	// Update fields
	if cmd.AccountBranch != nil {
		entry.AccountBranch = cmd.AccountBranch
	}

	if cmd.AccountNumber != nil {
		entry.AccountNumber = cmd.AccountNumber
	}

	if cmd.AccountType != nil {
		entry.AccountType = *cmd.AccountType
	}

	if cmd.AccountOpenedDate != nil {
		entry.AccountOpenedDate = cmd.AccountOpenedDate
	}

	if cmd.OwnerName != nil {
		entry.OwnerName = cmd.OwnerName
	}

	if cmd.OwnerTaxID != nil {
		entry.OwnerTaxID = cmd.OwnerTaxID
	}

	if cmd.BacenEntryID != nil {
		entry.BacenEntryID = cmd.BacenEntryID
	}

	// Update timestamp
	entry.UpdatedAt = time.Now()

	// Persist changes
	if err := h.repo.Update(ctx, entry); err != nil {
		h.logger.WithError(err).Error("Failed to update entry")
		return fmt.Errorf("failed to update entry: %w", err)
	}

	h.logger.WithFields(logrus.Fields{
		"entry_id": entry.EntryID,
		"status":   entry.Status,
	}).Info("Entry updated successfully")

	return nil
}

// validateCommand validates the UpdateEntryCommand inputs
func (h *UpdateEntryHandler) validateCommand(cmd *UpdateEntryCommand) error {
	if cmd.EntryID == "" {
		return fmt.Errorf("entry_id is required")
	}

	// Validate account type if provided
	if cmd.AccountType != nil {
		validAccountTypes := map[entities.AccountType]bool{
			entities.AccountTypeCACC: true,
			entities.AccountTypeSLRY: true,
			entities.AccountTypeSVGS: true,
			entities.AccountTypeTRAN: true,
		}
		if !validAccountTypes[*cmd.AccountType] {
			return fmt.Errorf("invalid account_type: %s", *cmd.AccountType)
		}
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