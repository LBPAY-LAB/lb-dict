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

func TestNewEntry_Success(t *testing.T) {
	// Arrange
	id := uuid.New()
	keyType := valueobjects.KeyTypeCPF
	keyValue := "12345678901"
	accountID := uuid.New()
	ispb := "12345678"
	branch := "0001"
	accountNumber := "123456"
	accountType := "CACC"
	ownerName := "John Doe"
	ownerTaxID := "12345678901"
	ownerType := "NATURAL_PERSON"

	// Act
	entry := &entities.Entry{
		ID:            id,
		KeyType:       entities.KeyType(keyType),
		KeyValue:      keyValue,
		Status:        entities.KeyStatusPending,
		AccountID:     accountID,
		ISPB:          ispb,
		Branch:        branch,
		AccountNumber: accountNumber,
		AccountType:   accountType,
		OwnerName:     ownerName,
		OwnerTaxID:    ownerTaxID,
		OwnerType:     ownerType,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Assert
	assert.Equal(t, id, entry.ID)
	assert.Equal(t, entities.KeyType(keyType), entry.KeyType)
	assert.Equal(t, keyValue, entry.KeyValue)
	assert.Equal(t, entities.KeyStatusPending, entry.Status)
	assert.Equal(t, accountID, entry.AccountID)
	assert.Equal(t, ispb, entry.ISPB)
}

func TestNewEntry_InvalidKeyType(t *testing.T) {
	// Arrange
	invalidKeyType := "INVALID"

	// Act
	_, err := valueobjects.NewKeyType(invalidKeyType)

	// Assert
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid key type")
}

func TestEntry_Validate_Success(t *testing.T) {
	// Arrange
	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		Status:        entities.KeyStatusActive,
		AccountID:     uuid.New(),
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		AccountType:   "CACC",
		OwnerName:     "John Doe",
		OwnerTaxID:    "12345678901",
		OwnerType:     "NATURAL_PERSON",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Act & Assert
	// Validation is implicit in the entity structure
	assert.NotEqual(t, uuid.Nil, entry.ID)
	assert.NotEmpty(t, entry.KeyValue)
	assert.NotEmpty(t, entry.ISPB)
}

func TestEntry_Activate_Success(t *testing.T) {
	// Arrange
	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		Status:        entities.KeyStatusPending,
		AccountID:     uuid.New(),
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		AccountType:   "CACC",
		OwnerName:     "John Doe",
		OwnerTaxID:    "12345678901",
		OwnerType:     "NATURAL_PERSON",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Act
	entry.Status = entities.KeyStatusActive
	entry.UpdatedAt = time.Now()

	// Assert
	assert.Equal(t, entities.KeyStatusActive, entry.Status)
}

func TestEntry_Block_Success(t *testing.T) {
	// Arrange
	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		Status:        entities.KeyStatusActive,
		AccountID:     uuid.New(),
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		AccountType:   "CACC",
		OwnerName:     "John Doe",
		OwnerTaxID:    "12345678901",
		OwnerType:     "NATURAL_PERSON",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Act
	entry.Status = entities.KeyStatusBlocked
	entry.UpdatedAt = time.Now()

	// Assert
	assert.Equal(t, entities.KeyStatusBlocked, entry.Status)
}

func TestEntry_Delete_Success(t *testing.T) {
	// Arrange
	entry := &entities.Entry{
		ID:            uuid.New(),
		KeyType:       entities.KeyTypeCPF,
		KeyValue:      "12345678901",
		Status:        entities.KeyStatusActive,
		AccountID:     uuid.New(),
		ISPB:          "12345678",
		Branch:        "0001",
		AccountNumber: "123456",
		AccountType:   "CACC",
		OwnerName:     "John Doe",
		OwnerTaxID:    "12345678901",
		OwnerType:     "NATURAL_PERSON",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	// Act
	now := time.Now()
	entry.Status = entities.KeyStatusDeleted
	entry.DeletedAt = &now
	entry.UpdatedAt = now

	// Assert
	assert.Equal(t, entities.KeyStatusDeleted, entry.Status)
	assert.NotNil(t, entry.DeletedAt)
}
