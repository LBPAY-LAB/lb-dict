package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/database"
	"github.com/sirupsen/logrus"
)

// ClaimRepository handles persistence operations for Claims
type ClaimRepository struct {
	db     *database.PostgresClient
	logger *logrus.Logger
}

// NewClaimRepository creates a new claim repository
func NewClaimRepository(db *database.PostgresClient, logger *logrus.Logger) *ClaimRepository {
	return &ClaimRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new claim into the database
func (r *ClaimRepository) Create(ctx context.Context, claim *entities.Claim) error {
	query := `
		INSERT INTO claims (
			id, claim_id, type, status, key, key_type,
			donor_participant, claimer_participant,
			claimer_account_branch, claimer_account_number, claimer_account_type,
			completion_period_end, claim_expiry_date,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15
		)
	`

	_, err := r.db.Exec(ctx, query,
		claim.ID,
		claim.ClaimID,
		claim.Type,
		claim.Status,
		claim.Key,
		claim.KeyType,
		claim.DonorParticipant,
		claim.ClaimerParticipant,
		claim.ClaimerAccountBranch,
		claim.ClaimerAccountNumber,
		claim.ClaimerAccountType,
		claim.CompletionPeriodEnd,
		claim.ClaimExpiryDate,
		claim.CreatedAt,
		claim.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to create claim: %s", claim.ClaimID)
		return fmt.Errorf("failed to insert claim: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"claim_id": claim.ClaimID,
		"key":      claim.Key,
		"status":   claim.Status,
	}).Info("Claim created successfully")

	return nil
}

// GetByID retrieves a claim by its UUID
func (r *ClaimRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Claim, error) {
	query := `
		SELECT
			id, claim_id, type, status, key, key_type,
			donor_participant, claimer_participant,
			claimer_account_branch, claimer_account_number, claimer_account_type,
			completion_period_end, claim_expiry_date,
			confirmed_at, completed_at, cancelled_at, expired_at,
			cancellation_reason, notes,
			created_at, updated_at, deleted_at
		FROM claims
		WHERE id = $1 AND deleted_at IS NULL
	`

	claim := &entities.Claim{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&claim.ID,
		&claim.ClaimID,
		&claim.Type,
		&claim.Status,
		&claim.Key,
		&claim.KeyType,
		&claim.DonorParticipant,
		&claim.ClaimerParticipant,
		&claim.ClaimerAccountBranch,
		&claim.ClaimerAccountNumber,
		&claim.ClaimerAccountType,
		&claim.CompletionPeriodEnd,
		&claim.ClaimExpiryDate,
		&claim.ConfirmedAt,
		&claim.CompletedAt,
		&claim.CancelledAt,
		&claim.ExpiredAt,
		&claim.CancellationReason,
		&claim.Notes,
		&claim.CreatedAt,
		&claim.UpdatedAt,
		&claim.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("claim not found: %s", id)
		}
		r.logger.WithError(err).Errorf("Failed to get claim by ID: %s", id)
		return nil, fmt.Errorf("failed to query claim: %w", err)
	}

	return claim, nil
}

// GetByClaimID retrieves a claim by its claim_id (BACEN identifier)
func (r *ClaimRepository) GetByClaimID(ctx context.Context, claimID string) (*entities.Claim, error) {
	query := `
		SELECT
			id, claim_id, type, status, key, key_type,
			donor_participant, claimer_participant,
			claimer_account_branch, claimer_account_number, claimer_account_type,
			completion_period_end, claim_expiry_date,
			confirmed_at, completed_at, cancelled_at, expired_at,
			cancellation_reason, notes,
			created_at, updated_at, deleted_at
		FROM claims
		WHERE claim_id = $1 AND deleted_at IS NULL
	`

	claim := &entities.Claim{}
	err := r.db.QueryRow(ctx, query, claimID).Scan(
		&claim.ID,
		&claim.ClaimID,
		&claim.Type,
		&claim.Status,
		&claim.Key,
		&claim.KeyType,
		&claim.DonorParticipant,
		&claim.ClaimerParticipant,
		&claim.ClaimerAccountBranch,
		&claim.ClaimerAccountNumber,
		&claim.ClaimerAccountType,
		&claim.CompletionPeriodEnd,
		&claim.ClaimExpiryDate,
		&claim.ConfirmedAt,
		&claim.CompletedAt,
		&claim.CancelledAt,
		&claim.ExpiredAt,
		&claim.CancellationReason,
		&claim.Notes,
		&claim.CreatedAt,
		&claim.UpdatedAt,
		&claim.DeletedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("claim not found: %s", claimID)
		}
		r.logger.WithError(err).Errorf("Failed to get claim by claim_id: %s", claimID)
		return nil, fmt.Errorf("failed to query claim: %w", err)
	}

	return claim, nil
}

