package database

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
)

// PostgresEntryRepository implements EntryRepository using PostgreSQL
type PostgresEntryRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresEntryRepository creates a new entry repository
func NewPostgresEntryRepository(pool *pgxpool.Pool) repositories.EntryRepository {
	return &PostgresEntryRepository{
		pool: pool,
	}
}

// FindByKey finds a PIX key by its value
func (r *PostgresEntryRepository) FindByKey(ctx context.Context, keyValue string) (*entities.Entry, error) {
	// Hash the key for LGPD-compliant search
	hash := hashKey(keyValue)

	query := `
		SELECT
			e.id, e.key_type, e.key_value, e.status,
			e.account_id, e.participant_ispb, e.participant_branch,
			e.created_at, e.updated_at, e.deleted_at,
			a.account_number, a.account_type, a.holder_name,
			a.holder_document, a.holder_document_type
		FROM core_dict.dict_entries e
		JOIN core_dict.accounts a ON e.account_id = a.id
		WHERE e.key_hash = $1 AND e.deleted_at IS NULL
		LIMIT 1
	`

	var entry entities.Entry

	err := r.pool.QueryRow(ctx, query, hash).Scan(
		&entry.ID,
		&entry.KeyType,
		&entry.KeyValue,
		&entry.Status,
		&entry.AccountID,
		&entry.ISPB,
		&entry.Branch,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&entry.DeletedAt,
		&entry.AccountNumber,
		&entry.AccountType,
		&entry.OwnerName,
		&entry.OwnerTaxID,
		&entry.OwnerType,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("key not found: %s", keyValue)
		}
		return nil, fmt.Errorf("failed to find key: %w", err)
	}

	return &entry, nil
}

// FindByID finds a PIX key by its ID
func (r *PostgresEntryRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Entry, error) {
	query := `
		SELECT
			e.id, e.key_type, e.key_value, e.status,
			e.account_id, e.participant_ispb, e.participant_branch,
			e.created_at, e.updated_at, e.deleted_at,
			a.account_number, a.account_type, a.holder_name,
			a.holder_document, a.holder_document_type
		FROM core_dict.dict_entries e
		JOIN core_dict.accounts a ON e.account_id = a.id
		WHERE e.id = $1 AND e.deleted_at IS NULL
		LIMIT 1
	`

	var entry entities.Entry

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&entry.ID,
		&entry.KeyType,
		&entry.KeyValue,
		&entry.Status,
		&entry.AccountID,
		&entry.ISPB,
		&entry.Branch,
		&entry.CreatedAt,
		&entry.UpdatedAt,
		&entry.DeletedAt,
		&entry.AccountNumber,
		&entry.AccountType,
		&entry.OwnerName,
		&entry.OwnerTaxID,
		&entry.OwnerType,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("entry not found: %s", id)
		}
		return nil, fmt.Errorf("failed to find entry: %w", err)
	}

	return &entry, nil
}

// List lists PIX keys with pagination
func (r *PostgresEntryRepository) List(ctx context.Context, accountID uuid.UUID, limit, offset int) ([]*entities.Entry, error) {
	query := `
		SELECT
			e.id, e.key_type, e.key_value, e.status,
			e.account_id, e.participant_ispb, e.participant_branch,
			e.created_at, e.updated_at, e.deleted_at,
			a.account_number, a.account_type, a.holder_name,
			a.holder_document, a.holder_document_type
		FROM core_dict.dict_entries e
		JOIN core_dict.accounts a ON e.account_id = a.id
		WHERE e.account_id = $1 AND e.deleted_at IS NULL
		ORDER BY e.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, accountID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list entries: %w", err)
	}
	defer rows.Close()

	var entries []*entities.Entry
	for rows.Next() {
		var entry entities.Entry

		err := rows.Scan(
			&entry.ID,
			&entry.KeyType,
			&entry.KeyValue,
			&entry.Status,
			&entry.AccountID,
			&entry.ISPB,
			&entry.Branch,
			&entry.CreatedAt,
			&entry.UpdatedAt,
			&entry.DeletedAt,
			&entry.AccountNumber,
			&entry.AccountType,
			&entry.OwnerName,
			&entry.OwnerTaxID,
			&entry.OwnerType,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan entry: %w", err)
		}

		entries = append(entries, &entry)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return entries, nil
}

// CountByAccount counts keys for an account
func (r *PostgresEntryRepository) CountByAccount(ctx context.Context, accountID uuid.UUID) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM core_dict.dict_entries
		WHERE account_id = $1 AND deleted_at IS NULL
	`

	var count int64
	err := r.pool.QueryRow(ctx, query, accountID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count entries: %w", err)
	}

	return count, nil
}

// Create creates a new PIX key entry
func (r *PostgresEntryRepository) Create(ctx context.Context, entry *entities.Entry) error {
	keyHash := hashKey(entry.KeyValue)

	query := `
		INSERT INTO core_dict.dict_entries (
			id, key_type, key_value, key_hash,
			status, account_id, participant_ispb, participant_branch,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	_, err := r.pool.Exec(ctx, query,
		entry.ID,
		entry.KeyType,
		entry.KeyValue,
		keyHash,
		entry.Status,
		entry.AccountID,
		entry.ISPB,
		entry.Branch,
		entry.CreatedAt,
		entry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create entry: %w", err)
	}

	return nil
}

// Update updates an existing PIX key entry
func (r *PostgresEntryRepository) Update(ctx context.Context, entry *entities.Entry) error {
	query := `
		UPDATE core_dict.dict_entries
		SET status = $2,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		entry.ID,
		entry.Status,
		entry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update entry: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("entry not found: %s", entry.ID)
	}

	return nil
}

// Delete performs soft delete on a PIX key entry
func (r *PostgresEntryRepository) Delete(ctx context.Context, entryID uuid.UUID) error {
	query := `
		UPDATE core_dict.dict_entries
		SET status = $2,
			deleted_at = $3,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, entryID, entities.KeyStatusDeleted, now)

	if err != nil {
		return fmt.Errorf("failed to delete entry: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("entry not found or already deleted: %s", entryID)
	}

	return nil
}

// UpdateStatus updates only the status of a PIX key entry
func (r *PostgresEntryRepository) UpdateStatus(ctx context.Context, entryID uuid.UUID, status entities.KeyStatus) error {
	query := `
		UPDATE core_dict.dict_entries
		SET status = $2,
			updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, entryID, status, now)

	if err != nil {
		return fmt.Errorf("failed to update entry status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("entry not found: %s", entryID)
	}

	return nil
}

// hashKey creates a SHA-256 hash of the key value (LGPD compliance)
func hashKey(keyValue string) string {
	hash := sha256.Sum256([]byte(keyValue))
	return hex.EncodeToString(hash[:])
}
