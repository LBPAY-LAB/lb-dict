package entities

import (
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/google/uuid"
)

// EntryStatus represents the status of a DICT entry
type EntryStatus string

const (
	EntryStatusActive                 EntryStatus = "ACTIVE"
	EntryStatusInactive               EntryStatus = "INACTIVE"
	EntryStatusBlocked                EntryStatus = "BLOCKED"
	EntryStatusPortabilityPending     EntryStatus = "PORTABILITY_PENDING"
	EntryStatusOwnershipChangePending EntryStatus = "OWNERSHIP_CHANGE_PENDING"
)

// KeyType represents the type of PIX key
type KeyType string

const (
	KeyTypeCPF   KeyType = "CPF"
	KeyTypeCNPJ  KeyType = "CNPJ"
	KeyTypeEMAIL KeyType = "EMAIL"
	KeyTypePHONE KeyType = "PHONE"
	KeyTypeEVP   KeyType = "EVP" // Random key
)

// AccountType represents the type of bank account
type AccountType string

const (
	AccountTypeCACC AccountType = "CACC" // Current account
	AccountTypeSLRY AccountType = "SLRY" // Salary account
	AccountTypeSVGS AccountType = "SVGS" // Savings account
	AccountTypeTRAN AccountType = "TRAN" // Transaction account
)

// OwnerType represents the type of account owner
type OwnerType string

const (
	OwnerTypeNaturalPerson OwnerType = "NATURAL_PERSON"
	OwnerTypeLegalPerson   OwnerType = "LEGAL_PERSON"
)

