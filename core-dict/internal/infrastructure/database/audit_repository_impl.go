package database

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/repositories"
)

// PostgresAuditRepository implements AuditRepository using PostgreSQL
type PostgresAuditRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresAuditRepository creates a new audit repository
func NewPostgresAuditRepository(pool *pgxpool.Pool) repositories.AuditRepository {
	return &PostgresAuditRepository{
		pool: pool,
	}
}

// FindByEntityID finds audit events by entity ID
func (r *PostgresAuditRepository) FindByEntityID(ctx context.Context, entityType entities.EntityType, entityID uuid.UUID, limit, offset int) ([]*entities.AuditEvent, error) {
	query := `
		SELECT
			event_id, entity_type, entity_id, event_type,
			user_id, old_values, new_values, metadata,
			ip_address, user_agent, occurred_at
		FROM audit.entry_events
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY occurred_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, entityType, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find audit events: %w", err)
	}
	defer rows.Close()

	var events []*entities.AuditEvent
	for rows.Next() {
		event, err := scanAuditEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// FindByActor finds audit logs by actor ID
func (r *PostgresAuditRepository) FindByActor(ctx context.Context, actorID uuid.UUID, limit, offset int) ([]*entities.AuditLog, error) {
	query := `
		SELECT
			event_id, entity_type, entity_id, event_type,
			user_id, old_values, new_values, metadata, occurred_at
		FROM audit.entry_events
		WHERE user_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, actorID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find audit logs: %w", err)
	}
	defer rows.Close()

	var logs []*entities.AuditLog
	for rows.Next() {
		var log entities.AuditLog
		var oldValuesJSON, newValuesJSON, metadataJSON []byte
		var userID *uuid.UUID

		err := rows.Scan(
			&log.ID,
			&log.EntityType,
			&log.EntityID,
			&log.Action,
			&userID,
			&oldValuesJSON,
			&newValuesJSON,
			&metadataJSON,
			&log.Timestamp,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %w", err)
		}

		if userID != nil {
			log.ActorID = *userID
		}

		// Unmarshal JSONB fields
		if len(oldValuesJSON) > 0 {
			if err := json.Unmarshal(oldValuesJSON, &log.Changes); err != nil {
				return nil, fmt.Errorf("failed to unmarshal old_values: %w", err)
			}
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &log.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		log.ActorType = "user"
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return logs, nil
}

// Create creates a new audit event
func (r *PostgresAuditRepository) Create(ctx context.Context, event *entities.AuditEvent) error {
	query := `
		INSERT INTO audit.entry_events (
			event_id, entity_type, entity_id, event_type,
			user_id, old_values, new_values, metadata,
			ip_address, user_agent, occurred_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`

	// Marshal maps to JSONB
	oldValuesJSON, err := json.Marshal(event.OldValues)
	if err != nil {
		return fmt.Errorf("failed to marshal old values: %w", err)
	}

	newValuesJSON, err := json.Marshal(event.NewValues)
	if err != nil {
		return fmt.Errorf("failed to marshal new values: %w", err)
	}

	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.pool.Exec(ctx, query,
		event.EventID,
		event.EntityType,
		event.EntityID,
		event.EventType,
		event.UserID,
		oldValuesJSON,
		newValuesJSON,
		metadataJSON,
		event.IPAddress,
		event.UserAgent,
		event.OccurredAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create audit event: %w", err)
	}

	return nil
}

// FindByID finds an audit event by ID
func (r *PostgresAuditRepository) FindByID(ctx context.Context, eventID uuid.UUID) (*entities.AuditEvent, error) {
	query := `
		SELECT
			event_id, entity_type, entity_id, event_type,
			user_id, old_values, new_values, metadata,
			ip_address, user_agent, occurred_at
		FROM audit.entry_events
		WHERE event_id = $1
		LIMIT 1
	`

	var event entities.AuditEvent
	var oldValuesJSON, newValuesJSON, metadataJSON []byte

	err := r.pool.QueryRow(ctx, query, eventID).Scan(
		&event.EventID,
		&event.EntityType,
		&event.EntityID,
		&event.EventType,
		&event.UserID,
		&oldValuesJSON,
		&newValuesJSON,
		&metadataJSON,
		&event.IPAddress,
		&event.UserAgent,
		&event.OccurredAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to find audit event: %w", err)
	}

	// Unmarshal JSONB fields
	if len(oldValuesJSON) > 0 {
		if err := json.Unmarshal(oldValuesJSON, &event.OldValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal old values: %w", err)
		}
	}

	if len(newValuesJSON) > 0 {
		if err := json.Unmarshal(newValuesJSON, &event.NewValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal new values: %w", err)
		}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	event.ID = event.EventID
	return &event, nil
}

// FindByEventType finds audit events by event type with pagination
func (r *PostgresAuditRepository) FindByEventType(ctx context.Context, eventType entities.EventType, limit, offset int) ([]*entities.AuditEvent, error) {
	query := `
		SELECT
			event_id, entity_type, entity_id, event_type,
			user_id, old_values, new_values, metadata,
			ip_address, user_agent, occurred_at
		FROM audit.entry_events
		WHERE event_type = $1
		ORDER BY occurred_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, eventType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find audit events by type: %w", err)
	}
	defer rows.Close()

	var events []*entities.AuditEvent
	for rows.Next() {
		event, err := scanAuditEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// FindByUserID finds audit events by user ID with pagination
func (r *PostgresAuditRepository) FindByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*entities.AuditEvent, error) {
	query := `
		SELECT
			event_id, entity_type, entity_id, event_type,
			user_id, old_values, new_values, metadata,
			ip_address, user_agent, occurred_at
		FROM audit.entry_events
		WHERE user_id = $1
		ORDER BY occurred_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find audit events by user: %w", err)
	}
	defer rows.Close()

	var events []*entities.AuditEvent
	for rows.Next() {
		event, err := scanAuditEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// FindByDateRange finds audit events within a date range
func (r *PostgresAuditRepository) FindByDateRange(ctx context.Context, from, to time.Time, limit, offset int) ([]*entities.AuditEvent, error) {
	query := `
		SELECT
			event_id, entity_type, entity_id, event_type,
			user_id, old_values, new_values, metadata,
			ip_address, user_agent, occurred_at
		FROM audit.entry_events
		WHERE occurred_at BETWEEN $1 AND $2
		ORDER BY occurred_at DESC
		LIMIT $3 OFFSET $4
	`

	rows, err := r.pool.Query(ctx, query, from, to, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to find audit events by date range: %w", err)
	}
	defer rows.Close()

	var events []*entities.AuditEvent
	for rows.Next() {
		event, err := scanAuditEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// List lists audit events with filters and pagination
func (r *PostgresAuditRepository) List(ctx context.Context, filters repositories.AuditFilters) ([]*entities.AuditEvent, error) {
	query := `
		SELECT
			event_id, entity_type, entity_id, event_type,
			user_id, old_values, new_values, metadata,
			ip_address, user_agent, occurred_at
		FROM audit.entry_events
		WHERE 1=1
	`

	args := []interface{}{}
	argPos := 1

	if filters.EventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argPos)
		args = append(args, *filters.EventType)
		argPos++
	}

	if filters.EntityType != nil {
		query += fmt.Sprintf(" AND entity_type = $%d", argPos)
		args = append(args, *filters.EntityType)
		argPos++
	}

	if filters.EntityID != nil {
		query += fmt.Sprintf(" AND entity_id = $%d", argPos)
		args = append(args, *filters.EntityID)
		argPos++
	}

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *filters.UserID)
		argPos++
	}

	if filters.IPAddress != nil {
		query += fmt.Sprintf(" AND ip_address = $%d", argPos)
		args = append(args, *filters.IPAddress)
		argPos++
	}

	if filters.OccurredAfter != nil {
		query += fmt.Sprintf(" AND occurred_at > $%d", argPos)
		args = append(args, *filters.OccurredAfter)
		argPos++
	}

	if filters.OccurredBefore != nil {
		query += fmt.Sprintf(" AND occurred_at < $%d", argPos)
		args = append(args, *filters.OccurredBefore)
		argPos++
	}

	query += " ORDER BY occurred_at DESC"

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
		return nil, fmt.Errorf("failed to list audit events: %w", err)
	}
	defer rows.Close()

	var events []*entities.AuditEvent
	for rows.Next() {
		event, err := scanAuditEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

// Count counts total audit events with filters
func (r *PostgresAuditRepository) Count(ctx context.Context, filters repositories.AuditFilters) (int64, error) {
	query := `SELECT COUNT(*) FROM audit.entry_events WHERE 1=1`

	args := []interface{}{}
	argPos := 1

	if filters.EventType != nil {
		query += fmt.Sprintf(" AND event_type = $%d", argPos)
		args = append(args, *filters.EventType)
		argPos++
	}

	if filters.EntityType != nil {
		query += fmt.Sprintf(" AND entity_type = $%d", argPos)
		args = append(args, *filters.EntityType)
		argPos++
	}

	if filters.EntityID != nil {
		query += fmt.Sprintf(" AND entity_id = $%d", argPos)
		args = append(args, *filters.EntityID)
		argPos++
	}

	if filters.UserID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argPos)
		args = append(args, *filters.UserID)
		argPos++
	}

	if filters.IPAddress != nil {
		query += fmt.Sprintf(" AND ip_address = $%d", argPos)
		args = append(args, *filters.IPAddress)
		argPos++
	}

	if filters.OccurredAfter != nil {
		query += fmt.Sprintf(" AND occurred_at > $%d", argPos)
		args = append(args, *filters.OccurredAfter)
		argPos++
	}

	if filters.OccurredBefore != nil {
		query += fmt.Sprintf(" AND occurred_at < $%d", argPos)
		args = append(args, *filters.OccurredBefore)
	}

	var count int64
	err := r.pool.QueryRow(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count audit events: %w", err)
	}

	return count, nil
}

// scanAuditEvent is a helper function to scan an audit event from database rows
func scanAuditEvent(rows pgxRows) (*entities.AuditEvent, error) {
	var event entities.AuditEvent
	var oldValuesJSON, newValuesJSON, metadataJSON []byte

	err := rows.Scan(
		&event.EventID,
		&event.EntityType,
		&event.EntityID,
		&event.EventType,
		&event.UserID,
		&oldValuesJSON,
		&newValuesJSON,
		&metadataJSON,
		&event.IPAddress,
		&event.UserAgent,
		&event.OccurredAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan audit event: %w", err)
	}

	// Unmarshal JSONB fields
	if len(oldValuesJSON) > 0 {
		if err := json.Unmarshal(oldValuesJSON, &event.OldValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal old values: %w", err)
		}
	}

	if len(newValuesJSON) > 0 {
		if err := json.Unmarshal(newValuesJSON, &event.NewValues); err != nil {
			return nil, fmt.Errorf("failed to unmarshal new values: %w", err)
		}
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &event.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	event.ID = event.EventID
	return &event, nil
}

// pgxRows is an interface for pgx.Rows to allow for easier testing
type pgxRows interface {
	Scan(dest ...interface{}) error
}
