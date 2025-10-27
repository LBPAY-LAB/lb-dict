package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

// PostgresClaimRepository implements ClaimRepository using PostgreSQL
type PostgresClaimRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresClaimRepository creates a new claim repository
func NewPostgresClaimRepository(pool *pgxpool.Pool) repositories.ClaimRepository {
	return &PostgresClaimRepository{
		pool: pool,
	}
}

// FindByID finds a claim by ID
func (r *PostgresClaimRepository) FindByID(ctx context.Context, id uuid.UUID) (*entities.Claim, error) {
	query := `
		SELECT
			c.id, c.claim_type, c.status,
			c.claimer_ispb, c.owner_ispb,
			c.claimer_account_id, c.owner_account_id,
			c.bacen_claim_id, c.workflow_id,
			c.completion_period_days, c.expires_at,
			c.resolution_type, c.resolution_reason, c.resolution_date,
			c.created_at, c.updated_at,
			c.entry_key
		FROM core_dict.claims c
		WHERE c.id = $1 AND c.deleted_at IS NULL
		LIMIT 1
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find claim: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("claim not found: %s", id)
	}

	claim, err := scanClaim(rows)
	if err != nil {
		return nil, err
	}

	return claim, nil
}

// Create creates a new claim
func (r *PostgresClaimRepository) Create(ctx context.Context, claim *entities.Claim) error {
	query := `
		INSERT INTO core_dict.claims (
			id, entry_key, claim_type, status,
			claimer_ispb, owner_ispb,
			claimer_account_id, owner_account_id,
			bacen_claim_id, workflow_id,
			completion_period_days, expires_at,
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err := r.pool.Exec(ctx, query,
		claim.ID,
		claim.EntryKey,
		claim.ClaimType,
		claim.Status,
		claim.ClaimerParticipant.ISPB,
		claim.DonorParticipant.ISPB,
		claim.ClaimerAccountID,
		claim.DonorAccountID,
		claim.BacenClaimID,
		claim.WorkflowID,
		claim.CompletionPeriodDays,
		claim.ExpiresAt,
		claim.CreatedAt,
		claim.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create claim: %w", err)
	}

	return nil
}

// Update updates an existing claim
func (r *PostgresClaimRepository) Update(ctx context.Context, claim *entities.Claim) error {
	query := `
		UPDATE core_dict.claims
		SET status = $2,
			resolution_type = $3,
			resolution_reason = $4,
			resolution_date = $5,
			workflow_id = $6,
			bacen_claim_id = $7,
			updated_at = $8
		WHERE id = $1 AND deleted_at IS NULL
	`

	result, err := r.pool.Exec(ctx, query,
		claim.ID,
		claim.Status,
		claim.ResolutionType,
		claim.ResolutionReason,
		claim.ResolutionDate,
		claim.WorkflowID,
		claim.BacenClaimID,
		claim.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update claim: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("claim not found: %s", claim.ID)
	}

	return nil
}

// Delete performs soft delete on a claim
func (r *PostgresClaimRepository) Delete(ctx context.Context, claimID uuid.UUID) error {
	query := `
		UPDATE core_dict.claims
		SET deleted_at = $2,
			updated_at = $2
		WHERE id = $1 AND deleted_at IS NULL
	`

	now := time.Now()
	result, err := r.pool.Exec(ctx, query, claimID, now)

	if err != nil {
		return fmt.Errorf("failed to delete claim: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("claim not found or already deleted: %s", claimID)
	}

	return nil
}

// FindByEntryKey finds claims by PIX key value
func (r *PostgresClaimRepository) FindByEntryKey(ctx context.Context, entryKey string) ([]*entities.Claim, error) {
	query := `
		SELECT
			c.id, c.claim_type, c.status,
			c.claimer_ispb, c.owner_ispb,
			c.claimer_account_id, c.owner_account_id,
			c.bacen_claim_id, c.workflow_id,
			c.completion_period_days, c.expires_at,
			c.resolution_type, c.resolution_reason, c.resolution_date,
			c.created_at, c.updated_at,
			c.entry_key
		FROM core_dict.claims c
		WHERE c.entry_key = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, entryKey)
	if err != nil {
		return nil, fmt.Errorf("failed to find claims by entry key: %w", err)
	}
	defer rows.Close()

	var claims []*entities.Claim
	for rows.Next() {
		claim, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return claims, nil
}

// FindByStatus finds claims by status with pagination
func (r *PostgresClaimRepository) FindByStatus(ctx context.Context, status valueobjects.ClaimStatus, limit, offset int) ([]*entities.Claim, error) {
	query := `
		SELECT
			c.id, c.claim_type, c.status,
			c.claimer_ispb, c.owner_ispb,
			c.claimer_account_id, c.owner_account_id,
			c.bacen_claim_id, c.workflow_id,
			c.completion_period_days, c.expires_at,
			c.resolution_type, c.resolution_reason, c.resolution_date,
			c.created_at, c.updated_at,
			c.entry_key
		FROM core_dict.claims c
		WHERE c.status = $1 AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find claims by status: %w", err)
	}
	defer rows.Close()

	var claims []*entities.Claim
	for rows.Next() {
		claim, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return claims, nil
}

// FindByParticipant finds claims for a participant with pagination
func (r *PostgresClaimRepository) FindByParticipant(ctx context.Context, ispb string, limit, offset int) ([]*entities.Claim, error) {
	query := `
		SELECT
			c.id, c.claim_type, c.status,
			c.claimer_ispb, c.owner_ispb,
			c.claimer_account_id, c.owner_account_id,
			c.bacen_claim_id, c.workflow_id,
			c.completion_period_days, c.expires_at,
			c.resolution_type, c.resolution_reason, c.resolution_date,
			c.created_at, c.updated_at,
			c.entry_key
		FROM core_dict.claims c
		WHERE (c.claimer_ispb = $1 OR c.owner_ispb = $1)
			AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, ispb, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find claims by participant: %w", err)
	}
	defer rows.Close()

	var claims []*entities.Claim
	for rows.Next() {
		claim, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return claims, nil
}

// FindExpired finds expired claims (expires_at < now)
// FindActiveByEntryID finds active claim by entry ID
func (r *PostgresClaimRepository) FindActiveByEntryID(ctx context.Context, entryID uuid.UUID) (*entities.Claim, error) {
	query := `
		SELECT
			c.id, c.claim_type, c.status,
			c.claimer_ispb, c.owner_ispb,
			c.claimer_account_id, c.owner_account_id,
			c.bacen_claim_id, c.workflow_id,
			c.completion_period_days, c.expires_at,
			c.resolution_type, c.resolution_reason, c.resolution_date,
			c.created_at, c.updated_at,
			c.entry_key
		FROM core_dict.claims c
		JOIN core_dict.dict_entries e ON c.entry_key = e.key_value
		WHERE e.id = $1
		  AND c.status IN ('OPEN', 'WAITING_RESOLUTION', 'CONFIRMED')
		  AND c.deleted_at IS NULL
		ORDER BY c.created_at DESC
		LIMIT 1
	`

	rows, err := r.pool.Query(ctx, query, entryID)
	if err != nil {
		return nil, fmt.Errorf("failed to find active claim by entry ID: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no active claim found for entry ID: %s", entryID)
	}

	claim, err := scanClaim(rows)
	if err != nil {
		return nil, err
	}

	return claim, nil
}

func (r *PostgresClaimRepository) FindExpired(ctx context.Context, limit int) ([]*entities.Claim, error) {
	query := `
		SELECT
			c.id, c.claim_type, c.status,
			c.claimer_ispb, c.owner_ispb,
			c.claimer_account_id, c.owner_account_id,
			c.bacen_claim_id, c.workflow_id,
			c.completion_period_days, c.expires_at,
			c.resolution_type, c.resolution_reason, c.resolution_date,
			c.created_at, c.updated_at,
			c.entry_key
		FROM core_dict.claims c
		WHERE c.expires_at < $1
			AND c.status NOT IN ('COMPLETED', 'CANCELLED', 'EXPIRED')
			AND c.deleted_at IS NULL
		ORDER BY c.expires_at ASC
		LIMIT $2
	`

	now := time.Now()
	rows, err := r.pool.Query(ctx, query, now, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find expired claims: %w", err)
	}
	defer rows.Close()

	var claims []*entities.Claim
	for rows.Next() {
		claim, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return claims, nil
}

// FindByWorkflowID finds a claim by Temporal workflow ID
func (r *PostgresClaimRepository) FindByWorkflowID(ctx context.Context, workflowID string) (*entities.Claim, error) {
	query := `
		SELECT
			c.id, c.claim_type, c.status,
			c.claimer_ispb, c.owner_ispb,
			c.claimer_account_id, c.owner_account_id,
			c.bacen_claim_id, c.workflow_id,
			c.completion_period_days, c.expires_at,
			c.resolution_type, c.resolution_reason, c.resolution_date,
			c.created_at, c.updated_at,
			c.entry_key
		FROM core_dict.claims c
		WHERE c.workflow_id = $1 AND c.deleted_at IS NULL
		LIMIT 1
	`

	rows, err := r.pool.Query(ctx, query, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find claim by workflow ID: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("claim not found for workflow ID: %s", workflowID)
	}

	claim, err := scanClaim(rows)
	if err != nil {
		return nil, err
	}

	return claim, nil
}

// FindPendingResolution finds claims waiting for resolution
func (r *PostgresClaimRepository) FindPendingResolution(ctx context.Context, limit int) ([]*entities.Claim, error) {
	query := `
		SELECT
			c.id, c.claim_type, c.status,
			c.claimer_ispb, c.owner_ispb,
			c.claimer_account_id, c.owner_account_id,
			c.bacen_claim_id, c.workflow_id,
			c.completion_period_days, c.expires_at,
			c.resolution_type, c.resolution_reason, c.resolution_date,
			c.created_at, c.updated_at,
			c.entry_key
		FROM core_dict.claims c
		WHERE c.status IN ('OPEN', 'WAITING_RESOLUTION')
			AND c.deleted_at IS NULL
		ORDER BY c.created_at ASC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to find pending resolution claims: %w", err)
	}
	defer rows.Close()

	var claims []*entities.Claim
	for rows.Next() {
		claim, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return claims, nil
}

// ExistsActiveClaim checks if there's an active claim for a key
func (r *PostgresClaimRepository) ExistsActiveClaim(ctx context.Context, entryKey string) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1
			FROM core_dict.claims
			WHERE entry_key = $1
				AND status IN ('OPEN', 'WAITING_RESOLUTION', 'CONFIRMED')
				AND deleted_at IS NULL
		)
	`

	var exists bool
	err := r.pool.QueryRow(ctx, query, entryKey).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check active claim existence: %w", err)
	}

	return exists, nil
}

// List lists claims with filters and pagination
func (r *PostgresClaimRepository) List(ctx context.Context, filters repositories.ClaimFilters) ([]*entities.Claim, error) {
	query := `
		SELECT
			c.id, c.claim_type, c.status,
			c.claimer_ispb, c.owner_ispb,
			c.claimer_account_id, c.owner_account_id,
			c.bacen_claim_id, c.workflow_id,
			c.completion_period_days, c.expires_at,
			c.resolution_type, c.resolution_reason, c.resolution_date,
			c.created_at, c.updated_at,
			c.entry_key
		FROM core_dict.claims c
		WHERE c.deleted_at IS NULL
	`

	args := []interface{}{}
	argPos := 1

	if filters.EntryKey != nil {
		query += fmt.Sprintf(" AND c.entry_key = $%d", argPos)
		args = append(args, *filters.EntryKey)
		argPos++
	}

	if filters.ClaimType != nil {
		query += fmt.Sprintf(" AND c.claim_type = $%d", argPos)
		args = append(args, *filters.ClaimType)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND c.status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	if filters.ClaimerISPB != nil {
		query += fmt.Sprintf(" AND c.claimer_ispb = $%d", argPos)
		args = append(args, *filters.ClaimerISPB)
		argPos++
	}

	if filters.DonorISPB != nil {
		query += fmt.Sprintf(" AND c.owner_ispb = $%d", argPos)
		args = append(args, *filters.DonorISPB)
		argPos++
	}

	if filters.ExpiresAfter != nil {
		query += fmt.Sprintf(" AND c.expires_at > $%d", argPos)
		args = append(args, *filters.ExpiresAfter)
		argPos++
	}

	if filters.ExpiresBefore != nil {
		query += fmt.Sprintf(" AND c.expires_at < $%d", argPos)
		args = append(args, *filters.ExpiresBefore)
		argPos++
	}

	query += " ORDER BY c.created_at DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argPos)
		args = append(args, filters.Limit)
		argPos++
	}

	if filters.Offset > 0 {
		query += fmt.Sprintf(" OFFSET $%d", argPos)
		args = append(args, filters.Offset)
	}

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list claims: %w", err)
	}
	defer rows.Close()

	var claims []*entities.Claim
	for rows.Next() {
		claim, err := scanClaim(rows)
		if err != nil {
			return nil, err
		}
		claims = append(claims, claim)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return claims, nil
}

// Count counts total claims with filters
func (r *PostgresClaimRepository) Count(ctx context.Context, filters repositories.ClaimFilters) (int64, error) {
	query := `SELECT COUNT(*) FROM core_dict.claims c WHERE c.deleted_at IS NULL`

	args := []interface{}{}
	argPos := 1

	if filters.EntryKey != nil {
		query += fmt.Sprintf(" AND c.entry_key = $%d", argPos)
		args = append(args, *filters.EntryKey)
		argPos++
	}

	if filters.ClaimType != nil {
		query += fmt.Sprintf(" AND c.claim_type = $%d", argPos)
		args = append(args, *filters.ClaimType)
		argPos++
	}

	if filters.Status != nil {
		query += fmt.Sprintf(" AND c.status = $%d", argPos)
		args = append(args, *filters.Status)
		argPos++
	}

	if filters.ClaimerISPB != nil {
		query += fmt.Sprintf(" AND c.claimer_ispb = $%d", argPos)
		args = append(args, *filters.ClaimerISPB)
		argPos++
	}

	if filters.DonorISPB != nil {
		query += fmt.Sprintf(" AND c.owner_ispb = $%d", argPos)
		args = append(args, *filters.DonorISPB)
		argPos++
	}

	if filters.ExpiresAfter != nil {
		query += fmt.Sprintf(" AND c.expires_at > $%d", argPos)
		args = append(args, *filters.ExpiresAfter)
		argPos++
	}

	if filters.ExpiresBefore != nil {
		query += fmt.Sprintf(" AND c.expires_at < $%d", argPos)
		args = append(args, *filters.ExpiresBefore)
	}

	var count int64
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count claims: %w", err)
	}

	return count, nil
}

// scanClaim is a helper function to scan a claim from database rows
func scanClaim(rows pgx.Rows) (*entities.Claim, error) {
	var claim entities.Claim
	var claimerISPB, donorISPB string
	var resolutionDate *time.Time
	var resolutionType, resolutionReason *string

	err := rows.Scan(
		&claim.ID,
		&claim.ClaimType,
		&claim.Status,
		&claimerISPB,
		&donorISPB,
		&claim.ClaimerAccountID,
		&claim.DonorAccountID,
		&claim.BacenClaimID,
		&claim.WorkflowID,
		&claim.CompletionPeriodDays,
		&claim.ExpiresAt,
		&resolutionType,
		&resolutionReason,
		&resolutionDate,
		&claim.CreatedAt,
		&claim.UpdatedAt,
		&claim.EntryKey,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan claim: %w", err)
	}

	// Map participant ISPBs (we don't have names in DB, so we use ISPB only)
	claim.ClaimerParticipant = valueobjects.Participant{ISPB: claimerISPB}
	claim.DonorParticipant = valueobjects.Participant{ISPB: donorISPB}

	if resolutionType != nil {
		claim.ResolutionType = *resolutionType
	}
	if resolutionReason != nil {
		claim.ResolutionReason = *resolutionReason
	}
	if resolutionDate != nil {
		claim.ResolutionDate = resolutionDate
	}

	return &claim, nil
}
