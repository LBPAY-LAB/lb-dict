package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor handles structured logging for gRPC requests
type LoggingInterceptor struct {
	enableRequestLogging  bool
	enableResponseLogging bool
	logPayload            bool
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	EnableRequestLogging  bool // Log incoming requests
	EnableResponseLogging bool // Log outgoing responses
	LogPayload            bool // Include request/response payload in logs (careful with PII)
}

// NewLoggingInterceptor creates a new logging interceptor
func NewLoggingInterceptor(config *LoggingConfig) *LoggingInterceptor {
	if config == nil {
		config = &LoggingConfig{
			EnableRequestLogging:  true,
			EnableResponseLogging: true,
			LogPayload:            false, // Default to false for security
		}
	}

	return &LoggingInterceptor{
		enableRequestLogging:  config.EnableRequestLogging,
		enableResponseLogging: config.EnableResponseLogging,
		logPayload:            config.LogPayload,
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp   string                 `json:"timestamp"`
	Level       string                 `json:"level"`       // INFO, WARN, ERROR
	Type        string                 `json:"type"`        // request, response
	Method      string                 `json:"method"`      // gRPC method
	UserID      string                 `json:"user_id,omitempty"`
	ISPB        string                 `json:"ispb,omitempty"`
	DurationMs  int64                  `json:"duration_ms,omitempty"`
	StatusCode  codes.Code             `json:"status_code"`
	StatusMsg   string                 `json:"status_msg,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Payload     interface{}            `json:"payload,omitempty"`
}

// Unary returns a unary server interceptor for logging
func (i *LoggingInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		startTime := time.Now()

		// Extract user info from context (set by auth interceptor)
		userID, _ := ctx.Value("user_id").(string)
		ispb, _ := ctx.Value("ispb").(string)

		// Log incoming request
		if i.enableRequestLogging {
			i.logRequest(ctx, info.FullMethod, userID, ispb, req)
		}

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(startTime)

		// Log outgoing response
		if i.enableResponseLogging {
			i.logResponse(ctx, info.FullMethod, userID, ispb, resp, err, duration)
		}

		return resp, err
	}
}

// logRequest logs incoming gRPC requests
func (i *LoggingInterceptor) logRequest(ctx context.Context, method string, userID, ispb string, req interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     "INFO",
		Type:      "request",
		Method:    method,
		UserID:    userID,
		ISPB:      ispb,
		Metadata:  i.extractMetadata(ctx),
	}

	if i.logPayload {
		entry.Payload = req
	}

	i.printJSON(entry)
}

// logResponse logs outgoing gRPC responses
func (i *LoggingInterceptor) logResponse(ctx context.Context, method string, userID, ispb string, resp interface{}, err error, duration time.Duration) {
	entry := LogEntry{
		Timestamp:  time.Now().UTC().Format(time.RFC3339Nano),
		Type:       "response",
		Method:     method,
		UserID:     userID,
		ISPB:       ispb,
		DurationMs: duration.Milliseconds(),
	}

	if err != nil {
		// Error response
		st, ok := status.FromError(err)
		if ok {
			entry.StatusCode = st.Code()
			entry.StatusMsg = st.Message()
		} else {
			entry.StatusCode = codes.Unknown
			entry.StatusMsg = err.Error()
		}
		entry.Level = i.getLevelForCode(entry.StatusCode)
		entry.Error = err.Error()
	} else {
		// Success response
		entry.Level = "INFO"
		entry.StatusCode = codes.OK
		entry.StatusMsg = "success"

		if i.logPayload {
			entry.Payload = resp
		}
	}

	i.printJSON(entry)
}

// extractMetadata extracts relevant metadata from gRPC context
func (i *LoggingInterceptor) extractMetadata(ctx context.Context) map[string]interface{} {
	meta := make(map[string]interface{})

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return meta
	}

	// Extract useful headers (avoid logging sensitive data like authorization tokens)
	for key, values := range md {
		if key == "authorization" {
			continue // Skip authorization header
		}
		if len(values) == 1 {
			meta[key] = values[0]
		} else {
			meta[key] = values
		}
	}

	return meta
}

// getLevelForCode returns appropriate log level based on gRPC status code
func (i *LoggingInterceptor) getLevelForCode(code codes.Code) string {
	switch code {
	case codes.OK:
		return "INFO"
	case codes.Canceled, codes.InvalidArgument, codes.NotFound, codes.AlreadyExists, codes.Unauthenticated:
		return "WARN"
	default:
		return "ERROR"
	}
}

// printJSON prints log entry as JSON to stdout
func (i *LoggingInterceptor) printJSON(entry LogEntry) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		// Fallback to plain text if JSON encoding fails
		fmt.Printf("[ERROR] Failed to encode log entry: %v\n", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

// Helper function to create structured logs from anywhere in the code

// LogInfo logs an informational message
func LogInfo(ctx context.Context, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     "INFO",
		StatusMsg: message,
		Metadata:  fields,
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		entry.UserID = userID
	}
	if ispb, ok := ctx.Value("ispb").(string); ok {
		entry.ISPB = ispb
	}

	jsonBytes, _ := json.Marshal(entry)
	fmt.Println(string(jsonBytes))
}

// LogError logs an error message
func LogError(ctx context.Context, message string, err error, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     "ERROR",
		StatusMsg: message,
		Error:     err.Error(),
		Metadata:  fields,
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		entry.UserID = userID
	}
	if ispb, ok := ctx.Value("ispb").(string); ok {
		entry.ISPB = ispb
	}

	jsonBytes, _ := json.Marshal(entry)
	fmt.Println(string(jsonBytes))
}

// LogWarn logs a warning message
func LogWarn(ctx context.Context, message string, fields map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     "WARN",
		StatusMsg: message,
		Metadata:  fields,
	}

	if userID, ok := ctx.Value("user_id").(string); ok {
		entry.UserID = userID
	}
	if ispb, ok := ctx.Value("ispb").(string); ok {
		entry.ISPB = ispb
	}

	jsonBytes, _ := json.Marshal(entry)
	fmt.Println(string(jsonBytes))
}
