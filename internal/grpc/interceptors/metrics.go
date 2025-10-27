package interceptors

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

var (
	// grpcRequestsTotal tracks total number of gRPC requests by method and status
	grpcRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total number of gRPC requests",
		},
		[]string{"method", "status"},
	)

	// grpcRequestDuration tracks gRPC request duration in seconds
	grpcRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Duration of gRPC requests in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "status"},
	)

	// grpcActiveRequests tracks currently active gRPC requests
	grpcActiveRequests = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "grpc_active_requests",
			Help: "Number of currently active gRPC requests",
		},
		[]string{"method"},
	)
)

// MetricsInterceptor creates a gRPC unary server interceptor for Prometheus metrics collection
func MetricsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := info.FullMethod
		startTime := time.Now()

		// Increment active requests
		grpcActiveRequests.WithLabelValues(method).Inc()
		defer grpcActiveRequests.WithLabelValues(method).Dec()

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(startTime).Seconds()

		// Determine status code
		statusCode := "OK"
		if err != nil {
			st, _ := status.FromError(err)
			statusCode = st.Code().String()
		}

		// Record metrics
		grpcRequestsTotal.WithLabelValues(method, statusCode).Inc()
		grpcRequestDuration.WithLabelValues(method, statusCode).Observe(duration)

		return resp, err
	}
}
