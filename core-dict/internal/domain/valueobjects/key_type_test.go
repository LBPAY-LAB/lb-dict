package valueobjects_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/valueobjects"
)

func TestKeyType_IsValid(t *testing.T) {
	tests := []struct {
		name     string
		keyType  valueobjects.KeyType
		expected bool
	}{
		{
			name:     "Valid CPF",
			keyType:  valueobjects.KeyTypeCPF,
			expected: true,
		},
		{
			name:     "Valid CNPJ",
			keyType:  valueobjects.KeyTypeCNPJ,
			expected: true,
		},
		{
			name:     "Valid EMAIL",
			keyType:  valueobjects.KeyTypeEmail,
			expected: true,
		},
		{
			name:     "Valid PHONE",
			keyType:  valueobjects.KeyTypePhone,
			expected: true,
		},
		{
			name:     "Valid EVP",
			keyType:  valueobjects.KeyTypeEVP,
			expected: true,
		},
		{
			name:     "Invalid type",
			keyType:  valueobjects.KeyType("INVALID"),
			expected: false,
		},
		{
			name:     "Empty type",
			keyType:  valueobjects.KeyType(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.keyType.IsValid()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestKeyType_Format_CPF(t *testing.T) {
	// Arrange
	keyType := valueobjects.KeyTypeCPF

	// Act
	result := keyType.String()

	// Assert
	assert.Equal(t, "CPF", result)
}

func TestKeyType_Format_EVP(t *testing.T) {
	tests := []struct {
		name     string
		keyType  valueobjects.KeyType
		expected string
	}{
		{
			name:     "CPF format",
			keyType:  valueobjects.KeyTypeCPF,
			expected: "CPF",
		},
		{
			name:     "CNPJ format",
			keyType:  valueobjects.KeyTypeCNPJ,
			expected: "CNPJ",
		},
		{
			name:     "EMAIL format",
			keyType:  valueobjects.KeyTypeEmail,
			expected: "EMAIL",
		},
		{
			name:     "PHONE format",
			keyType:  valueobjects.KeyTypePhone,
			expected: "PHONE",
		},
		{
			name:     "EVP format",
			keyType:  valueobjects.KeyTypeEVP,
			expected: "EVP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := tt.keyType.String()

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewKeyType_Success(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		expected  valueobjects.KeyType
		wantError bool
	}{
		{
			name:      "Create CPF",
			input:     "CPF",
			expected:  valueobjects.KeyTypeCPF,
			wantError: false,
		},
		{
			name:      "Create CNPJ",
			input:     "CNPJ",
			expected:  valueobjects.KeyTypeCNPJ,
			wantError: false,
		},
		{
			name:      "Create EMAIL",
			input:     "EMAIL",
			expected:  valueobjects.KeyTypeEmail,
			wantError: false,
		},
		{
			name:      "Create PHONE",
			input:     "PHONE",
			expected:  valueobjects.KeyTypePhone,
			wantError: false,
		},
		{
			name:      "Create EVP",
			input:     "EVP",
			expected:  valueobjects.KeyTypeEVP,
			wantError: false,
		},
		{
			name:      "Invalid type",
			input:     "INVALID",
			expected:  "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result, err := valueobjects.NewKeyType(tt.input)

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
