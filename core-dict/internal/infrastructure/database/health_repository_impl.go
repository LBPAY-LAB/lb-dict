package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// PostgresHealthRepository implementa HealthRepository usando PostgreSQL
type PostgresHealthRepository struct {
	pool        *pgxpool.Pool
	redisClient *redis.Client
}

// NewPostgresHealthRepository cria um novo PostgresHealthRepository
func NewPostgresHealthRepository(pool *pgxpool.Pool, redisClient *redis.Client) *PostgresHealthRepository {
	return &PostgresHealthRepository{
		pool:        pool,
		redisClient: redisClient,
	}
}

// CheckDatabase verifica conectividade com PostgreSQL
func (r *PostgresHealthRepository) CheckDatabase(ctx context.Context) error {
	// Simple ping query
	var result int
	err := r.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}
	if result != 1 {
		return fmt.Errorf("database health check returned unexpected result: %d", result)
	}
	return nil
}

// CheckRedis verifica conectividade com Redis
func (r *PostgresHealthRepository) CheckRedis(ctx context.Context) error {
	if r.redisClient == nil {
		return fmt.Errorf("redis client not configured")
	}
	if err := r.redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}
	return nil
}

// CheckPulsar verifica conectividade com Pulsar
func (r *PostgresHealthRepository) CheckPulsar(ctx context.Context) error {
	// TODO: Implement when Pulsar client is available
	// For now, return nil (no error) as Pulsar is optional
	return nil
}
