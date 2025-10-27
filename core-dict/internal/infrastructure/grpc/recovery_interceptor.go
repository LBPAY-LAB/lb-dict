package grpc

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RecoveryInterceptor handles panic recovery for gRPC requests
// It catches panics in handlers and converts them to gRPC errors
type RecoveryInterceptor struct {
	logPanics       bool
	notifyOnPanic   PanicNotifier
}

// RecoveryConfig holds recovery interceptor configuration
type RecoveryConfig struct {
	LogPanics     bool          // Whether to log panics (default: true)
	NotifyOnPanic PanicNotifier // Optional callback for panic notifications (e.g., Sentry)
}

// PanicNotifier is a callback function for panic notifications
type PanicNotifier func(ctx context.Context, panicValue interface{}, stackTrace string)

// NewRecoveryInterceptor creates a new recovery interceptor
func NewRecoveryInterceptor(config *RecoveryConfig) *RecoveryInterceptor {
	if config == nil {
		config = &RecoveryConfig{
			LogPanics: true,
		}
	}

	return &RecoveryInterceptor{
		logPanics:     config.LogPanics,
		notifyOnPanic: config.NotifyOnPanic,
	}
}

// Unary returns a unary server interceptor for panic recovery
func (i *RecoveryInterceptor) Unary() grpc.UnaryServerInterceptor {
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

				// Log panic if enabled
				if i.logPanics {
					i.logPanic(ctx, info.FullMethod, r, stackTrace)
				}

				// Notify if callback is set
				if i.notifyOnPanic != nil {
					i.notifyOnPanic(ctx, r, stackTrace)
				}

				// Convert panic to gRPC error
				err = status.Errorf(
					codes.Internal,
					"internal server error: panic recovered: %v",
					r,
				)
			}
		}()

		// Call the handler
		resp, err = handler(ctx, req)
		return resp, err
	}
}

// logPanic logs panic information in structured format
func (i *RecoveryInterceptor) logPanic(ctx context.Context, method string, panicValue interface{}, stackTrace string) {
	// Extract user info from context if available
	userID, _ := ctx.Value("user_id").(string)
	ispb, _ := ctx.Value("ispb").(string)

	// Create structured log entry
	logEntry := map[string]interface{}{
		"timestamp":   time.Now().UTC().Format(time.RFC3339Nano),
		"level":       "ERROR",
		"type":        "panic_recovered",
		"method":      method,
		"panic_value": fmt.Sprintf("%v", panicValue),
		"stack_trace": stackTrace,
	}

	if userID != "" {
		logEntry["user_id"] = userID
	}
	if ispb != "" {
		logEntry["ispb"] = ispb
	}

	// Print structured log
	fmt.Printf("[PANIC RECOVERED] %+v\n", logEntry)
	fmt.Printf("Stack Trace:\n%s\n", stackTrace)
}

// Helper function to simulate panic notification to external service (e.g., Sentry)
func DefaultPanicNotifier(sentryDSN string) PanicNotifier {
	return func(ctx context.Context, panicValue interface{}, stackTrace string) {
		// In production, send to Sentry or similar error tracking service
		// For now, just log
		fmt.Printf("[SENTRY] Would send panic to %s: %v\n", sentryDSN, panicValue)

		// Example Sentry integration (pseudo-code):
		// sentry.CaptureException(fmt.Errorf("panic: %v", panicValue))
		// sentry.ConfigureScope(func(scope *sentry.Scope) {
		//     scope.SetContext("stack_trace", map[string]interface{}{
		//         "trace": stackTrace,
		//     })
		//     if userID, ok := ctx.Value("user_id").(string); ok {
		//         scope.SetUser(sentry.User{ID: userID})
		//     }
		// })
	}
}

// PanicInfo holds information about a recovered panic
type PanicInfo struct {
	Timestamp  time.Time
	Method     string
	PanicValue interface{}
	StackTrace string
	UserID     string
	ISPB       string
}

// SafeExecute wraps a function call with panic recovery
// This can be used in non-gRPC contexts
func SafeExecute(fn func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())
			fmt.Printf("[PANIC] %v\n%s\n", r, stackTrace)
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	return fn()
}

// SafeExecuteWithContext wraps a function call with panic recovery and context
func SafeExecuteWithContext(ctx context.Context, fn func(context.Context) error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			stackTrace := string(debug.Stack())

			userID, _ := ctx.Value("user_id").(string)
			ispb, _ := ctx.Value("ispb").(string)

			fmt.Printf("[PANIC] User: %s, ISPB: %s, Panic: %v\n%s\n", userID, ispb, r, stackTrace)
			err = fmt.Errorf("panic recovered: %v", r)
		}
	}()

	return fn(ctx)
}
