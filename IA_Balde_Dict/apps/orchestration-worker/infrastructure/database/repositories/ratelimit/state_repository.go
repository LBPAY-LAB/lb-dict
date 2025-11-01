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

// stateRepository implements ports.StateRepository using PostgreSQL with partitioning support
type stateRepository struct {
	pool *pgxpool.Pool
}

// NewStateRepository creates a new state repository
func NewStateRepository(pool *pgxpool.Pool) ports.StateRepository {
	return &stateRepository{
		pool: pool,
	}
}

// Save inserts a new state snapshot
func (r *stateRepository) Save(ctx context.Context, state *ratelimit.PolicyState) error {
	ctx, span := tracer.Start(ctx, "StateRepository.Save")
	defer span.End()

	span.SetAttributes(
		attribute.String("endpoint_id", state.EndpointID),
		attribute.Int("available_tokens", state.AvailableTokens),
	)

	query := `
		INSERT INTO dict_rate_limit_states
		(endpoint_id, available_tokens, capacity, refill_tokens, refill_period_sec,
		 psp_category, consumption_rate_per_minute, recovery_eta_seconds,
		 exhaustion_projection_seconds, error_404_rate, response_timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	var id int64
	err := r.pool.QueryRow(ctx, query,
		state.EndpointID,
		state.AvailableTokens,
		state.Capacity,
		state.RefillTokens,
		state.RefillPeriodSec,
		nullableString(state.PSPCategory),
		state.ConsumptionRatePerMinute,
		state.RecoveryETASeconds,
		state.ExhaustionProjectionSeconds,
		state.Error404Rate,
		state.ResponseTimestamp.UTC(),
		state.CreatedAt.UTC(),
	).Scan(&id)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Insert failed")
		return fmt.Errorf("failed to save state: %w", err)
	}

	state.ID = id
	span.SetAttributes(attribute.Int64("state_id", id))
	span.SetStatus(codes.Ok, "State saved successfully")

	return nil
}

// SaveBatch inserts multiple state snapshots in a single transaction
func (r *stateRepository) SaveBatch(ctx context.Context, states []*ratelimit.PolicyState) error {
	ctx, span := tracer.Start(ctx, "StateRepository.SaveBatch")
	defer span.End()

	span.SetAttributes(attribute.Int("state_count", len(states)))

	if len(states) == 0 {
		return nil
	}

	// Use transaction for batch operation
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to begin transaction")
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO dict_rate_limit_states
		(endpoint_id, available_tokens, capacity, refill_tokens, refill_period_sec,
		 psp_category, consumption_rate_per_minute, recovery_eta_seconds,
		 exhaustion_projection_seconds, error_404_rate, response_timestamp, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id
	`

	for _, state := range states {
		var id int64
		err := tx.QueryRow(ctx, query,
			state.EndpointID,
			state.AvailableTokens,
			state.Capacity,
			state.RefillTokens,
			state.RefillPeriodSec,
			nullableString(state.PSPCategory),
			state.ConsumptionRatePerMinute,
			state.RecoveryETASeconds,
			state.ExhaustionProjectionSeconds,
			state.Error404Rate,
			state.ResponseTimestamp.UTC(),
			state.CreatedAt.UTC(),
		).Scan(&id)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Batch insert failed")
			return fmt.Errorf("failed to save state %s: %w", state.EndpointID, err)
		}

		state.ID = id
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	span.SetStatus(codes.Ok, "States saved successfully")

	return nil
}

