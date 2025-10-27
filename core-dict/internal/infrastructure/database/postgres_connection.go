package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresConfig holds database configuration
type PostgresConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	Database        string
	Schema          string
	MaxConnections  int32
	MinConnections  int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
	HealthCheckPeriod time.Duration
	ConnectTimeout  time.Duration
	// Multi-tenant support
	CurrentISPB     string
}

// DefaultPostgresConfig returns default configuration
func DefaultPostgresConfig() *PostgresConfig {
	return &PostgresConfig{
		Host:              "localhost",
		Port:              5432,
		User:              "dict_app",
		Password:          "dict_password",
		Database:          "lbpay_core_dict",
		Schema:            "core_dict",
		MaxConnections:    20,
		MinConnections:    5,
		MaxConnLifetime:   time.Hour,
		MaxConnIdleTime:   time.Minute * 30,
		HealthCheckPeriod: time.Second * 30,
		ConnectTimeout:    time.Second * 10,
	}
}

// PostgresConnectionPool manages PostgreSQL connection pool
type PostgresConnectionPool struct {
	pool   *pgxpool.Pool
	config *PostgresConfig
}

// NewPostgresConnectionPool creates a new connection pool
func NewPostgresConnectionPool(ctx context.Context, config *PostgresConfig) (*PostgresConnectionPool, error) {
	if config == nil {
		config = DefaultPostgresConfig()
	}

	// Build connection string
	connString := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable search_path=%s",
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.Database,
		config.Schema,
	)

	// Parse config
	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connection string: %w", err)
	}

	// Configure pool
	poolConfig.MaxConns = config.MaxConnections
	poolConfig.MinConns = config.MinConnections
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod
	poolConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout

	// Set default ISPB if provided (for Row-Level Security)
	if config.CurrentISPB != "" {
		poolConfig.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
			_, err := conn.Exec(ctx, fmt.Sprintf("SET app.current_ispb = '%s'", config.CurrentISPB))
			return err
		}
	}

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Ping to verify connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &PostgresConnectionPool{
		pool:   pool,
		config: config,
	}, nil
}

// Pool returns the underlying pgxpool.Pool
func (p *PostgresConnectionPool) Pool() *pgxpool.Pool {
	return p.pool
}

// Close closes the connection pool
func (p *PostgresConnectionPool) Close() {
	p.pool.Close()
}

// Stats returns pool statistics
func (p *PostgresConnectionPool) Stats() *pgxpool.Stat {
	return p.pool.Stat()
}

// HealthCheck performs a health check
func (p *PostgresConnectionPool) HealthCheck(ctx context.Context) error {
	// Ping database
	if err := p.pool.Ping(ctx); err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}

	// Execute simple query
	var result int
	err := p.pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		return fmt.Errorf("query failed: %w", err)
	}

	if result != 1 {
		return fmt.Errorf("unexpected result: got %d, want 1", result)
	}

	return nil
}

// WithTransaction executes a function within a transaction
func (p *PostgresConnectionPool) WithTransaction(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			// Log rollback error (in production, use proper logger)
			fmt.Printf("failed to rollback transaction: %v\n", err)
		}
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// SetISPB sets the current ISPB for Row-Level Security
func (p *PostgresConnectionPool) SetISPB(ctx context.Context, ispb string) error {
	_, err := p.pool.Exec(ctx, fmt.Sprintf("SET app.current_ispb = '%s'", ispb))
	return err
}

// ResetISPB resets the ISPB session variable
func (p *PostgresConnectionPool) ResetISPB(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, "RESET app.current_ispb")
	return err
}
