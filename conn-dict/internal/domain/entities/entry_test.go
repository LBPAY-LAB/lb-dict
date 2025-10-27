package entities

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// =============================================================================
// Constructor Tests
// =============================================================================

func TestNewEntry_Success(t *testing.T) {
	// Arrange
	entryID := "ENTRY-001"
	key := "12345678901"
	keyType := KeyTypeCPF
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, entryID, entry.EntryID)
	assert.Equal(t, key, entry.Key)
	assert.Equal(t, keyType, entry.KeyType)
	assert.Equal(t, ispb, entry.Participant)
	assert.Equal(t, AccountTypeCACC, entry.AccountType)
	assert.Equal(t, OwnerTypeNaturalPerson, entry.OwnerType)
	assert.Equal(t, EntryStatusActive, entry.Status)
	assert.NotNil(t, entry.RegisteredAt)
	assert.NotNil(t, entry.ActivatedAt)
	assert.NotEqual(t, uuid.Nil, entry.ID)
}

func TestNewEntry_InvalidISPB_TooShort(t *testing.T) {
	// Arrange
	entryID := "ENTRY-002"
	key := "12345678901"
	keyType := KeyTypeCPF
	ispb := "1234567" // Only 7 digits

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "invalid ISPB")
}

func TestNewEntry_InvalidISPB_TooLong(t *testing.T) {
	// Arrange
	entryID := "ENTRY-003"
	key := "12345678901"
	keyType := KeyTypeCPF
	ispb := "123456789" // 9 digits

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "invalid ISPB")
}

func TestNewEntry_InvalidISPB_NonNumeric(t *testing.T) {
	// Arrange
	entryID := "ENTRY-004"
	key := "12345678901"
	keyType := KeyTypeCPF
	ispb := "1234567A" // Contains letter

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "invalid ISPB")
}

func TestNewEntry_InvalidKeyFormat_CPF_TooShort(t *testing.T) {
	// Arrange
	entryID := "ENTRY-005"
	key := "1234567890" // Only 10 digits
	keyType := KeyTypeCPF
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "CPF must be 11 digits")
}

func TestNewEntry_InvalidKeyFormat_CPF_TooLong(t *testing.T) {
	// Arrange
	entryID := "ENTRY-006"
	key := "123456789012" // 12 digits
	keyType := KeyTypeCPF
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "CPF must be 11 digits")
}

func TestNewEntry_InvalidKeyFormat_CNPJ_TooShort(t *testing.T) {
	// Arrange
	entryID := "ENTRY-007"
	key := "1234567890123" // Only 13 digits
	keyType := KeyTypeCNPJ
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeLegalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "CNPJ must be 14 digits")
}

func TestNewEntry_InvalidKeyFormat_CNPJ_TooLong(t *testing.T) {
	// Arrange
	entryID := "ENTRY-008"
	key := "123456789012345" // 15 digits
	keyType := KeyTypeCNPJ
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeLegalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "CNPJ must be 14 digits")
}

func TestNewEntry_InvalidKeyFormat_Email_MissingAt(t *testing.T) {
	// Arrange
	entryID := "ENTRY-009"
	key := "user.example.com" // Missing @
	keyType := KeyTypeEMAIL
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestNewEntry_InvalidKeyFormat_Email_MissingDomain(t *testing.T) {
	// Arrange
	entryID := "ENTRY-010"
	key := "user@" // Missing domain
	keyType := KeyTypeEMAIL
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "invalid email format")
}

func TestNewEntry_ValidKeyFormat_Email(t *testing.T) {
	// Arrange
	entryID := "ENTRY-011"
	key := "user@example.com"
	keyType := KeyTypeEMAIL
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, key, entry.Key)
}

func TestNewEntry_InvalidKeyFormat_Phone_MissingCountryCode(t *testing.T) {
	// Arrange
	entryID := "ENTRY-012"
	key := "11987654321" // Missing +55
	keyType := KeyTypePHONE
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "phone must be in format +55DDNNNNNNNN")
}

