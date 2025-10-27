package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/infrastructure/database"
)

func TestAccountRepo_Create_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := database.NewPostgresAccountRepository(pool)

	account := &entities.Account{
		ID:            uuid.New(),
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		AccountType:   "CHECKING",
		Status:        "ACTIVE",
		Owner: entities.Owner{
			Name:   "John Doe",
			TaxID:  "12345678901",
			Type:   entities.OwnerTypeNaturalPerson,
		},
		OpenedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), account)
	assert.NoError(t, err)

	// Verify account was created
	found, err := repo.FindByID(context.Background(), account.ID)
	assert.NoError(t, err)
	assert.Equal(t, account.ID, found.ID)
	assert.Equal(t, account.ISPB, found.ISPB)
}

func TestAccountRepo_FindByID_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := database.NewPostgresAccountRepository(pool)

	account := &entities.Account{
		ID:            uuid.New(),
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		AccountType:   "CHECKING",
		Status:        "ACTIVE",
		Owner: entities.Owner{
			Name:   "Jane Doe",
			TaxID:  "98765432109",
			Type:   entities.OwnerTypeNaturalPerson,
		},
		OpenedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), account)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), account.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, account.Owner.Name, found.Owner.Name)
	assert.Equal(t, account.Owner.TaxID, found.Owner.TaxID)
}

func TestAccountRepo_UpdateStatus(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := database.NewPostgresAccountRepository(pool)

	account := &entities.Account{
		ID:            uuid.New(),
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		AccountType:   "CHECKING",
		Status:        "ACTIVE",
		Owner: entities.Owner{
			Name:   "Test User",
			TaxID:  "11111111111",
			Type:   entities.OwnerTypeNaturalPerson,
		},
		OpenedAt:  time.Now(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	err := repo.Create(context.Background(), account)
	require.NoError(t, err)

	// Update status
	account.Status = "BLOCKED"
	account.UpdatedAt = time.Now()

	err = repo.Update(context.Background(), account)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(context.Background(), account.ID)
	assert.NoError(t, err)
	assert.Equal(t, "BLOCKED", found.Status)
}

func TestAccountRepo_List_ByISPB(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	repo := database.NewPostgresAccountRepository(pool)

	ispb := "87654321"

	// Create 3 accounts for same ISPB
	for i := 0; i < 3; i++ {
		account := &entities.Account{
			ID:            uuid.New(),
			ISPB:          ispb,
			Branch:        "0001",
			AccountNumber: uuid.New().String(),
			AccountType:   "CHECKING",
			Status:        "ACTIVE",
			Owner: entities.Owner{
				Name:   "Test User",
				TaxID:  "11111111111",
				Type:   entities.OwnerTypeNaturalPerson,
			},
			OpenedAt:  time.Now(),
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err := repo.Create(context.Background(), account)
		require.NoError(t, err)
	}

	// List accounts by ISPB
	accounts, err := repo.FindByISPB(context.Background(), ispb, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, accounts, 3)

	for _, acc := range accounts {
		assert.Equal(t, ispb, acc.ISPB)
	}
}
