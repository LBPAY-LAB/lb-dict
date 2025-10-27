package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

func TestClaimType_IsValid(t *testing.T) {
	tests := []struct {
		name      string
		claimType valueobjects.ClaimType
		expected  bool
	}{
		{
			name:      "Valid OWNERSHIP",
			claimType: valueobjects.ClaimTypeOwnership,
			expected:  true,
		},
		{
			name:      "Valid PORTABILITY",
			claimType: valueobjects.ClaimTypePortability,
			expected:  true,
		},
		{
			name:      "Invalid type",
			claimType: valueobjects.ClaimType("INVALID"),
			expected:  false,
		},
		{
			name:      "Empty type",
			claimType: valueobjects.ClaimType(""),
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.claimType.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClaimType_String(t *testing.T) {
	tests := []struct {
		name      string
		claimType valueobjects.ClaimType
		expected  string
	}{
		{
			name:      "OWNERSHIP to string",
			claimType: valueobjects.ClaimTypeOwnership,
			expected:  "OWNERSHIP",
		},
		{
			name:      "PORTABILITY to string",
			claimType: valueobjects.ClaimTypePortability,
			expected:  "PORTABILITY",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.claimType.String()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewClaimType_Success(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  valueobjects.ClaimType
		wantError bool
	}{
		{
			name:      "Create OWNERSHIP",
			input:     "OWNERSHIP",
			expected:  valueobjects.ClaimTypeOwnership,
			wantError: false,
		},
		{
			name:      "Create PORTABILITY",
			input:     "PORTABILITY",
			expected:  valueobjects.ClaimTypePortability,
			wantError: false,
		},
		{
			name:      "Invalid type",
			input:     "INVALID",
			expected:  "",
			wantError: true,
		},
		{
			name:      "Empty type",
			input:     "",
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := valueobjects.NewClaimType(tt.input)

			// Assert
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "invalid claim type")
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}