func TestNewEntry_InvalidKeyFormat_Phone_TooShort(t *testing.T) {
	// Arrange
	entryID := "ENTRY-013"
	key := "+5511987654" // Too short
	keyType := KeyTypePHONE
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "phone must be in format +55DDNNNNNNNN")
}

func TestNewEntry_ValidKeyFormat_Phone(t *testing.T) {
	// Arrange
	entryID := "ENTRY-014"
	key := "+5511987654321" // Valid 11 digits after +55
	keyType := KeyTypePHONE
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, key, entry.Key)
}

func TestNewEntry_InvalidKeyFormat_EVP_NotUUID(t *testing.T) {
	// Arrange
	entryID := "ENTRY-015"
	key := "not-a-uuid-123456" // Not a valid UUID
	keyType := KeyTypeEVP
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, entry)
	assert.Contains(t, err.Error(), "EVP must be a valid UUID")
}

func TestNewEntry_ValidKeyFormat_EVP(t *testing.T) {
	// Arrange
	entryID := "ENTRY-016"
	key := uuid.New().String() // Valid UUID
	keyType := KeyTypeEVP
	ispb := "12345678"

	// Act
	entry, err := NewEntry(entryID, key, keyType, ispb, AccountTypeCACC, OwnerTypeNaturalPerson)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	assert.Equal(t, key, entry.Key)
}

// =============================================================================
// Activation Tests
// =============================================================================

func TestEntry_Activate_Success(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusInactive
	entry.ActivatedAt = nil

	// Act
	err := entry.Activate()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusActive, entry.Status)
	assert.NotNil(t, entry.ActivatedAt)
	assert.Nil(t, entry.DeactivatedAt)
}

func TestEntry_Activate_AlreadyActive(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive

	// Act
	err := entry.Activate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already active")
}

func TestEntry_Activate_FromBlocked(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusBlocked

	// Act
	err := entry.Activate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot activate blocked entry")
	assert.Equal(t, EntryStatusBlocked, entry.Status)
}

// =============================================================================
// Deactivation Tests
// =============================================================================

func TestEntry_Deactivate_Success(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive
	reason := "User requested deactivation"

	// Act
	err := entry.Deactivate(reason)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusInactive, entry.Status)
	assert.NotNil(t, entry.DeactivatedAt)
	assert.NotNil(t, entry.ReasonForStatusChange)
	assert.Equal(t, reason, *entry.ReasonForStatusChange)
}

func TestEntry_Deactivate_AlreadyInactive(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusInactive
	reason := "User requested deactivation"

	// Act
	err := entry.Deactivate(reason)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already inactive")
}

func TestEntry_Deactivate_FromBlocked(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusBlocked
	reason := "User requested deactivation"

	// Act
	err := entry.Deactivate(reason)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot deactivate blocked entry")
	assert.Equal(t, EntryStatusBlocked, entry.Status)
}

func TestEntry_Deactivate_WithEmptyReason(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive
	reason := "" // Empty reason (allowed by implementation)

	// Act
	err := entry.Deactivate(reason)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusInactive, entry.Status)
}

// =============================================================================
// Block Tests
// =============================================================================

func TestEntry_Block_Success(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive
	reason := "Fraudulent activity detected"

	// Act
	err := entry.Block(reason)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusBlocked, entry.Status)
	assert.NotNil(t, entry.ReasonForStatusChange)
	assert.Equal(t, reason, *entry.ReasonForStatusChange)
}

func TestEntry_Block_AlreadyBlocked(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusBlocked
	reason := "Fraudulent activity detected"

	// Act
	err := entry.Block(reason)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already blocked")
}

func TestEntry_Block_FromInactive(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusInactive
	reason := "Fraudulent activity detected"

	// Act
	err := entry.Block(reason)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusBlocked, entry.Status)
}

// =============================================================================
// Unblock Tests
// =============================================================================

func TestEntry_Unblock_Success(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusBlocked
	reason := "Fraud investigation complete"
	entry.ReasonForStatusChange = &reason

	// Act
	err := entry.Unblock()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusActive, entry.Status)
	assert.Nil(t, entry.ReasonForStatusChange)
}

func TestEntry_Unblock_NotBlocked(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive

	// Act
	err := entry.Unblock()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not blocked")
}

