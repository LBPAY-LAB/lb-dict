package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
	"github.com/lb-conn/connector-dict/domain/ratelimit"
)

// alertRepository implements ports.AlertRepository using PostgreSQL
type alertRepository struct {
	pool *pgxpool.Pool
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(pool *pgxpool.Pool) ports.AlertRepository {
	return &alertRepository{
		pool: pool,
	}
}

// Save inserts a new alert
func (r *alertRepository) Save(ctx context.Context, alert *ratelimit.Alert) error {
	ctx, span := tracer.Start(ctx, "AlertRepository.Save")
	defer span.End()

	span.SetAttributes(
		attribute.String("endpoint_id", alert.EndpointID),
		attribute.String("severity", string(alert.Severity)),
	)

	query := `
		INSERT INTO dict_rate_limit_alerts
		(endpoint_id, severity, threshold_percent, available_tokens, capacity,
		 utilization_percent, consumption_rate_per_minute, recovery_eta_seconds,
		 exhaustion_projection_seconds, psp_category, message, resolved,
		 resolved_at, resolution_notes, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id
	`

	var id int64
	err := r.pool.QueryRow(ctx, query,
		alert.EndpointID,
		string(alert.Severity),
		alert.ThresholdPercent,
		alert.AvailableTokens,
		alert.Capacity,
		alert.UtilizationPercent,
		alert.ConsumptionRatePerMinute,
		alert.RecoveryETASeconds,
		alert.ExhaustionProjectionSeconds,
		nullableString(alert.PSPCategory),
		alert.Message,
		alert.Resolved,
		nullableTime(alert.ResolvedAt),
		nullableString(alert.ResolutionNotes),
		alert.CreatedAt.UTC(),
	).Scan(&id)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Insert failed")
		return fmt.Errorf("failed to save alert: %w", err)
	}

	alert.ID = id
	span.SetAttributes(attribute.Int64("alert_id", id))
	span.SetStatus(codes.Ok, "Alert saved successfully")

	return nil
}

// GetUnresolved retrieves all unresolved alerts
func (r *alertRepository) GetUnresolved(ctx context.Context) ([]*ratelimit.Alert, error) {
	ctx, span := tracer.Start(ctx, "AlertRepository.GetUnresolved")
	defer span.End()

	query := `
		SELECT id, endpoint_id, severity, threshold_percent, available_tokens,
		       capacity, utilization_percent, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, psp_category,
		       message, resolved, resolved_at, resolution_notes, created_at
		FROM dict_rate_limit_alerts
		WHERE NOT resolved
		ORDER BY severity DESC, created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query unresolved alerts: %w", err)
	}
	defer rows.Close()

	alerts := make([]*ratelimit.Alert, 0)

	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("alert_count", len(alerts)))
	span.SetStatus(codes.Ok, "Unresolved alerts retrieved successfully")

	return alerts, nil
}

// GetUnresolvedByEndpoint retrieves unresolved alerts for a specific endpoint
func (r *alertRepository) GetUnresolvedByEndpoint(ctx context.Context, endpointID string) ([]*ratelimit.Alert, error) {
	ctx, span := tracer.Start(ctx, "AlertRepository.GetUnresolvedByEndpoint")
	defer span.End()

	span.SetAttributes(attribute.String("endpoint_id", endpointID))

	query := `
		SELECT id, endpoint_id, severity, threshold_percent, available_tokens,
		       capacity, utilization_percent, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, psp_category,
		       message, resolved, resolved_at, resolution_notes, created_at
		FROM dict_rate_limit_alerts
		WHERE endpoint_id = $1 AND NOT resolved
		ORDER BY severity DESC, created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, endpointID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query unresolved alerts by endpoint: %w", err)
	}
	defer rows.Close()

	alerts := make([]*ratelimit.Alert, 0)

	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("alert_count", len(alerts)))
	span.SetStatus(codes.Ok, "Unresolved alerts by endpoint retrieved successfully")

	return alerts, nil
}

