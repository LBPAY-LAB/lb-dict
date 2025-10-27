package testhelpers

import (
	"time"

	"github.com/google/uuid"
)

// Test fixtures for common test data

// ValidEntry represents a valid test entry
type ValidEntry struct {
	ID        string
	KeyType   string
	KeyValue  string
	AccountID string
	ISPB      string
	Status    string
	UserID    string
}

// NewValidEntry creates a valid test entry
func NewValidEntry() ValidEntry {
	return ValidEntry{
		ID:        uuid.NewString(),
		KeyType:   "CPF",
		KeyValue:  "12345678901",
		AccountID: uuid.NewString(),
		ISPB:      "12345678",
		Status:    "ACTIVE",
		UserID:    "test-user-123",
	}
}

// NewValidCPFEntry creates a CPF entry
func NewValidCPFEntry(cpf string) ValidEntry {
	entry := NewValidEntry()
	entry.KeyType = "CPF"
	entry.KeyValue = cpf
	return entry
}

// NewValidEVPEntry creates an EVP entry
func NewValidEVPEntry() ValidEntry {
	entry := NewValidEntry()
	entry.KeyType = "EVP"
	entry.KeyValue = uuid.NewString()
	return entry
}

// ValidClaim represents a valid test claim
type ValidClaim struct {
	ID          string
	EntryID     string
	ClaimType   string
	Status      string
	DonorISPB   string
	ClaimerISPB string
	UserID      string
	ExpiresAt   time.Time
}

// NewValidClaim creates a valid test claim
func NewValidClaim(entryID string) ValidClaim {
	return ValidClaim{
		ID:          uuid.NewString(),
		EntryID:     entryID,
		ClaimType:   "OWNERSHIP",
		Status:      "OPEN",
		DonorISPB:   "12345678",
		ClaimerISPB: "87654321",
		UserID:      "test-user-123",
		ExpiresAt:   time.Now().Add(30 * 24 * time.Hour),
	}
}

// NewPortabilityClaim creates a portability claim
func NewPortabilityClaim(entryID string) ValidClaim {
	claim := NewValidClaim(entryID)
	claim.ClaimType = "PORTABILITY"
	return claim
}

// ValidAccount represents a valid test account
type ValidAccount struct {
	ID            string
	ISPB          string
	AccountNumber string
	AccountType   string
	Branch        string
	OwnerName     string
	OwnerDocument string
}

// NewValidAccount creates a valid test account
func NewValidAccount() ValidAccount {
	return ValidAccount{
		ID:            uuid.NewString(),
		ISPB:          "12345678",
		AccountNumber: "1234567890",
		AccountType:   "CACC",
		Branch:        "0001",
		OwnerName:     "Test User",
		OwnerDocument: "12345678901",
	}
}

// TestISPBs for testing
var (
	TestISPBBankA = "12345678"
	TestISPBBankB = "87654321"
	TestISPBBankC = "11111111"
)

// TestCPFs for testing
var (
	TestCPF1 = "12345678901"
	TestCPF2 = "98765432109"
	TestCPF3 = "11122233344"
	TestCPF4 = "55566677788"
	TestCPF5 = "99988877766"
	TestCPF6 = "00011122233" // For max keys test
)

// TestEmails for testing
var (
	TestEmail1 = "test1@example.com"
	TestEmail2 = "test2@example.com"
)

// TestPhones for testing
var (
	TestPhone1 = "+5511999999999"
	TestPhone2 = "+5521888888888"
)
