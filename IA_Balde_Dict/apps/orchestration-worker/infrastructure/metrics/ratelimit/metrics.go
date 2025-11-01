package ratelimit

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// Metrics contains all Prometheus metrics for rate limit monitoring
type Metrics struct {
	// Gauges for current state
	AvailableTokens           *prometheus.GaugeVec
	Capacity                  *prometheus.GaugeVec
	UtilizationPercent        *prometheus.GaugeVec
	ConsumptionRatePerMinute  *prometheus.GaugeVec
	RecoveryETASeconds        *prometheus.GaugeVec
	ExhaustionProjectionSeconds *prometheus.GaugeVec
	Error404Rate              *prometheus.GaugeVec

	// Counters for events
	AlertsCreatedTotal  *prometheus.CounterVec
	AlertsResolvedTotal *prometheus.CounterVec

	// Histogram for operation duration
	MonitoringDuration *prometheus.HistogramVec
}

// NewMetrics creates and registers all Prometheus metrics
func NewMetrics(registry prometheus.Registerer) *Metrics {
	return &Metrics{
		// Gauges
		AvailableTokens: promauto.With(registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "dict_rate_limit_available_tokens",
				Help: "Current available tokens in the rate limit bucket",
			},
			[]string{"endpoint_id", "psp_category"},
		),

		Capacity: promauto.With(registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "dict_rate_limit_capacity",
				Help: "Maximum capacity of the rate limit bucket",
			},
			[]string{"endpoint_id", "psp_category"},
		),

		UtilizationPercent: promauto.With(registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "dict_rate_limit_utilization_percent",
				Help: "Percentage of rate limit capacity utilized",
			},
			[]string{"endpoint_id", "psp_category"},
		),

		ConsumptionRatePerMinute: promauto.With(registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "dict_rate_limit_consumption_rate_per_minute",
				Help: "Token consumption rate in tokens per minute",
			},
			[]string{"endpoint_id", "psp_category"},
		),

		RecoveryETASeconds: promauto.With(registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "dict_rate_limit_recovery_eta_seconds",
				Help: "Estimated seconds until full token recovery",
			},
			[]string{"endpoint_id", "psp_category"},
		),

		ExhaustionProjectionSeconds: promauto.With(registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "dict_rate_limit_exhaustion_projection_seconds",
				Help: "Estimated seconds until token exhaustion if current trend continues",
			},
			[]string{"endpoint_id", "psp_category"},
		),

		Error404Rate: promauto.With(registry).NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "dict_rate_limit_error_404_rate",
				Help: "Percentage of 404 errors in recent requests",
			},
			[]string{"endpoint_id", "psp_category"},
		),

		// Counters
		AlertsCreatedTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Name: "dict_rate_limit_alerts_created_total",
				Help: "Total number of rate limit alerts created",
			},
			[]string{"endpoint_id", "severity", "psp_category"},
		),

		AlertsResolvedTotal: promauto.With(registry).NewCounterVec(
			prometheus.CounterOpts{
				Name: "dict_rate_limit_alerts_resolved_total",
				Help: "Total number of rate limit alerts resolved",
			},
			[]string{"endpoint_id", "severity", "psp_category"},
		),

		// Histogram
		MonitoringDuration: promauto.With(registry).NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "dict_rate_limit_monitoring_duration_seconds",
				Help:    "Duration of rate limit monitoring operations",
				Buckets: []float64{0.01, 0.05, 0.1, 0.5, 1, 5, 10},
			},
			[]string{"operation"},
		),
	}
}
