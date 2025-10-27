package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// PostgresClient wraps pgx connection pool with application-specific methods
type PostgresClient struct {
	pool   *pgxpool.Pool
	logger *logrus.Logger
	config *PostgresConfig
}

// PostgresConfig holds PostgreSQL connection configuration
type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	SSLMode         string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	HealthCheckPeriod time.Duration
}

// NewPostgresClient creates a new PostgreSQL client with connection pooling
func NewPostgresClient(config *PostgresConfig, logger *logrus.Logger) (*PostgresClient, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}
	if logger == nil {
		return nil, fmt.Errorf("logger cannot be nil")
	}

	// Build connection string
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.SSLMode,
	)

	// Parse connection config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool
	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

	// Create connection pool
	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"host":     config.Host,
		"port":     config.Port,
		"database": config.Database,
		"max_conns": config.MaxConns,
	}).Info("PostgreSQL connection pool created successfully")

	return &PostgresClient{
		pool:   pool,
		logger: logger,
		config: config,
	}, nil
}

// GetPool returns the underlying connection pool
func (c *PostgresClient) GetPool() *pgxpool.Pool {
	return c.pool
}

// Close closes the connection pool
func (c *PostgresClient) Close() {
	if c.pool != nil {
		c.pool.Close()
		c.logger.Info("PostgreSQL connection pool closed")
	}
}

// Ping checks if the database is reachable
func (c *PostgresClient) Ping(ctx context.Context) error {
	return c.pool.Ping(ctx)
}

// HealthCheck performs a comprehensive health check
func (c *PostgresClient) HealthCheck(ctx context.Context) error {
	// Check ping
	if err := c.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Check if we can acquire a connection
	conn, err := c.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Execute a simple query
	var result int
	if err := conn.QueryRow(ctx, "SELECT 1").Scan(&result); err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected query result: %d", result)
	}

	return nil
}

// GetStats returns connection pool statistics
func (c *PostgresClient) GetStats() *pgxpool.Stat {
	return c.pool.Stat()
}

// BeginTx starts a new transaction
func (c *PostgresClient) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return c.pool.Begin(ctx)
}

// BeginTxWithOptions starts a new transaction with custom options
func (c *PostgresClient) BeginTxWithOptions(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return c.pool.BeginTx(ctx, opts)
}

// ExecuteInTransaction executes a function within a transaction
// Automatically commits if fn returns nil, rolls back on error
func (c *PostgresClient) ExecuteInTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := c.BeginTx(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			c.logger.Errorf("Failed to rollback transaction: %v", rbErr)
		}
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Query executes a query that returns rows
func (c *PostgresClient) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return c.pool.Query(ctx, sql, args...)
}

// QueryRow executes a query that returns at most one row
func (c *PostgresClient) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return c.pool.QueryRow(ctx, sql, args...)
}

// Exec executes a query without returning any rows
func (c *PostgresClient) Exec(ctx context.Context, sql string, args ...interface{}) (pgconn.CommandTag, error) {
	return c.pool.Exec(ctx, sql, args...)
}

// DefaultPostgresConfig returns default PostgreSQL configuration
func DefaultPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:              "localhost",
		Port:              5432,
		User:              "conn_dict_user",
		Password:          "conn_dict_password_dev",
		Database:          "conn_dict",
		SSLMode:           "disable",
		MaxConns:          25,
		MinConns:          5,
		MaxConnLifetime:   5 * time.Minute,
		MaxConnIdleTime:   1 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}
}

// ConfigFromEnv creates PostgresConfig from environment variables
func ConfigFromEnv() *PostgresConfig {
	// This will be populated from .env
	// For now, returning defaults
	return DefaultPostgresConfig()
}