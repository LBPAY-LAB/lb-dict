package activities

import (
	"context"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/pulsar"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/repositories"
	"github.com/sirupsen/logrus"
)

// EntryActivities contains all Temporal activities for entry operations
type EntryActivities struct {
	logger         *logrus.Logger
	entryRepo      *repositories.EntryRepository
	pulsarProducer *pulsar.Producer
}

// NewEntryActivities creates a new instance of EntryActivities
func NewEntryActivities(
	logger *logrus.Logger,
	entryRepo *repositories.EntryRepository,
	pulsarProducer *pulsar.Producer,
) *EntryActivities {
	return &EntryActivities{
		logger:         logger,
		entryRepo:      entryRepo,
		pulsarProducer: pulsarProducer,
	}
}

// CreateEntryInput is the input for CreateEntryActivity
type CreateEntryInput struct {
	EntryID         string
	Key             string
	KeyType         string
	ParticipantISPB string
	AccountBranch   string
	AccountNumber   string
	AccountType     string
	OwnerType       string
	OwnerName       string
	OwnerTaxID      string
}

// CreateEntryActivity creates a new entry in the database
func (a *EntryActivities) CreateEntryActivity(ctx context.Context, input CreateEntryInput) error {
	a.logger.WithFields(logrus.Fields{
		"entry_id":    input.EntryID,
		"key":         input.Key,
		"key_type":    input.KeyType,
		"participant": input.ParticipantISPB,
	}).Info("Creating entry")

	// Convert string to KeyType enum
	var keyType entities.KeyType
	switch input.KeyType {
	case "CPF":
		keyType = entities.KeyTypeCPF
	case "CNPJ":
		keyType = entities.KeyTypeCNPJ
	case "EMAIL":
		keyType = entities.KeyTypeEMAIL
	case "PHONE":
		keyType = entities.KeyTypePHONE
	case "EVP":
		keyType = entities.KeyTypeEVP
	default:
		return fmt.Errorf("invalid key type: %s", input.KeyType)
	}

	// Convert string to AccountType enum
	var accountType entities.AccountType
	switch input.AccountType {
	case "CACC":
		accountType = entities.AccountTypeCACC
	case "SLRY":
		accountType = entities.AccountTypeSLRY
	case "SVGS":
		accountType = entities.AccountTypeSVGS
	case "TRAN":
		accountType = entities.AccountTypeTRAN
	default:
		return fmt.Errorf("invalid account type: %s", input.AccountType)
	}

	// Convert string to OwnerType enum
	var ownerType entities.OwnerType
	switch input.OwnerType {
	case "NATURAL_PERSON":
		ownerType = entities.OwnerTypeNaturalPerson
	case "LEGAL_PERSON":
		ownerType = entities.OwnerTypeLegalPerson
	default:
		return fmt.Errorf("invalid owner type: %s", input.OwnerType)
	}

	// Create entry entity with validation
	entry, err := entities.NewEntry(
		input.EntryID,
		input.Key,
		keyType,
		input.ParticipantISPB,
		accountType,
		ownerType,
	)
	if err != nil {
		a.logger.WithError(err).Error("Failed to create entry entity")
		return fmt.Errorf("invalid entry data: %w", err)
	}

	// Set optional account information
	if input.AccountBranch != "" {
		entry.AccountBranch = &input.AccountBranch
	}
	if input.AccountNumber != "" {
		entry.AccountNumber = &input.AccountNumber
	}

	// Set owner information
	if input.OwnerName != "" {
		entry.OwnerName = &input.OwnerName
	}
	if input.OwnerTaxID != "" {
		entry.OwnerTaxID = &input.OwnerTaxID
	}

	// Insert into database
	if err := a.entryRepo.Create(ctx, entry); err != nil {
		a.logger.WithError(err).Errorf("Failed to insert entry: %s", entry.EntryID)
		return fmt.Errorf("database error: %w", err)
	}

	// Publish entry created event to Pulsar
	event := map[string]interface{}{
		"event_type":   "entry_created",
		"entry_id":     entry.EntryID,
		"key":          entry.Key,
		"key_type":     entry.KeyType,
		"participant":  entry.Participant,
		"status":       entry.Status,
		"owner_name":   entry.OwnerName,
		"owner_tax_id": entry.OwnerTaxID,
		"created_at":   entry.CreatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, entry.EntryID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish entry created event (non-critical)")
		// Don't fail the activity if event publishing fails
	}

	a.logger.WithField("entry_id", entry.EntryID).Info("Entry created successfully")

	return nil
}

