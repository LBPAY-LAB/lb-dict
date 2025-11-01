package ratelimit

import (
	"time"
)

// Policy represents a DICT API rate limit policy configuration
// This is static reference data retrieved from DICT BACEN
type Policy struct {
	// EndpointID is the unique identifier for this rate-limited endpoint
	EndpointID string

	// EndpointPath is the HTTP path (e.g., "/api/v1/entries")
	EndpointPath string

	// HTTPMethod is the HTTP method (GET, POST, PUT, DELETE, PATCH)
	HTTPMethod string

	// Capacity is the maximum number of tokens in the bucket
	Capacity int

	// RefillTokens is the number of tokens added per refill period
	RefillTokens int

	// RefillPeriodSec is the refill period in seconds
	RefillPeriodSec int

	// PSPCategory is the PSP category (A-H) if endpoint-specific limits exist
	// May be empty if policy applies to all categories
	PSPCategory string

	// CreatedAt is when this policy was first stored
	CreatedAt time.Time

	// UpdatedAt is when this policy was last updated
	UpdatedAt time.Time
}

// NewPolicy creates a new Policy with validation
func NewPolicy(
	endpointID string,
	endpointPath string,
	httpMethod string,
	capacity int,
	refillTokens int,
	refillPeriodSec int,
	pspCategory string,
) (*Policy, error) {
	if capacity <= 0 {
		return nil, ErrInvalidCapacity
	}

	if refillTokens <= 0 || refillPeriodSec <= 0 {
		return nil, ErrInvalidRefillRate
	}

	if pspCategory != "" {
		if err := validatePSPCategory(pspCategory); err != nil {
			return nil, err
		}
	}

	now := time.Now().UTC()

	return &Policy{
		EndpointID:      endpointID,
		EndpointPath:    endpointPath,
		HTTPMethod:      httpMethod,
		Capacity:        capacity,
		RefillTokens:    refillTokens,
		RefillPeriodSec: refillPeriodSec,
		PSPCategory:     pspCategory,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

// GetRefillRate returns tokens per second refill rate
func (p *Policy) GetRefillRate() float64 {
	if p.RefillPeriodSec == 0 {
		return 0
	}
	return float64(p.RefillTokens) / float64(p.RefillPeriodSec)
}

// Validate performs validation on Policy fields
func (p *Policy) Validate() error {
	if p.Capacity <= 0 {
		return NewValidationError("Capacity", p.Capacity, "must be greater than zero")
	}

	if p.RefillTokens <= 0 {
		return NewValidationError("RefillTokens", p.RefillTokens, "must be greater than zero")
	}

	if p.RefillPeriodSec <= 0 {
		return NewValidationError("RefillPeriodSec", p.RefillPeriodSec, "must be greater than zero")
	}

	if p.PSPCategory != "" {
		if err := validatePSPCategory(p.PSPCategory); err != nil {
			return err
		}
	}

	return nil
}

// validatePSPCategory checks if category is valid (A-H)
func validatePSPCategory(category string) error {
	validCategories := map[string]bool{
		"A": true, "B": true, "C": true, "D": true,
		"E": true, "F": true, "G": true, "H": true,
	}

	if !validCategories[category] {
		return ErrInvalidCategory
	}

	return nil
}
