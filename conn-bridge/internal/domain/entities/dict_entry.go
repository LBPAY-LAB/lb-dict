package entities

import (
	"time"
)

// DictEntry represents a DICT entry entity
type DictEntry struct {
	Key           string
	Type          KeyType
	Participant   string
	Account       Account
	Owner         Owner
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Status        EntryStatus
	ClaimID       string
	DonationCause string
}

// KeyType represents the type of DICT key
type KeyType string

const (
	KeyTypeCPF   KeyType = "CPF"
	KeyTypeCNPJ  KeyType = "CNPJ"
	KeyTypePhone KeyType = "PHONE"
	KeyTypeEmail KeyType = "EMAIL"
	KeyTypeEVP   KeyType = "EVP"
)

// EntryStatus represents the status of a DICT entry
type EntryStatus string

const (
	StatusActive   EntryStatus = "ACTIVE"
	StatusInactive EntryStatus = "INACTIVE"
	StatusClaimed  EntryStatus = "CLAIMED"
	StatusDeleted  EntryStatus = "DELETED"
)

// Account represents bank account information
type Account struct {
	ISPB          string
	Branch        string
	Number        string
	Type          AccountType
	OpeningDate   time.Time
}

// AccountType represents the type of bank account
type AccountType string

const (
	AccountTypeChecking AccountType = "CHECKING"
	AccountTypeSavings  AccountType = "SAVINGS"
	AccountTypePayment  AccountType = "PAYMENT"
)

// Owner represents the owner of a DICT entry
type Owner struct {
	Type     OwnerType
	Document string
	Name     string
}

// OwnerType represents the type of owner
type OwnerType string

const (
	OwnerTypePerson OwnerType = "PERSON"
	OwnerTypeEntity OwnerType = "ENTITY"
)

// Validate validates the DICT entry
func (d *DictEntry) Validate() error {
	// TODO: Implement validation logic
	return nil
}

// IsActive returns true if the entry is active
func (d *DictEntry) IsActive() bool {
	return d.Status == StatusActive
}

// CanBeClaimed returns true if the entry can be claimed
func (d *DictEntry) CanBeClaimed() bool {
	return d.Status == StatusActive && d.ClaimID == ""
}
