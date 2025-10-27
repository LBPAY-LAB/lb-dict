package grpc

import (
	"context"
	"fmt"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// MetricsInterceptor handles Prometheus metrics collection for gRPC requests
type MetricsInterceptor struct {
	mu sync.RWMutex

	// Request counters by method and status
	requestCount map[string]map[string]int64 // method -> status -> count

	// Request duration histogram buckets (in milliseconds)
	requestDuration map[string][]int64 // method -> durations

	// Active requests gauge
	activeRequests map[string]int64 // method -> count

	// Error counters
	errorCount map[string]int64 // method -> count
}

// NewMetricsInterceptor creates a new metrics interceptor
func NewMetricsInterceptor() *MetricsInterceptor {
	return &MetricsInterceptor{
		requestCount:    make(map[string]map[string]int64),
		requestDuration: make(map[string][]int64),
		activeRequests:  make(map[string]int64),
		errorCount:      make(map[string]int64),
	}
}

// Unary returns a unary server interceptor for metrics collection
func (i *MetricsInterceptor) Unary() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		method := info.FullMethod
		startTime := time.Now()

		// Increment active requests
		i.incrementActiveRequests(method)
		defer i.decrementActiveRequests(method)

		// Call the handler
		resp, err := handler(ctx, req)

		// Calculate duration
		duration := time.Since(startTime)

		// Get status code
		code := codes.OK
		if err != nil {
			if st, ok := status.FromError(err); ok {
				code = st.Code()
			} else {
				code = codes.Unknown
			}
		}

		// Record metrics
		i.recordRequest(method, code, duration)

		if err != nil {
			i.incrementErrorCount(method)
		}

		return resp, err
	}
}

// recordRequest records a completed request
func (i *MetricsInterceptor) recordRequest(method string, code codes.Code, duration time.Duration) {
	i.mu.Lock()
	defer i.mu.Unlock()

	// Initialize map for method if not exists
	if i.requestCount[method] == nil {
		i.requestCount[method] = make(map[string]int64)
	}

	// Increment counter for this method and status
	statusStr := code.String()
	i.requestCount[method][statusStr]++

	// Record duration (in milliseconds)
	if i.requestDuration[method] == nil {
		i.requestDuration[method] = make([]int64, 0)
	}
	i.requestDuration[method] = append(i.requestDuration[method], duration.Milliseconds())

	// Keep only last 1000 durations to avoid memory bloat
	if len(i.requestDuration[method]) > 1000 {
		i.requestDuration[method] = i.requestDuration[method][len(i.requestDuration[method])-1000:]
	}
}

// incrementActiveRequests increments active requests counter
func (i *MetricsInterceptor) incrementActiveRequests(method string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.activeRequests[method]++
}

// decrementActiveRequests decrements active requests counter
func (i *MetricsInterceptor) decrementActiveRequests(method string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.activeRequests[method]--
}

// incrementErrorCount increments error counter
func (i *MetricsInterceptor) incrementErrorCount(method string) {
	i.mu.Lock()
	defer i.mu.Unlock()
	i.errorCount[method]++
}

// GetMetrics returns current metrics snapshot
func (i *MetricsInterceptor) GetMetrics() *MetricsSnapshot {
	i.mu.RLock()
	defer i.mu.RUnlock()

	snapshot := &MetricsSnapshot{
		Timestamp:       time.Now(),
		RequestCount:    make(map[string]map[string]int64),
		RequestDuration: make(map[string]*DurationStats),
		ActiveRequests:  make(map[string]int64),
		ErrorCount:      make(map[string]int64),
	}

	// Copy request counts
	for method, statuses := range i.requestCount {
		snapshot.RequestCount[method] = make(map[string]int64)
		for status, count := range statuses {
			snapshot.RequestCount[method][status] = count
		}
	}

	// Calculate duration stats
	for method, durations := range i.requestDuration {
		if len(durations) > 0 {
			snapshot.RequestDuration[method] = calculateDurationStats(durations)
		}
	}

	// Copy active requests
	for method, count := range i.activeRequests {
		snapshot.ActiveRequests[method] = count
	}

	// Copy error counts
	for method, count := range i.errorCount {
		snapshot.ErrorCount[method] = count
	}

	return snapshot
}

