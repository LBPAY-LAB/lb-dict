package interceptors

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	requestIDKey = "request-id"
	traceIDKey   = "trace-id"
)

// LoggingInterceptor creates a gRPC unary server interceptor for request/response logging
func LoggingInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()

		// Extract or generate request ID
		requestID := extractOrGenerateRequestID(ctx)

		// Create logger with base fields
		logEntry := logger.WithFields(logrus.Fields{
			"request_id": requestID,
			"method":     info.FullMethod,
			"start_time": start.Format(time.RFC3339Nano),
		})

		// Add trace ID if present
		if traceID := extractTraceID(ctx); traceID != "" {
			logEntry = logEntry.WithField("trace_id", traceID)
		}

		// Add metadata to context for downstream use
		ctx = contextWithRequestID(ctx, requestID)

		// Log incoming request
		logEntry.Info("gRPC request started")

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start)

		// Determine status code
		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			} else {
				statusCode = codes.Unknown
			}
		}

		// Create response log entry
		responseLog := logEntry.WithFields(logrus.Fields{
			"duration_ms": duration.Milliseconds(),
			"duration":    duration.String(),
			"status_code": statusCode.String(),
			"success":     err == nil,
		})

		// Log based on status
		switch {
		case err != nil && statusCode >= codes.Internal:
			// Server errors (5xx equivalent)
			responseLog.WithError(err).Error("gRPC request completed with server error")
		case err != nil:
			// Client errors (4xx equivalent)
			responseLog.WithError(err).Warn("gRPC request completed with client error")
		case duration > 5*time.Second:
			// Slow requests
			responseLog.Warn("gRPC request completed (slow)")
		default:
			// Successful requests
			responseLog.Info("gRPC request completed")
		}

		return resp, err
	}
}

// extractOrGenerateRequestID extracts request ID from metadata or generates a new one
func extractOrGenerateRequestID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return generateRequestID()
	}

	values := md.Get(requestIDKey)
	if len(values) > 0 && values[0] != "" {
		return values[0]
	}

	values = md.Get("x-request-id")
	if len(values) > 0 && values[0] != "" {
		return values[0]
	}

	return generateRequestID()
}

// extractTraceID extracts trace ID from metadata
func extractTraceID(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(traceIDKey)
	if len(values) > 0 {
		return values[0]
	}

	values = md.Get("x-trace-id")
	if len(values) > 0 {
		return values[0]
	}

	return ""
}

// generateRequestID generates a new UUID for request tracking
func generateRequestID() string {
	return uuid.New().String()
}

// contextWithRequestID adds request ID to context metadata
func contextWithRequestID(ctx context.Context, requestID string) context.Context {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		md = metadata.New(map[string]string{})
	} else {
		md = md.Copy()
	}

	md.Set(requestIDKey, requestID)
	return metadata.NewIncomingContext(ctx, md)
}

// GetRequestIDFromContext retrieves request ID from context (utility function)
func GetRequestIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}

	values := md.Get(requestIDKey)
	if len(values) > 0 {
		return values[0]
	}

	return ""
}