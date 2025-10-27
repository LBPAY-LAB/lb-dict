package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

func TestNewParticipant_Success(t *testing.T) {
	tests := []struct {
		name      string
		ispb      string
		partName  string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "Valid participant",
			ispb:      "12345678",
			partName:  "Bank A",
			wantError: false,
		},
		{
			name:      "Another valid participant",
			ispb:      "87654321",
			partName:  "Bank B",
			wantError: false,
		},
		{
			name:      "Invalid ISPB - too short",
			ispb:      "1234567",
			partName:  "Bank A",
			wantError: true,
			errorMsg:  "invalid ISPB",
		},
		{
			name:      "Invalid ISPB - too long",
			ispb:      "123456789",
			partName:  "Bank A",
			wantError: true,
			errorMsg:  "invalid ISPB",
		},
		{
			name:      "Invalid ISPB - contains letters",
			ispb:      "1234567A",
			partName:  "Bank A",
			wantError: true,
			errorMsg:  "invalid ISPB",
		},
		{
			name:      "Empty name",
			ispb:      "12345678",
			partName:  "",
			wantError: true,
			errorMsg:  "name cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			participant, err := valueobjects.NewParticipant(tt.ispb, tt.partName)

			// Assert
			if tt.wantError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.ispb, participant.ISPB)
				assert.Equal(t, tt.partName, participant.Name)
			}
		})
	}
}

func TestParticipant_Validate(t *testing.T) {
	tests := []struct {
		name      string
		ispb      string
		partName  string
		wantError bool
	}{
		{
			name:      "Valid participant",
			ispb:      "12345678",
			partName:  "Bank A",
			wantError: false,
		},
		{
			name:      "Invalid ISPB",
			ispb:      "123",
			partName:  "Bank A",
			wantError: true,
		},
		{
			name:      "Empty name",
			ispb:      "12345678",
			partName:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			_, err := valueobjects.NewParticipant(tt.ispb, tt.partName)

			// Assert
			if tt.wantError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestParticipant_Equals(t *testing.T) {
	// Arrange
	participant1, _ := valueobjects.NewParticipant("12345678", "Bank A")
	participant2, _ := valueobjects.NewParticipant("12345678", "Bank A (Different Name)")
	participant3, _ := valueobjects.NewParticipant("87654321", "Bank B")

	tests := []struct {
		name     string
		p1       valueobjects.Participant
		p2       valueobjects.Participant
		expected bool
	}{
		{
			name:     "Same ISPB - equal",
			p1:       participant1,
			p2:       participant2,
			expected: true,
		},
		{
			name:     "Different ISPB - not equal",
			p1:       participant1,
			p2:       participant3,
			expected: false,
		},
		{
			name:     "Same participant - equal",
			p1:       participant1,
			p2:       participant1,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.p1.Equals(tt.p2)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParticipant_String(t *testing.T) {
	tests := []struct {
		name     string
		ispb     string
		partName string
		expected string
	}{
		{
			name:     "Bank A",
			ispb:     "12345678",
			partName: "Bank A",
			expected: "12345678 - Bank A",
		},
		{
			name:     "Bank B",
			ispb:     "87654321",
			partName: "Banco do Brasil",
			expected: "87654321 - Banco do Brasil",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			participant, _ := valueobjects.NewParticipant(tt.ispb, tt.partName)

			// Act
			result := participant.String()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}
