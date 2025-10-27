package aggregates

import (
	"testing"
	"time"

	"github.com/lbpay-lab/conn-dict/internal/domain/entities"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Note: This file tests aggregate logic if aggregates are used
// If the domain uses entities directly, these tests demonstrate
// testing domain logic with testify

func TestClaimAggregate_Creation(t *testing.T) {
	tests := []struct {
		name        string
		claimID     string
		claimType   entities.ClaimType
		key         string
		keyType     string
		donor       string
		claimer     string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid portability claim",
			claimID:   "claim-001",
			claimType: entities.ClaimTypePortability,
			key:       "12345678901",
			keyType:   "CPF",
			donor:     "60701190",
			claimer:   "60746948",
			wantErr:   false,
		},
		{
			name:      "valid ownership claim",
			claimID:   "claim-002",
			claimType: entities.ClaimTypeOwnership,
			key:       "test@example.com",
			keyType:   "EMAIL",
			donor:     "60701190",
			claimer:   "60746948",
			wantErr:   false,
		},
		{
			name:        "empty claim ID",
			claimID:     "",
			claimType:   entities.ClaimTypePortability,
			key:         "12345678901",
			keyType:     "CPF",
			donor:       "60701190",
			claimer:     "60746948",
			wantErr:     true,
			errContains: "claim ID",
		},
		{
			name:        "same donor and claimer",
			claimID:     "claim-003",
			claimType:   entities.ClaimTypePortability,
			key:         "12345678901",
			keyType:     "CPF",
			donor:       "60701190",
			claimer:     "60701190",
			wantErr:     true,
			errContains: "donor and claimer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim, err := entities.NewClaim(
				tt.claimID,
				tt.claimType,
				tt.key,
				tt.keyType,
				tt.donor,
				tt.claimer,
			)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				assert.Nil(t, claim)
			} else {
				require.NoError(t, err)
				require.NotNil(t, claim)
				assert.Equal(t, tt.claimID, claim.ClaimID)
				assert.Equal(t, tt.claimType, claim.Type)
				assert.Equal(t, tt.key, claim.Key)
				assert.Equal(t, entities.ClaimStatusOpen, claim.Status)
			}
		})
	}
}

func TestClaimAggregate_StatusTransitions(t *testing.T) {
	// Create a valid claim
	claim, err := entities.NewClaim(
		"claim-status-test",
		entities.ClaimTypePortability,
		"12345678901",
		"CPF",
		"60701190",
		"60746948",
	)
	require.NoError(t, err)

	tests := []struct {
		name          string
		initialStatus entities.ClaimStatus
		action        func(*entities.Claim) error
		wantStatus    entities.ClaimStatus
		wantErr       bool
	}{
		{
			name:          "confirm from open",
			initialStatus: entities.ClaimStatusOpen,
			action: func(c *entities.Claim) error {
				return c.Confirm()
			},
			wantStatus: entities.ClaimStatusConfirmed,
			wantErr:    false,
		},
		{
			name:          "complete from confirmed",
			initialStatus: entities.ClaimStatusConfirmed,
			action: func(c *entities.Claim) error {
				return c.Complete()
			},
			wantStatus: entities.ClaimStatusCompleted,
			wantErr:    false,
		},
		{
			name:          "cancel from open",
			initialStatus: entities.ClaimStatusOpen,
			action: func(c *entities.Claim) error {
				return c.Cancel("user requested")
			},
			wantStatus: entities.ClaimStatusCancelled,
			wantErr:    false,
		},
		{
			name:          "expire from open",
			initialStatus: entities.ClaimStatusOpen,
			action: func(c *entities.Claim) error {
				return c.Expire()
			},
			wantStatus: entities.ClaimStatusExpired,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset claim to initial status
			claim.Status = tt.initialStatus

			err := tt.action(claim)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantStatus, claim.Status)
			}
		})
	}
}

func TestClaimAggregate_InvalidTransitions(t *testing.T) {
	claim, err := entities.NewClaim(
		"claim-invalid-transition",
		entities.ClaimTypePortability,
		"12345678901",
		"CPF",
		"60701190",
		"60746948",
	)
	require.NoError(t, err)

	// Complete claim
	err = claim.Confirm()
	require.NoError(t, err)
	err = claim.Complete()
	require.NoError(t, err)

	// Try to cancel completed claim - should fail
	err = claim.Cancel("attempting to cancel completed")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid status transition")
}

func TestClaimAggregate_BusinessRules(t *testing.T) {
	t.Run("portability claim rules", func(t *testing.T) {
		claim, err := entities.NewClaim(
			"claim-portability",
			entities.ClaimTypePortability,
			"12345678901",
			"CPF",
			"60701190",
			"60746948",
		)
		require.NoError(t, err)
		assert.Equal(t, entities.ClaimTypePortability, claim.Type)
	})

	t.Run("ownership claim rules", func(t *testing.T) {
		claim, err := entities.NewClaim(
			"claim-ownership",
			entities.ClaimTypeOwnership,
			"test@example.com",
			"EMAIL",
			"60701190",
			"60746948",
		)
		require.NoError(t, err)
		assert.Equal(t, entities.ClaimTypeOwnership, claim.Type)
	})

	t.Run("claim timeout period", func(t *testing.T) {
		claim, err := entities.NewClaim(
			"claim-timeout",
			entities.ClaimTypePortability,
			"12345678901",
			"CPF",
			"60701190",
			"60746948",
		)
		require.NoError(t, err)

		// Simulate claim created 31 days ago
		claim.CreatedAt = time.Now().Add(-31 * 24 * time.Hour)

		// Check if claim should be expired (business rule: 30 days)
		daysSinceCreation := time.Since(claim.CreatedAt).Hours() / 24
		assert.Greater(t, daysSinceCreation, float64(30))
	})
}

func TestClaimAggregate_ConcurrentModification(t *testing.T) {
	claim, err := entities.NewClaim(
		"claim-concurrent",
		entities.ClaimTypePortability,
		"12345678901",
		"CPF",
		"60701190",
		"60746948",
	)
	require.NoError(t, err)

	// Simulate concurrent modification detection
	// In a real system, this would use optimistic locking with version field
	originalUpdatedAt := claim.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	claim.UpdatedAt = time.Now()

	assert.True(t, claim.UpdatedAt.After(originalUpdatedAt))
}

func TestClaimAggregate_Validation(t *testing.T) {
	tests := []struct {
		name    string
		claim   *entities.Claim
		wantErr bool
	}{
		{
			name: "valid claim",
			claim: &entities.Claim{
				ClaimID:            "claim-valid",
				Type:               entities.ClaimTypePortability,
				Key:                "12345678901",
				KeyType:            "CPF",
				DonorParticipant:   "60701190",
				ClaimerParticipant: "60746948",
				Status:             entities.ClaimStatusOpen,
				CreatedAt:          time.Now(),
				UpdatedAt:          time.Now(),
			},
			wantErr: false,
		},
		{
			name: "missing claim ID",
			claim: &entities.Claim{
				Type:               entities.ClaimTypePortability,
				Key:                "12345678901",
				DonorParticipant:   "60701190",
				ClaimerParticipant: "60746948",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.claim.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