// GetUnresolvedBySeverity retrieves unresolved alerts by severity
func (r *alertRepository) GetUnresolvedBySeverity(ctx context.Context, severity ratelimit.AlertSeverity) ([]*ratelimit.Alert, error) {
	ctx, span := tracer.Start(ctx, "AlertRepository.GetUnresolvedBySeverity")
	defer span.End()

	span.SetAttributes(attribute.String("severity", string(severity)))

	query := `
		SELECT id, endpoint_id, severity, threshold_percent, available_tokens,
		       capacity, utilization_percent, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, psp_category,
		       message, resolved, resolved_at, resolution_notes, created_at
		FROM dict_rate_limit_alerts
		WHERE severity = $1 AND NOT resolved
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, string(severity))
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query unresolved alerts by severity: %w", err)
	}
	defer rows.Close()

	alerts := make([]*ratelimit.Alert, 0)

	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("alert_count", len(alerts)))
	span.SetStatus(codes.Ok, "Unresolved alerts by severity retrieved successfully")

	return alerts, nil
}

// Resolve marks an alert as resolved
func (r *alertRepository) Resolve(ctx context.Context, alertID int64, notes string) error {
	ctx, span := tracer.Start(ctx, "AlertRepository.Resolve")
	defer span.End()

	span.SetAttributes(attribute.Int64("alert_id", alertID))

	query := `
		UPDATE dict_rate_limit_alerts
		SET resolved = TRUE,
		    resolved_at = $1,
		    resolution_notes = $2
		WHERE id = $3 AND NOT resolved
	`

	now := time.Now().UTC()

	result, err := r.pool.Exec(ctx, query, now, notes, alertID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Update failed")
		return fmt.Errorf("failed to resolve alert: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		span.SetStatus(codes.Ok, "Alert not found or already resolved")
		return fmt.Errorf("alert %d not found or already resolved", alertID)
	}

	span.SetStatus(codes.Ok, "Alert resolved successfully")

	return nil
}

// ResolveBulk marks multiple alerts as resolved
func (r *alertRepository) ResolveBulk(ctx context.Context, alertIDs []int64, notes string) error {
	ctx, span := tracer.Start(ctx, "AlertRepository.ResolveBulk")
	defer span.End()

	span.SetAttributes(attribute.Int("alert_count", len(alertIDs)))

	if len(alertIDs) == 0 {
		return nil
	}

	// Use transaction for bulk operation
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE dict_rate_limit_alerts
		SET resolved = TRUE,
		    resolved_at = $1,
		    resolution_notes = $2
		WHERE id = ANY($3) AND NOT resolved
	`

	now := time.Now().UTC()

	result, err := tx.Exec(ctx, query, now, notes, alertIDs)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Bulk update failed")
		return fmt.Errorf("failed to resolve alerts in bulk: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	rowsAffected := result.RowsAffected()
	span.SetAttributes(attribute.Int64("alerts_resolved", rowsAffected))
	span.SetStatus(codes.Ok, "Alerts resolved in bulk successfully")

	return nil
}

// AutoResolve automatically resolves alerts based on current state
// Returns the number of alerts resolved
// Calls the database function auto_resolve_alerts()
func (r *alertRepository) AutoResolve(ctx context.Context, endpointID string, availableTokens, capacity int) (int, error) {
	ctx, span := tracer.Start(ctx, "AlertRepository.AutoResolve")
	defer span.End()

	span.SetAttributes(
		attribute.String("endpoint_id", endpointID),
		attribute.Int("available_tokens", availableTokens),
		attribute.Int("capacity", capacity),
	)

	// Call database function that implements auto-resolution logic
	query := `SELECT auto_resolve_alerts($1, $2, $3)`

	var resolvedCount int
	err := r.pool.QueryRow(ctx, query, endpointID, availableTokens, capacity).Scan(&resolvedCount)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Auto-resolve function failed")
		return 0, fmt.Errorf("failed to auto-resolve alerts: %w", err)
	}

	span.SetAttributes(attribute.Int("alerts_resolved", resolvedCount))
	span.SetStatus(codes.Ok, "Auto-resolve completed successfully")

	return resolvedCount, nil
}

