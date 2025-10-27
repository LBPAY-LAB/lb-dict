package domain

import "errors"

// Domain-level errors
var (
	// Key errors
	ErrInvalidKeyType  = errors.New("invalid key type")
	ErrInvalidKeyValue = errors.New("invalid key value")
	ErrDuplicateKey    = errors.New("duplicate key")
	ErrEntryNotFound   = errors.New("entry not found")
	ErrInvalidStatus   = errors.New("invalid status")
	ErrMaxKeysExceeded = errors.New("maximum number of keys exceeded")

	// Claim errors
	ErrInvalidClaim = errors.New("invalid claim")
	ErrClaimExpired = errors.New("claim expired")

	// Account errors
	ErrInvalidAccount = errors.New("invalid account")

	// Participant errors
	ErrInvalidParticipant = errors.New("invalid participant")

	// Authorization errors
	ErrUnauthorized = errors.New("unauthorized")
)
