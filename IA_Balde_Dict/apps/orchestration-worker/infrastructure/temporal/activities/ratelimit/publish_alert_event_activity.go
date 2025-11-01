package ratelimit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/google/uuid"
	"go.temporal.io/sdk/activity"

	domainRL "github.com/lb-conn/connector-dict/domain/ratelimit"
)

// PublishAlertEventActivity publishes alert events to Pulsar
type PublishAlertEventActivity struct {
	pulsarProducer pulsar.Producer
}

// PublishAlertEventInput contains the input for publishing alert events
type PublishAlertEventInput struct {
	Alerts []*domainRL.Alert `json:"alerts"`
}

// RateLimitAlertEvent represents the alert event published to Pulsar
type RateLimitAlertEvent struct {
	EventID                     string    `json:"event_id"`
	EventType                   string    `json:"event_type"`
	Timestamp                   time.Time `json:"timestamp"`
	EndpointID                  string    `json:"endpoint_id"`
	Severity                    string    `json:"severity"`
	AvailableTokens             int       `json:"available_tokens"`
	Capacity                    int       `json:"capacity"`
	UtilizationPercent          float64   `json:"utilization_percent"`
	RecoveryETASeconds          int       `json:"recovery_eta_seconds"`
	ExhaustionProjectionSeconds int       `json:"exhaustion_projection_seconds"`
	PSPCategory                 string    `json:"psp_category"`
	Message                     string    `json:"message"`
}

// PublishAlertEventResult contains the result of publishing events
type PublishAlertEventResult struct {
	PublishedCount int `json:"published_count"`
	FailedCount    int `json:"failed_count"`
}

// NewPublishAlertEventActivity creates a new PublishAlertEventActivity
func NewPublishAlertEventActivity(pulsarProducer pulsar.Producer) *PublishAlertEventActivity {
	return &PublishAlertEventActivity{
		pulsarProducer: pulsarProducer,
	}
}

// Execute publishes alert events to Pulsar for Core-Dict consumption
func (a *PublishAlertEventActivity) Execute(ctx context.Context, input PublishAlertEventInput) (*PublishAlertEventResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("PublishAlertEventActivity started", "alerts", len(input.Alerts))

	if len(input.Alerts) == 0 {
		logger.Info("No alerts to publish")
		return &PublishAlertEventResult{PublishedCount: 0, FailedCount: 0}, nil
	}

	publishedCount := 0
	failedCount := 0

	for _, alert := range input.Alerts {
		// Create event
		event := RateLimitAlertEvent{
			EventID:                     uuid.New().String(),
			EventType:                   determineEventType(alert),
			Timestamp:                   time.Now().UTC(),
			EndpointID:                  alert.EndpointID,
			Severity:                    string(alert.Severity),
			AvailableTokens:             alert.AvailableTokens,
			Capacity:                    alert.Capacity,
			UtilizationPercent:          alert.UtilizationPercent,
			RecoveryETASeconds:          alert.RecoveryETASeconds,
			ExhaustionProjectionSeconds: alert.ExhaustionProjectionSeconds,
			PSPCategory:                 alert.PSPCategory,
			Message:                     alert.Message,
		}

		// Serialize to JSON
		payload, err := json.Marshal(event)
		if err != nil {
			logger.Error("Failed to marshal event",
				"alert_id", alert.ID,
				"error", err)
			failedCount++
			continue
		}

		// Publish to Pulsar
		_, err = a.pulsarProducer.Send(ctx, &pulsar.ProducerMessage{
			Payload: payload,
			Key:     alert.EndpointID,
			Properties: map[string]string{
				"event_type": event.EventType,
				"severity":   event.Severity,
				"category":   event.PSPCategory,
			},
		})

		if err != nil {
			logger.Error("Failed to publish event to Pulsar",
				"alert_id", alert.ID,
				"event_id", event.EventID,
				"error", err)
			failedCount++
			continue
		}

		logger.Info("Published alert event",
			"alert_id", alert.ID,
			"event_id", event.EventID,
			"event_type", event.EventType,
			"endpoint_id", alert.EndpointID,
			"severity", alert.Severity)

		publishedCount++
	}

	result := &PublishAlertEventResult{
		PublishedCount: publishedCount,
		FailedCount:    failedCount,
	}

	logger.Info("PublishAlertEventActivity completed",
		"published", publishedCount,
		"failed", failedCount)

	return result, nil
}

// determineEventType determines the event type based on alert state
func determineEventType(alert *domainRL.Alert) string {
	if alert.Resolved {
		return "rate_limit.alert.resolved"
	}
	return "rate_limit.alert.created"
}
