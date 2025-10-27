package grpc

import (
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Domain errors for Connect client
var (
	ErrConnectUnavailable  = errors.New("connect service unavailable")
	ErrConnectTimeout      = errors.New("connect service timeout")
	ErrEntryNotFound       = errors.New("entry not found")
	ErrClaimNotFound       = errors.New("claim not found")
	ErrInfractionNotFound  = errors.New("infraction not found")
	ErrDuplicateEntry      = errors.New("duplicate entry")
	ErrInvalidInput        = errors.New("invalid input")
	ErrCircuitOpen         = errors.New("circuit breaker is open")
)

// mapGRPCError converts gRPC errors to domain errors
func mapGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		// Try to determine which entity wasn't found based on message
		msg := st.Message()
		if len(msg) > 0 {
			switch msg[0] {
			case 'e', 'E':
				return ErrEntryNotFound
			case 'c', 'C':
				return ErrClaimNotFound
			case 'i', 'I':
				return ErrInfractionNotFound
			}
		}
		return ErrEntryNotFound // Default to entry
	case codes.AlreadyExists:
		return ErrDuplicateEntry
	case codes.InvalidArgument:
		return ErrInvalidInput
	case codes.Unavailable:
		return ErrConnectUnavailable
	case codes.DeadlineExceeded:
		return ErrConnectTimeout
	case codes.ResourceExhausted:
		return ErrConnectUnavailable
	default:
		return err
	}
}