// MetricsSnapshot represents a snapshot of metrics at a point in time
type MetricsSnapshot struct {
	Timestamp       time.Time
	RequestCount    map[string]map[string]int64 // method -> status -> count
	RequestDuration map[string]*DurationStats   // method -> stats
	ActiveRequests  map[string]int64            // method -> count
	ErrorCount      map[string]int64            // method -> count
}

// DurationStats holds statistical information about request durations
type DurationStats struct {
	Count      int64   // Total number of requests
	Sum        int64   // Sum of all durations (ms)
	Min        int64   // Minimum duration (ms)
	Max        int64   // Maximum duration (ms)
	Mean       float64 // Average duration (ms)
	P50        int64   // 50th percentile (median)
	P95        int64   // 95th percentile
	P99        int64   // 99th percentile
}

// calculateDurationStats calculates statistics from a list of durations
func calculateDurationStats(durations []int64) *DurationStats {
	if len(durations) == 0 {
		return &DurationStats{}
	}

	stats := &DurationStats{
		Count: int64(len(durations)),
		Min:   durations[0],
		Max:   durations[0],
	}

	// Calculate sum, min, max
	for _, d := range durations {
		stats.Sum += d
		if d < stats.Min {
			stats.Min = d
		}
		if d > stats.Max {
			stats.Max = d
		}
	}

	// Calculate mean
	stats.Mean = float64(stats.Sum) / float64(stats.Count)

	// Calculate percentiles (simple implementation - sort and pick)
	// Note: This modifies the input slice, but since we're working with a copy, it's okay
	sorted := make([]int64, len(durations))
	copy(sorted, durations)

	// Simple bubble sort (fine for small arrays, use quicksort for production)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	stats.P50 = sorted[int(float64(len(sorted))*0.50)]
	stats.P95 = sorted[int(float64(len(sorted))*0.95)]
	if len(sorted) > 1 {
		stats.P99 = sorted[int(float64(len(sorted))*0.99)]
	} else {
		stats.P99 = stats.P95
	}

	return stats
}

// PrometheusExporter exports metrics in Prometheus format
func (i *MetricsInterceptor) PrometheusExporter() string {
	snapshot := i.GetMetrics()

	var output string

	// HELP and TYPE for grpc_request_total
	output += "# HELP grpc_request_total Total number of gRPC requests\n"
	output += "# TYPE grpc_request_total counter\n"
	for method, statuses := range snapshot.RequestCount {
		for status, count := range statuses {
			output += "grpc_request_total{method=\"" + method + "\",status=\"" + status + "\"} " + toString(count) + "\n"
		}
	}

	// HELP and TYPE for grpc_request_duration_milliseconds
	output += "# HELP grpc_request_duration_milliseconds gRPC request duration in milliseconds\n"
	output += "# TYPE grpc_request_duration_milliseconds summary\n"
	for method, stats := range snapshot.RequestDuration {
		output += "grpc_request_duration_milliseconds{method=\"" + method + "\",quantile=\"0.5\"} " + toString(stats.P50) + "\n"
		output += "grpc_request_duration_milliseconds{method=\"" + method + "\",quantile=\"0.95\"} " + toString(stats.P95) + "\n"
		output += "grpc_request_duration_milliseconds{method=\"" + method + "\",quantile=\"0.99\"} " + toString(stats.P99) + "\n"
		output += "grpc_request_duration_milliseconds_sum{method=\"" + method + "\"} " + toString(stats.Sum) + "\n"
		output += "grpc_request_duration_milliseconds_count{method=\"" + method + "\"} " + toString(stats.Count) + "\n"
	}

	// HELP and TYPE for grpc_active_requests
	output += "# HELP grpc_active_requests Number of active gRPC requests\n"
	output += "# TYPE grpc_active_requests gauge\n"
	for method, count := range snapshot.ActiveRequests {
		output += "grpc_active_requests{method=\"" + method + "\"} " + toString(count) + "\n"
	}

	// HELP and TYPE for grpc_errors_total
	output += "# HELP grpc_errors_total Total number of gRPC errors\n"
	output += "# TYPE grpc_errors_total counter\n"
	for method, count := range snapshot.ErrorCount {
		output += "grpc_errors_total{method=\"" + method + "\"} " + toString(count) + "\n"
	}

	return output
}

// toString converts an int64 to string
func toString(i int64) string {
	return fmt.Sprintf("%d", i)
}
