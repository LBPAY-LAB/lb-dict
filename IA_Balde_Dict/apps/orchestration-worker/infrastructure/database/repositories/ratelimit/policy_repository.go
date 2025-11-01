package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
	"github.com/lb-conn/connector-dict/domain/ratelimit"
)

const tracerName = "github.com/lb-conn/connector-dict/infrastructure/database/repositories/ratelimit"

var tracer = otel.Tracer(tracerName)

// policyRepository implements ports.PolicyRepository using PostgreSQL with pgx
type policyRepository struct {
	pool *pgxpool.Pool
}

// NewPolicyRepository creates a new policy repository
func NewPolicyRepository(pool *pgxpool.Pool) ports.PolicyRepository {
	return &policyRepository{
		pool: pool,
	}
}

// GetAll retrieves all rate limit policies
func (r *policyRepository) GetAll(ctx context.Context) ([]*ratelimit.Policy, error) {
	ctx, span := tracer.Start(ctx, "PolicyRepository.GetAll")
	defer span.End()

	query := `
		SELECT endpoint_id, endpoint_path, http_method, capacity, refill_tokens,
		       refill_period_sec, psp_category, created_at, updated_at
		FROM dict_rate_limit_policies
		ORDER BY endpoint_id
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query policies: %w", err)
	}
	defer rows.Close()

	policies := make([]*ratelimit.Policy, 0)

	for rows.Next() {
		policy, err := scanPolicy(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		policies = append(policies, policy)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("policy_count", len(policies)))
	span.SetStatus(codes.Ok, "Policies retrieved successfully")

	return policies, nil
}

// GetByID retrieves a specific policy by endpoint ID
func (r *policyRepository) GetByID(ctx context.Context, endpointID string) (*ratelimit.Policy, error) {
	ctx, span := tracer.Start(ctx, "PolicyRepository.GetByID")
	defer span.End()

	span.SetAttributes(attribute.String("endpoint_id", endpointID))

	query := `
		SELECT endpoint_id, endpoint_path, http_method, capacity, refill_tokens,
		       refill_period_sec, psp_category, created_at, updated_at
		FROM dict_rate_limit_policies
		WHERE endpoint_id = $1
	`

	row := r.pool.QueryRow(ctx, query, endpointID)

	policy, err := scanPolicy(row)
	if err != nil {
		if err == pgx.ErrNoRows {
			span.SetStatus(codes.Ok, "Policy not found")
			return nil, fmt.Errorf("policy not found: %s", endpointID)
		}
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query policy: %w", err)
	}

	span.SetStatus(codes.Ok, "Policy retrieved successfully")

	return policy, nil
}

// GetByCategory retrieves policies for a specific PSP category
func (r *policyRepository) GetByCategory(ctx context.Context, category string) ([]*ratelimit.Policy, error) {
	ctx, span := tracer.Start(ctx, "PolicyRepository.GetByCategory")
	defer span.End()

	span.SetAttributes(attribute.String("category", category))

	query := `
		SELECT endpoint_id, endpoint_path, http_method, capacity, refill_tokens,
		       refill_period_sec, psp_category, created_at, updated_at
		FROM dict_rate_limit_policies
		WHERE psp_category = $1 OR psp_category IS NULL
		ORDER BY endpoint_id
	`

	rows, err := r.pool.Query(ctx, query, category)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Query failed")
		return nil, fmt.Errorf("failed to query policies by category: %w", err)
	}
	defer rows.Close()

	policies := make([]*ratelimit.Policy, 0)

	for rows.Next() {
		policy, err := scanPolicy(rows)
		if err != nil {
			span.RecordError(err)
			return nil, fmt.Errorf("failed to scan policy: %w", err)
		}
		policies = append(policies, policy)
	}

	if err := rows.Err(); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Row iteration error")
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	span.SetAttributes(attribute.Int("policy_count", len(policies)))
	span.SetStatus(codes.Ok, "Policies retrieved successfully")

	return policies, nil
}

// Upsert inserts or updates a policy
func (r *policyRepository) Upsert(ctx context.Context, policy *ratelimit.Policy) error {
	ctx, span := tracer.Start(ctx, "PolicyRepository.Upsert")
	defer span.End()

	span.SetAttributes(attribute.String("endpoint_id", policy.EndpointID))

	query := `
		INSERT INTO dict_rate_limit_policies
		(endpoint_id, endpoint_path, http_method, capacity, refill_tokens,
		 refill_period_sec, psp_category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (endpoint_id)
		DO UPDATE SET
			endpoint_path = EXCLUDED.endpoint_path,
			http_method = EXCLUDED.http_method,
			capacity = EXCLUDED.capacity,
			refill_tokens = EXCLUDED.refill_tokens,
			refill_period_sec = EXCLUDED.refill_period_sec,
			psp_category = EXCLUDED.psp_category,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now().UTC()

	_, err := r.pool.Exec(ctx, query,
		policy.EndpointID,
		policy.EndpointPath,
		policy.HTTPMethod,
		policy.Capacity,
		policy.RefillTokens,
		policy.RefillPeriodSec,
		nullableString(policy.PSPCategory),
		policy.CreatedAt,
		now,
	)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Upsert failed")
		return fmt.Errorf("failed to upsert policy: %w", err)
	}

	span.SetStatus(codes.Ok, "Policy upserted successfully")

	return nil
}

