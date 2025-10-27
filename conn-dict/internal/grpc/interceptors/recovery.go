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
		// Recover from panics
		defer func() {
			if r := recover(); r != nil {
				// Capture stack trace
				stackTrace := debug.Stack()

				// Extract request ID if available
				requestID := GetRequestIDFromContext(ctx)

				// Log panic with full details
				logFields := logrus.Fields{
					"method":      info.FullMethod,
					"panic_value": r,
					"stack_trace": string(stackTrace),
				}

				if requestID != "" {
					logFields["request_id"] = requestID
				}

				logger.WithFields(logFields).Error("gRPC handler panic recovered")

				// Convert panic to gRPC error
				err = convertPanicToError(r)

				// Ensure response is nil on panic
				resp = nil
			}
		}()

		// Call the actual handler
		return handler(ctx, req)
	}
}

// convertPanicToError converts a panic value to an appropriate gRPC error
func convertPanicToError(r interface{}) error {
	var msg string

	switch v := r.(type) {
	case error:
		msg = v.Error()
	case string:
		msg = v
	default:
		msg = fmt.Sprintf("panic: %v", r)
	}

	// Return Internal error (gRPC code 13, equivalent to HTTP 500)
	return status.Errorf(codes.Internal, "internal server error: %s", msg)
}

// RecoveryWithCustomHandler creates a recovery interceptor with custom error handling
func RecoveryWithCustomHandler(
	logger *logrus.Logger,
	customHandler func(context.Context, interface{}) error,
) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := debug.Stack()
				requestID := GetRequestIDFromContext(ctx)

				logFields := logrus.Fields{
					"method":      info.FullMethod,
					"panic_value": r,
					"stack_trace": string(stackTrace),
				}

				if requestID != "" {
					logFields["request_id"] = requestID
				}

				logger.WithFields(logFields).Error("gRPC handler panic recovered (custom handler)")

				// Use custom error handler if provided
				if customHandler != nil {
					err = customHandler(ctx, r)
				} else {
					err = convertPanicToError(r)
				}

				resp = nil
			}
		}()

		return handler(ctx, req)
	}
}

// SafeRecoveryInterceptor creates a recovery interceptor that never returns nil error
// This ensures a panic always results in a proper gRPC error response
func SafeRecoveryInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				stackTrace := debug.Stack()
				requestID := GetRequestIDFromContext(ctx)

				logFields := logrus.Fields{
					"method":      info.FullMethod,
					"panic_value": r,
					"stack_trace": string(stackTrace),
				}

				if requestID != "" {
					logFields["request_id"] = requestID
				}

				logger.WithFields(logFields).Error("gRPC handler panic recovered (safe mode)")

				// Always return a valid error
				err = status.Error(codes.Internal, "internal server error occurred")
				resp = nil
			}
		}()

		resp, err = handler(ctx, req)

		// Safety check: ensure we never return (nil, nil) after panic recovery
		if resp == nil && err == nil {
			logger.WithField("method", info.FullMethod).Warn("handler returned (nil, nil), converting to error")
			err = status.Error(codes.Internal, "handler returned invalid response")
		}

		return resp, err
	}
}