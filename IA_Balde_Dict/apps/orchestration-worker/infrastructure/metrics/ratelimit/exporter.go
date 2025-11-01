package ratelimit

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/lb-conn/connector-dict/apps/orchestration-worker/application/ports"
	domainRL "github.com/lb-conn/connector-dict/domain/ratelimit"
)

// Exporter exports rate limit metrics to Prometheus
type Exporter struct {
	metrics   *Metrics
	stateRepo ports.StateRepository
	alertRepo ports.AlertRepository
}

// NewExporter creates a new metrics exporter
func NewExporter(
	registry prometheus.Registerer,
	stateRepo ports.StateRepository,
	alertRepo ports.AlertRepository,
) *Exporter {
	return &Exporter{
		metrics:   NewMetrics(registry),
		stateRepo: stateRepo,
		alertRepo: alertRepo,
	}
}

// UpdateMetrics updates all Prometheus metrics with current state
func (e *Exporter) UpdateMetrics(ctx context.Context) error {
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		e.metrics.MonitoringDuration.WithLabelValues("update_metrics").Observe(duration)
	}()

	// Get all latest states
	states, err := e.stateRepo.GetLatestAll(ctx)
	if err != nil {
		return err
	}

	// Update gauges for each state
	for _, state := range states {
		labels := prometheus.Labels{
			"endpoint_id":  state.EndpointID,
			"psp_category": state.PSPCategory,
		}

		e.metrics.AvailableTokens.With(labels).Set(float64(state.AvailableTokens))
		e.metrics.Capacity.With(labels).Set(float64(state.Capacity))
		e.metrics.UtilizationPercent.With(labels).Set(state.GetUtilizationPercent())
		e.metrics.ConsumptionRatePerMinute.With(labels).Set(state.ConsumptionRatePerMinute)
		e.metrics.RecoveryETASeconds.With(labels).Set(float64(state.RecoveryETASeconds))
		e.metrics.ExhaustionProjectionSeconds.With(labels).Set(float64(state.ExhaustionProjectionSeconds))
		e.metrics.Error404Rate.With(labels).Set(state.Error404Rate)
	}

	return nil
}

// RecordAlertCreated increments the alert created counter
func (e *Exporter) RecordAlertCreated(endpointID string, severity domainRL.AlertSeverity, category string) {
	e.metrics.AlertsCreatedTotal.WithLabelValues(endpointID, string(severity), category).Inc()
}

// RecordAlertResolved increments the alert resolved counter
func (e *Exporter) RecordAlertResolved(endpointID string, severity domainRL.AlertSeverity, category string) {
	e.metrics.AlertsResolvedTotal.WithLabelValues(endpointID, string(severity), category).Inc()
}

// RecordOperationDuration records the duration of a monitoring operation
func (e *Exporter) RecordOperationDuration(operation string, duration time.Duration) {
	e.metrics.MonitoringDuration.WithLabelValues(operation).Observe(duration.Seconds())
}
