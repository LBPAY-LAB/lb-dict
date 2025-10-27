package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNewClaim_Success_Portability(t *testing.T) {
	// Arrange
	claimID := "CLAIM123456"
	claimType := ClaimTypePortability
	key := "+5511999999999"
	keyType := "PHONE"
	donorISPB := "12345678"
	claimerISPB := "87654321"

	// Act
	claim, err := NewClaim(claimID, claimType, key, keyType, donorISPB, claimerISPB)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, claim)
	assert.Equal(t, claimID, claim.ClaimID)
	assert.Equal(t, ClaimTypePortability, claim.Type)
	assert.Equal(t, ClaimStatusOpen, claim.Status)
	assert.Equal(t, key, claim.Key)
	assert.Equal(t, keyType, claim.KeyType)
	assert.Equal(t, donorISPB, claim.DonorParticipant)
	assert.Equal(t, claimerISPB, claim.ClaimerParticipant)
	assert.NotEmpty(t, claim.ID)
	assert.False(t, claim.CreatedAt.IsZero())
	assert.False(t, claim.UpdatedAt.IsZero())

	// Verify time periods
	expectedCompletionEnd := claim.CreatedAt.Add(7 * 24 * time.Hour)
	expectedExpiryDate := claim.CreatedAt.Add(30 * 24 * time.Hour)
	assert.WithinDuration(t, expectedCompletionEnd, claim.CompletionPeriodEnd, 1*time.Second)
	assert.WithinDuration(t, expectedExpiryDate, claim.ClaimExpiryDate, 1*time.Second)
}

func TestNewClaim_Success_Ownership(t *testing.T) {
	// Arrange
	claimID := "CLAIM789012"
	claimType := ClaimTypeOwnership
	key := "test@example.com"
	keyType := "EMAIL"
	donorISPB := "11111111"
	claimerISPB := "22222222"

	// Act
	claim, err := NewClaim(claimID, claimType, key, keyType, donorISPB, claimerISPB)

	// Assert
	require.NoError(t, err)
	require.NotNil(t, claim)
	assert.Equal(t, claimID, claim.ClaimID)
	assert.Equal(t, ClaimTypeOwnership, claim.Type)
	assert.Equal(t, ClaimStatusOpen, claim.Status)
	assert.Equal(t, key, claim.Key)
	assert.Equal(t, donorISPB, claim.DonorParticipant)
	assert.Equal(t, claimerISPB, claim.ClaimerParticipant)
}

