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

// InfractionRepository handles persistence of Infraction entities
type InfractionRepository struct {
	db     *database.PostgresClient
	logger *logrus.Logger
}

// NewInfractionRepository creates a new InfractionRepository
func NewInfractionRepository(db *database.PostgresClient, logger *logrus.Logger) *InfractionRepository {
	return &InfractionRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new infraction into the database
func (r *InfractionRepository) Create(ctx context.Context, infraction *entities.Infraction) error {
	query := `
		INSERT INTO infractions (
			id, infraction_id, entry_id, claim_id, key,
			type, description, evidence_urls,
			reporter_participant, reported_participant,
			status, resolution_notes,
			reported_at, investigated_at, resolved_at,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10,
			$11, $12,
			$13, $14, $15,
			$16, $17
		)
	`

	_, err := r.db.Exec(ctx, query,
		infraction.ID, infraction.InfractionID, infraction.EntryID, infraction.ClaimID, infraction.Key,
		infraction.Type, infraction.Description, infraction.EvidenceURLs,
		infraction.ReporterParticipant, infraction.ReportedParticipant,
		infraction.Status, infraction.ResolutionNotes,
		infraction.ReportedAt, infraction.InvestigatedAt, infraction.ResolvedAt,
		infraction.CreatedAt, infraction.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to create infraction: %s", infraction.InfractionID)
		return fmt.Errorf("failed to insert infraction: %w", err)
	}

	r.logger.WithField("infraction_id", infraction.InfractionID).Info("Infraction created successfully")
	return nil
}

// GetByID retrieves an infraction by UUID
func (r *InfractionRepository) GetByID(ctx context.Context, id string) (*entities.Infraction, error) {
	query := `
		SELECT 
			id, infraction_id, entry_id, claim_id, key,
			type, description, evidence_urls,
			reporter_participant, reported_participant,
			status, resolution_notes,
			reported_at, investigated_at, resolved_at,
			created_at, updated_at, deleted_at
		FROM infractions
		WHERE id = $1 AND deleted_at IS NULL
	`

	infraction := &entities.Infraction{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&infraction.ID, &infraction.InfractionID, &infraction.EntryID, &infraction.ClaimID, &infraction.Key,
		&infraction.Type, &infraction.Description, &infraction.EvidenceURLs,
		&infraction.ReporterParticipant, &infraction.ReportedParticipant,
		&infraction.Status, &infraction.ResolutionNotes,
		&infraction.ReportedAt, &infraction.InvestigatedAt, &infraction.ResolvedAt,
		&infraction.CreatedAt, &infraction.UpdatedAt, &infraction.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("infraction not found: %s", id)
	}

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to get infraction by ID: %s", id)
		return nil, fmt.Errorf("failed to query infraction: %w", err)
	}

	return infraction, nil
}

// GetByInfractionID retrieves an infraction by infraction_id
func (r *InfractionRepository) GetByInfractionID(ctx context.Context, infractionID string) (*entities.Infraction, error) {
	query := `
		SELECT 
			id, infraction_id, entry_id, claim_id, key,
			type, description, evidence_urls,
			reporter_participant, reported_participant,
			status, resolution_notes,
			reported_at, investigated_at, resolved_at,
			created_at, updated_at, deleted_at
		FROM infractions
		WHERE infraction_id = $1 AND deleted_at IS NULL
	`

	infraction := &entities.Infraction{}
	err := r.db.QueryRow(ctx, query, infractionID).Scan(
		&infraction.ID, &infraction.InfractionID, &infraction.EntryID, &infraction.ClaimID, &infraction.Key,
		&infraction.Type, &infraction.Description, &infraction.EvidenceURLs,
		&infraction.ReporterParticipant, &infraction.ReportedParticipant,
		&infraction.Status, &infraction.ResolutionNotes,
		&infraction.ReportedAt, &infraction.InvestigatedAt, &infraction.ResolvedAt,
		&infraction.CreatedAt, &infraction.UpdatedAt, &infraction.DeletedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("infraction not found: %s", infractionID)
	}

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to get infraction by infraction_id: %s", infractionID)
		return nil, fmt.Errorf("failed to query infraction: %w", err)
	}

	return infraction, nil
}

// Update updates an existing infraction
func (r *InfractionRepository) Update(ctx context.Context, infraction *entities.Infraction) error {
	query := `
		UPDATE infractions SET
			entry_id = $1, claim_id = $2, key = $3,
			type = $4, description = $5, evidence_urls = $6,
			reporter_participant = $7, reported_participant = $8,
			status = $9, resolution_notes = $10,
			reported_at = $11, investigated_at = $12, resolved_at = $13,
			updated_at = $14
		WHERE infraction_id = $15 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query,
		infraction.EntryID, infraction.ClaimID, infraction.Key,
		infraction.Type, infraction.Description, infraction.EvidenceURLs,
		infraction.ReporterParticipant, infraction.ReportedParticipant,
		infraction.Status, infraction.ResolutionNotes,
		infraction.ReportedAt, infraction.InvestigatedAt, infraction.ResolvedAt,
		infraction.UpdatedAt,
		infraction.InfractionID,
	)

	if err != nil {
		r.logger.WithError(err).Errorf("Failed to update infraction: %s", infraction.InfractionID)
		return fmt.Errorf("failed to update infraction: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("infraction not found or already deleted: %s", infraction.InfractionID)
	}

	r.logger.WithField("infraction_id", infraction.InfractionID).Info("Infraction updated successfully")
	return nil
}

// UpdateStatus updates the status of an infraction with resolution notes
func (r *InfractionRepository) UpdateStatus(
	ctx context.Context,
	infractionID string,
	status entities.InfractionStatus,
	notes *string,
) error {
	query := `
		UPDATE infractions SET
			status = $1,
			resolution_notes = $2,
			updated_at = NOW()
		WHERE infraction_id = $3 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query, status, notes, infractionID)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to update infraction status: %s", infractionID)
		return fmt.Errorf("failed to update infraction status: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("infraction not found or already deleted: %s", infractionID)
	}

	r.logger.WithFields(logrus.Fields{
		"infraction_id": infractionID,
		"status":        status,
	}).Info("Infraction status updated successfully")

	return nil
}

