package entities_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lbpay-lab/core-dict/internal/domain/entities"
)

func TestNewAccount_Success(t *testing.T) {
	// Arrange
	ispb := "12345678"
	branch := "0001"
	accountNumber := "123456"
	accountType := entities.AccountTypeCACC
	owner := entities.Owner{
		TaxID: "12345678901",
		Type:  entities.OwnerTypeNaturalPerson,
		Name:  "John Doe",
	}

	// Act
	account, err := entities.NewAccount(ispb, branch, accountNumber, accountType, owner)

	// Assert
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, account.ID)
	assert.Equal(t, ispb, account.ISPB)
	assert.Equal(t, branch, account.Branch)
	assert.Equal(t, accountNumber, account.AccountNumber)
	assert.Equal(t, accountType, account.AccountType)
	assert.Equal(t, "ACTIVE", account.Status)
	assert.Equal(t, owner.TaxID, account.Owner.TaxID)
}

func TestAccount_Validate_Success(t *testing.T) {
	// Arrange
	owner := entities.Owner{
		TaxID: "12345678901",
		Type:  entities.OwnerTypeNaturalPerson,
		Name:  "John Doe",
	}
	account, _ := entities.NewAccount("12345678", "0001", "123456", entities.AccountTypeCACC, owner)

	// Act
	err := account.Validate()

	// Assert
	require.NoError(t, err)
}

func TestAccount_UpdateStatus(t *testing.T) {
	// Arrange
	owner := entities.Owner{
		TaxID: "12345678901",
		Type:  entities.OwnerTypeNaturalPerson,
		Name:  "John Doe",
	}
	account, _ := entities.NewAccount("12345678", "0001", "123456", entities.AccountTypeCACC, owner)

	tests := []struct {
		name          string
		initialStatus string
		newStatus     string
		shouldBlock   bool
		shouldClose   bool
		wantErr       bool
	}{
		{
			name:          "Block active account",
			initialStatus: "ACTIVE",
			shouldBlock:   true,
			wantErr:       false,
		},
		{
			name:          "Close active account",
			initialStatus: "ACTIVE",
			shouldClose:   true,
			wantErr:       false,
		},
		{
			name:          "Unblock blocked account",
			initialStatus: "BLOCKED",
			shouldBlock:   false,
			wantErr:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			account.Status = tt.initialStatus
			account.ClosedAt = nil

			// Act
			var err error
			if tt.shouldBlock {
				err = account.Block()
			} else if tt.shouldClose {
				err = account.Close()
			} else {
				err = account.Unblock()
			}

			// Assert
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestAccount_IsClosed(t *testing.T) {
	// Arrange
	owner := entities.Owner{
		TaxID: "12345678901",
		Type:  entities.OwnerTypeNaturalPerson,
		Name:  "John Doe",
	}
	account, _ := entities.NewAccount("12345678", "0001", "123456", entities.AccountTypeCACC, owner)

	// Test 1: Active account is not closed
	t.Run("Active account is not closed", func(t *testing.T) {
		assert.False(t, account.IsClosed())
	})

	// Test 2: Closed account is closed
	t.Run("Closed account is closed", func(t *testing.T) {
		now := time.Now()
		account.ClosedAt = &now
		account.Status = "CLOSED"
		assert.True(t, account.IsClosed())
	})

	// Test 3: Account with CLOSED status is closed
	t.Run("Account with CLOSED status", func(t *testing.T) {
		account2, _ := entities.NewAccount("12345678", "0001", "123456", entities.AccountTypeCACC, owner)
		account2.Status = "CLOSED"
		assert.True(t, account2.IsClosed())
	})
}
