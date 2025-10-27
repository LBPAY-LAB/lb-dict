package entities

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDictEntry_Validate(t *testing.T) {
	tests := []struct {
		name    string
		entry   *DictEntry
		wantErr bool
	}{
		{
			name: "valid CPF entry",
			entry: &DictEntry{
				Key:         "12345678901",
				Type:        KeyTypeCPF,
				Participant: "60701190",
				Account: Account{
					ISPB:        "60701190",
					Branch:      "0001",
					Number:      "123456",
					Type:        AccountTypeChecking,
					OpeningDate: time.Now(),
				},
				Owner: Owner{
					Type:     OwnerTypePerson,
					Document: "12345678901",
					Name:     "John Doe",
				},
				Status:    StatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
		{
			name: "valid email entry",
			entry: &DictEntry{
				Key:         "test@example.com",
				Type:        KeyTypeEmail,
				Participant: "60701190",
				Account: Account{
					ISPB:        "60701190",
					Branch:      "0001",
					Number:      "123456",
					Type:        AccountTypeSavings,
					OpeningDate: time.Now(),
				},
				Owner: Owner{
					Type:     OwnerTypePerson,
					Document: "12345678901",
					Name:     "Jane Doe",
				},
				Status:    StatusActive,
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.entry.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDictEntry_IsActive(t *testing.T) {
	tests := []struct {
		name   string
		status EntryStatus
		want   bool
	}{
		{
			name:   "active status",
			status: StatusActive,
			want:   true,
		},
		{
			name:   "inactive status",
			status: StatusInactive,
			want:   false,
		},
		{
			name:   "claimed status",
			status: StatusClaimed,
			want:   false,
		},
		{
			name:   "deleted status",
			status: StatusDeleted,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &DictEntry{
				Status: tt.status,
			}
			got := entry.IsActive()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestDictEntry_CanBeClaimed(t *testing.T) {
	tests := []struct {
		name    string
		status  EntryStatus
		claimID string
		want    bool
	}{
		{
			name:    "active with no claim",
			status:  StatusActive,
			claimID: "",
			want:    true,
		},
		{
			name:    "active with existing claim",
			status:  StatusActive,
			claimID: "claim-123",
			want:    false,
		},
		{
			name:    "inactive with no claim",
			status:  StatusInactive,
			claimID: "",
			want:    false,
		},
		{
			name:    "claimed status",
			status:  StatusClaimed,
			claimID: "claim-123",
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := &DictEntry{
				Status:  tt.status,
				ClaimID: tt.claimID,
			}
			got := entry.CanBeClaimed()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestKeyType_Constants(t *testing.T) {
	// Test that key type constants are correctly defined
	assert.Equal(t, KeyType("CPF"), KeyTypeCPF)
	assert.Equal(t, KeyType("CNPJ"), KeyTypeCNPJ)
	assert.Equal(t, KeyType("PHONE"), KeyTypePhone)
	assert.Equal(t, KeyType("EMAIL"), KeyTypeEmail)
	assert.Equal(t, KeyType("EVP"), KeyTypeEVP)
}

func TestEntryStatus_Constants(t *testing.T) {
	// Test that status constants are correctly defined
	assert.Equal(t, EntryStatus("ACTIVE"), StatusActive)
	assert.Equal(t, EntryStatus("INACTIVE"), StatusInactive)
	assert.Equal(t, EntryStatus("CLAIMED"), StatusClaimed)
	assert.Equal(t, EntryStatus("DELETED"), StatusDeleted)
}

func TestAccountType_Constants(t *testing.T) {
	// Test that account type constants are correctly defined
	assert.Equal(t, AccountType("CHECKING"), AccountTypeChecking)
	assert.Equal(t, AccountType("SAVINGS"), AccountTypeSavings)
	assert.Equal(t, AccountType("PAYMENT"), AccountTypePayment)
}

func TestOwnerType_Constants(t *testing.T) {
	// Test that owner type constants are correctly defined
	assert.Equal(t, OwnerType("PERSON"), OwnerTypePerson)
	assert.Equal(t, OwnerType("ENTITY"), OwnerTypeEntity)
}

func TestDictEntry_CompleteFlow(t *testing.T) {
	// Test a complete flow of entry lifecycle
	entry := &DictEntry{
		Key:         "test@example.com",
		Type:        KeyTypeEmail,
		Participant: "60701190",
		Account: Account{
			ISPB:        "60701190",
			Branch:      "0001",
			Number:      "123456",
			Type:        AccountTypeChecking,
			OpeningDate: time.Now(),
		},
		Owner: Owner{
			Type:     OwnerTypePerson,
			Document: "12345678901",
			Name:     "Test User",
		},
		Status:    StatusActive,
		ClaimID:   "",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Should be active and claimable
	require.True(t, entry.IsActive())
	require.True(t, entry.CanBeClaimed())

	// Simulate claim
	entry.ClaimID = "claim-456"
	entry.Status = StatusClaimed

	// Should not be active or claimable
	assert.False(t, entry.IsActive())
	assert.False(t, entry.CanBeClaimed())
}
