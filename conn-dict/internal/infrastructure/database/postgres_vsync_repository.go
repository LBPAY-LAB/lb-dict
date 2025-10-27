package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/lbpay-lab/conn-dict/internal/domain/aggregates"
	"github.com/lbpay-lab/conn-dict/internal/domain/interfaces"
)

// PostgresVsyncRepository implements VsyncRepository using PostgreSQL
type PostgresVsyncRepository struct {
	db *sql.DB
}

// NewPostgresVsyncRepository creates a new PostgresVsyncRepository
func NewPostgresVsyncRepository(db *sql.DB) *PostgresVsyncRepository {
	return &PostgresVsyncRepository{
		db: db,
	}
}

// Save persists a vsync entry
func (r *PostgresVsyncRepository) Save(ctx context.Context, entry *aggregates.VsyncEntry) error {
	query := `
		INSERT INTO vsync_entries (
			key, key_type, ispb, branch, account, account_type,
			owner_name, owner_document, status, synced_at, created_at,
			updated_at, attempts, last_error, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		ON CONFLICT (key, key_type) DO UPDATE SET
			ispb = EXCLUDED.ispb,
			branch = EXCLUDED.branch,
			account = EXCLUDED.account,
			account_type = EXCLUDED.account_type,
			owner_name = EXCLUDED.owner_name,
			owner_document = EXCLUDED.owner_document,
			status = EXCLUDED.status,
			synced_at = EXCLUDED.synced_at,
			updated_at = EXCLUDED.updated_at,
			attempts = EXCLUDED.attempts,
			last_error = EXCLUDED.last_error,
			version = EXCLUDED.version
	`

	_, err := r.db.ExecContext(ctx, query,
		entry.Key, entry.KeyType, entry.ISPB, entry.Branch,
		entry.Account, entry.AccountType, entry.OwnerName, entry.OwnerDocument,
		entry.Status, entry.SyncedAt, entry.CreatedAt, entry.UpdatedAt,
		entry.Attempts, entry.LastError, entry.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to save vsync entry: %w", err)
	}

	return nil
}

// FindByKey retrieves a vsync entry by key and key type
func (r *PostgresVsyncRepository) FindByKey(ctx context.Context, key, keyType string) (*aggregates.VsyncEntry, error) {
	query := `
		SELECT key, key_type, ispb, branch, account, account_type,
			   owner_name, owner_document, status, synced_at, created_at,
			   updated_at, attempts, last_error, version
		FROM vsync_entries
		WHERE key = $1 AND key_type = $2
	`

	entry := &aggregates.VsyncEntry{}

	err := r.db.QueryRowContext(ctx, query, key, keyType).Scan(
		&entry.Key, &entry.KeyType, &entry.ISPB, &entry.Branch,
		&entry.Account, &entry.AccountType, &entry.OwnerName, &entry.OwnerDocument,
		&entry.Status, &entry.SyncedAt, &entry.CreatedAt, &entry.UpdatedAt,
		&entry.Attempts, &entry.LastError, &entry.Version,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("vsync entry not found for key: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find vsync entry: %w", err)
	}

	return entry, nil
}

// FindPendingEntries retrieves all pending vsync entries
func (r *PostgresVsyncRepository) FindPendingEntries(ctx context.Context) ([]*aggregates.VsyncEntry, error) {
	query := `
		SELECT key, key_type, ispb, branch, account, account_type,
			   owner_name, owner_document, status, synced_at, created_at,
			   updated_at, attempts, last_error, version
		FROM vsync_entries
		WHERE status = $1
		ORDER BY created_at ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, aggregates.VsyncStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending entries: %w", err)
	}
	defer rows.Close()

	entries := make([]*aggregates.VsyncEntry, 0)
	for rows.Next() {
		entry := &aggregates.VsyncEntry{}

		err := rows.Scan(
			&entry.Key, &entry.KeyType, &entry.ISPB, &entry.Branch,
			&entry.Account, &entry.AccountType, &entry.OwnerName, &entry.OwnerDocument,
			&entry.Status, &entry.SyncedAt, &entry.CreatedAt, &entry.UpdatedAt,
			&entry.Attempts, &entry.LastError, &entry.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vsync entry: %w", err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// FindFailedEntries retrieves all failed vsync entries that can be retried
func (r *PostgresVsyncRepository) FindFailedEntries(ctx context.Context) ([]*aggregates.VsyncEntry, error) {
	query := `
		SELECT key, key_type, ispb, branch, account, account_type,
			   owner_name, owner_document, status, synced_at, created_at,
			   updated_at, attempts, last_error, version
		FROM vsync_entries
		WHERE status = $1 AND attempts < 3
		ORDER BY updated_at ASC
		LIMIT 100
	`

	rows, err := r.db.QueryContext(ctx, query, aggregates.VsyncStatusFailed)
	if err != nil {
		return nil, fmt.Errorf("failed to query failed entries: %w", err)
	}
	defer rows.Close()

	entries := make([]*aggregates.VsyncEntry, 0)
	for rows.Next() {
		entry := &aggregates.VsyncEntry{}

		err := rows.Scan(
			&entry.Key, &entry.KeyType, &entry.ISPB, &entry.Branch,
			&entry.Account, &entry.AccountType, &entry.OwnerName, &entry.OwnerDocument,
			&entry.Status, &entry.SyncedAt, &entry.CreatedAt, &entry.UpdatedAt,
			&entry.Attempts, &entry.LastError, &entry.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan vsync entry: %w", err)
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// Delete removes a vsync entry
func (r *PostgresVsyncRepository) Delete(ctx context.Context, key, keyType string) error {
	query := `DELETE FROM vsync_entries WHERE key = $1 AND key_type = $2`

	_, err := r.db.ExecContext(ctx, query, key, keyType)
	if err != nil {
		return fmt.Errorf("failed to delete vsync entry: %w", err)
	}

	return nil
}

// Ensure PostgresVsyncRepository implements VsyncRepository
var _ interfaces.VsyncRepository = (*PostgresVsyncRepository)(nil)