// GetHistory retrieves alert history in a time range
func (r *alertRepository) GetHistory(ctx context.Context, since, until time.Time) ([]*ratelimit.Alert, error) {
	ctx, span := tracer.Start(ctx, "AlertRepository.GetHistory")
	defer span.End()

	span.SetAttributes(
		attribute.String("since", since.Format(time.RFC3339)),
		attribute.String("until", until.Format(time.RFC3339)),
	)

	query := `
		SELECT id, endpoint_id, severity, threshold_percent, available_tokens,
		       capacity, utilization_percent, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, psp_category,
		       message, resolved, resolved_at, resolution_notes, created_at
		FROM dict_rate_limit_alerts
		WHERE created_at >= $1 AND created_at <= $2
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, since.UTC(), until.UTC())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query alert history: %w", err)
	}
	defer rows.Close()

	alerts := make([]*ratelimit.Alert, 0)

	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("alert_count", len(alerts)))
	span.SetStatus(codes.Ok, "Alert history retrieved successfully")

	return alerts, nil
}

// GetHistoryByEndpoint retrieves alert history for a specific endpoint
func (r *alertRepository) GetHistoryByEndpoint(ctx context.Context, endpointID string, since, until time.Time) ([]*ratelimit.Alert, error) {
	ctx, span := tracer.Start(ctx, "AlertRepository.GetHistoryByEndpoint")
	defer span.End()

	span.SetAttributes(
		attribute.String("endpoint_id", endpointID),
		attribute.String("since", since.Format(time.RFC3339)),
		attribute.String("until", until.Format(time.RFC3339)),
	)

	query := `
		SELECT id, endpoint_id, severity, threshold_percent, available_tokens,
		       capacity, utilization_percent, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, psp_category,
		       message, resolved, resolved_at, resolution_notes, created_at
		FROM dict_rate_limit_alerts
		WHERE endpoint_id = $1
		  AND created_at >= $2
		  AND created_at <= $3
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query, endpointID, since.UTC(), until.UTC())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query alert history by endpoint: %w", err)
	}
	defer rows.Close()

	alerts := make([]*ratelimit.Alert, 0)

	for rows.Next() {
		alert, err := scanAlert(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan alert: %w", err)
		}
		alerts = append(alerts, alert)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("alert_count", len(alerts)))
	span.SetStatus(codes.Ok, "Alert history by endpoint retrieved successfully")

	return alerts, nil
}

// Helper function to scan Alert from row
func scanAlert(s scanner) (*ratelimit.Alert, error) {
	var (
		id                          int64
		endpointID                  string
		severity                    string
		thresholdPercent            int
		availableTokens             int
		capacity                    int
		utilizationPercent          float64
		consumptionRatePerMinute    float64
		recoveryETASeconds          int
		exhaustionProjectionSeconds int
		pspCategory                 *string
		message                     string
		resolved                    bool
		resolvedAt                  *time.Time
		resolutionNotes             *string
		createdAt                   time.Time
	)

	err := s.Scan(
		&id,
		&endpointID,
		&severity,
		&thresholdPercent,
		&availableTokens,
		&capacity,
		&utilizationPercent,
		&consumptionRatePerMinute,
		&recoveryETASeconds,
		&exhaustionProjectionSeconds,
		&pspCategory,
		&message,
		&resolved,
		&resolvedAt,
		&resolutionNotes,
		&createdAt,
	)

	if err != nil {
		return nil, err
	}

	category := ""
	if pspCategory != nil {
		category = *pspCategory
	}

	notes := ""
	if resolutionNotes != nil {
		notes = *resolutionNotes
	}

	alert := &ratelimit.Alert{
		ID:                          id,
		EndpointID:                  endpointID,
		Severity:                    ratelimit.AlertSeverity(severity),
		ThresholdPercent:            thresholdPercent,
		AvailableTokens:             availableTokens,
		Capacity:                    capacity,
		UtilizationPercent:          utilizationPercent,
		ConsumptionRatePerMinute:    consumptionRatePerMinute,
		RecoveryETASeconds:          recoveryETASeconds,
		ExhaustionProjectionSeconds: exhaustionProjectionSeconds,
		PSPCategory:                 category,
		Message:                     message,
		Resolved:                    resolved,
		ResolvedAt:                  resolvedAt,
		ResolutionNotes:             notes,
		CreatedAt:                   createdAt.UTC(),
	}

	return alert, nil
}

// Helper function for nullable time.Time
func nullableTime(t *time.Time) *time.Time {
	if t == nil {
		return nil
	}
	utc := t.UTC()
	return &utc
}
