package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/database"
	"github.com/sirupsen/logrus"
)

// EntryRepository handles persistence of Entry entities
type EntryRepository struct {
	db     *database.PostgresClient
	logger *logrus.Logger
}

// NewEntryRepository creates a new EntryRepository
func NewEntryRepository(db *database.PostgresClient, logger *logrus.Logger) *EntryRepository {
	return &EntryRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new entry into the database
func (r *EntryRepository) Create(ctx context.Context, entry *entities.Entry) error {
	query := `
		INSERT INTO entries (
			id, entry_id, key, key_type, participant,
			account_branch, account_number, account_type, account_opened_date,
			owner_type, owner_name, owner_tax_id,
			status, reason_for_status_change, bacen_entry_id,
			registered_at, activated_at, deactivated_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11, $12,
			$13, $14, $15,
			$16, $17, $18,
			$19, $20
		)
	`

	_, err := r.db.Exec(ctx, query,
		entry.ID, entry.EntryID, entry.Key, entry.KeyType, entry.Participant,
		entry.AccountBranch, entry.AccountNumber, entry.AccountType, entry.AccountOpenedDate,
		entry.OwnerType, entry.OwnerName, entry.OwnerTaxID,
		entry.Status, entry.ReasonForStatusChange, entry.BacenEntryID,
		entry.RegisteredAt, entry.ActivatedAt, entry.DeactivatedAt,
		entry.CreatedAt, entry.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to create entry: %s", entry.EntryID)
		return fmt.Errorf("failed to insert entry: %w", err)
	}

	r.logger.WithField("entry_id", entry.EntryID).Info("Entry created successfully")
	return nil
}

// GetByID retrieves an entry by UUID
func (r *EntryRepository) GetByID(ctx context.Context, id string) (*entities.Entry, error) {
	query := `
		SELECT 
			id, entry_id, key, key_type, participant,
			account_branch, account_number, account_type, account_opened_date,
			owner_type, owner_name, owner_tax_id,
			status, reason_for_status_change, bacen_entry_id,
			registered_at, activated_at, deactivated_at,
			created_at, updated_at, deleted_at
		FROM entries
		WHERE id = $1 AND deleted_at IS NULL
	`

	entry := &entities.Entry{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&entry.ID, &entry.EntryID, &entry.Key, &entry.KeyType, &entry.Participant,
		&entry.AccountBranch, &entry.AccountNumber, &entry.AccountType, &entry.AccountOpenedDate,
		&entry.OwnerType, &entry.OwnerName, &entry.OwnerTaxID,
		&entry.Status, &entry.ReasonForStatusChange, &entry.BacenEntryID,
		&entry.RegisteredAt, &entry.ActivatedAt, &entry.DeactivatedAt,
		&entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("entry not found: %s", id)
	}

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to get entry by ID: %s", id)
		return nil, fmt.Errorf("failed to query entry: %w", err)
	}

	return entry, nil
}

// GetByEntryID retrieves an entry by entry_id
func (r *EntryRepository) GetByEntryID(ctx context.Context, entryID string) (*entities.Entry, error) {
	query := `
		SELECT 
			id, entry_id, key, key_type, participant,
			account_branch, account_number, account_type, account_opened_date,
			owner_type, owner_name, owner_tax_id,
			status, reason_for_status_change, bacen_entry_id,
			registered_at, activated_at, deactivated_at,
			created_at, updated_at, deleted_at
		FROM entries
		WHERE entry_id = $1 AND deleted_at IS NULL
	`

	entry := &entities.Entry{}
	err := r.db.QueryRow(ctx, query, entryID).Scan(
		&entry.ID, &entry.EntryID, &entry.Key, &entry.KeyType, &entry.Participant,
		&entry.AccountBranch, &entry.AccountNumber, &entry.AccountType, &entry.AccountOpenedDate,
		&entry.OwnerType, &entry.OwnerName, &entry.OwnerTaxID,
		&entry.Status, &entry.ReasonForStatusChange, &entry.BacenEntryID,
		&entry.RegisteredAt, &entry.ActivatedAt, &entry.DeactivatedAt,
		&entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("entry not found: %s", entryID)
	}

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to get entry by entry_id: %s", entryID)
		return nil, fmt.Errorf("failed to query entry: %w", err)
	}

	return entry, nil
}

// GetByKey retrieves an entry by PIX key
func (r *EntryRepository) GetByKey(ctx context.Context, key string) (*entities.Entry, error) {
	query := `
		SELECT 
			id, entry_id, key, key_type, participant,
			account_branch, account_number, account_type, account_opened_date,
			owner_type, owner_name, owner_tax_id,
			status, reason_for_status_change, bacen_entry_id,
			registered_at, activated_at, deactivated_at,
			created_at, updated_at, deleted_at
		FROM entries
		WHERE key = $1 AND deleted_at IS NULL
	`

	entry := &entities.Entry{}
	err := r.db.QueryRow(ctx, query, key).Scan(
		&entry.ID, &entry.EntryID, &entry.Key, &entry.KeyType, &entry.Participant,
		&entry.AccountBranch, &entry.AccountNumber, &entry.AccountType, &entry.AccountOpenedDate,
		&entry.OwnerType, &entry.OwnerName, &entry.OwnerTaxID,
		&entry.Status, &entry.ReasonForStatusChange, &entry.BacenEntryID,
		&entry.RegisteredAt, &entry.ActivatedAt, &entry.DeactivatedAt,
		&entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("entry not found for key: %s", key)
	}

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to get entry by key: %s", key)
		return nil, fmt.Errorf("failed to query entry: %w", err)
	}

	return entry, nil
}

