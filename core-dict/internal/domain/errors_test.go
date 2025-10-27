package domain_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lbpay-lab/core-dict/internal/domain"
)

func TestDomainErrors_All(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrInvalidKeyType",
			err:      domain.ErrInvalidKeyType,
			expected: "invalid key type",
		},
		{
			name:     "ErrInvalidKeyValue",
			err:      domain.ErrInvalidKeyValue,
			expected: "invalid key value",
		},
		{
			name:     "ErrDuplicateKey",
			err:      domain.ErrDuplicateKey,
			expected: "duplicate key",
		},
		{
			name:     "ErrEntryNotFound",
			err:      domain.ErrEntryNotFound,
			expected: "entry not found",
		},
		{
			name:     "ErrInvalidStatus",
			err:      domain.ErrInvalidStatus,
			expected: "invalid status",
		},
		{
			name:     "ErrMaxKeysExceeded",
			err:      domain.ErrMaxKeysExceeded,
			expected: "maximum number of keys exceeded",
		},
		{
			name:     "ErrInvalidClaim",
			err:      domain.ErrInvalidClaim,
			expected: "invalid claim",
		},
		{
			name:     "ErrClaimExpired",
			err:      domain.ErrClaimExpired,
			expected: "claim expired",
		},
		{
			name:     "ErrInvalidAccount",
			err:      domain.ErrInvalidAccount,
			expected: "invalid account",
		},
		{
			name:     "ErrInvalidParticipant",
			err:      domain.ErrInvalidParticipant,
			expected: "invalid participant",
		},
		{
			name:     "ErrUnauthorized",
			err:      domain.ErrUnauthorized,
			expected: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Assert
			assert.NotNil(t, tt.err)
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestErrInvalidKeyType(t *testing.T) {
	// Act
	err := domain.ErrInvalidKeyType

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "invalid key type", err.Error())
	assert.True(t, errors.Is(err, domain.ErrInvalidKeyType))
}

func TestErrInvalidKeyValue(t *testing.T) {
	// Act
	err := domain.ErrInvalidKeyValue

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "invalid key value", err.Error())
	assert.True(t, errors.Is(err, domain.ErrInvalidKeyValue))
}

func TestErrDuplicateKey(t *testing.T) {
	// Act
	err := domain.ErrDuplicateKey

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "duplicate key", err.Error())
	assert.True(t, errors.Is(err, domain.ErrDuplicateKey))
}

func TestErrEntryNotFound(t *testing.T) {
	// Act
	err := domain.ErrEntryNotFound

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "entry not found", err.Error())
	assert.True(t, errors.Is(err, domain.ErrEntryNotFound))
}

func TestErrInvalidStatus(t *testing.T) {
	// Act
	err := domain.ErrInvalidStatus

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "invalid status", err.Error())
	assert.True(t, errors.Is(err, domain.ErrInvalidStatus))
}

func TestErrInvalidClaim(t *testing.T) {
	// Act
	err := domain.ErrInvalidClaim

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "invalid claim", err.Error())
	assert.True(t, errors.Is(err, domain.ErrInvalidClaim))
}

func TestErrClaimExpired(t *testing.T) {
	// Act
	err := domain.ErrClaimExpired

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "claim expired", err.Error())
	assert.True(t, errors.Is(err, domain.ErrClaimExpired))
}

func TestErrUnauthorized(t *testing.T) {
	// Act
	err := domain.ErrUnauthorized

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "unauthorized", err.Error())
	assert.True(t, errors.Is(err, domain.ErrUnauthorized))
}

func TestErrMaxKeysExceeded(t *testing.T) {
	// Act
	err := domain.ErrMaxKeysExceeded

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "maximum number of keys exceeded", err.Error())
	assert.True(t, errors.Is(err, domain.ErrMaxKeysExceeded))
}

func TestErrInvalidAccount(t *testing.T) {
	// Act
	err := domain.ErrInvalidAccount

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "invalid account", err.Error())
	assert.True(t, errors.Is(err, domain.ErrInvalidAccount))
}

func TestErrInvalidParticipant(t *testing.T) {
	// Act
	err := domain.ErrInvalidParticipant

	// Assert
	assert.NotNil(t, err)
	assert.Equal(t, "invalid participant", err.Error())
	assert.True(t, errors.Is(err, domain.ErrInvalidParticipant))
}