func TestEntry_Unblock_FromInactive(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusInactive

	// Act
	err := entry.Unblock()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not blocked")
	assert.Equal(t, EntryStatusInactive, entry.Status)
}

// =============================================================================
// Claim Status Tests
// =============================================================================

func TestEntry_SetPortabilityPending_Success(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive

	// Act
	err := entry.SetPortabilityPending()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusPortabilityPending, entry.Status)
}

func TestEntry_SetPortabilityPending_AlreadyPending(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusPortabilityPending

	// Act
	err := entry.SetPortabilityPending()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already has portability pending")
}

func TestEntry_SetPortabilityPending_Blocked(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusBlocked

	// Act
	err := entry.SetPortabilityPending()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot process portability for blocked entry")
	assert.Equal(t, EntryStatusBlocked, entry.Status)
}

func TestEntry_SetOwnershipChangePending_Success(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive

	// Act
	err := entry.SetOwnershipChangePending()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusOwnershipChangePending, entry.Status)
}

func TestEntry_SetOwnershipChangePending_AlreadyPending(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusOwnershipChangePending

	// Act
	err := entry.SetOwnershipChangePending()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already has ownership change pending")
}

func TestEntry_SetOwnershipChangePending_Blocked(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusBlocked

	// Act
	err := entry.SetOwnershipChangePending()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot process ownership change for blocked entry")
	assert.Equal(t, EntryStatusBlocked, entry.Status)
}

// =============================================================================
// Ownership Update Tests
// =============================================================================

func TestEntry_UpdateOwnership_Success(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	newISPB := "87654321"
	newOwnerName := "New Owner Name"
	newOwnerTaxID := "98765432100" // CPF (11 digits)

	// Act
	err := entry.UpdateOwnership(newISPB, newOwnerName, newOwnerTaxID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, newISPB, entry.Participant)
	assert.NotNil(t, entry.OwnerName)
	assert.Equal(t, newOwnerName, *entry.OwnerName)
	assert.NotNil(t, entry.OwnerTaxID)
	assert.Equal(t, newOwnerTaxID, *entry.OwnerTaxID)
}

func TestEntry_UpdateOwnership_InvalidISPB(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	newISPB := "1234567" // Invalid (7 digits)
	newOwnerName := "New Owner Name"
	newOwnerTaxID := "98765432100"

	// Act
	err := entry.UpdateOwnership(newISPB, newOwnerName, newOwnerTaxID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid ISPB")
}

func TestEntry_UpdateOwnership_InvalidTaxIDLength_TooShort(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	newISPB := "87654321"
	newOwnerName := "New Owner Name"
	newOwnerTaxID := "9876543210" // Only 10 digits

	// Act
	err := entry.UpdateOwnership(newISPB, newOwnerName, newOwnerTaxID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tax ID length")
}

func TestEntry_UpdateOwnership_InvalidTaxIDLength_TooLong(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	newISPB := "87654321"
	newOwnerName := "New Owner Name"
	newOwnerTaxID := "987654321012345" // 15 digits

	// Act
	err := entry.UpdateOwnership(newISPB, newOwnerName, newOwnerTaxID)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid tax ID length")
}

func TestEntry_UpdateOwnership_ValidCNPJ(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	newISPB := "87654321"
	newOwnerName := "New Company Name"
	newOwnerTaxID := "12345678901234" // CNPJ (14 digits)

	// Act
	err := entry.UpdateOwnership(newISPB, newOwnerName, newOwnerTaxID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, newISPB, entry.Participant)
	assert.Equal(t, newOwnerTaxID, *entry.OwnerTaxID)
}

// =============================================================================
// Helper Methods Tests
// =============================================================================

func TestEntry_IsActive_True(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive

	// Act
	result := entry.IsActive()

	// Assert
	assert.True(t, result)
}

func TestEntry_IsActive_False_Inactive(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusInactive

	// Act
	result := entry.IsActive()

	// Assert
	assert.False(t, result)
}

func TestEntry_IsActive_False_Blocked(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusBlocked

	// Act
	result := entry.IsActive()

	// Assert
	assert.False(t, result)
}

func TestEntry_IsBlocked_True(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusBlocked

	// Act
	result := entry.IsBlocked()

	// Assert
	assert.True(t, result)
}

func TestEntry_IsBlocked_False(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive

	// Act
	result := entry.IsBlocked()

	// Assert
	assert.False(t, result)
}

func TestEntry_HasPendingClaim_True_Portability(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusPortabilityPending

	// Act
	result := entry.HasPendingClaim()

	// Assert
	assert.True(t, result)
}

func TestEntry_HasPendingClaim_True_OwnershipChange(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusOwnershipChangePending

	// Act
	result := entry.HasPendingClaim()

	// Assert
	assert.True(t, result)
}

func TestEntry_HasPendingClaim_False_Active(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusActive

	// Act
	result := entry.HasPendingClaim()

	// Assert
	assert.False(t, result)
}

func TestEntry_HasPendingClaim_False_Blocked(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	entry.Status = EntryStatusBlocked

	// Act
	result := entry.HasPendingClaim()

	// Assert
	assert.False(t, result)
}

// =============================================================================
// Edge Cases and Additional Tests
// =============================================================================

func TestEntry_MultipleStatusTransitions(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)

	// Act & Assert - Active -> Portability Pending
	err := entry.SetPortabilityPending()
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusPortabilityPending, entry.Status)

	// Cannot block while portability pending (need to check actual behavior)
	err = entry.Block("Fraud detected")
	assert.NoError(t, err) // Block can happen from any non-blocked state
	assert.Equal(t, EntryStatusBlocked, entry.Status)

	// Unblock returns to Active
	err = entry.Unblock()
	assert.NoError(t, err)
	assert.Equal(t, EntryStatusActive, entry.Status)
}