// UpdateEntryInput is the input for UpdateEntryActivity
type UpdateEntryInput struct {
	AccountBranch *string
	AccountNumber *string
	OwnerName     *string
	OwnerTaxID    *string
}

// UpdateEntryActivity updates an existing entry in the database
func (a *EntryActivities) UpdateEntryActivity(ctx context.Context, entryID string, updates UpdateEntryInput) error {
	a.logger.WithField("entry_id", entryID).Info("Updating entry")

	// Get entry from database
	entry, err := a.entryRepo.GetByEntryID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	// Apply updates to entry fields
	if updates.AccountBranch != nil {
		entry.AccountBranch = updates.AccountBranch
	}
	if updates.AccountNumber != nil {
		entry.AccountNumber = updates.AccountNumber
	}
	if updates.OwnerName != nil {
		entry.OwnerName = updates.OwnerName
	}
	if updates.OwnerTaxID != nil {
		entry.OwnerTaxID = updates.OwnerTaxID
	}

	// Update in database
	if err := a.entryRepo.Update(ctx, entry); err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	// Publish entry updated event
	event := map[string]interface{}{
		"event_type":   "entry_updated",
		"entry_id":     entry.EntryID,
		"key":          entry.Key,
		"updated_at":   entry.UpdatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, entry.EntryID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish entry updated event")
	}

	a.logger.WithField("entry_id", entryID).Info("Entry updated successfully")

	return nil
}

// DeleteEntryActivity performs a soft delete on an entry
func (a *EntryActivities) DeleteEntryActivity(ctx context.Context, entryID string) error {
	a.logger.WithField("entry_id", entryID).Info("Deleting entry")

	// Get entry first to publish complete event data
	entry, err := a.entryRepo.GetByEntryID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	// Soft delete
	if err := a.entryRepo.Delete(ctx, entryID); err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	// Publish entry deleted event
	event := map[string]interface{}{
		"event_type": "entry_deleted",
		"entry_id":   entry.EntryID,
		"key":        entry.Key,
		"deleted_at": entry.DeletedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, entry.EntryID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish entry deleted event")
	}

	a.logger.WithField("entry_id", entryID).Info("Entry deleted successfully")

	return nil
}

// ActivateEntryActivity activates an entry
func (a *EntryActivities) ActivateEntryActivity(ctx context.Context, entryID string) error {
	a.logger.WithField("entry_id", entryID).Info("Activating entry")

	// Get entry from database
	entry, err := a.entryRepo.GetByEntryID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	// Activate entry
	if err := entry.Activate(); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update in database
	if err := a.entryRepo.Update(ctx, entry); err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	// Publish entry activated event
	event := map[string]interface{}{
		"event_type":   "entry_activated",
		"entry_id":     entry.EntryID,
		"key":          entry.Key,
		"status":       entry.Status,
		"activated_at": entry.ActivatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, entry.EntryID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish entry activated event")
	}

	a.logger.WithField("entry_id", entryID).Info("Entry activated successfully")

	return nil
}

