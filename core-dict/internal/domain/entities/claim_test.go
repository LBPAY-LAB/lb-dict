package entities_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

func TestNewClaim_Success(t *testing.T) {
	// Arrange
	entryKey := "12345678901"
	claimType := valueobjects.ClaimTypeOwnership
	claimer, _ := valueobjects.NewParticipant("12345678", "Bank A")
	donor, _ := valueobjects.NewParticipant("87654321", "Bank B")
	claimerAccountID := uuid.New()
	donorAccountID := uuid.New()

	// Act
	claim, err := entities.NewClaim(entryKey, claimType, claimer, donor, claimerAccountID, donorAccountID)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, claim.ID)
	assert.Equal(t, entryKey, claim.EntryKey)
	assert.Equal(t, claimType, claim.ClaimType)
	assert.Equal(t, valueobjects.ClaimStatusOpen, claim.Status)
	assert.Equal(t, claimer.ISPB, claim.ClaimerParticipant.ISPB)
	assert.Equal(t, donor.ISPB, claim.DonorParticipant.ISPB)
	assert.Equal(t, 30, claim.CompletionPeriodDays)
	assert.True(t, claim.ExpiresAt.After(time.Now()))
}

func TestClaim_Confirm_Success(t *testing.T) {
	// Arrange
	claimer, _ := valueobjects.NewParticipant("12345678", "Bank A")
	donor, _ := valueobjects.NewParticipant("87654321", "Bank B")
	claim, _ := entities.NewClaim("12345678901", valueobjects.ClaimTypeOwnership, claimer, donor, uuid.New(), uuid.New())
	reason := "Owner confirmed"

	// Act
	err := claim.Confirm(reason)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, valueobjects.ClaimStatusConfirmed, claim.Status)
	assert.Equal(t, "APPROVED", claim.ResolutionType)
	assert.Equal(t, reason, claim.ResolutionReason)
	assert.NotNil(t, claim.ResolutionDate)
}

func TestClaim_Cancel_Success(t *testing.T) {
	// Arrange
	claimer, _ := valueobjects.NewParticipant("12345678", "Bank A")
	donor, _ := valueobjects.NewParticipant("87654321", "Bank B")
	claim, _ := entities.NewClaim("12345678901", valueobjects.ClaimTypeOwnership, claimer, donor, uuid.New(), uuid.New())
	reason := "Cancelled by claimer"

	// Act
	err := claim.Cancel(reason)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, valueobjects.ClaimStatusCancelled, claim.Status)
	assert.Equal(t, "CANCELLED", claim.ResolutionType)
	assert.Equal(t, reason, claim.ResolutionReason)
	assert.NotNil(t, claim.ResolutionDate)
}

func TestClaim_Complete_Success(t *testing.T) {
	// Arrange
	claimer, _ := valueobjects.NewParticipant("12345678", "Bank A")
	donor, _ := valueobjects.NewParticipant("87654321", "Bank B")
	claim, _ := entities.NewClaim("12345678901", valueobjects.ClaimTypeOwnership, claimer, donor, uuid.New(), uuid.New())

	// First confirm the claim
	_ = claim.Confirm("Confirmed")

	// Act
	err := claim.Complete()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, valueobjects.ClaimStatusCompleted, claim.Status)
	assert.Equal(t, "APPROVED", claim.ResolutionType)
	assert.NotNil(t, claim.ResolutionDate)
}

func TestClaim_Expire_Success(t *testing.T) {
	// Arrange
	claimer, _ := valueobjects.NewParticipant("12345678", "Bank A")
	donor, _ := valueobjects.NewParticipant("87654321", "Bank B")
	claim, _ := entities.NewClaim("12345678901", valueobjects.ClaimTypeOwnership, claimer, donor, uuid.New(), uuid.New())

	// Set expiration date to the past
	claim.ExpiresAt = time.Now().Add(-1 * time.Hour)

	// Act
	err := claim.Expire()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, valueobjects.ClaimStatusExpired, claim.Status)
	assert.Equal(t, "TIMEOUT", claim.ResolutionType)
	assert.Equal(t, "No response within completion period", claim.ResolutionReason)
	assert.NotNil(t, claim.ResolutionDate)
}

func TestClaim_AutoConfirm_30Days(t *testing.T) {
	// Arrange
	claimer, _ := valueobjects.NewParticipant("12345678", "Bank A")
	donor, _ := valueobjects.NewParticipant("87654321", "Bank B")
	claim, _ := entities.NewClaim("12345678901", valueobjects.ClaimTypeOwnership, claimer, donor, uuid.New(), uuid.New())

	// Set status to waiting resolution and expiration date to the past
	claim.Status = valueobjects.ClaimStatusWaitingResolution
	claim.ExpiresAt = time.Now().Add(-1 * time.Hour)

	// Act
	err := claim.AutoConfirm()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, valueobjects.ClaimStatusAutoConfirmed, claim.Status)
	assert.Equal(t, "TIMEOUT", claim.ResolutionType)
	assert.Equal(t, "Auto-confirmed after completion period", claim.ResolutionReason)
	assert.NotNil(t, claim.ResolutionDate)
}

func TestClaim_IsExpired(t *testing.T) {
	// Arrange
	claimer, _ := valueobjects.NewParticipant("12345678", "Bank A")
	donor, _ := valueobjects.NewParticipant("87654321", "Bank B")

	tests := []struct {
		name       string
		expiresAt  time.Time
		wantExpired bool
	}{
		{
			name:       "Not expired - future date",
			expiresAt:  time.Now().Add(24 * time.Hour),
			wantExpired: false,
		},
		{
			name:       "Expired - past date",
			expiresAt:  time.Now().Add(-1 * time.Hour),
			wantExpired: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim, _ := entities.NewClaim("12345678901", valueobjects.ClaimTypeOwnership, claimer, donor, uuid.New(), uuid.New())
			claim.ExpiresAt = tt.expiresAt

			// Act
			result := claim.IsExpired()

			// Assert
			assert.Equal(t, tt.wantExpired, result)
		})
	}
}

func TestClaim_CanBeCancelled(t *testing.T) {
	// Arrange
	claimer, _ := valueobjects.NewParticipant("12345678", "Bank A")
	donor, _ := valueobjects.NewParticipant("87654321", "Bank B")

	tests := []struct {
		name         string
		status       valueobjects.ClaimStatus
		canBeCancelled bool
	}{
		{
			name:         "Open claim can be cancelled",
			status:       valueobjects.ClaimStatusOpen,
			canBeCancelled: true,
		},
		{
			name:         "Waiting resolution can be cancelled",
			status:       valueobjects.ClaimStatusWaitingResolution,
			canBeCancelled: true,
		},
		{
			name:         "Completed claim cannot be cancelled",
			status:       valueobjects.ClaimStatusCompleted,
			canBeCancelled: false,
		},
		{
			name:         "Already cancelled claim cannot be cancelled again",
			status:       valueobjects.ClaimStatusCancelled,
			canBeCancelled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim, _ := entities.NewClaim("12345678901", valueobjects.ClaimTypeOwnership, claimer, donor, uuid.New(), uuid.New())
			claim.Status = tt.status

			// Act
			err := claim.Cancel("Test cancellation")

			// Assert
			if tt.canBeCancelled {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}