// Entry represents a PIX key registered in DICT
type Entry struct {
	ID      uuid.UUID
	EntryID string // External entry ID

	// Key information
	Key     string
	KeyType KeyType

	// Account information
	Participant      string      // ISPB (8 digits)
	AccountBranch    *string     // Optional
	AccountNumber    *string     // Optional
	AccountType      AccountType
	AccountOpenedDate *time.Time

	// Owner information
	OwnerType   OwnerType
	OwnerName   *string
	OwnerTaxID  *string // CPF (11) or CNPJ (14)

	// Status management
	Status                   EntryStatus
	ReasonForStatusChange    *string
	BacenEntryID             *string

	// Timestamps
	RegisteredAt   *time.Time
	ActivatedAt    *time.Time
	DeactivatedAt  *time.Time

	// Audit
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// NewEntry creates a new Entry with validation
func NewEntry(
	entryID, key string,
	keyType KeyType,
	participant string,
	accountType AccountType,
	ownerType OwnerType,
) (*Entry, error) {
	// Validate ISPB
	if !isValidISPB(participant) {
		return nil, fmt.Errorf("invalid ISPB: must be 8 digits, got %s", participant)
	}

	// Validate key format based on type
	if err := validateKeyFormat(key, keyType); err != nil {
		return nil, fmt.Errorf("invalid key format: %w", err)
	}

	now := time.Now()

	entry := &Entry{
		ID:           uuid.New(),
		EntryID:      entryID,
		Key:          key,
		KeyType:      keyType,
		Participant:  participant,
		AccountType:  accountType,
		OwnerType:    ownerType,
		Status:       EntryStatusActive,
		RegisteredAt: &now,
		ActivatedAt:  &now,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	return entry, nil
}

// Activate activates the entry
func (e *Entry) Activate() error {
	if e.Status == EntryStatusActive {
		return errors.New("entry is already active")
	}

	if e.Status == EntryStatusBlocked {
		return errors.New("cannot activate blocked entry - unblock first")
	}

	now := time.Now()
	e.Status = EntryStatusActive
	e.ActivatedAt = &now
	e.DeactivatedAt = nil
	e.UpdatedAt = now

	return nil
}

// Deactivate deactivates the entry
func (e *Entry) Deactivate(reason string) error {
	if e.Status == EntryStatusInactive {
		return errors.New("entry is already inactive")
	}

	if e.Status == EntryStatusBlocked {
		return errors.New("cannot deactivate blocked entry - unblock first")
	}

	now := time.Now()
	e.Status = EntryStatusInactive
	e.DeactivatedAt = &now
	e.ReasonForStatusChange = &reason
	e.UpdatedAt = now

	return nil
}

// Block blocks the entry
func (e *Entry) Block(reason string) error {
	if e.Status == EntryStatusBlocked {
		return errors.New("entry is already blocked")
	}

	now := time.Now()
	e.Status = EntryStatusBlocked
	e.ReasonForStatusChange = &reason
	e.UpdatedAt = now

	return nil
}

// Unblock unblocks the entry and activates it
func (e *Entry) Unblock() error {
	if e.Status != EntryStatusBlocked {
		return errors.New("entry is not blocked")
	}

	now := time.Now()
	e.Status = EntryStatusActive
	e.ReasonForStatusChange = nil
	e.UpdatedAt = now

	return nil
}

// SetPortabilityPending marks entry as having a portability claim
func (e *Entry) SetPortabilityPending() error {
	if e.Status == EntryStatusPortabilityPending {
		return errors.New("entry already has portability pending")
	}

	if e.Status == EntryStatusBlocked {
		return errors.New("cannot process portability for blocked entry")
	}

	now := time.Now()
	e.Status = EntryStatusPortabilityPending
	e.UpdatedAt = now

	return nil
}

// SetOwnershipChangePending marks entry as having an ownership claim
func (e *Entry) SetOwnershipChangePending() error {
	if e.Status == EntryStatusOwnershipChangePending {
		return errors.New("entry already has ownership change pending")
	}

	if e.Status == EntryStatusBlocked {
		return errors.New("cannot process ownership change for blocked entry")
	}

	now := time.Now()
	e.Status = EntryStatusOwnershipChangePending
	e.UpdatedAt = now

	return nil
}

// UpdateOwnership updates the owner information
func (e *Entry) UpdateOwnership(participant string, ownerName, ownerTaxID string) error {
	if !isValidISPB(participant) {
		return fmt.Errorf("invalid ISPB: must be 8 digits, got %s", participant)
	}

	if len(ownerTaxID) != 11 && len(ownerTaxID) != 14 {
		return fmt.Errorf("invalid tax ID length: must be 11 (CPF) or 14 (CNPJ), got %d", len(ownerTaxID))
	}

	e.Participant = participant
	e.OwnerName = &ownerName
	e.OwnerTaxID = &ownerTaxID
	e.UpdatedAt = time.Now()

	return nil
}

// IsActive checks if entry is active
func (e *Entry) IsActive() bool {
	return e.Status == EntryStatusActive
}

// IsBlocked checks if entry is blocked
func (e *Entry) IsBlocked() bool {
	return e.Status == EntryStatusBlocked
}

// HasPendingClaim checks if entry has a pending claim
func (e *Entry) HasPendingClaim() bool {
	return e.Status == EntryStatusPortabilityPending || e.Status == EntryStatusOwnershipChangePending
}

// validateKeyFormat validates the key format based on key type
func validateKeyFormat(key string, keyType KeyType) error {
	switch keyType {
	case KeyTypeCPF:
		// CPF: 11 digits
		if !regexp.MustCompile(`^\d{11}$`).MatchString(key) {
			return errors.New("CPF must be 11 digits")
		}
	case KeyTypeCNPJ:
		// CNPJ: 14 digits
		if !regexp.MustCompile(`^\d{14}$`).MatchString(key) {
			return errors.New("CNPJ must be 14 digits")
		}
	case KeyTypeEMAIL:
		// Email: basic validation
		if !regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`).MatchString(key) {
			return errors.New("invalid email format")
		}
	case KeyTypePHONE:
		// Phone: +55DDNNNNNNNN (Brazilian format)
		if !regexp.MustCompile(`^\+55\d{10,11}$`).MatchString(key) {
			return errors.New("phone must be in format +55DDNNNNNNNN")
		}
	case KeyTypeEVP:
		// EVP: UUID format
		if _, err := uuid.Parse(key); err != nil {
			return errors.New("EVP must be a valid UUID")
		}
	default:
		return fmt.Errorf("unknown key type: %s", keyType)
	}

	return nil
}