func TestNewClaim_InvalidDonorISPB(t *testing.T) {
	tests := []struct {
		name      string
		donorISPB string
		wantErr   string
	}{
		{
			name:      "too_short",
			donorISPB: "1234567",
			wantErr:   "donor ISPB must be 8 digits",
		},
		{
			name:      "too_long",
			donorISPB: "123456789",
			wantErr:   "donor ISPB must be 8 digits",
		},
		{
			name:      "empty",
			donorISPB: "",
			wantErr:   "donor ISPB must be 8 digits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claimID := "CLAIM123"
			claimType := ClaimTypePortability
			key := "+5511999999999"
			keyType := "PHONE"
			claimerISPB := "87654321"

			// Act
			claim, err := NewClaim(claimID, claimType, key, keyType, tt.donorISPB, claimerISPB)

			// Assert
			require.Error(t, err)
			assert.Nil(t, claim)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestNewClaim_InvalidClaimerISPB(t *testing.T) {
	tests := []struct {
		name        string
		claimerISPB string
		wantErr     string
	}{
		{
			name:        "too_short",
			claimerISPB: "1234567",
			wantErr:     "claimer ISPB must be 8 digits",
		},
		{
			name:        "too_long",
			claimerISPB: "123456789",
			wantErr:     "claimer ISPB must be 8 digits",
		},
		{
			name:        "empty",
			claimerISPB: "",
			wantErr:     "claimer ISPB must be 8 digits",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claimID := "CLAIM123"
			claimType := ClaimTypePortability
			key := "+5511999999999"
			keyType := "PHONE"
			donorISPB := "12345678"

			// Act
			claim, err := NewClaim(claimID, claimType, key, keyType, donorISPB, tt.claimerISPB)

			// Assert
			require.Error(t, err)
			assert.Nil(t, claim)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestNewClaim_SameDonorAndClaimer(t *testing.T) {
	// Arrange
	claimID := "CLAIM123"
	claimType := ClaimTypePortability
	key := "+5511999999999"
	keyType := "PHONE"
	sameISPB := "12345678"

	// Act
	claim, err := NewClaim(claimID, claimType, key, keyType, sameISPB, sameISPB)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claim)
	assert.Contains(t, err.Error(), "donor and claimer must be different")
}

func TestNewClaim_EmptyKey(t *testing.T) {
	// Arrange
	claimID := "CLAIM123"
	claimType := ClaimTypePortability
	key := ""
	keyType := "PHONE"
	donorISPB := "12345678"
	claimerISPB := "87654321"

	// Act
	claim, err := NewClaim(claimID, claimType, key, keyType, donorISPB, claimerISPB)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claim)
	assert.Contains(t, err.Error(), "key cannot be empty")
}

func TestNewClaim_EmptyClaimID(t *testing.T) {
	// Arrange
	claimID := ""
	claimType := ClaimTypePortability
	key := "+5511999999999"
	keyType := "PHONE"
	donorISPB := "12345678"
	claimerISPB := "87654321"

	// Act
	claim, err := NewClaim(claimID, claimType, key, keyType, donorISPB, claimerISPB)

	// Assert
	require.Error(t, err)
	assert.Nil(t, claim)
	assert.Contains(t, err.Error(), "claim_id cannot be empty")
}

// =============================================================================
// Status Transition Tests
// =============================================================================

func TestClaim_Confirm_Success(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	require.Equal(t, ClaimStatusOpen, claim.Status)

	// Act
	err = claim.Confirm()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, ClaimStatusConfirmed, claim.Status)
	assert.NotNil(t, claim.ConfirmedAt)
	assert.False(t, claim.ConfirmedAt.IsZero())
	assert.False(t, claim.UpdatedAt.IsZero())
}

func TestClaim_Confirm_FromWaitingResolution(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)

	err = claim.MoveToWaitingResolution()
	require.NoError(t, err)
	require.Equal(t, ClaimStatusWaitingResolution, claim.Status)

	// Act
	err = claim.Confirm()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, ClaimStatusConfirmed, claim.Status)
	assert.NotNil(t, claim.ConfirmedAt)
}

func TestClaim_Confirm_WrongStatus(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus ClaimStatus
	}{
		{
			name:          "from_cancelled",
			initialStatus: ClaimStatusCancelled,
		},
		{
			name:          "from_completed",
			initialStatus: ClaimStatusCompleted,
		},
		{
			name:          "from_expired",
			initialStatus: ClaimStatusExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
			require.NoError(t, err)
			claim.Status = tt.initialStatus // Force status

			// Act
			err = claim.Confirm()

			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), "claim can only be confirmed from OPEN or WAITING_RESOLUTION status")
			assert.Equal(t, tt.initialStatus, claim.Status) // Status unchanged
		})
	}
}

func TestClaim_Cancel_Success(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	reason := "User requested cancellation"

	// Act
	err = claim.Cancel(reason)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, ClaimStatusCancelled, claim.Status)
	assert.Equal(t, reason, claim.CancellationReason)
	assert.NotNil(t, claim.CancelledAt)
	assert.False(t, claim.CancelledAt.IsZero())
	assert.False(t, claim.UpdatedAt.IsZero())
}

func TestClaim_Cancel_WithoutReason(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	reason := "" // Empty reason should be allowed

	// Act
	err = claim.Cancel(reason)

	// Assert
	require.NoError(t, err)
	assert.Equal(t, ClaimStatusCancelled, claim.Status)
	assert.Equal(t, "", claim.CancellationReason)
}

func TestClaim_Cancel_FromCompleted_Fails(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	claim.Status = ClaimStatusCompleted // Force status

	// Act
	err = claim.Cancel("Some reason")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel a completed or expired claim")
	assert.Equal(t, ClaimStatusCompleted, claim.Status) // Status unchanged
}