// GetLatest retrieves the most recent state for an endpoint
func (r *stateRepository) GetLatest(ctx context.Context, endpointID string) (*ratelimit.PolicyState, error) {
	ctx, span := tracer.Start(ctx, "StateRepository.GetLatest")
	defer span.End()

	span.SetAttributes(attribute.String("endpoint_id", endpointID))

	query := `
		SELECT id, endpoint_id, available_tokens, capacity, refill_tokens,
		       refill_period_sec, psp_category, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, error_404_rate,
		       response_timestamp, created_at
		FROM dict_rate_limit_states
		WHERE endpoint_id = $1
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, endpointID)

	state, err := scanState(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			span.SetStatus(codes.Ok, "State not found")
			return nil, fmt.Errorf("no state found for endpoint: %s", endpointID)
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query latest state: %w", err)
	}

	span.SetStatus(codes.Ok, "Latest state retrieved successfully")

	return state, nil
}

// GetLatestAll retrieves the most recent state for all endpoints
func (r *stateRepository) GetLatestAll(ctx context.Context) ([]*ratelimit.PolicyState, error) {
	ctx, span := tracer.Start(ctx, "StateRepository.GetLatestAll")
	defer span.End()

	// Use DISTINCT ON to get latest state per endpoint (PostgreSQL-specific, efficient)
	query := `
		SELECT DISTINCT ON (endpoint_id)
		       id, endpoint_id, available_tokens, capacity, refill_tokens,
		       refill_period_sec, psp_category, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, error_404_rate,
		       response_timestamp, created_at
		FROM dict_rate_limit_states
		ORDER BY endpoint_id, created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query latest states: %w", err)
	}
	defer rows.Close()

	states := make([]*ratelimit.PolicyState, 0)

	for rows.Next() {
		state, err := scanState(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan state: %w", err)
		}
		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("state_count", len(states)))
	span.SetStatus(codes.Ok, "Latest states retrieved successfully")

	return states, nil
}

// GetHistory retrieves historical states for an endpoint in a time range
func (r *stateRepository) GetHistory(ctx context.Context, endpointID string, since, until time.Time) ([]*ratelimit.PolicyState, error) {
	ctx, span := tracer.Start(ctx, "StateRepository.GetHistory")
	defer span.End()

	span.SetAttributes(
		attribute.String("endpoint_id", endpointID),
		attribute.String("since", since.Format(time.RFC3339)),
		attribute.String("until", until.Format(time.RFC3339)),
	)

	query := `
		SELECT id, endpoint_id, available_tokens, capacity, refill_tokens,
		       refill_period_sec, psp_category, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, error_404_rate,
		       response_timestamp, created_at
		FROM dict_rate_limit_states
		WHERE endpoint_id = $1
		  AND created_at >= $2
		  AND created_at <= $3
		ORDER BY created_at ASC
	`

	rows, err := r.pool.Query(ctx, query, endpointID, since.UTC(), until.UTC())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query history: %w", err)
	}
	defer rows.Close()

	states := make([]*ratelimit.PolicyState, 0)

	for rows.Next() {
		state, err := scanState(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan state: %w", err)
		}
		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("state_count", len(states)))
	span.SetStatus(codes.Ok, "History retrieved successfully")

	return states, nil
}

