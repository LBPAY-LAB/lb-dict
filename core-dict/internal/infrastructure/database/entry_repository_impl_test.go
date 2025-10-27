package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/infrastructure/database"
)

func setupTestDB(t *testing.T) (*pgxpool.Pool, func()) {
	ctx := context.Background()

	// Start PostgreSQL container
	req := testcontainers.ContainerRequest{
		Image:        "postgres:16",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "core_dict_test",
		},
		WaitingFor: wait.ForLog("database system is ready to accept connections").WithStartupTimeout(60 * time.Second),
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

	// Connect to DB
	connString := fmt.Sprintf("postgres://postgres:test@%s:%s/core_dict_test?sslmode=disable", host, port.Port())
	pool, err := pgxpool.New(ctx, connString)
	require.NoError(t, err)

	// Run schema creation
	createSchema(t, pool)

	cleanup := func() {
		pool.Close()
		container.Terminate(ctx)
	}

	return pool, cleanup
}

func createSchema(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	// Create schema
	_, err := pool.Exec(ctx, `CREATE SCHEMA IF NOT EXISTS core_dict`)
	require.NoError(t, err)

	// Create accounts table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS core_dict.accounts (
			id UUID PRIMARY KEY,
			participant_ispb VARCHAR(8) NOT NULL,
			branch_code VARCHAR(4) NOT NULL,
			account_number VARCHAR(20) NOT NULL,
			account_type VARCHAR(20) NOT NULL,
			account_status VARCHAR(20) NOT NULL,
			holder_name VARCHAR(255) NOT NULL,
			holder_document VARCHAR(14) NOT NULL,
			holder_document_type VARCHAR(4) NOT NULL,
			opened_at TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP,
			closed_at TIMESTAMP
		)
	`)
	require.NoError(t, err)

	// Create dict_entries table
	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS core_dict.dict_entries (
			id UUID PRIMARY KEY,
			key_type VARCHAR(10) NOT NULL,
			key_value VARCHAR(255) NOT NULL,
			key_hash VARCHAR(64) NOT NULL,
			status VARCHAR(20) NOT NULL,
			account_id UUID NOT NULL,
			participant_ispb VARCHAR(8) NOT NULL,
			participant_branch VARCHAR(4) NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP,
			CONSTRAINT fk_account FOREIGN KEY (account_id) REFERENCES core_dict.accounts(id)
		)
	`)
	require.NoError(t, err)

	// Create index on key_hash
	_, err = pool.Exec(ctx, `
		CREATE INDEX IF NOT EXISTS idx_entries_key_hash ON core_dict.dict_entries(key_hash)
	`)
	require.NoError(t, err)
}

