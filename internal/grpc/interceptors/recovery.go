package interceptors

import (
	"context"
	"fmt"
	"runtime/debug"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor creates a gRPC unary server interceptor for panic recovery
func RecoveryInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		// Defer panic recovery
		defer func() {
			if r := recover(); r != nil {
				// Capture stack trace
				stackTrace := string(debug.Stack())

				// Extract request ID if available
				requestID := ""
				if reqID := ctx.Value("request_id"); reqID != nil {
					if id, ok := reqID.(string); ok {
						requestID = id
					}
				}

				// Log panic with full details
				logger.WithFields(logrus.Fields{
					"request_id":  requestID,
					"method":      info.FullMethod,
					"panic_value": r,
					"stack_trace": stackTrace,
					"event":       "grpc_panic_recovered",
				}).Error("Panic recovered in gRPC handler")

				// Convert panic to gRPC error
				err = status.Errorf(
					codes.Internal,
					"Internal server error: %v",
					r,
				)

				// Set resp to nil to ensure clean error return
				resp = nil
			}
		}()

		// Call the actual handler
		return handler(ctx, req)
	}
}

// panicToError converts a panic value to a user-friendly error message
func panicToError(p interface{}) error {
	switch v := p.(type) {
	case error:
		return status.Errorf(codes.Internal, "Internal error: %v", v)
	case string:
		return status.Errorf(codes.Internal, "Internal error: %s", v)
	default:
		return status.Errorf(codes.Internal, "Internal error: %v", fmt.Sprintf("%v", v))
	}
}