// Update updates an existing entry
func (r *EntryRepository) Update(ctx context.Context, entry *entities.Entry) error {
	query := `
		UPDATE entries SET
			key = $1, key_type = $2, participant = $3,
			account_branch = $4, account_number = $5, account_type = $6, account_opened_date = $7,
			owner_type = $8, owner_name = $9, owner_tax_id = $10,
			status = $11, reason_for_status_change = $12, bacen_entry_id = $13,
			registered_at = $14, activated_at = $15, deactivated_at = $16,
			updated_at = $17
		WHERE entry_id = $18 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query,
		entry.Key, entry.KeyType, entry.Participant,
		entry.AccountBranch, entry.AccountNumber, entry.AccountType, entry.AccountOpenedDate,
		entry.OwnerType, entry.OwnerName, entry.OwnerTaxID,
		entry.Status, entry.ReasonForStatusChange, entry.BacenEntryID,
		entry.RegisteredAt, entry.ActivatedAt, entry.DeactivatedAt,
		entry.UpdatedAt,
		entry.EntryID,
	)

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to update entry: %s", entry.EntryID)
		return fmt.Errorf("failed to update entry: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("entry not found or already deleted: %s", entry.EntryID)
	}

	r.logger.WithField("entry_id", entry.EntryID).Info("Entry updated successfully")
	return nil
}

// UpdateStatus updates only the status of an entry
func (r *EntryRepository) UpdateStatus(ctx context.Context, entryID string, status entities.EntryStatus) error {
	query := `
		UPDATE entries SET
			status = $1,
			updated_at = NOW()
		WHERE entry_id = $2 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query, status, entryID)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to update entry status: %s", entryID)
		return fmt.Errorf("failed to update entry status: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("entry not found or already deleted: %s", entryID)
	}

	r.logger.WithFields(logrus.Fields{
		"entry_id": entryID,
		"status":   status,
	}).Info("Entry status updated successfully")

	return nil
}

// Delete performs a soft delete on an entry
func (r *EntryRepository) Delete(ctx context.Context, entryID string) error {
	query := `
		UPDATE entries SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE entry_id = $1 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query, entryID)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to delete entry: %s", entryID)
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("entry not found or already deleted: %s", entryID)
	}

	r.logger.WithField("entry_id", entryID).Info("Entry deleted successfully")
	return nil
}

// ListByParticipant lists all entries for a given participant (ISPB)
func (r *EntryRepository) ListByParticipant(ctx context.Context, ispb string, limit, offset int) ([]*entities.Entry, error) {
	query := `
		SELECT 
			id, entry_id, key, key_type, participant,
			account_branch, account_number, account_type, account_opened_date,
			owner_type, owner_name, owner_tax_id,
			status, reason_for_status_change, bacen_entry_id,
			registered_at, activated_at, deactivated_at,
			created_at, updated_at, deleted_at
		FROM entries
		WHERE participant = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, ispb, limit, offset)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to list entries for participant: %s", ispb)
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}
	defer rows.Close()

	var entries []*entities.Entry
	for rows.Next() {
		entry := &entities.Entry{}
		err := rows.Scan(
			&entry.ID, &entry.EntryID, &entry.Key, &entry.KeyType, &entry.Participant,
			&entry.AccountBranch, &entry.AccountNumber, &entry.AccountType, &entry.AccountOpenedDate,
			&entry.OwnerType, &entry.OwnerName, &entry.OwnerTaxID,
			&entry.Status, &entry.ReasonForStatusChange, &entry.BacenEntryID,
			&entry.RegisteredAt, &entry.ActivatedAt, &entry.DeactivatedAt,
			&entry.CreatedAt, &entry.UpdatedAt, &entry.DeletedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan entry row")
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

// CountByParticipant counts total entries for a participant (ISPB)
func (r *EntryRepository) CountByParticipant(ctx context.Context, ispb string) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM entries
		WHERE participant = $1 AND deleted_at IS NULL
	`

	var count int64
	err := r.db.QueryRow(ctx, query, ispb).Scan(&count)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to count entries for participant: %s", ispb)
		return 0, fmt.Errorf("failed to count entries: %w", err)
	}

	return count, nil
}

// HasActiveKey checks if a PIX key already exists and is active
func (r *EntryRepository) HasActiveKey(ctx context.Context, key string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM entries
		WHERE key = $1
		  AND status = 'ACTIVE'
		  AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, key).Scan(&count)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to check active key: %s", key)
		return false, fmt.Errorf("failed to check active key: %w", err)
	}

	return count > 0, nil
}