// UpdateStatus updates the claim status
func (r *ClaimRepository) UpdateStatus(ctx context.Context, claimID string, status entities.ClaimStatus) error {
	query := `
		UPDATE claims
		SET status = $1, updated_at = $2
		WHERE claim_id = $3 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, status, time.Now(), claimID)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to update claim status: %s", claimID)
		return fmt.Errorf("failed to update status: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("claim not found or already deleted: %s", claimID)
	}

	r.logger.WithFields(logrus.Fields{
		"claim_id": claimID,
		"status":   status,
	}).Info("Claim status updated")

	return nil
}

// Update updates the entire claim entity
func (r *ClaimRepository) Update(ctx context.Context, claim *entities.Claim) error {
	query := `
		UPDATE claims
		SET
			status = $1,
			confirmed_at = $2,
			completed_at = $3,
			cancelled_at = $4,
			expired_at = $5,
			cancellation_reason = $6,
			notes = $7,
			updated_at = $8
		WHERE claim_id = $9 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query,
		claim.Status,
		claim.ConfirmedAt,
		claim.CompletedAt,
		claim.CancelledAt,
		claim.ExpiredAt,
		claim.CancellationReason,
		claim.Notes,
		time.Now(),
		claim.ClaimID,
	)

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to update claim: %s", claim.ClaimID)
		return fmt.Errorf("failed to update claim: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("claim not found or already deleted: %s", claim.ClaimID)
	}

	r.logger.WithFields(logrus.Fields{
		"claim_id": claim.ClaimID,
		"status":   claim.Status,
	}).Info("Claim updated successfully")

	return nil
}

// Delete soft deletes a claim
func (r *ClaimRepository) Delete(ctx context.Context, claimID string) error {
	query := `
		UPDATE claims
		SET deleted_at = $1, updated_at = $1
		WHERE claim_id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), claimID)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to delete claim: %s", claimID)
		return fmt.Errorf("failed to delete claim: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("claim not found or already deleted: %s", claimID)
	}

	r.logger.WithField("claim_id", claimID).Info("Claim deleted")

	return nil
}

// ListByKey lists all claims for a specific key
func (r *ClaimRepository) ListByKey(ctx context.Context, key string) ([]*entities.Claim, error) {
	query := `
		SELECT
			id, claim_id, type, status, key, key_type,
			donor_participant, claimer_participant,
			claimer_account_branch, claimer_account_number, claimer_account_type,
			completion_period_end, claim_expiry_date,
			confirmed_at, completed_at, cancelled_at, expired_at,
			cancellation_reason, notes,
			created_at, updated_at, deleted_at
		FROM claims
		WHERE key = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, key)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to list claims for key: %s", key)
		return nil, fmt.Errorf("failed to query claims: %w", err)
	}
	defer rows.Close()

	claims := make([]*entities.Claim, 0)
	for rows.Next() {
		claim := &entities.Claim{}
		err := rows.Scan(
			&claim.ID,
			&claim.ClaimID,
			&claim.Type,
			&claim.Status,
			&claim.Key,
			&claim.KeyType,
			&claim.DonorParticipant,
			&claim.ClaimerParticipant,
			&claim.ClaimerAccountBranch,
			&claim.ClaimerAccountNumber,
			&claim.ClaimerAccountType,
			&claim.CompletionPeriodEnd,
			&claim.ClaimExpiryDate,
			&claim.ConfirmedAt,
			&claim.CompletedAt,
			&claim.CancelledAt,
			&claim.ExpiredAt,
			&claim.CancellationReason,
			&claim.Notes,
			&claim.CreatedAt,
			&claim.UpdatedAt,
			&claim.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan claim: %w", err)
		}
		claims = append(claims, claim)
	}

	return claims, nil
}

// ListExpired lists all expired claims that need to be processed
func (r *ClaimRepository) ListExpired(ctx context.Context, limit int) ([]*entities.Claim, error) {
	query := `
		SELECT
			id, claim_id, type, status, key, key_type,
			donor_participant, claimer_participant,
			claimer_account_branch, claimer_account_number, claimer_account_type,
			completion_period_end, claim_expiry_date,
			confirmed_at, completed_at, cancelled_at, expired_at,
			cancellation_reason, notes,
			created_at, updated_at, deleted_at
		FROM claims
		WHERE claim_expiry_date < NOW()
		  AND status IN ('OPEN', 'WAITING_RESOLUTION')
		  AND deleted_at IS NULL
		ORDER BY claim_expiry_date ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list expired claims")
		return nil, fmt.Errorf("failed to query expired claims: %w", err)
	}
	defer rows.Close()

	claims := make([]*entities.Claim, 0)
	for rows.Next() {
		claim := &entities.Claim{}
		err := rows.Scan(
			&claim.ID,
			&claim.ClaimID,
			&claim.Type,
			&claim.Status,
			&claim.Key,
			&claim.KeyType,
			&claim.DonorParticipant,
			&claim.ClaimerParticipant,
			&claim.ClaimerAccountBranch,
			&claim.ClaimerAccountNumber,
			&claim.ClaimerAccountType,
			&claim.CompletionPeriodEnd,
			&claim.ClaimExpiryDate,
			&claim.ConfirmedAt,
			&claim.CompletedAt,
			&claim.CancelledAt,
			&claim.ExpiredAt,
			&claim.CancellationReason,
			&claim.Notes,
			&claim.CreatedAt,
			&claim.UpdatedAt,
			&claim.DeletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan claim: %w", err)
		}
		claims = append(claims, claim)
	}

	return claims, nil
}

// HasActiveClaim checks if there's an active claim for a key
func (r *ClaimRepository) HasActiveClaim(ctx context.Context, key string) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM claims
		WHERE key = $1
		  AND status IN ('OPEN', 'WAITING_RESOLUTION', 'CONFIRMED')
		  AND deleted_at IS NULL
	`

	var count int
	err := r.db.QueryRow(ctx, query, key).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check active claims: %w", err)
	}

	return count > 0, nil
}