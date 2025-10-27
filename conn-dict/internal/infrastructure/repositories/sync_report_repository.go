package repositories

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/sirupsen/logrus"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/database"
)

// SyncReportRepository handles persistence of VSYNC reports
type SyncReportRepository struct {
	db     *database.PostgresClient
	logger *logrus.Logger
}

// NewSyncReportRepository creates a new SyncReportRepository
func NewSyncReportRepository(db *database.PostgresClient, logger *logrus.Logger) *SyncReportRepository {
	return &SyncReportRepository{
		db:     db,
		logger: logger,
	}
}

// Create inserts a new sync report
func (r *SyncReportRepository) Create(ctx context.Context, report *entities.SyncReport) error {
	metadataJSON, err := json.Marshal(report.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO sync_reports (
			id, sync_id, sync_type, sync_timestamp, participant_ispb,
			entries_fetched, entries_compared, entries_synced,
			entries_created, entries_updated, entries_deleted,
			discrepancies_found, discrepancies_missing_local,
			discrepancies_outdated_local, discrepancies_missing_bacen,
			status, duration_ms, error_message, error_code,
			metadata, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15,
			$16, $17, $18, $19,
			$20, $21, $22
		)
	`

	_, err = r.db.Exec(ctx, query,
		report.ID, report.SyncID, report.SyncType, report.SyncTimestamp, report.ParticipantISPB,
		report.EntriesFetched, report.EntriesCompared, report.EntriesSynced,
		report.EntriesCreated, report.EntriesUpdated, report.EntriesDeleted,
		report.DiscrepanciesFound, report.DiscrepanciesMissingLocal,
		report.DiscrepanciesOutdatedLocal, report.DiscrepanciesMissingBacen,
		report.Status, report.DurationMS, report.ErrorMessage, report.ErrorCode,
		metadataJSON, report.CreatedAt, report.UpdatedAt,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to create sync report")
		return fmt.Errorf("failed to create sync report: %w", err)
	}

	r.logger.WithFields(logrus.Fields{
		"sync_id":          report.SyncID,
		"participant_ispb": report.ParticipantISPB,
		"status":           report.Status,
		"discrepancies":    report.DiscrepanciesFound,
	}).Info("Sync report created")

	return nil
}

// GetByID retrieves a sync report by ID
func (r *SyncReportRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.SyncReport, error) {
	query := `
		SELECT
			id, sync_id, sync_type, sync_timestamp, participant_ispb,
			entries_fetched, entries_compared, entries_synced,
			entries_created, entries_updated, entries_deleted,
			discrepancies_found, discrepancies_missing_local,
			discrepancies_outdated_local, discrepancies_missing_bacen,
			status, duration_ms, error_message, error_code,
			metadata, created_at, updated_at
		FROM sync_reports
		WHERE id = $1
	`

	var report entities.SyncReport
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(
		&report.ID, &report.SyncID, &report.SyncType, &report.SyncTimestamp, &report.ParticipantISPB,
		&report.EntriesFetched, &report.EntriesCompared, &report.EntriesSynced,
		&report.EntriesCreated, &report.EntriesUpdated, &report.EntriesDeleted,
		&report.DiscrepanciesFound, &report.DiscrepanciesMissingLocal,
		&report.DiscrepanciesOutdatedLocal, &report.DiscrepanciesMissingBacen,
		&report.Status, &report.DurationMS, &report.ErrorMessage, &report.ErrorCode,
		&metadataJSON, &report.CreatedAt, &report.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("sync report not found: %s", id)
		}
		r.logger.WithError(err).Error("Failed to get sync report")
		return nil, fmt.Errorf("failed to get sync report: %w", err)
	}

	// Unmarshal metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &report.Metadata); err != nil {
			r.logger.WithError(err).Warn("Failed to unmarshal metadata")
		}
	}

	return &report, nil
}

// GetBySyncID retrieves a sync report by Temporal workflow execution ID
func (r *SyncReportRepository) GetBySyncID(ctx context.Context, syncID string) (*entities.SyncReport, error) {
	query := `
		SELECT
			id, sync_id, sync_type, sync_timestamp, participant_ispb,
			entries_fetched, entries_compared, entries_synced,
			entries_created, entries_updated, entries_deleted,
			discrepancies_found, discrepancies_missing_local,
			discrepancies_outdated_local, discrepancies_missing_bacen,
			status, duration_ms, error_message, error_code,
			metadata, created_at, updated_at
		FROM sync_reports
		WHERE sync_id = $1
	`

	var report entities.SyncReport
	var metadataJSON []byte

	err := r.db.QueryRow(ctx, query, syncID).Scan(
		&report.ID, &report.SyncID, &report.SyncType, &report.SyncTimestamp, &report.ParticipantISPB,
		&report.EntriesFetched, &report.EntriesCompared, &report.EntriesSynced,
		&report.EntriesCreated, &report.EntriesUpdated, &report.EntriesDeleted,
		&report.DiscrepanciesFound, &report.DiscrepanciesMissingLocal,
		&report.DiscrepanciesOutdatedLocal, &report.DiscrepanciesMissingBacen,
		&report.Status, &report.DurationMS, &report.ErrorMessage, &report.ErrorCode,
		&metadataJSON, &report.CreatedAt, &report.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("sync report not found: %s", syncID)
		}
		r.logger.WithError(err).Error("Failed to get sync report")
		return nil, fmt.Errorf("failed to get sync report: %w", err)
	}

	// Unmarshal metadata
	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &report.Metadata); err != nil {
			r.logger.WithError(err).Warn("Failed to unmarshal metadata")
		}
	}

	return &report, nil
}

// List retrieves sync reports with pagination
func (r *SyncReportRepository) List(ctx context.Context, limit, offset int) ([]*entities.SyncReport, error) {
	query := `
		SELECT
			id, sync_id, sync_type, sync_timestamp, participant_ispb,
			entries_fetched, entries_compared, entries_synced,
			entries_created, entries_updated, entries_deleted,
			discrepancies_found, discrepancies_missing_local,
			discrepancies_outdated_local, discrepancies_missing_bacen,
			status, duration_ms, error_message, error_code,
			metadata, created_at, updated_at
		FROM sync_reports
		ORDER BY sync_timestamp DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list sync reports")
		return nil, fmt.Errorf("failed to list sync reports: %w", err)
	}
	defer rows.Close()

	var reports []*entities.SyncReport
	for rows.Next() {
		var report entities.SyncReport
		var metadataJSON []byte

		err := rows.Scan(
			&report.ID, &report.SyncID, &report.SyncType, &report.SyncTimestamp, &report.ParticipantISPB,
			&report.EntriesFetched, &report.EntriesCompared, &report.EntriesSynced,
			&report.EntriesCreated, &report.EntriesUpdated, &report.EntriesDeleted,
			&report.DiscrepanciesFound, &report.DiscrepanciesMissingLocal,
			&report.DiscrepanciesOutdatedLocal, &report.DiscrepanciesMissingBacen,
			&report.Status, &report.DurationMS, &report.ErrorMessage, &report.ErrorCode,
			&metadataJSON, &report.CreatedAt, &report.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan sync report")
			continue
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &report.Metadata); err != nil {
				r.logger.WithError(err).Warn("Failed to unmarshal metadata")
			}
		}

		reports = append(reports, &report)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating sync reports")
		return nil, fmt.Errorf("error iterating sync reports: %w", err)
	}

	return reports, nil
}

