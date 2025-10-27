package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

func TestKeyStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   valueobjects.KeyStatus
		expected bool
	}{
		{
			name:     "Valid PENDING",
			status:   valueobjects.KeyStatusPending,
			expected: true,
		},
		{
			name:     "Valid ACTIVE",
			status:   valueobjects.KeyStatusActive,
			expected: true,
		},
		{
			name:     "Valid BLOCKED",
			status:   valueobjects.KeyStatusBlocked,
			expected: true,
		},
		{
			name:     "Valid DELETED",
			status:   valueobjects.KeyStatusDeleted,
			expected: true,
		},
		{
			name:     "Valid CLAIM_PENDING",
			status:   valueobjects.KeyStatusClaimPending,
			expected: true,
		},
		{
			name:     "Valid PORTABILITY_REQUESTED",
			status:   valueobjects.KeyStatusPortabilityRequested,
			expected: true,
		},
		{
			name:     "Valid OWNERSHIP_CONFIRMED",
			status:   valueobjects.KeyStatusOwnershipConfirmed,
			expected: true,
		},
		{
			name:     "Valid FAILED",
			status:   valueobjects.KeyStatusFailed,
			expected: true,
		},
		{
			name:     "Invalid status",
			status:   valueobjects.KeyStatus("INVALID"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.status.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKeyStatus_CanTransitionTo_Valid(t *testing.T) {
	tests := []struct {
		name        string
		fromStatus  valueobjects.KeyStatus
		toStatus    valueobjects.KeyStatus
		canTransition bool
	}{
		// PENDING transitions
		{
			name:        "PENDING to ACTIVE",
			fromStatus:  valueobjects.KeyStatusPending,
			toStatus:    valueobjects.KeyStatusActive,
			canTransition: true,
		},
		{
			name:        "PENDING to FAILED",
			fromStatus:  valueobjects.KeyStatusPending,
			toStatus:    valueobjects.KeyStatusFailed,
			canTransition: true,
		},
		{
			name:        "PENDING to DELETED",
			fromStatus:  valueobjects.KeyStatusPending,
			toStatus:    valueobjects.KeyStatusDeleted,
			canTransition: true,
		},
		// ACTIVE transitions
		{
			name:        "ACTIVE to BLOCKED",
			fromStatus:  valueobjects.KeyStatusActive,
			toStatus:    valueobjects.KeyStatusBlocked,
			canTransition: true,
		},
		{
			name:        "ACTIVE to DELETED",
			fromStatus:  valueobjects.KeyStatusActive,
			toStatus:    valueobjects.KeyStatusDeleted,
			canTransition: true,
		},
		{
			name:        "ACTIVE to CLAIM_PENDING",
			fromStatus:  valueobjects.KeyStatusActive,
			toStatus:    valueobjects.KeyStatusClaimPending,
			canTransition: true,
		},
		{
			name:        "ACTIVE to PORTABILITY_REQUESTED",
			fromStatus:  valueobjects.KeyStatusActive,
			toStatus:    valueobjects.KeyStatusPortabilityRequested,
			canTransition: true,
		},
		// BLOCKED transitions
		{
			name:        "BLOCKED to ACTIVE",
			fromStatus:  valueobjects.KeyStatusBlocked,
			toStatus:    valueobjects.KeyStatusActive,
			canTransition: true,
		},
		{
			name:        "BLOCKED to DELETED",
			fromStatus:  valueobjects.KeyStatusBlocked,
			toStatus:    valueobjects.KeyStatusDeleted,
			canTransition: true,
		},
		// CLAIM_PENDING transitions
		{
			name:        "CLAIM_PENDING to ACTIVE",
			fromStatus:  valueobjects.KeyStatusClaimPending,
			toStatus:    valueobjects.KeyStatusActive,
			canTransition: true,
		},
		{
			name:        "CLAIM_PENDING to OWNERSHIP_CONFIRMED",
			fromStatus:  valueobjects.KeyStatusClaimPending,
			toStatus:    valueobjects.KeyStatusOwnershipConfirmed,
			canTransition: true,
		},
		// OWNERSHIP_CONFIRMED transitions
		{
			name:        "OWNERSHIP_CONFIRMED to ACTIVE",
			fromStatus:  valueobjects.KeyStatusOwnershipConfirmed,
			toStatus:    valueobjects.KeyStatusActive,
			canTransition: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.fromStatus.CanTransitionTo(tt.toStatus)

			// Assert
			assert.Equal(t, tt.canTransition, result)
		})
	}
}

func TestKeyStatus_CanTransitionTo_Invalid(t *testing.T) {
	tests := []struct {
		name       string
		fromStatus valueobjects.KeyStatus
		toStatus   valueobjects.KeyStatus
	}{
		{
			name:       "PENDING to BLOCKED (invalid)",
			fromStatus: valueobjects.KeyStatusPending,
			toStatus:   valueobjects.KeyStatusBlocked,
		},
		{
			name:       "ACTIVE to PENDING (invalid)",
			fromStatus: valueobjects.KeyStatusActive,
			toStatus:   valueobjects.KeyStatusPending,
		},
		{
			name:       "DELETED to ACTIVE (invalid - final state)",
			fromStatus: valueobjects.KeyStatusDeleted,
			toStatus:   valueobjects.KeyStatusActive,
		},
		{
			name:       "FAILED to ACTIVE (invalid - final state)",
			fromStatus: valueobjects.KeyStatusFailed,
			toStatus:   valueobjects.KeyStatusActive,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.fromStatus.CanTransitionTo(tt.toStatus)

			// Assert
			assert.False(t, result, "Transition should not be allowed")
		})
	}
}

func TestNewKeyStatus_Success(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  valueobjects.KeyStatus
		wantError bool
	}{
		{
			name:      "Create PENDING",
			input:     "PENDING",
			expected:  valueobjects.KeyStatusPending,
			wantError: false,
		},
		{
			name:      "Create ACTIVE",
			input:     "ACTIVE",
			expected:  valueobjects.KeyStatusActive,
			wantError: false,
		},
		{
			name:      "Create BLOCKED",
			input:     "BLOCKED",
			expected:  valueobjects.KeyStatusBlocked,
			wantError: false,
		},
		{
			name:      "Invalid status",
			input:     "INVALID",
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := valueobjects.NewKeyStatus(tt.input)

			// Assert
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
