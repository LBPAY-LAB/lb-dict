package cache

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

// CacheMetrics provides observability for cache operations
type CacheMetrics struct {
	logger *logrus.Logger
}

// NewCacheMetrics creates a new CacheMetrics instance
func NewCacheMetrics(logger *logrus.Logger) *CacheMetrics {
	return &CacheMetrics{
		logger: logger,
	}
}

// RecordCacheHit records a cache hit
func (cm *CacheMetrics) RecordCacheHit(strategy CacheStrategy) {
	cacheHitsTotal.WithLabelValues(string(strategy)).Inc()
}

// RecordCacheMiss records a cache miss
func (cm *CacheMetrics) RecordCacheMiss(strategy CacheStrategy) {
	cacheMissesTotal.WithLabelValues(string(strategy)).Inc()
}

// RecordOperation records a cache operation with duration
func (cm *CacheMetrics) RecordOperation(operation string, strategy CacheStrategy, duration time.Duration) {
	cacheOperationDuration.WithLabelValues(operation, string(strategy)).Observe(duration.Seconds())
}

// RecordError records a cache error
func (cm *CacheMetrics) RecordError(operation, errorType string) {
	cacheErrorsTotal.WithLabelValues(operation, errorType).Inc()
}

// GetCacheHitRate calculates the cache hit rate for a strategy
func (cm *CacheMetrics) GetCacheHitRate(strategy CacheStrategy) (float64, error) {
	// This is a helper for monitoring - in production you'd query Prometheus
	hitsMetric := cacheHitsTotal.WithLabelValues(string(strategy))
	missesMetric := cacheMissesTotal.WithLabelValues(string(strategy))

	// Note: This returns the metric objects, not values
	// In production, query Prometheus API for actual values
	_ = hitsMetric
	_ = missesMetric

	return 0.0, nil // Placeholder
}

// LogCacheStatistics logs cache statistics for monitoring
func (cm *CacheMetrics) LogCacheStatistics(ctx context.Context) {
	// This would typically gather metrics from Prometheus
	cm.logger.Info("Cache statistics logged (query Prometheus for detailed metrics)")
}

// CacheStatsCollector collects and logs cache statistics periodically
type CacheStatsCollector struct {
	metrics  *CacheMetrics
	logger   *logrus.Logger
	interval time.Duration
	stop     chan bool
}

// NewCacheStatsCollector creates a new cache statistics collector
func NewCacheStatsCollector(metrics *CacheMetrics, logger *logrus.Logger, interval time.Duration) *CacheStatsCollector {
	return &CacheStatsCollector{
		metrics:  metrics,
		logger:   logger,
		interval: interval,
		stop:     make(chan bool),
	}
}

// Start begins collecting cache statistics
func (csc *CacheStatsCollector) Start() {
	ticker := time.NewTicker(csc.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				ctx := context.Background()
				csc.metrics.LogCacheStatistics(ctx)
			case <-csc.stop:
				ticker.Stop()
				csc.logger.Info("Cache stats collector stopped")
				return
			}
		}
	}()
	csc.logger.Infof("Cache stats collector started: interval=%v", csc.interval)
}

// Stop stops the collector
func (csc *CacheStatsCollector) Stop() {
	close(csc.stop)
}

// GetMetricsRegistry returns the Prometheus registry for cache metrics
func GetMetricsRegistry() prometheus.Gatherer {
	return prometheus.DefaultGatherer
}

// ResetCacheMetrics resets all cache metrics (useful for testing)
func ResetCacheMetrics() {
	cacheHitsTotal.Reset()
	cacheMissesTotal.Reset()
	cacheOperationDuration.Reset()
	cacheErrorsTotal.Reset()
	writeBehindQueueSize.Set(0)
	// Note: Histogram doesn't have Reset() method in prometheus client
	// writeBehindFlushDuration is automatically reset by prometheus
}

// MetricLabels provides label constants for metrics
var MetricLabels = struct {
	Strategy  string
	Operation string
	ErrorType string
}{
	Strategy:  "strategy",
	Operation: "operation",
	ErrorType: "error_type",
}

// StrategyNames maps strategy types to human-readable names
var StrategyNames = map[CacheStrategy]string{
	StrategyCacheAside:    "Cache-Aside (Lazy Loading)",
	StrategyWriteThrough:  "Write-Through",
	StrategyWriteBehind:   "Write-Behind (Write-Back)",
	StrategyReadThrough:   "Read-Through",
	StrategyWriteAround:   "Write-Around",
}

// GetStrategyName returns the human-readable name for a strategy
func GetStrategyName(strategy CacheStrategy) string {
	if name, ok := StrategyNames[strategy]; ok {
		return name
	}
	return string(strategy)
}
