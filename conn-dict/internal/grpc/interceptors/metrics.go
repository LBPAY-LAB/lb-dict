package interceptors

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Prometheus metrics
var (
	// grpcRequestsTotal tracks total number of gRPC requests
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc",
			Name:      "requests_total",
			Help:      "Total number of gRPC requests",
		},
		[]string{"method", "status_code"},
	)

	// grpcRequestDuration tracks request duration in seconds
	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc",
			Name:      "request_duration_seconds",
			Help:      "gRPC request duration in seconds",
			Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1.0, 2.5, 5.0, 10.0},
		},
		[]string{"method", "status_code"},
	)

	// grpcActiveRequests tracks number of active requests
	grpcActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc",
			Name:      "active_requests",
			Help:      "Number of active gRPC requests",
		},
		[]string{"method"},
	)

	// grpcRequestSize tracks request payload size in bytes
	grpcRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc",
			Name:      "request_size_bytes",
			Help:      "gRPC request payload size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 7), // 100B to 100MB
		},
		[]string{"method"},
	)

	// grpcResponseSize tracks response payload size in bytes
	grpcResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc",
			Name:      "response_size_bytes",
			Help:      "gRPC response payload size in bytes",
			Buckets:   prometheus.ExponentialBuckets(100, 10, 7), // 100B to 100MB
		},
		[]string{"method"},
	)

	// grpcErrorsTotal tracks total number of errors by type
	grpcErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "conn_dict",
			Subsystem: "grpc",
			Name:      "errors_total",
			Help:      "Total number of gRPC errors by type",
		},
		[]string{"method", "error_type"},
	)
)

// MetricsInterceptor creates a gRPC unary server interceptor for Prometheus metrics
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := info.FullMethod

		// Increment active requests
		grpcActiveRequests.WithLabelValues(method).Inc()
		defer grpcActiveRequests.WithLabelValues(method).Dec()

		// Track request size (approximate)
		if sized, ok := req.(interface{ Size() int }); ok {
			grpcRequestSize.WithLabelValues(method).Observe(float64(sized.Size()))
		}

		// Start timer
		start := time.Now()

		// Call handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(start).Seconds()

		// Determine status code
		statusCode := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				statusCode = st.Code()
			} else {
				statusCode = codes.Unknown
			}

			// Track error type
			errorType := categorizeError(statusCode)
			grpcErrorsTotal.WithLabelValues(method, errorType).Inc()
		}

		statusStr := statusCode.String()

		// Record metrics
		grpcRequestsTotal.WithLabelValues(method, statusStr).Inc()
		grpcRequestDuration.WithLabelValues(method, statusStr).Observe(duration)

		// Track response size (approximate)
		if resp != nil {
			if sized, ok := resp.(interface{ Size() int }); ok {
				grpcResponseSize.WithLabelValues(method).Observe(float64(sized.Size()))
			}
		}

		return resp, err
	}
}

// categorizeError categorizes gRPC status codes into error types
func categorizeError(code codes.Code) string {
	switch code {
	case codes.OK:
		return "success"
	case codes.Canceled:
		return "client_error"
	case codes.Unknown:
		return "unknown_error"
	case codes.InvalidArgument:
		return "client_error"
	case codes.DeadlineExceeded:
		return "timeout"
	case codes.NotFound:
		return "not_found"
	case codes.AlreadyExists:
		return "conflict"
	case codes.PermissionDenied:
		return "permission_denied"
	case codes.ResourceExhausted:
		return "rate_limit"
	case codes.FailedPrecondition:
		return "client_error"
	case codes.Aborted:
		return "aborted"
	case codes.OutOfRange:
		return "client_error"
	case codes.Unimplemented:
		return "not_implemented"
	case codes.Internal:
		return "server_error"
	case codes.Unavailable:
		return "unavailable"
	case codes.DataLoss:
		return "data_loss"
	case codes.Unauthenticated:
		return "unauthenticated"
	default:
		return "unknown_error"
	}
}

// GetMetricsRegistry returns the Prometheus registry (for testing/inspection)
func GetMetricsRegistry() prometheus.Gatherer {
	return prometheus.DefaultGatherer
}

// ResetMetrics resets all metrics (useful for testing)
func ResetMetrics() {
	grpcRequestsTotal.Reset()
	grpcRequestDuration.Reset()
	grpcActiveRequests.Reset()
	grpcRequestSize.Reset()
	grpcResponseSize.Reset()
	grpcErrorsTotal.Reset()
}