func TestClaim_Cancel_FromExpired_Fails(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	claim.Status = ClaimStatusExpired // Force status

	// Act
	err = claim.Cancel("Some reason")

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot cancel a completed or expired claim")
	assert.Equal(t, ClaimStatusExpired, claim.Status) // Status unchanged
}

func TestClaim_Complete_Success(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	err = claim.Confirm()
	require.NoError(t, err)
	require.Equal(t, ClaimStatusConfirmed, claim.Status)

	// Act
	err = claim.Complete()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, ClaimStatusCompleted, claim.Status)
	assert.NotNil(t, claim.CompletedAt)
	assert.False(t, claim.CompletedAt.IsZero())
	assert.False(t, claim.UpdatedAt.IsZero())
}

func TestClaim_Complete_NotConfirmed(t *testing.T) {
	tests := []struct {
		name          string
		initialStatus ClaimStatus
	}{
		{
			name:          "from_open",
			initialStatus: ClaimStatusOpen,
		},
		{
			name:          "from_cancelled",
			initialStatus: ClaimStatusCancelled,
		},
		{
			name:          "from_waiting_resolution",
			initialStatus: ClaimStatusWaitingResolution,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
			require.NoError(t, err)
			claim.Status = tt.initialStatus // Force status

			// Act
			err = claim.Complete()

			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), "claim must be confirmed before completion")
			assert.Equal(t, tt.initialStatus, claim.Status) // Status unchanged
		})
	}
}

func TestClaim_Expire_Success(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)

	// Act
	err = claim.Expire()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, ClaimStatusExpired, claim.Status)
	assert.NotNil(t, claim.ExpiredAt)
	assert.False(t, claim.ExpiredAt.IsZero())
	assert.False(t, claim.UpdatedAt.IsZero())
}

func TestClaim_Expire_FromCompleted_Fails(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	claim.Status = ClaimStatusCompleted // Force status

	// Act
	err = claim.Expire()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cannot expire a completed claim")
	assert.Equal(t, ClaimStatusCompleted, claim.Status) // Status unchanged
}

func TestClaim_MoveToWaitingResolution_Success(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	require.Equal(t, ClaimStatusOpen, claim.Status)

	// Act
	err = claim.MoveToWaitingResolution()

	// Assert
	require.NoError(t, err)
	assert.Equal(t, ClaimStatusWaitingResolution, claim.Status)
	assert.False(t, claim.UpdatedAt.IsZero())
}

func TestClaim_MoveToWaitingResolution_NotOpen_Fails(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	claim.Status = ClaimStatusConfirmed // Force status

	// Act
	err = claim.MoveToWaitingResolution()

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "claim must be in OPEN status")
	assert.Equal(t, ClaimStatusConfirmed, claim.Status) // Status unchanged
}

// =============================================================================
// Validation Tests
// =============================================================================

func TestClaim_ValidateStatusTransition_ValidTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus ClaimStatus
		newStatus     ClaimStatus
	}{
		// From OPEN
		{
			name:          "open_to_waiting_resolution",
			currentStatus: ClaimStatusOpen,
			newStatus:     ClaimStatusWaitingResolution,
		},
		{
			name:          "open_to_cancelled",
			currentStatus: ClaimStatusOpen,
			newStatus:     ClaimStatusCancelled,
		},
		// From WAITING_RESOLUTION
		{
			name:          "waiting_resolution_to_confirmed",
			currentStatus: ClaimStatusWaitingResolution,
			newStatus:     ClaimStatusConfirmed,
		},
		{
			name:          "waiting_resolution_to_cancelled",
			currentStatus: ClaimStatusWaitingResolution,
			newStatus:     ClaimStatusCancelled,
		},
		{
			name:          "waiting_resolution_to_expired",
			currentStatus: ClaimStatusWaitingResolution,
			newStatus:     ClaimStatusExpired,
		},
		// From CONFIRMED
		{
			name:          "confirmed_to_completed",
			currentStatus: ClaimStatusConfirmed,
			newStatus:     ClaimStatusCompleted,
		},
		{
			name:          "confirmed_to_cancelled",
			currentStatus: ClaimStatusConfirmed,
			newStatus:     ClaimStatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
			require.NoError(t, err)
			claim.Status = tt.currentStatus

			// Act
			err = claim.ValidateStatusTransition(tt.newStatus)

			// Assert
			require.NoError(t, err)
		})
	}
}

