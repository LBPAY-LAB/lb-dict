package ratelimit

import (
	"errors"
	"fmt"
)

// Domain-specific errors for rate limit monitoring

var (
	// ErrInvalidCapacity indicates capacity is zero or negative
	ErrInvalidCapacity = errors.New("capacity must be greater than zero")

	// ErrInvalidRefillRate indicates refill configuration is invalid
	ErrInvalidRefillRate = errors.New("refill tokens and period must be greater than zero")

	// ErrNegativeTokens indicates available tokens cannot be negative
	ErrNegativeTokens = errors.New("available tokens cannot be negative")

	// ErrTokensExceedCapacity indicates available tokens exceed capacity
	ErrTokensExceedCapacity = errors.New("available tokens cannot exceed capacity")

	// ErrInvalidThreshold indicates threshold percentage is not 10 or 20
	ErrInvalidThreshold = errors.New("threshold must be 10 (CRITICAL) or 20 (WARNING)")

	// ErrInvalidSeverity indicates severity is not WARNING or CRITICAL
	ErrInvalidSeverity = errors.New("severity must be WARNING or CRITICAL")

	// ErrInvalidCategory indicates PSP category is not A-H
	ErrInvalidCategory = errors.New("PSP category must be A, B, C, D, E, F, G, or H")

	// ErrInsufficientData indicates not enough historical data for calculation
	ErrInsufficientData = errors.New("insufficient data for calculation")

	// ErrDivisionByZero indicates a calculation would result in division by zero
	ErrDivisionByZero = errors.New("division by zero in calculation")
)

// ValidationError wraps a field-specific validation error
type ValidationError struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s (value: %v)", e.Field, e.Message, e.Value)
}

// NewValidationError creates a new validation error
func NewValidationError(field string, value interface{}, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Value:   value,
		Message: message,
	}
}