// UpsertBatch inserts or updates multiple policies in a single transaction
func (r *policyRepository) UpsertBatch(ctx context.Context, policies []*ratelimit.Policy) error {
	ctx, span := tracer.Start(ctx, "PolicyRepository.UpsertBatch")
	defer span.End()

	span.SetAttributes(attribute.Int("policy_count", len(policies)))

	if len(policies) == 0 {
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
		INSERT INTO dict_rate_limit_policies
		(endpoint_id, endpoint_path, http_method, capacity, refill_tokens,
		 refill_period_sec, psp_category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (endpoint_id)
		DO UPDATE SET
			endpoint_path = EXCLUDED.endpoint_path,
			http_method = EXCLUDED.http_method,
			capacity = EXCLUDED.capacity,
			refill_tokens = EXCLUDED.refill_tokens,
			refill_period_sec = EXCLUDED.refill_period_sec,
			psp_category = EXCLUDED.psp_category,
			updated_at = EXCLUDED.updated_at
	`

	now := time.Now().UTC()

	for _, policy := range policies {
		_, err := tx.Exec(ctx, query,
			policy.EndpointID,
			policy.EndpointPath,
			policy.HTTPMethod,
			policy.Capacity,
			policy.RefillTokens,
			policy.RefillPeriodSec,
			nullableString(policy.PSPCategory),
			policy.CreatedAt,
			now,
		)

		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, "Batch upsert failed")
			return fmt.Errorf("failed to upsert policy %s: %w", policy.EndpointID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, "Failed to commit transaction")
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	span.SetStatus(codes.Ok, "Policies upserted successfully")

	return nil
}

// Helper functions

type scanner interface {
	Scan(dest ...interface{}) error
}

func scanPolicy(s scanner) (*ratelimit.Policy, error) {
	var (
		endpointID      string
		endpointPath    string
		httpMethod      string
		capacity        int
		refillTokens    int
		refillPeriodSec int
		pspCategory     *string
		createdAt       time.Time
		updatedAt       time.Time
	)

	err := s.Scan(
		&endpointID,
		&endpointPath,
		&httpMethod,
		&capacity,
		&refillTokens,
		&refillPeriodSec,
		&pspCategory,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	category := ""
	if pspCategory != nil {
		category = *pspCategory
	}

	policy := &ratelimit.Policy{
		EndpointID:      endpointID,
		EndpointPath:    endpointPath,
		HTTPMethod:      httpMethod,
		Capacity:        capacity,
		RefillTokens:    refillTokens,
		RefillPeriodSec: refillPeriodSec,
		PSPCategory:     category,
		CreatedAt:       createdAt.UTC(),
		UpdatedAt:       updatedAt.UTC(),
	}

	return policy, nil
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
