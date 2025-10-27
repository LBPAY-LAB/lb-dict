package helpers

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// SetupTestDB creates a test PostgreSQL database using testcontainers
func SetupTestDB(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()

	// Create PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_DB":       "dict_test",
			"POSTGRES_USER":     "dict_test_user",
			"POSTGRES_PASSWORD": "dict_test_password",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get container host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "5432")
	require.NoError(t, err)

	// Create connection string
	connStr := fmt.Sprintf(
		"postgres://dict_test_user:dict_test_password@%s:%s/dict_test?sslmode=disable",
		host, port.Port(),
	)

	// Create connection pool
	pool, err := pgxpool.New(ctx, connStr)
	require.NoError(t, err)

	// Verify connection
	err = pool.Ping(ctx)
	require.NoError(t, err)

	// Cleanup
	t.Cleanup(func() {
		pool.Close()
		container.Terminate(ctx)
	})

	return pool
}

// SetupTestRedis creates a test Redis instance using testcontainers
func SetupTestRedis(t *testing.T) *redis.Client {
	ctx := context.Background()

	// Create Redis container
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err)

	// Get container host and port
	host, err := container.Host(ctx)
	require.NoError(t, err)

	port, err := container.MappedPort(ctx, "6379")
	require.NoError(t, err)

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	// Verify connection
	_, err = client.Ping(ctx).Result()
	require.NoError(t, err)

	// Cleanup
	t.Cleanup(func() {
		client.Close()
		container.Terminate(ctx)
	})

	return client
}

// CleanupTest performs cleanup after tests
func CleanupTest(t *testing.T) {
	t.Log("Test cleanup completed")
}

// LoadFixture loads a test fixture by name
func LoadFixture(t *testing.T, name string) interface{} {
	fixtures := map[string]interface{}{
		"valid_claim_portability": CreateValidPortabilityClaim(),
		"valid_claim_ownership":   CreateValidOwnershipClaim(),
		"valid_entry":             CreateValidEntry(),
	}

	fixture, ok := fixtures[name]
	require.True(t, ok, "Fixture %s not found", name)

	return fixture
}

// CreateValidPortabilityClaim creates a valid portability claim for testing
func CreateValidPortabilityClaim() *entities.Claim {
	claim, _ := entities.NewClaim(
		fmt.Sprintf("claim-test-%d", time.Now().UnixNano()),
		entities.ClaimTypePortability,
		"12345678901",
		"CPF",
		"60701190",
		"60746948",
	)
	claim.ClaimerAccountBranch = "0001"
	claim.ClaimerAccountNumber = "123456"
	claim.ClaimerAccountType = "CHECKING"
	return claim
}

// CreateValidOwnershipClaim creates a valid ownership claim for testing
func CreateValidOwnershipClaim() *entities.Claim {
	claim, _ := entities.NewClaim(
		fmt.Sprintf("claim-test-%d", time.Now().UnixNano()),
		entities.ClaimTypeOwnership,
		"test@example.com",
		"EMAIL",
		"60701190",
		"60746948",
	)
	claim.ClaimerAccountBranch = "0001"
	claim.ClaimerAccountNumber = "654321"
	claim.ClaimerAccountType = "SAVINGS"
	return claim
}

// CreateValidEntry creates a valid entry for testing
func CreateValidEntry() *entities.Entry {
	branch := "0001"
	accountNumber := "123456"
	ownerName := "Test User"
	ownerTaxID := "12345678901"

	return &entities.Entry{
		EntryID:       fmt.Sprintf("entry-%d", time.Now().UnixNano()),
		Key:           fmt.Sprintf("test-%d@example.com", time.Now().UnixNano()),
		KeyType:       entities.KeyTypeEMAIL,
		Participant:   "60701190",
		AccountBranch: &branch,
		AccountNumber: &accountNumber,
		AccountType:   entities.AccountTypeCACC,
		OwnerType:     entities.OwnerTypeNaturalPerson,
		OwnerName:     &ownerName,
		OwnerTaxID:    &ownerTaxID,
		Status:        entities.EntryStatusActive,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// WaitForCondition waits for a condition to be true or timeout
func WaitForCondition(t *testing.T, condition func() bool, timeout time.Duration) {
	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	require.Fail(t, "Timeout waiting for condition")
}

// CreateTestContext creates a context with a reasonable timeout for tests
func CreateTestContext(t *testing.T) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// CreateTestContextWithTimeout creates a context with a custom timeout
func CreateTestContextWithTimeout(t *testing.T, timeout time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	t.Cleanup(cancel)
	return ctx
}

// GenerateTestClaimID generates a unique claim ID for testing
func GenerateTestClaimID() string {
	return fmt.Sprintf("claim-test-%d", time.Now().UnixNano())
}

// GenerateTestKey generates a test key based on type
func GenerateTestKey(keyType string) string {
	switch keyType {
	case "CPF":
		return fmt.Sprintf("%011d", time.Now().Unix()%100000000000)
	case "EMAIL":
		return fmt.Sprintf("test_%d@example.com", time.Now().UnixNano())
	case "PHONE":
		return fmt.Sprintf("+5511%09d", time.Now().Unix()%1000000000)
	case "CNPJ":
		return fmt.Sprintf("%014d", time.Now().Unix()%100000000000000)
	default:
		return fmt.Sprintf("key-%d", time.Now().UnixNano())
	}
}

// SetupTestDatabase runs migrations and prepares the test database
func SetupTestDatabase(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Create tables (simplified for testing)
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS claims (
			claim_id VARCHAR(255) PRIMARY KEY,
			type VARCHAR(50) NOT NULL,
			key VARCHAR(255) NOT NULL,
			key_type VARCHAR(50) NOT NULL,
			donor_participant VARCHAR(8) NOT NULL,
			claimer_participant VARCHAR(8) NOT NULL,
			claimer_account_branch VARCHAR(4),
			claimer_account_number VARCHAR(20),
			claimer_account_type VARCHAR(20),
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			completed_at TIMESTAMP,
			cancelled_at TIMESTAMP,
			expired_at TIMESTAMP,
			cancellation_reason TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS entries (
			key VARCHAR(255) PRIMARY KEY,
			key_type VARCHAR(50) NOT NULL,
			participant VARCHAR(8) NOT NULL,
			branch VARCHAR(4),
			account_number VARCHAR(20),
			account_type VARCHAR(20),
			owner_type VARCHAR(50),
			owner_tax_id VARCHAR(14),
			owner_name VARCHAR(255),
			status VARCHAR(50) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
	}

	for _, migration := range migrations {
		_, err := pool.Exec(ctx, migration)
		require.NoError(t, err)
	}

	// Cleanup tables after test
	t.Cleanup(func() {
		pool.Exec(ctx, "DROP TABLE IF EXISTS claims CASCADE")
		pool.Exec(ctx, "DROP TABLE IF EXISTS entries CASCADE")
	})
}

// TruncateTables truncates all test tables
func TruncateTables(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()
	_, err := pool.Exec(ctx, "TRUNCATE claims, entries CASCADE")
	require.NoError(t, err)
}