func TestClaim_ValidateStatusTransition_InvalidTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus ClaimStatus
		newStatus     ClaimStatus
	}{
		// From OPEN (invalid)
		{
			name:          "open_to_confirmed",
			currentStatus: ClaimStatusOpen,
			newStatus:     ClaimStatusConfirmed,
		},
		{
			name:          "open_to_completed",
			currentStatus: ClaimStatusOpen,
			newStatus:     ClaimStatusCompleted,
		},
		// From CANCELLED (terminal)
		{
			name:          "cancelled_to_open",
			currentStatus: ClaimStatusCancelled,
			newStatus:     ClaimStatusOpen,
		},
		{
			name:          "cancelled_to_confirmed",
			currentStatus: ClaimStatusCancelled,
			newStatus:     ClaimStatusConfirmed,
		},
		// From COMPLETED (terminal)
		{
			name:          "completed_to_open",
			currentStatus: ClaimStatusCompleted,
			newStatus:     ClaimStatusOpen,
		},
		{
			name:          "completed_to_cancelled",
			currentStatus: ClaimStatusCompleted,
			newStatus:     ClaimStatusCancelled,
		},
		// From EXPIRED (terminal)
		{
			name:          "expired_to_open",
			currentStatus: ClaimStatusExpired,
			newStatus:     ClaimStatusOpen,
		},
		{
			name:          "expired_to_confirmed",
			currentStatus: ClaimStatusExpired,
			newStatus:     ClaimStatusConfirmed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
			require.NoError(t, err)
			claim.Status = tt.currentStatus

			// Act
			err = claim.ValidateStatusTransition(tt.newStatus)

			// Assert
			require.Error(t, err)
			assert.Contains(t, err.Error(), "invalid status transition")
		})
	}
}

// =============================================================================
// Helper Methods Tests
// =============================================================================

func TestClaim_IsExpired_True(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)

	// Set expiry date to the past
	claim.ClaimExpiryDate = time.Now().Add(-1 * time.Hour)

	// Act
	isExpired := claim.IsExpired()

	// Assert
	assert.True(t, isExpired)
}

func TestClaim_IsExpired_False(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)

	// Set expiry date to the future (default is 30 days from now)
	require.True(t, claim.ClaimExpiryDate.After(time.Now()))

	// Act
	isExpired := claim.IsExpired()

	// Assert
	assert.False(t, isExpired)
}

func TestClaim_IsCompletionPeriodExpired_True(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)

	// Set completion period end to the past
	claim.CompletionPeriodEnd = time.Now().Add(-1 * time.Hour)

	// Act
	isExpired := claim.IsCompletionPeriodExpired()

	// Assert
	assert.True(t, isExpired)
}

func TestClaim_IsCompletionPeriodExpired_False(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)

	// Set completion period end to the future (default is 7 days from now)
	require.True(t, claim.CompletionPeriodEnd.After(time.Now()))

	// Act
	isExpired := claim.IsCompletionPeriodExpired()

	// Assert
	assert.False(t, isExpired)
}

func TestClaim_IsActive_True(t *testing.T) {
	tests := []struct {
		name   string
		status ClaimStatus
	}{
		{
			name:   "open",
			status: ClaimStatusOpen,
		},
		{
			name:   "waiting_resolution",
			status: ClaimStatusWaitingResolution,
		},
		{
			name:   "confirmed",
			status: ClaimStatusConfirmed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
			require.NoError(t, err)
			claim.Status = tt.status

			// Act
			isActive := claim.IsActive()

			// Assert
			assert.True(t, isActive)
		})
	}
}

