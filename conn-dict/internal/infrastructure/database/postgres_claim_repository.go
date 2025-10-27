package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/domain/aggregates"
	"github.com/lbpay-lab/conn-dict/internal/domain/interfaces"
)

// PostgresClaimRepository implements ClaimRepository using PostgreSQL
type PostgresClaimRepository struct {
	db *sql.DB
}

// NewPostgresClaimRepository creates a new PostgresClaimRepository
func NewPostgresClaimRepository(db *sql.DB) *PostgresClaimRepository {
	return &PostgresClaimRepository{
		db: db,
	}
}

// Save persists a claim aggregate
func (r *PostgresClaimRepository) Save(ctx context.Context, claim *aggregates.Claim) error {
	query := `
		INSERT INTO claims (
			id, key, key_type, ispb, branch, account, account_type,
			owner_name, owner_document, status, created_at, updated_at,
			confirmed_at, cancelled_at, expires_at, version
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			updated_at = EXCLUDED.updated_at,
			confirmed_at = EXCLUDED.confirmed_at,
			cancelled_at = EXCLUDED.cancelled_at,
			version = EXCLUDED.version
	`

	_, err := r.db.ExecContext(ctx, query,
		claim.ID, claim.Key, claim.KeyType, claim.ISPB, claim.Branch,
		claim.Account, claim.AccountType, claim.OwnerName, claim.OwnerDocument,
		claim.Status, claim.CreatedAt, claim.UpdatedAt, claim.ConfirmedAt,
		claim.CancelledAt, claim.ExpiresAt, claim.Version,
	)
	if err != nil {
		return fmt.Errorf("failed to save claim: %w", err)
	}

	return nil
}

// FindByID retrieves a claim by ID
func (r *PostgresClaimRepository) FindByID(ctx context.Context, id string) (*aggregates.Claim, error) {
	query := `
		SELECT id, key, key_type, ispb, branch, account, account_type,
			   owner_name, owner_document, status, created_at, updated_at,
			   confirmed_at, cancelled_at, expires_at, version
		FROM claims
		WHERE id = $1
	`

	claim := &aggregates.Claim{}
	var confirmedAt, cancelledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&claim.ID, &claim.Key, &claim.KeyType, &claim.ISPB, &claim.Branch,
		&claim.Account, &claim.AccountType, &claim.OwnerName, &claim.OwnerDocument,
		&claim.Status, &claim.CreatedAt, &claim.UpdatedAt, &confirmedAt,
		&cancelledAt, &claim.ExpiresAt, &claim.Version,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("claim not found: %s", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find claim: %w", err)
	}

	if confirmedAt.Valid {
		claim.ConfirmedAt = &confirmedAt.Time
	}
	if cancelledAt.Valid {
		claim.CancelledAt = &cancelledAt.Time
	}

	return claim, nil
}

// FindByKey retrieves a claim by key and key type
func (r *PostgresClaimRepository) FindByKey(ctx context.Context, key, keyType string) (*aggregates.Claim, error) {
	query := `
		SELECT id, key, key_type, ispb, branch, account, account_type,
			   owner_name, owner_document, status, created_at, updated_at,
			   confirmed_at, cancelled_at, expires_at, version
		FROM claims
		WHERE key = $1 AND key_type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	claim := &aggregates.Claim{}
	var confirmedAt, cancelledAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, key, keyType).Scan(
		&claim.ID, &claim.Key, &claim.KeyType, &claim.ISPB, &claim.Branch,
		&claim.Account, &claim.AccountType, &claim.OwnerName, &claim.OwnerDocument,
		&claim.Status, &claim.CreatedAt, &claim.UpdatedAt, &confirmedAt,
		&cancelledAt, &claim.ExpiresAt, &claim.Version,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("claim not found for key: %s", key)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find claim: %w", err)
	}

	if confirmedAt.Valid {
		claim.ConfirmedAt = &confirmedAt.Time
	}
	if cancelledAt.Valid {
		claim.CancelledAt = &cancelledAt.Time
	}

	return claim, nil
}

// FindPendingClaims retrieves all pending claims
func (r *PostgresClaimRepository) FindPendingClaims(ctx context.Context) ([]*aggregates.Claim, error) {
	query := `
		SELECT id, key, key_type, ispb, branch, account, account_type,
			   owner_name, owner_document, status, created_at, updated_at,
			   confirmed_at, cancelled_at, expires_at, version
		FROM claims
		WHERE status = $1
		ORDER BY created_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, aggregates.ClaimStatusPending)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending claims: %w", err)
	}
	defer rows.Close()

	claims := make([]*aggregates.Claim, 0)
	for rows.Next() {
		claim := &aggregates.Claim{}
		var confirmedAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&claim.ID, &claim.Key, &claim.KeyType, &claim.ISPB, &claim.Branch,
			&claim.Account, &claim.AccountType, &claim.OwnerName, &claim.OwnerDocument,
			&claim.Status, &claim.CreatedAt, &claim.UpdatedAt, &confirmedAt,
			&cancelledAt, &claim.ExpiresAt, &claim.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan claim: %w", err)
		}

		if confirmedAt.Valid {
			claim.ConfirmedAt = &confirmedAt.Time
		}
		if cancelledAt.Valid {
			claim.CancelledAt = &cancelledAt.Time
		}

		claims = append(claims, claim)
	}

	return claims, nil
}

// FindExpiredClaims retrieves all expired claims
func (r *PostgresClaimRepository) FindExpiredClaims(ctx context.Context) ([]*aggregates.Claim, error) {
	query := `
		SELECT id, key, key_type, ispb, branch, account, account_type,
			   owner_name, owner_document, status, created_at, updated_at,
			   confirmed_at, cancelled_at, expires_at, version
		FROM claims
		WHERE status = $1 AND expires_at < $2
		ORDER BY expires_at ASC
	`

	rows, err := r.db.QueryContext(ctx, query, aggregates.ClaimStatusPending, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query expired claims: %w", err)
	}
	defer rows.Close()

	claims := make([]*aggregates.Claim, 0)
	for rows.Next() {
		claim := &aggregates.Claim{}
		var confirmedAt, cancelledAt sql.NullTime

		err := rows.Scan(
			&claim.ID, &claim.Key, &claim.KeyType, &claim.ISPB, &claim.Branch,
			&claim.Account, &claim.AccountType, &claim.OwnerName, &claim.OwnerDocument,
			&claim.Status, &claim.CreatedAt, &claim.UpdatedAt, &confirmedAt,
			&cancelledAt, &claim.ExpiresAt, &claim.Version,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan claim: %w", err)
		}

		if confirmedAt.Valid {
			claim.ConfirmedAt = &confirmedAt.Time
		}
		if cancelledAt.Valid {
			claim.CancelledAt = &cancelledAt.Time
		}

		claims = append(claims, claim)
	}

	return claims, nil
}

// Delete removes a claim
func (r *PostgresClaimRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM claims WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete claim: %w", err)
	}

	return nil
}

// Ensure PostgresClaimRepository implements ClaimRepository
var _ interfaces.ClaimRepository = (*PostgresClaimRepository)(nil)
