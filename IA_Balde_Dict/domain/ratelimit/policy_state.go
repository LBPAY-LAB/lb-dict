package ratelimit

import (
	"time"
)

// PolicyState represents a snapshot of rate limit token bucket state at a point in time
// This is time-series data collected every 5 minutes from DICT
type PolicyState struct {
	// ID is the database primary key
	ID int64

	// EndpointID references the policy
	EndpointID string

	// Token bucket state at snapshot time
	AvailableTokens int
	Capacity        int
	RefillTokens    int
	RefillPeriodSec int

	// PSPCategory at snapshot time (may change over time)
	PSPCategory string

	// Calculated metrics (computed by domain calculators)
	ConsumptionRatePerMinute      float64 // tokens/min consumption rate
	RecoveryETASeconds            int     // estimated seconds to full recovery
	ExhaustionProjectionSeconds   int     // estimated seconds to exhaustion
	Error404Rate                  float64 // percentage of 404 errors

	// ResponseTimestamp is the authoritative timestamp from DICT <ResponseTime>
	ResponseTimestamp time.Time

	// CreatedAt is when we stored this snapshot (partition key)
	CreatedAt time.Time
}

// NewPolicyState creates a new PolicyState with validation
func NewPolicyState(
	endpointID string,
	availableTokens int,
	capacity int,
	refillTokens int,
	refillPeriodSec int,
	pspCategory string,
	responseTimestamp time.Time,
) (*PolicyState, error) {
	if availableTokens < 0 {
		return nil, ErrNegativeTokens
	}

	if availableTokens > capacity {
		return nil, ErrTokensExceedCapacity
	}

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

	return &PolicyState{
		EndpointID:        endpointID,
		AvailableTokens:   availableTokens,
		Capacity:          capacity,
		RefillTokens:      refillTokens,
		RefillPeriodSec:   refillPeriodSec,
		PSPCategory:       pspCategory,
		ResponseTimestamp: responseTimestamp.UTC(),
		CreatedAt:         time.Now().UTC(),
	}, nil
}

// GetUtilizationPercent returns the percentage of capacity utilized (100% - remaining%)
func (s *PolicyState) GetUtilizationPercent() float64 {
	if s.Capacity == 0 {
		return 0
	}
	return 100.0 - (float64(s.AvailableTokens) / float64(s.Capacity) * 100.0)
}

// GetRemainingPercent returns the percentage of tokens remaining
func (s *PolicyState) GetRemainingPercent() float64 {
	if s.Capacity == 0 {
		return 0
	}
	return float64(s.AvailableTokens) / float64(s.Capacity) * 100.0
}

// GetRefillRate returns tokens per second refill rate
func (s *PolicyState) GetRefillRate() float64 {
	if s.RefillPeriodSec == 0 {
		return 0
	}
	return float64(s.RefillTokens) / float64(s.RefillPeriodSec)
}

// Validate performs validation on PolicyState fields
func (s *PolicyState) Validate() error {
	if s.AvailableTokens < 0 {
		return NewValidationError("AvailableTokens", s.AvailableTokens, "cannot be negative")
	}

	if s.AvailableTokens > s.Capacity {
		return NewValidationError("AvailableTokens", s.AvailableTokens, "cannot exceed capacity")
	}

	if s.Capacity <= 0 {
		return NewValidationError("Capacity", s.Capacity, "must be greater than zero")
	}

	if s.RefillTokens <= 0 {
		return NewValidationError("RefillTokens", s.RefillTokens, "must be greater than zero")
	}

	if s.RefillPeriodSec <= 0 {
		return NewValidationError("RefillPeriodSec", s.RefillPeriodSec, "must be greater than zero")
	}

	if s.PSPCategory != "" {
		if err := validatePSPCategory(s.PSPCategory); err != nil {
			return err
		}
	}

	return nil
}