// DeactivateEntryActivity deactivates an entry
func (a *EntryActivities) DeactivateEntryActivity(ctx context.Context, entryID, reason string) error {
	a.logger.WithFields(logrus.Fields{
		"entry_id": entryID,
		"reason":   reason,
	}).Info("Deactivating entry")

	// Get entry from database
	entry, err := a.entryRepo.GetByEntryID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	// Deactivate entry
	if err := entry.Deactivate(reason); err != nil {
		return fmt.Errorf("invalid status transition: %w", err)
	}

	// Update in database
	if err := a.entryRepo.Update(ctx, entry); err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	// Publish entry deactivated event
	event := map[string]interface{}{
		"event_type":     "entry_deactivated",
		"entry_id":       entry.EntryID,
		"key":            entry.Key,
		"status":         entry.Status,
		"reason":         reason,
		"deactivated_at": entry.DeactivatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, entry.EntryID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish entry deactivated event")
	}

	a.logger.WithField("entry_id", entryID).Info("Entry deactivated successfully")

	return nil
}

// GetEntryStatusActivity retrieves the current status of an entry
func (a *EntryActivities) GetEntryStatusActivity(ctx context.Context, entryID string) (string, error) {
	a.logger.Infof("Getting entry status: %s", entryID)

	entry, err := a.entryRepo.GetByEntryID(ctx, entryID)
	if err != nil {
		a.logger.WithError(err).Errorf("Failed to get entry: %s", entryID)
		return "", fmt.Errorf("entry not found: %w", err)
	}

	return string(entry.Status), nil
}

// ValidateEntryActivity validates an entry
func (a *EntryActivities) ValidateEntryActivity(ctx context.Context, entryID string) error {
	a.logger.WithField("entry_id", entryID).Info("Validating entry")

	// Get entry from database
	entry, err := a.entryRepo.GetByEntryID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	// Run validation checks
	if entry.Key == "" {
		return fmt.Errorf("entry has no key")
	}

	if entry.Participant == "" {
		return fmt.Errorf("entry has no participant ISPB")
	}

	if entry.Status == entities.EntryStatusBlocked {
		return fmt.Errorf("entry is blocked and cannot be used")
	}

	if entry.DeletedAt != nil {
		return fmt.Errorf("entry is deleted")
	}

	// Check if key is in valid format (already validated by entity)
	// Additional business validations can be added here

	a.logger.WithField("entry_id", entryID).Info("Entry validation successful")

	return nil
}

// UpdateEntryOwnershipActivity updates entry ownership information
func (a *EntryActivities) UpdateEntryOwnershipActivity(ctx context.Context, entryID, newOwnerISPB, ownerName, ownerTaxID string) error {
	a.logger.WithFields(logrus.Fields{
		"entry_id":     entryID,
		"new_owner":    newOwnerISPB,
		"owner_name":   ownerName,
		"owner_tax_id": ownerTaxID,
	}).Info("Updating entry ownership")

	// Get entry from database
	entry, err := a.entryRepo.GetByEntryID(ctx, entryID)
	if err != nil {
		return fmt.Errorf("failed to get entry: %w", err)
	}

	// Update ownership using entity method
	if err := entry.UpdateOwnership(newOwnerISPB, ownerName, ownerTaxID); err != nil {
		return fmt.Errorf("invalid ownership update: %w", err)
	}

	// Update in database
	if err := a.entryRepo.Update(ctx, entry); err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	// Publish ownership updated event
	event := map[string]interface{}{
		"event_type":     "ownership_updated",
		"entry_id":       entry.EntryID,
		"key":            entry.Key,
		"new_owner":      newOwnerISPB,
		"owner_name":     ownerName,
		"owner_tax_id":   ownerTaxID,
		"updated_at":     entry.UpdatedAt,
	}

	if err := a.pulsarProducer.PublishEvent(ctx, event, entry.EntryID); err != nil {
		a.logger.WithError(err).Warn("Failed to publish ownership updated event")
	}

	a.logger.WithField("entry_id", entryID).Info("Entry ownership updated successfully")

	return nil
}