// GetByCategory retrieves latest states for endpoints of a specific category
func (r *stateRepository) GetByCategory(ctx context.Context, category string, limit int) ([]*ratelimit.PolicyState, error) {
	ctx, span := tracer.Start(ctx, "StateRepository.GetByCategory")
	defer span.End()

	span.SetAttributes(
		attribute.String("category", category),
		attribute.Int("limit", limit),
	)

	query := `
		SELECT DISTINCT ON (endpoint_id)
		       id, endpoint_id, available_tokens, capacity, refill_tokens,
		       refill_period_sec, psp_category, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, error_404_rate,
		       response_timestamp, created_at
		FROM dict_rate_limit_states
		WHERE psp_category = $1
		ORDER BY endpoint_id, created_at DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, category, limit)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query by category: %w", err)
	}
	defer rows.Close()

	states := make([]*ratelimit.PolicyState, 0)

	for rows.Next() {
		state, err := scanState(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan state: %w", err)
		}
		states = append(states, state)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("state_count", len(states)))
	span.SetStatus(codes.Ok, "States by category retrieved successfully")

	return states, nil
}

// GetPreviousState retrieves the state immediately before the given timestamp
// Used for calculating consumption rates
func (r *stateRepository) GetPreviousState(ctx context.Context, endpointID string, before time.Time) (*ratelimit.PolicyState, error) {
	ctx, span := tracer.Start(ctx, "StateRepository.GetPreviousState")
	defer span.End()

	span.SetAttributes(
		attribute.String("endpoint_id", endpointID),
		attribute.String("before", before.Format(time.RFC3339)),
	)

	query := `
		SELECT id, endpoint_id, available_tokens, capacity, refill_tokens,
		       refill_period_sec, psp_category, consumption_rate_per_minute,
		       recovery_eta_seconds, exhaustion_projection_seconds, error_404_rate,
		       response_timestamp, created_at
		FROM dict_rate_limit_states
		WHERE endpoint_id = $1
		  AND created_at < $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	row := r.pool.QueryRow(ctx, query, endpointID, before.UTC())

	state, err := scanState(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			span.SetStatus(codes.Ok, "No previous state found")
			return nil, nil // Not an error - just no previous state exists
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query previous state: %w", err)
	}

	span.SetStatus(codes.Ok, "Previous state retrieved successfully")

	return state, nil
}

// DeleteOlderThan deletes states older than the specified timestamp
// Returns the number of records deleted (for 13-month retention policy)
func (r *stateRepository) DeleteOlderThan(ctx context.Context, timestamp time.Time) (int64, error) {
	ctx, span := tracer.Start(ctx, "StateRepository.DeleteOlderThan")
	defer span.End()

	span.SetAttributes(attribute.String("timestamp", timestamp.Format(time.RFC3339)))

	// Note: This deletes from partitioned table
	// For better performance, use database function drop_old_partitions() instead
	query := `
		DELETE FROM dict_rate_limit_states
		WHERE created_at < $1
	`

	result, err := r.pool.Exec(ctx, query, timestamp.UTC())
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Delete failed")
		return 0, fmt.Errorf("failed to delete old states: %w", err)
	}

	rowsAffected := result.RowsAffected()

	span.SetAttributes(attribute.Int64("rows_deleted", rowsAffected))
	span.SetStatus(codes.Ok, "Old states deleted successfully")

	return rowsAffected, nil
}

// Helper function to scan PolicyState from row
func scanState(s scanner) (*ratelimit.PolicyState, error) {
	var (
		id                          int64
		endpointID                  string
		availableTokens             int
		capacity                    int
		refillTokens                int
		refillPeriodSec             int
		pspCategory                 *string
		consumptionRatePerMinute    float64
		recoveryETASeconds          int
		exhaustionProjectionSeconds int
		error404Rate                float64
		responseTimestamp           time.Time
		createdAt                   time.Time
	)

	err := s.Scan(
		&id,
		&endpointID,
		&availableTokens,
		&capacity,
		&refillTokens,
		&refillPeriodSec,
		&pspCategory,
		&consumptionRatePerMinute,
		&recoveryETASeconds,
		&exhaustionProjectionSeconds,
		&error404Rate,
		&responseTimestamp,
		&createdAt,
	)

	if err != nil {
		return nil, err
	}

	category := ""
	if pspCategory != nil {
		category = *pspCategory
	}

	state := &ratelimit.PolicyState{
		ID:                          id,
		EndpointID:                  endpointID,
		AvailableTokens:             availableTokens,
		Capacity:                    capacity,
		RefillTokens:                refillTokens,
		RefillPeriodSec:             refillPeriodSec,
		PSPCategory:                 category,
		ConsumptionRatePerMinute:    consumptionRatePerMinute,
		RecoveryETASeconds:          recoveryETASeconds,
		ExhaustionProjectionSeconds: exhaustionProjectionSeconds,
		Error404Rate:                error404Rate,
		ResponseTimestamp:           responseTimestamp.UTC(),
		CreatedAt:                   createdAt.UTC(),
	}

	return state, nil
}
