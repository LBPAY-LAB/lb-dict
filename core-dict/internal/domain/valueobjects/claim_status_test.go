package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

func TestClaimStatus_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		status   valueobjects.ClaimStatus
		expected bool
	}{
		{
			name:     "Valid OPEN",
			status:   valueobjects.ClaimStatusOpen,
			expected: true,
		},
		{
			name:     "Valid WAITING_RESOLUTION",
			status:   valueobjects.ClaimStatusWaitingResolution,
			expected: true,
		},
		{
			name:     "Valid CONFIRMED",
			status:   valueobjects.ClaimStatusConfirmed,
			expected: true,
		},
		{
			name:     "Valid CANCELLED",
			status:   valueobjects.ClaimStatusCancelled,
			expected: true,
		},
		{
			name:     "Valid COMPLETED",
			status:   valueobjects.ClaimStatusCompleted,
			expected: true,
		},
		{
			name:     "Valid EXPIRED",
			status:   valueobjects.ClaimStatusExpired,
			expected: true,
		},
		{
			name:     "Valid AUTO_CONFIRMED",
			status:   valueobjects.ClaimStatusAutoConfirmed,
			expected: true,
		},
		{
			name:     "Invalid status",
			status:   valueobjects.ClaimStatus("INVALID"),
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

func TestClaimStatus_IsFinal(t *testing.T) {
	tests := []struct {
		name     string
		status   valueobjects.ClaimStatus
		isFinal  bool
	}{
		{
			name:     "OPEN is not final",
			status:   valueobjects.ClaimStatusOpen,
			isFinal:  false,
		},
		{
			name:     "WAITING_RESOLUTION is not final",
			status:   valueobjects.ClaimStatusWaitingResolution,
			isFinal:  false,
		},
		{
			name:     "CONFIRMED is not final",
			status:   valueobjects.ClaimStatusConfirmed,
			isFinal:  false,
		},
		{
			name:     "COMPLETED is final",
			status:   valueobjects.ClaimStatusCompleted,
			isFinal:  true,
		},
		{
			name:     "CANCELLED is final",
			status:   valueobjects.ClaimStatusCancelled,
			isFinal:  true,
		},
		{
			name:     "EXPIRED is final",
			status:   valueobjects.ClaimStatusExpired,
			isFinal:  true,
		},
		{
			name:     "AUTO_CONFIRMED is final",
			status:   valueobjects.ClaimStatusAutoConfirmed,
			isFinal:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.status.IsFinal()

			// Assert
			assert.Equal(t, tt.isFinal, result)
		})
	}
}

func TestClaimStatus_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name          string
		fromStatus    valueobjects.ClaimStatus
		toStatus      valueobjects.ClaimStatus
		canTransition bool
	}{
		// OPEN transitions
		{
			name:          "OPEN to WAITING_RESOLUTION",
			fromStatus:    valueobjects.ClaimStatusOpen,
			toStatus:      valueobjects.ClaimStatusWaitingResolution,
			canTransition: true,
		},
		{
			name:          "OPEN to CONFIRMED",
			fromStatus:    valueobjects.ClaimStatusOpen,
			toStatus:      valueobjects.ClaimStatusConfirmed,
			canTransition: true,
		},
		{
			name:          "OPEN to CANCELLED",
			fromStatus:    valueobjects.ClaimStatusOpen,
			toStatus:      valueobjects.ClaimStatusCancelled,
			canTransition: true,
		},
		{
			name:          "OPEN to EXPIRED",
			fromStatus:    valueobjects.ClaimStatusOpen,
			toStatus:      valueobjects.ClaimStatusExpired,
			canTransition: true,
		},
		{
			name:          "OPEN to COMPLETED (invalid)",
			fromStatus:    valueobjects.ClaimStatusOpen,
			toStatus:      valueobjects.ClaimStatusCompleted,
			canTransition: false,
		},
		// WAITING_RESOLUTION transitions
		{
			name:          "WAITING_RESOLUTION to CONFIRMED",
			fromStatus:    valueobjects.ClaimStatusWaitingResolution,
			toStatus:      valueobjects.ClaimStatusConfirmed,
			canTransition: true,
		},
		{
			name:          "WAITING_RESOLUTION to CANCELLED",
			fromStatus:    valueobjects.ClaimStatusWaitingResolution,
			toStatus:      valueobjects.ClaimStatusCancelled,
			canTransition: true,
		},
		{
			name:          "WAITING_RESOLUTION to COMPLETED",
			fromStatus:    valueobjects.ClaimStatusWaitingResolution,
			toStatus:      valueobjects.ClaimStatusCompleted,
			canTransition: true,
		},
		{
			name:          "WAITING_RESOLUTION to EXPIRED",
			fromStatus:    valueobjects.ClaimStatusWaitingResolution,
			toStatus:      valueobjects.ClaimStatusExpired,
			canTransition: true,
		},
		{
			name:          "WAITING_RESOLUTION to AUTO_CONFIRMED",
			fromStatus:    valueobjects.ClaimStatusWaitingResolution,
			toStatus:      valueobjects.ClaimStatusAutoConfirmed,
			canTransition: true,
		},
		// CONFIRMED transitions
		{
			name:          "CONFIRMED to COMPLETED",
			fromStatus:    valueobjects.ClaimStatusConfirmed,
			toStatus:      valueobjects.ClaimStatusCompleted,
			canTransition: true,
		},
		{
			name:          "CONFIRMED to CANCELLED (invalid)",
			fromStatus:    valueobjects.ClaimStatusConfirmed,
			toStatus:      valueobjects.ClaimStatusCancelled,
			canTransition: false,
		},
		// Final states cannot transition
		{
			name:          "COMPLETED to OPEN (invalid - final state)",
			fromStatus:    valueobjects.ClaimStatusCompleted,
			toStatus:      valueobjects.ClaimStatusOpen,
			canTransition: false,
		},
		{
			name:          "CANCELLED to OPEN (invalid - final state)",
			fromStatus:    valueobjects.ClaimStatusCancelled,
			toStatus:      valueobjects.ClaimStatusOpen,
			canTransition: false,
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

func TestNewClaimStatus_Success(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  valueobjects.ClaimStatus
		wantError bool
	}{
		{
			name:      "Create OPEN",
			input:     "OPEN",
			expected:  valueobjects.ClaimStatusOpen,
			wantError: false,
		},
		{
			name:      "Create WAITING_RESOLUTION",
			input:     "WAITING_RESOLUTION",
			expected:  valueobjects.ClaimStatusWaitingResolution,
			wantError: false,
		},
		{
			name:      "Create CONFIRMED",
			input:     "CONFIRMED",
			expected:  valueobjects.ClaimStatusConfirmed,
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
			result, err := valueobjects.NewClaimStatus(tt.input)

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