// Delete performs a soft delete on an infraction
func (r *InfractionRepository) Delete(ctx context.Context, infractionID string) error {
	query := `
		UPDATE infractions SET
			deleted_at = NOW(),
			updated_at = NOW()
		WHERE infraction_id = $1 AND deleted_at IS NULL
	`

	cmdTag, err := r.db.Exec(ctx, query, infractionID)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to delete infraction: %s", infractionID)
		return fmt.Errorf("failed to delete infraction: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("infraction not found or already deleted: %s", infractionID)
	}

	r.logger.WithField("infraction_id", infractionID).Info("Infraction deleted successfully")
	return nil
}

// ListByKey lists all infractions for a given PIX key
func (r *InfractionRepository) ListByKey(ctx context.Context, key string, limit, offset int) ([]*entities.Infraction, error) {
	query := `
		SELECT 
			id, infraction_id, entry_id, claim_id, key,
			type, description, evidence_urls,
			reporter_participant, reported_participant,
			status, resolution_notes,
			reported_at, investigated_at, resolved_at,
			created_at, updated_at, deleted_at
		FROM infractions
		WHERE key = $1 AND deleted_at IS NULL
		ORDER BY reported_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, key, limit, offset)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to list infractions for key: %s", key)
		return nil, fmt.Errorf("failed to list infractions: %w", err)
	}
	defer rows.Close()

	return r.scanInfractions(rows)
}

// ListByReporter lists infractions reported by a participant
func (r *InfractionRepository) ListByReporter(ctx context.Context, ispb string, limit, offset int) ([]*entities.Infraction, error) {
	query := `
		SELECT 
			id, infraction_id, entry_id, claim_id, key,
			type, description, evidence_urls,
			reporter_participant, reported_participant,
			status, resolution_notes,
			reported_at, investigated_at, resolved_at,
			created_at, updated_at, deleted_at
		FROM infractions
		WHERE reporter_participant = $1 AND deleted_at IS NULL
		ORDER BY reported_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, ispb, limit, offset)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to list infractions for reporter: %s", ispb)
		return nil, fmt.Errorf("failed to list infractions: %w", err)
	}
	defer rows.Close()

	return r.scanInfractions(rows)
}

// ListByStatus lists infractions by status
func (r *InfractionRepository) ListByStatus(
	ctx context.Context,
	status entities.InfractionStatus,
	limit, offset int,
) ([]*entities.Infraction, error) {
	query := `
		SELECT 
			id, infraction_id, entry_id, claim_id, key,
			type, description, evidence_urls,
			reporter_participant, reported_participant,
			status, resolution_notes,
			reported_at, investigated_at, resolved_at,
			created_at, updated_at, deleted_at
		FROM infractions
		WHERE status = $1 AND deleted_at IS NULL
		ORDER BY reported_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, status, limit, offset)
	if err != nil {
		r.logger.WithError(err).Errorf("Failed to list infractions by status: %s", status)
		return nil, fmt.Errorf("failed to list infractions: %w", err)
	}
	defer rows.Close()

	return r.scanInfractions(rows)
}

// ListOpen lists open infractions needing investigation
func (r *InfractionRepository) ListOpen(ctx context.Context, limit int) ([]*entities.Infraction, error) {
	query := `
		SELECT 
			id, infraction_id, entry_id, claim_id, key,
			type, description, evidence_urls,
			reporter_participant, reported_participant,
			status, resolution_notes,
			reported_at, investigated_at, resolved_at,
			created_at, updated_at, deleted_at
		FROM infractions
		WHERE status IN ('OPEN', 'UNDER_INVESTIGATION') 
		  AND deleted_at IS NULL
		ORDER BY reported_at ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list open infractions")
		return nil, fmt.Errorf("failed to list open infractions: %w", err)
	}
	defer rows.Close()

	return r.scanInfractions(rows)
}

// scanInfractions is a helper method to scan multiple infraction rows
func (r *InfractionRepository) scanInfractions(rows pgx.Rows) ([]*entities.Infraction, error) {
	var infractions []*entities.Infraction

	for rows.Next() {
		infraction := &entities.Infraction{}
		err := rows.Scan(
			&infraction.ID, &infraction.InfractionID, &infraction.EntryID, &infraction.ClaimID, &infraction.Key,
			&infraction.Type, &infraction.Description, &infraction.EvidenceURLs,
			&infraction.ReporterParticipant, &infraction.ReportedParticipant,
			&infraction.Status, &infraction.ResolutionNotes,
			&infraction.ReportedAt, &infraction.InvestigatedAt, &infraction.ResolvedAt,
			&infraction.CreatedAt, &infraction.UpdatedAt, &infraction.DeletedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan infraction row")
			return nil, fmt.Errorf("failed to scan infraction: %w", err)
		}
		infractions = append(infractions, infraction)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating infraction rows")
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return infractions, nil
}
