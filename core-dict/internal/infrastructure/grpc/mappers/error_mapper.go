package mappers

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/lbpay-lab/core-dict/internal/domain"
)

// ============================================================================
// Domain Errors → gRPC Status Codes
// ============================================================================

// MapDomainErrorToGRPC maps domain errors to appropriate gRPC status codes
func MapDomainErrorToGRPC(err error) error {
	if err == nil {
		return nil
	}

	// Map specific domain errors
	switch {
	// Validation Errors → InvalidArgument
	case errors.Is(err, domain.ErrInvalidKeyType):
		return status.Error(codes.InvalidArgument, "Invalid key type: "+err.Error())
	case errors.Is(err, domain.ErrInvalidKeyValue):
		return status.Error(codes.InvalidArgument, "Invalid key value: "+err.Error())
	case errors.Is(err, domain.ErrInvalidStatus):
		return status.Error(codes.InvalidArgument, "Invalid status: "+err.Error())
	case errors.Is(err, domain.ErrInvalidAccount):
		return status.Error(codes.InvalidArgument, "Invalid account: "+err.Error())
	case errors.Is(err, domain.ErrInvalidClaim):
		return status.Error(codes.InvalidArgument, "Invalid claim: "+err.Error())
	case errors.Is(err, domain.ErrInvalidParticipant):
		return status.Error(codes.InvalidArgument, "Invalid participant: "+err.Error())

	// Not Found Errors → NotFound
	case errors.Is(err, domain.ErrEntryNotFound):
		return status.Error(codes.NotFound, "Entry not found: "+err.Error())
	// TODO: Add when implemented in domain/errors.go
	// case errors.Is(err, domain.ErrClaimNotFound):
	// 	return status.Error(codes.NotFound, "Claim not found: "+err.Error())
	// case errors.Is(err, domain.ErrAccountNotFound):
	// 	return status.Error(codes.NotFound, "Account not found: "+err.Error())

	// Conflict Errors → AlreadyExists
	case errors.Is(err, domain.ErrDuplicateKey):
		return status.Error(codes.AlreadyExists, "Key already registered. You may initiate a portability claim if you believe this is your key.")
	// TODO: Add when implemented in domain/errors.go
	// case errors.Is(err, domain.ErrDuplicateKeyGlobal):
	// 	return status.Error(codes.AlreadyExists, "Key already registered in RSFN. You may initiate a portability claim.")

	// Permission Errors → PermissionDenied
	case errors.Is(err, domain.ErrUnauthorized):
		return status.Error(codes.PermissionDenied, "Unauthorized: "+err.Error())
	// TODO: Add when implemented in domain/errors.go
	// case errors.Is(err, domain.ErrNotOwner):
	// 	return status.Error(codes.PermissionDenied, "You are not the owner of this resource.")

	// Resource Exhausted → ResourceExhausted
	case errors.Is(err, domain.ErrMaxKeysExceeded):
		return status.Error(codes.ResourceExhausted, "Maximum number of keys exceeded. Please delete an existing key before creating a new one.")

	// Precondition Failed → FailedPrecondition
	// TODO: Add when implemented in domain/errors.go
	// case errors.Is(err, domain.ErrCannotDeleteActiveKey):
	// 	return status.Error(codes.FailedPrecondition, "Cannot delete active key. Block it first or initiate proper deletion workflow.")
	// case errors.Is(err, domain.ErrCannotBlockDeletedKey):
	// 	return status.Error(codes.FailedPrecondition, "Cannot block a deleted key.")
	// case errors.Is(err, domain.ErrClaimAlreadyExists):
	// 	return status.Error(codes.FailedPrecondition, "An active claim already exists for this key.")

	// Deadline Exceeded → DeadlineExceeded
	case errors.Is(err, domain.ErrClaimExpired):
		return status.Error(codes.DeadlineExceeded, "Claim has expired (>30 days). Please initiate a new claim.")

	// Default → Internal
	default:
		return status.Error(codes.Internal, "Internal server error. Please try again later.")
	}
}

// ============================================================================
// gRPC Status → User-Friendly Message
// ============================================================================

// FormatUserFriendlyError converts gRPC status code to user-friendly message
func FormatUserFriendlyError(code codes.Code, msg string) string {
	switch code {
	case codes.InvalidArgument:
		return "Invalid input: " + msg
	case codes.NotFound:
		return "Resource not found: " + msg
	case codes.AlreadyExists:
		return "Resource already exists: " + msg
	case codes.PermissionDenied:
		return "Permission denied: " + msg
	case codes.Unauthenticated:
		return "Authentication required. Please log in."
	case codes.ResourceExhausted:
		return "Limit exceeded: " + msg
	case codes.FailedPrecondition:
		return "Operation not allowed: " + msg
	case codes.DeadlineExceeded:
		return "Request timeout: " + msg
	case codes.Unavailable:
		return "Service temporarily unavailable. Please try again."
	case codes.Internal:
		return "An error occurred. Please contact support if the problem persists."
	default:
		return msg
	}
}

// ============================================================================
// Context Errors
// ============================================================================

// MapContextError maps context errors to gRPC status
func MapContextError(err error) error {
	if err == nil {
		return nil
	}

	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return status.Error(codes.DeadlineExceeded, "Request timeout")
	case errors.Is(err, context.Canceled):
		return status.Error(codes.Canceled, "Request canceled")
	default:
		return err
	}
}
