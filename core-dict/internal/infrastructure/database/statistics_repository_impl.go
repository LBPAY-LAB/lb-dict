package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

// PostgresStatisticsRepository implementa StatisticsRepository usando PostgreSQL
type PostgresStatisticsRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresStatisticsRepository cria um novo PostgresStatisticsRepository
func NewPostgresStatisticsRepository(pool *pgxpool.Pool) *PostgresStatisticsRepository {
	return &PostgresStatisticsRepository{
		pool: pool,
	}
}

// GetStatistics retorna estatísticas agregadas do sistema
func (r *PostgresStatisticsRepository) GetStatistics(ctx context.Context) (*entities.Statistics, error) {
	stats := &entities.Statistics{}

	// Query for entry statistics
	entryQuery := `
		SELECT
			COUNT(*) as total_keys,
			COUNT(CASE WHEN status = 'ACTIVE' THEN 1 END) as active_keys,
			COUNT(CASE WHEN status = 'BLOCKED' THEN 1 END) as blocked_keys,
			COUNT(CASE WHEN status = 'DELETED' THEN 1 END) as deleted_keys
		FROM entries
	`

	err := r.pool.QueryRow(ctx, entryQuery).Scan(
		&stats.TotalKeys,
		&stats.ActiveKeys,
		&stats.BlockedKeys,
		&stats.DeletedKeys,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get entry statistics: %w", err)
	}

	// Query for claim statistics
	claimQuery := `
		SELECT
			COUNT(*) as total_claims,
			COUNT(CASE WHEN status = 'OPEN' THEN 1 END) as pending_claims,
			COUNT(CASE WHEN status = 'COMPLETED' THEN 1 END) as completed_claims
		FROM claims
	`

	err = r.pool.QueryRow(ctx, claimQuery).Scan(
		&stats.TotalClaims,
		&stats.PendingClaims,
		&stats.CompletedClaims,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim statistics: %w", err)
	}

	return stats, nil
}
