package interceptors

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor creates a gRPC unary server interceptor for request/response logging
func LoggingInterceptor(logger *logrus.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()

		// Generate or extract request ID
		requestID := extractRequestID(ctx)
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Add request ID to context
		ctx = context.WithValue(ctx, "request_id", requestID)

		// Extract metadata
		md, _ := metadata.FromIncomingContext(ctx)

		// Log incoming request
		logger.WithFields(logrus.Fields{
			"request_id":  requestID,
			"method":      info.FullMethod,
			"metadata":    md,
			"client_ip":   extractClientIP(ctx),
			"event":       "grpc_request_started",
		}).Info("Incoming gRPC request")

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(startTime)

		// Prepare log fields
		logFields := logrus.Fields{
			"request_id":     requestID,
			"method":         info.FullMethod,
			"duration_ms":    duration.Milliseconds(),
			"duration_human": duration.String(),
			"event":          "grpc_request_completed",
		}

		// Add status code
		if err != nil {
			st, _ := status.FromError(err)
			logFields["status_code"] = st.Code().String()
			logFields["error"] = err.Error()

			logger.WithFields(logFields).Error("gRPC request failed")
		} else {
			logFields["status_code"] = "OK"
			logger.WithFields(logFields).Info("gRPC request completed successfully")
		}

		return resp, err
	}
}

// extractRequestID extracts request ID from gRPC metadata or context
func extractRequestID(ctx context.Context) string {
	// Try to get from metadata first
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		if values := md.Get("x-request-id"); len(values) > 0 {
			return values[0]
		}
		if values := md.Get("request-id"); len(values) > 0 {
			return values[0]
		}
	}

	// Try to get from context value
	if reqID := ctx.Value("request_id"); reqID != nil {
		if id, ok := reqID.(string); ok {
			return id
		}
	}

	return ""
}

// extractClientIP extracts client IP from gRPC context
func extractClientIP(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "unknown"
	}

	// Common headers for client IP
	headers := []string{"x-forwarded-for", "x-real-ip", "x-client-ip"}
	for _, header := range headers {
		if values := md.Get(header); len(values) > 0 {
			return values[0]
		}
	}

	return "unknown"
}
