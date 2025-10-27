package testhelpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestEnvironment holds all dependencies for integration tests
type TestEnvironment struct {
	Ctx          context.Context
	T            *testing.T
	PG           *pgxpool.Pool
	Redis        *redis.Client
	PulsarMock   *PulsarMock
	ConnectMock  *ConnectMock
	PGContainer  testcontainers.Container
	RedisContainer testcontainers.Container
}

// SetupIntegrationTest initializes a complete test environment
func SetupIntegrationTest(t *testing.T) *TestEnvironment {
	ctx := context.Background()

	// Start PostgreSQL container
	pgContainer, pgPool := startPostgresContainer(ctx, t)

	// Start Redis container
	redisContainer, redisClient := startRedisContainer(ctx, t)

	// Start mocks
	pulsarMock := NewPulsarMock()
	connectMock := NewConnectMock(t)

	env := &TestEnvironment{
		Ctx:            ctx,
		T:              t,
		PG:             pgPool,
		Redis:          redisClient,
		PulsarMock:     pulsarMock,
		ConnectMock:    connectMock,
		PGContainer:    pgContainer,
		RedisContainer: redisContainer,
	}

	// Run migrations
	runMigrations(t, pgPool)

	// Cleanup
	t.Cleanup(func() {
		pgPool.Close()
		redisClient.Close()
		pulsarMock.Stop()
		connectMock.Stop()
		if pgContainer != nil {
			_ = pgContainer.Terminate(ctx)
		}
		if redisContainer != nil {
			_ = redisContainer.Terminate(ctx)
		}
	})

	return env
}

func startPostgresContainer(ctx context.Context, t *testing.T) (testcontainers.Container, *pgxpool.Pool) {
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "core_dict_test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").
			WithOccurrence(2).
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	dsn := fmt.Sprintf("postgres://test:test@%s:%s/core_dict_test?sslmode=disable", host, port.Port())

	var pool *pgxpool.Pool
	// Retry connection
	for i := 0; i < 10; i++ {
		pool, err = pgxpool.New(ctx, dsn)
		if err == nil {
			if err := pool.Ping(ctx); err == nil {
				break
			}
		}
		time.Sleep(1 * time.Second)
	}
	require.NoError(t, err)

	return container, pool
}

func startRedisContainer(ctx context.Context, t *testing.T) (testcontainers.Container, *redis.Client) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections").WithStartupTimeout(30 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	// Wait for Redis to be ready
	for i := 0; i < 10; i++ {
		if err := client.Ping(ctx).Err(); err == nil {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}

	return container, client
}

func runMigrations(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Create basic schema for testing
	schemas := []string{
		`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
		`CREATE TABLE IF NOT EXISTS entries (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			key_type VARCHAR(20) NOT NULL,
			key_value VARCHAR(77) NOT NULL,
			account_id UUID NOT NULL,
			ispb VARCHAR(8) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'ACTIVE',
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP,
			user_id VARCHAR(50) NOT NULL,
			UNIQUE(key_type, key_value)
		)`,
		`CREATE TABLE IF NOT EXISTS claims (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			entry_id UUID NOT NULL REFERENCES entries(id),
			claim_type VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'OPEN',
			donor_ispb VARCHAR(8) NOT NULL,
			claimer_ispb VARCHAR(8) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			confirmed_at TIMESTAMP,
			completed_at TIMESTAMP,
			cancelled_at TIMESTAMP,
			expires_at TIMESTAMP NOT NULL,
			user_id VARCHAR(50) NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS audit_events (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			entity_type VARCHAR(50) NOT NULL,
			entity_id UUID NOT NULL,
			action VARCHAR(50) NOT NULL,
			user_id VARCHAR(50) NOT NULL,
			metadata JSONB,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS accounts (
			id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
			ispb VARCHAR(8) NOT NULL,
			account_number VARCHAR(20) NOT NULL,
			account_type VARCHAR(20) NOT NULL,
			branch VARCHAR(10) NOT NULL,
			owner_name VARCHAR(200) NOT NULL,
			owner_document VARCHAR(14) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
	}

	for _, schema := range schemas {
		_, err := pool.Exec(ctx, schema)
		require.NoError(t, err)
	}
}

// CleanDatabase truncates all tables
func (env *TestEnvironment) CleanDatabase() {
	tables := []string{"audit_events", "claims", "entries", "accounts"}
	for _, table := range tables {
		_, err := env.PG.Exec(env.Ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		require.NoError(env.T, err)
	}
}

// CleanRedis flushes all Redis data
func (env *TestEnvironment) CleanRedis() {
	err := env.Redis.FlushAll(env.Ctx).Err()
	require.NoError(env.T, err)
}

// CleanAll cleans both database and cache
func (env *TestEnvironment) CleanAll() {
	env.CleanDatabase()
	env.CleanRedis()
	env.PulsarMock.Reset()
	env.ConnectMock.Reset()
}