func TestEntry_TimestampUpdates(t *testing.T) {
	// Arrange
	entry := createTestEntry(t)
	originalUpdatedAt := entry.UpdatedAt

	// Wait a tiny bit to ensure timestamp difference
	time.Sleep(1 * time.Millisecond)

	// Act - Perform an update
	err := entry.Block("Testing timestamp update")

	// Assert
	assert.NoError(t, err)
	assert.True(t, entry.UpdatedAt.After(originalUpdatedAt))
}

func TestEntry_AllAccountTypes(t *testing.T) {
	accountTypes := []AccountType{
		AccountTypeCACC,
		AccountTypeSLRY,
		AccountTypeSVGS,
		AccountTypeTRAN,
	}

	for _, accountType := range accountTypes {
		// Act
		entry, err := NewEntry(
			"ENTRY-TEST",
			"12345678901",
			KeyTypeCPF,
			"12345678",
			accountType,
			OwnerTypeNaturalPerson,
		)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, entry)
		assert.Equal(t, accountType, entry.AccountType)
	}
}

func TestEntry_AllKeyTypes(t *testing.T) {
	testCases := []struct {
		name    string
		key     string
		keyType KeyType
	}{
		{"CPF", "12345678901", KeyTypeCPF},
		{"CNPJ", "12345678901234", KeyTypeCNPJ},
		{"Email", "test@example.com", KeyTypeEMAIL},
		{"Phone", "+5511987654321", KeyTypePHONE},
		{"EVP", uuid.New().String(), KeyTypeEVP},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Act
			entry, err := NewEntry(
				"ENTRY-TEST",
				tc.key,
				tc.keyType,
				"12345678",
				AccountTypeCACC,
				OwnerTypeNaturalPerson,
			)

			// Assert
			assert.NoError(t, err)
			assert.NotNil(t, entry)
			assert.Equal(t, tc.keyType, entry.KeyType)
			assert.Equal(t, tc.key, entry.Key)
		})
	}
}

// =============================================================================
// Test Helpers
// =============================================================================

// createTestEntry creates a valid entry for testing
func createTestEntry(t *testing.T) *Entry {
	entry, err := NewEntry(
		"ENTRY-TEST-001",
		"12345678901",
		KeyTypeCPF,
		"12345678",
		AccountTypeCACC,
		OwnerTypeNaturalPerson,
	)
	assert.NoError(t, err)
	assert.NotNil(t, entry)
	return entry
}