// ListByParticipant retrieves sync reports for a specific participant
func (r *SyncReportRepository) ListByParticipant(
	ctx context.Context,
	participantISPB string,
	limit, offset int,
) ([]*entities.SyncReport, error) {
	query := `
		SELECT
			id, sync_id, sync_type, sync_timestamp, participant_ispb,
			entries_fetched, entries_compared, entries_synced,
			entries_created, entries_updated, entries_deleted,
			discrepancies_found, discrepancies_missing_local,
			discrepancies_outdated_local, discrepancies_missing_bacen,
			status, duration_ms, error_message, error_code,
			metadata, created_at, updated_at
		FROM sync_reports
		WHERE participant_ispb = $1
		ORDER BY sync_timestamp DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, participantISPB, limit, offset)
	if err != nil {
		r.logger.WithError(err).Error("Failed to list sync reports by participant")
		return nil, fmt.Errorf("failed to list sync reports: %w", err)
	}
	defer rows.Close()

	var reports []*entities.SyncReport
	for rows.Next() {
		var report entities.SyncReport
		var metadataJSON []byte

		err := rows.Scan(
			&report.ID, &report.SyncID, &report.SyncType, &report.SyncTimestamp, &report.ParticipantISPB,
			&report.EntriesFetched, &report.EntriesCompared, &report.EntriesSynced,
			&report.EntriesCreated, &report.EntriesUpdated, &report.EntriesDeleted,
			&report.DiscrepanciesFound, &report.DiscrepanciesMissingLocal,
			&report.DiscrepanciesOutdatedLocal, &report.DiscrepanciesMissingBacen,
			&report.Status, &report.DurationMS, &report.ErrorMessage, &report.ErrorCode,
			&metadataJSON, &report.CreatedAt, &report.UpdatedAt,
		)
		if err != nil {
			r.logger.WithError(err).Error("Failed to scan sync report")
			continue
		}

		// Unmarshal metadata
		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &report.Metadata); err != nil {
				r.logger.WithError(err).Warn("Failed to unmarshal metadata")
			}
		}

		reports = append(reports, &report)
	}

	if err := rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating sync reports")
		return nil, fmt.Errorf("error iterating sync reports: %w", err)
	}

	return reports, nil
}

// Update updates an existing sync report
func (r *SyncReportRepository) Update(ctx context.Context, report *entities.SyncReport) error {
	metadataJSON, err := json.Marshal(report.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE sync_reports SET
			sync_type = $2,
			sync_timestamp = $3,
			participant_ispb = $4,
			entries_fetched = $5,
			entries_compared = $6,
			entries_synced = $7,
			entries_created = $8,
			entries_updated = $9,
			entries_deleted = $10,
			discrepancies_found = $11,
			discrepancies_missing_local = $12,
			discrepancies_outdated_local = $13,
			discrepancies_missing_bacen = $14,
			status = $15,
			duration_ms = $16,
			error_message = $17,
			error_code = $18,
			metadata = $19
		WHERE id = $1
	`

	_, err = r.db.Exec(ctx, query,
		report.ID,
		report.SyncType, report.SyncTimestamp, report.ParticipantISPB,
		report.EntriesFetched, report.EntriesCompared, report.EntriesSynced,
		report.EntriesCreated, report.EntriesUpdated, report.EntriesDeleted,
		report.DiscrepanciesFound, report.DiscrepanciesMissingLocal,
		report.DiscrepanciesOutdatedLocal, report.DiscrepanciesMissingBacen,
		report.Status, report.DurationMS, report.ErrorMessage, report.ErrorCode,
		metadataJSON,
	)

	if err != nil {
		r.logger.WithError(err).Error("Failed to update sync report")
		return fmt.Errorf("failed to update sync report: %w", err)
	}

	return nil
}