func TestClaim_IsActive_False(t *testing.T) {
	tests := []struct {
		name   string
		status ClaimStatus
	}{
		{
			name:   "completed",
			status: ClaimStatusCompleted,
		},
		{
			name:   "cancelled",
			status: ClaimStatusCancelled,
		},
		{
			name:   "expired",
			status: ClaimStatusExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
			require.NoError(t, err)
			claim.Status = tt.status

			// Act
			isActive := claim.IsActive()

			// Assert
			assert.False(t, isActive)
		})
	}
}

func TestClaim_CanBeCancelled_True(t *testing.T) {
	tests := []struct {
		name   string
		status ClaimStatus
	}{
		{
			name:   "open",
			status: ClaimStatusOpen,
		},
		{
			name:   "waiting_resolution",
			status: ClaimStatusWaitingResolution,
		},
		{
			name:   "confirmed",
			status: ClaimStatusConfirmed,
		},
		{
			name:   "cancelled", // Already cancelled, but method still returns true
			status: ClaimStatusCancelled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
			require.NoError(t, err)
			claim.Status = tt.status

			// Act
			canCancel := claim.CanBeCancelled()

			// Assert - CanBeCancelled returns true if NOT completed or expired
			expected := tt.status != ClaimStatusCompleted && tt.status != ClaimStatusExpired
			assert.Equal(t, expected, canCancel)
		})
	}
}

func TestClaim_CanBeCancelled_False(t *testing.T) {
	tests := []struct {
		name   string
		status ClaimStatus
	}{
		{
			name:   "completed",
			status: ClaimStatusCompleted,
		},
		{
			name:   "expired",
			status: ClaimStatusExpired,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
			require.NoError(t, err)
			claim.Status = tt.status

			// Act
			canCancel := claim.CanBeCancelled()

			// Assert
			assert.False(t, canCancel)
		})
	}
}

// =============================================================================
// Time Tests
// =============================================================================

func TestClaim_ExpiryDate_Is30Days(t *testing.T) {
	// Arrange
	beforeCreation := time.Now()
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	afterCreation := time.Now()

	// Act
	expectedMinExpiry := beforeCreation.Add(30 * 24 * time.Hour)
	expectedMaxExpiry := afterCreation.Add(30 * 24 * time.Hour)

	// Assert
	assert.True(t, claim.ClaimExpiryDate.After(expectedMinExpiry) || claim.ClaimExpiryDate.Equal(expectedMinExpiry))
	assert.True(t, claim.ClaimExpiryDate.Before(expectedMaxExpiry) || claim.ClaimExpiryDate.Equal(expectedMaxExpiry))

	// More precise check
	duration := claim.ClaimExpiryDate.Sub(claim.CreatedAt)
	assert.Equal(t, 30*24*time.Hour, duration)
}

func TestClaim_CompletionPeriod_Is7Days(t *testing.T) {
	// Arrange
	beforeCreation := time.Now()
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	afterCreation := time.Now()

	// Act
	expectedMinCompletion := beforeCreation.Add(7 * 24 * time.Hour)
	expectedMaxCompletion := afterCreation.Add(7 * 24 * time.Hour)

	// Assert
	assert.True(t, claim.CompletionPeriodEnd.After(expectedMinCompletion) || claim.CompletionPeriodEnd.Equal(expectedMinCompletion))
	assert.True(t, claim.CompletionPeriodEnd.Before(expectedMaxCompletion) || claim.CompletionPeriodEnd.Equal(expectedMaxCompletion))

	// More precise check
	duration := claim.CompletionPeriodEnd.Sub(claim.CreatedAt)
	assert.Equal(t, 7*24*time.Hour, duration)
}

func TestClaim_TimestampsUpdated_OnStatusChange(t *testing.T) {
	// Arrange
	claim, err := NewClaim("CLAIM123", ClaimTypePortability, "+5511999999999", "PHONE", "12345678", "87654321")
	require.NoError(t, err)
	originalUpdatedAt := claim.UpdatedAt

	// Wait a bit to ensure timestamp difference
	time.Sleep(10 * time.Millisecond)

	// Act - Confirm claim
	err = claim.Confirm()
	require.NoError(t, err)

	// Assert
	assert.True(t, claim.UpdatedAt.After(originalUpdatedAt))
	assert.NotNil(t, claim.ConfirmedAt)
}