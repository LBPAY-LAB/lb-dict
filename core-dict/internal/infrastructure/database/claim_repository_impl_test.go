package database_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
	"github.com/lbpay-lab/core-dict/internal/infrastructure/database"
)

func createClaimsTable(t *testing.T, pool *pgxpool.Pool) {
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS core_dict.claims (
			id UUID PRIMARY KEY,
			entry_key VARCHAR(255) NOT NULL,
			claim_type VARCHAR(50) NOT NULL,
			status VARCHAR(50) NOT NULL,
			claimer_ispb VARCHAR(8) NOT NULL,
			owner_ispb VARCHAR(8) NOT NULL,
			claimer_account_id UUID NOT NULL,
			owner_account_id UUID NOT NULL,
			bacen_claim_id VARCHAR(100),
			workflow_id VARCHAR(100),
			completion_period_days INTEGER NOT NULL,
			expires_at TIMESTAMP NOT NULL,
			resolution_type VARCHAR(50),
			resolution_reason VARCHAR(255),
			resolution_date TIMESTAMP,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
			deleted_at TIMESTAMP
		)
	`)
	require.NoError(t, err)
}

func TestClaimRepo_Create_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createClaimsTable(t, pool)

	accountID1 := createTestAccount(t, pool)
	accountID2 := createTestAccount(t, pool)

	repo := database.NewPostgresClaimRepository(pool)

	claim := &entities.Claim{
		ID:                   uuid.New(),
		EntryKey:             "test@example.com",
		ClaimType:            valueobjects.ClaimTypeOwnership,
		Status:               valueobjects.ClaimStatusOpen,
		ClaimerParticipant:   valueobjects.Participant{ISPB: "12345678", Name: "Claimer Bank"},
		DonorParticipant:     valueobjects.Participant{ISPB: "87654321", Name: "Donor Bank"},
		ClaimerAccountID:     accountID1,
		DonorAccountID:       accountID2,
		BacenClaimID:         "BCN123456",
		WorkflowID:           "WF-" + uuid.New().String(),
		CompletionPeriodDays: 30,
		ExpiresAt:            time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := repo.Create(context.Background(), claim)
	assert.NoError(t, err)

	// Verify claim was created
	found, err := repo.FindByID(context.Background(), claim.ID)
	assert.NoError(t, err)
	assert.Equal(t, claim.ID, found.ID)
	assert.Equal(t, claim.EntryKey, found.EntryKey)
}

func TestClaimRepo_FindByID_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createClaimsTable(t, pool)

	accountID1 := createTestAccount(t, pool)
	accountID2 := createTestAccount(t, pool)

	repo := database.NewPostgresClaimRepository(pool)

	claim := &entities.Claim{
		ID:                   uuid.New(),
		EntryKey:             "test@example.com",
		ClaimType:            valueobjects.ClaimTypeOwnership,
		Status:               valueobjects.ClaimStatusOpen,
		ClaimerParticipant:   valueobjects.Participant{ISPB: "12345678"},
		DonorParticipant:     valueobjects.Participant{ISPB: "87654321"},
		ClaimerAccountID:     accountID1,
		DonorAccountID:       accountID2,
		CompletionPeriodDays: 30,
		ExpiresAt:            time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := repo.Create(context.Background(), claim)
	require.NoError(t, err)

	found, err := repo.FindByID(context.Background(), claim.ID)
	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, claim.EntryKey, found.EntryKey)
}

func TestClaimRepo_Update_Success(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createClaimsTable(t, pool)

	accountID1 := createTestAccount(t, pool)
	accountID2 := createTestAccount(t, pool)

	repo := database.NewPostgresClaimRepository(pool)

	claim := &entities.Claim{
		ID:                   uuid.New(),
		EntryKey:             "test@example.com",
		ClaimType:            valueobjects.ClaimTypeOwnership,
		Status:               valueobjects.ClaimStatusOpen,
		ClaimerParticipant:   valueobjects.Participant{ISPB: "12345678"},
		DonorParticipant:     valueobjects.Participant{ISPB: "87654321"},
		ClaimerAccountID:     accountID1,
		DonorAccountID:       accountID2,
		CompletionPeriodDays: 30,
		ExpiresAt:            time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := repo.Create(context.Background(), claim)
	require.NoError(t, err)

	// Update status
	claim.Status = valueobjects.ClaimStatusCompleted
	claim.ResolutionType = "CONFIRMED"
	claim.ResolutionReason = "Approved by donor"
	now := time.Now()
	claim.ResolutionDate = &now
	claim.UpdatedAt = now

	err = repo.Update(context.Background(), claim)
	assert.NoError(t, err)

	// Verify update
	found, err := repo.FindByID(context.Background(), claim.ID)
	assert.NoError(t, err)
	assert.Equal(t, valueobjects.ClaimStatusCompleted, found.Status)
	assert.Equal(t, "CONFIRMED", found.ResolutionType)
}

func TestClaimRepo_FindExpired_30Days(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createClaimsTable(t, pool)

	accountID1 := createTestAccount(t, pool)
	accountID2 := createTestAccount(t, pool)

	repo := database.NewPostgresClaimRepository(pool)

	// Create expired claim
	expiredClaim := &entities.Claim{
		ID:                   uuid.New(),
		EntryKey:             "expired@example.com",
		ClaimType:            valueobjects.ClaimTypeOwnership,
		Status:               valueobjects.ClaimStatusOpen,
		ClaimerParticipant:   valueobjects.Participant{ISPB: "12345678"},
		DonorParticipant:     valueobjects.Participant{ISPB: "87654321"},
		ClaimerAccountID:     accountID1,
		DonorAccountID:       accountID2,
		CompletionPeriodDays: 30,
		ExpiresAt:            time.Now().Add(-1 * time.Hour), // Expired 1 hour ago
		CreatedAt:            time.Now().Add(-31 * 24 * time.Hour),
		UpdatedAt:            time.Now(),
	}

	err := repo.Create(context.Background(), expiredClaim)
	require.NoError(t, err)

	// Create non-expired claim
	validClaim := &entities.Claim{
		ID:                   uuid.New(),
		EntryKey:             "valid@example.com",
		ClaimType:            valueobjects.ClaimTypeOwnership,
		Status:               valueobjects.ClaimStatusOpen,
		ClaimerParticipant:   valueobjects.Participant{ISPB: "12345678"},
		DonorParticipant:     valueobjects.Participant{ISPB: "87654321"},
		ClaimerAccountID:     accountID1,
		DonorAccountID:       accountID2,
		CompletionPeriodDays: 30,
		ExpiresAt:            time.Now().Add(15 * 24 * time.Hour),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err = repo.Create(context.Background(), validClaim)
	require.NoError(t, err)

	// Find expired claims
	expired, err := repo.FindExpired(context.Background(), 10)
	assert.NoError(t, err)
	assert.Len(t, expired, 1)
	assert.Equal(t, expiredClaim.ID, expired[0].ID)
}

func TestClaimRepo_ExistsActiveClaim_True(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createClaimsTable(t, pool)

	accountID1 := createTestAccount(t, pool)
	accountID2 := createTestAccount(t, pool)

	repo := database.NewPostgresClaimRepository(pool)

	entryKey := "active@example.com"

	claim := &entities.Claim{
		ID:                   uuid.New(),
		EntryKey:             entryKey,
		ClaimType:            valueobjects.ClaimTypeOwnership,
		Status:               valueobjects.ClaimStatusOpen,
		ClaimerParticipant:   valueobjects.Participant{ISPB: "12345678"},
		DonorParticipant:     valueobjects.Participant{ISPB: "87654321"},
		ClaimerAccountID:     accountID1,
		DonorAccountID:       accountID2,
		CompletionPeriodDays: 30,
		ExpiresAt:            time.Now().Add(30 * 24 * time.Hour),
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	err := repo.Create(context.Background(), claim)
	require.NoError(t, err)

	exists, err := repo.ExistsActiveClaim(context.Background(), entryKey)
	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestClaimRepo_ExistsActiveClaim_False(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createClaimsTable(t, pool)

	repo := database.NewPostgresClaimRepository(pool)

	exists, err := repo.ExistsActiveClaim(context.Background(), "nonexistent@example.com")
	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestClaimRepo_List_FilterByStatus(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createClaimsTable(t, pool)

	accountID1 := createTestAccount(t, pool)
	accountID2 := createTestAccount(t, pool)

	repo := database.NewPostgresClaimRepository(pool)

	// Create claims with different statuses
	statuses := []valueobjects.ClaimStatus{
		valueobjects.ClaimStatusOpen,
		valueobjects.ClaimStatusOpen,
		valueobjects.ClaimStatusCompleted,
	}

	for _, status := range statuses {
		claim := &entities.Claim{
			ID:                   uuid.New(),
			EntryKey:             "test@example.com",
			ClaimType:            valueobjects.ClaimTypeOwnership,
			Status:               status,
			ClaimerParticipant:   valueobjects.Participant{ISPB: "12345678"},
			DonorParticipant:     valueobjects.Participant{ISPB: "87654321"},
			ClaimerAccountID:     accountID1,
			DonorAccountID:       accountID2,
			CompletionPeriodDays: 30,
			ExpiresAt:            time.Now().Add(30 * 24 * time.Hour),
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}
		err := repo.Create(context.Background(), claim)
		require.NoError(t, err)
	}

	// Filter by OPEN status
	openClaims, err := repo.FindByStatus(context.Background(), valueobjects.ClaimStatusOpen, 10, 0)
	assert.NoError(t, err)
	assert.Len(t, openClaims, 2)
}

func TestClaimRepo_List_Pagination(t *testing.T) {
	pool, cleanup := setupTestDB(t)
	defer cleanup()

	createClaimsTable(t, pool)

	accountID1 := createTestAccount(t, pool)
	accountID2 := createTestAccount(t, pool)

	repo := database.NewPostgresClaimRepository(pool)

	// Create 5 claims
	for i := 0; i < 5; i++ {
		claim := &entities.Claim{
			ID:                   uuid.New(),
			EntryKey:             "test@example.com",
			ClaimType:            valueobjects.ClaimTypeOwnership,
			Status:               valueobjects.ClaimStatusOpen,
			ClaimerParticipant:   valueobjects.Participant{ISPB: "12345678"},
			DonorParticipant:     valueobjects.Participant{ISPB: "87654321"},
			ClaimerAccountID:     accountID1,
			DonorAccountID:       accountID2,
			CompletionPeriodDays: 30,
			ExpiresAt:            time.Now().Add(30 * 24 * time.Hour),
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		}
		err := repo.Create(context.Background(), claim)
		require.NoError(t, err)
	}

	// List with pagination
	claims, err := repo.FindByStatus(context.Background(), valueobjects.ClaimStatusOpen, 3, 0)
	assert.NoError(t, err)
	assert.Len(t, claims, 3)

	// Page 2
	claims2, err := repo.FindByStatus(context.Background(), valueobjects.ClaimStatusOpen, 3, 3)
	assert.NoError(t, err)
	assert.Len(t, claims2, 2)
}