func createTestAccount(t *testing.T, pool *pgxpool.Pool) uuid.UUID {
	ctx := context.Background()
	accountID := uuid.New()

	query := `
		INSERT INTO core_dict.accounts (
			id, participant_ispb, branch_code, account_number,
			account_type, account_status, holder_name,
			holder_document, holder_document_type, opened_at, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err := pool.Exec(ctx, query,
		accountID,
		"12345678",
		"0001",
		"123456",
		"CHECKING",
		"ACTIVE",
		"John Doe",
		"12345678901",
		"CPF",
		time.Now(),
		time.Now(),
		time.Now(),
	)
	require.NoError(t, err)

	return accountID
}

func TestEntryRepo_Create_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	accountID := createTestAccount(t, pool)
	repo := database.NewPostgresEntryRepository(pool)

	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		Status:        entities.KeyStatusActive,
		AccountID:     accountID,
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		AccountType:   "CHECKING",
		OwnerName:     "John Doe",
		OwnerTaxID:    "12345678901",
		OwnerType:     "NATURAL_PERSON",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(context.Background(), entry)
	assert.NoError(t, err)

	// Verify entry was created
	found, err := repo.FindByID(context.Background(), entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, entry.ID, found.ID)
	assert.Equal(t, entry.KeyValue, found.KeyValue)
}

func TestEntryRepo_FindByID_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	accountID := createTestAccount(t, pool)
	repo := database.NewPostgresEntryRepository(pool)

	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		Status:        entities.KeyStatusActive,
		AccountID:     accountID,
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(context.Background(), entry)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), entry.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, entry.ID, found.ID)
	assert.Equal(t, entry.KeyValue, found.KeyValue)
}

func TestEntryRepo_FindByID_NotFound(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := database.NewPostgresEntryRepository(pool)

	_, err := repo.FindByID(context.Background(), uuid.New())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entry not found")
}

func TestEntryRepo_FindByKey_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	accountID := createTestAccount(t, pool)
	repo := database.NewPostgresEntryRepository(pool)

	keyValue := "test@example.com"
	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeEmail,
		KeyValue:      keyValue,
		Status:        entities.KeyStatusActive,
		AccountID:     accountID,
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(context.Background(), entry)
	require.NoError(t, err)

	found, err := repo.FindByKey(context.Background(), keyValue)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, entry.KeyValue, found.KeyValue)
}

func TestEntryRepo_Update_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	accountID := createTestAccount(t, pool)
	repo := database.NewPostgresEntryRepository(pool)

	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		Status:        entities.KeyStatusPending,
		AccountID:     accountID,
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(context.Background(), entry)
	require.NoError(t, err)

	// Update status
	entry.Status = entities.KeyStatusActive
	entry.UpdatedAt = time.Now()

	err = repo.Update(context.Background(), entry)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(context.Background(), entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, entities.KeyStatusActive, found.Status)
}

func TestEntryRepo_Delete_SoftDelete(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	accountID := createTestAccount(t, pool)
	repo := database.NewPostgresEntryRepository(pool)

	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		Status:        entities.KeyStatusActive,
		AccountID:     accountID,
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(context.Background(), entry)
	require.NoError(t, err)

	// Soft delete
	err = repo.Delete(context.Background(), entry.ID)
	assert.NoError(t, err)

	// Entry should not be found (soft deleted)
	_, err = repo.FindByID(context.Background(), entry.ID)
	assert.Error(t, err)
}

func TestEntryRepo_List_Paginated(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	accountID := createTestAccount(t, pool)
	repo := database.NewPostgresEntryRepository(pool)

	// Create 5 entries
	for i := 0; i < 5; i++ {
		entry := &entities.Entry{
			ID:            uuid.New(),
			KeyType:       entities.KeyTypeEmail,
			KeyValue:      fmt.Sprintf("test%d@example.com", i),
			Status:        entities.KeyStatusActive,
			AccountID:     accountID,
			ISPB:          "12345678",
			Branch:        "0001",
			AccountNumber: "123456",
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		err := repo.Create(context.Background(), entry)
		require.NoError(t, err)
	}

	// List with pagination
	entries, err := repo.List(context.Background(), accountID, 3, 0)
	assert.NoError(t, err)
	assert.Len(t, entries, 3)

	// List page 2
	entries2, err := repo.List(context.Background(), accountID, 3, 3)
	assert.NoError(t, err)
	assert.Len(t, entries2, 2)
}

func TestEntryRepo_TransferOwnership(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	accountID := createTestAccount(t, pool)
	repo := database.NewPostgresEntryRepository(pool)

	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		Status:        entities.KeyStatusActive,
		AccountID:     accountID,
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err := repo.Create(context.Background(), entry)
	require.NoError(t, err)

	// Create new account for transfer
	newAccountID := createTestAccount(t, pool)

	// Update account ownership
	entry.AccountID = newAccountID
	entry.UpdatedAt = time.Now()

	err = repo.Update(context.Background(), entry)
	assert.NoError(t, err)

	// Verify transfer
	found, err := repo.FindByID(context.Background(), entry.ID)
	assert.NoError(t, err)
	assert.Equal(t, newAccountID, found.AccountID)
}
