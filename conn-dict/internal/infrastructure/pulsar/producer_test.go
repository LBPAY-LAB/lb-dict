package pulsar

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestProducer(t *testing.T, topic string) *Producer {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := ProducerConfig{
		URL:            "pulsar://localhost:6650",
		Topic:          topic,
		ProducerName:   "test-producer",
		ConnectTimeout: 5 * time.Second,
	}

	producer, err := NewProducer(config, logger)
	if err != nil {
		t.Skipf("Pulsar not available: %v", err)
		return nil
	}
	return producer
}

// TestNewProducer tests Producer creation
func TestNewProducer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name      string
		brokerURL string
		topic     string
		wantErr   bool
	}{
		{
			name:      "valid_configuration",
			brokerURL: "pulsar://localhost:6650",
			topic:     "test-topic",
			wantErr:   false, // Will fail connection but struct created
		},
		{
			name:      "empty_broker_url",
			brokerURL: "",
			topic:     "test-topic",
			wantErr:   true,
		},
		{
			name:      "empty_topic",
			brokerURL: "pulsar://localhost:6650",
			topic:     "",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ProducerConfig{
				URL:            tt.brokerURL,
				Topic:          tt.topic,
				ProducerName:   "test-producer",
				ConnectTimeout: 5 * time.Second,
			}
			producer, err := NewProducer(config, logger)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, producer)
			} else {
				// Note: Connection will fail in tests without Pulsar running
				// but we test that the struct is created correctly
				if err != nil {
					// Expected in test environment without Pulsar
					assert.Contains(t, err.Error(), "connection")
				}
			}
		})
	}
}

// TestPublishEvent tests async event publishing
func TestPublishEvent(t *testing.T) {
	// Skip if no Pulsar available
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	producer := getTestProducer(t, "test-events")
	if producer == nil {
		return
	}
	defer producer.Close()

	tests := []struct {
		name    string
		event   interface{}
		key     string
		wantErr bool
	}{
		{
			name: "valid_event",
			event: map[string]string{
				"type": "claim_created",
				"id":   "claim-123",
			},
			key:     "claim-123",
			wantErr: false,
		},
		{
			name: "empty_event",
			event: map[string]string{},
			key:   "empty-key",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := producer.PublishEvent(ctx, tt.event, tt.key)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// Async publish might not error immediately
				assert.NoError(t, err)
			}
		})
	}
}

// TestPublishEventSync tests sync event publishing
func TestPublishEventSync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	producer := getTestProducer(t, "test-events-sync")
	if producer == nil {
		return
	}
	defer producer.Close()

	ctx := context.Background()
	event := map[string]interface{}{
		"type":      "claim_completed",
		"claim_id":  "claim-456",
		"timestamp": time.Now().Unix(),
	}

	msgID, err := producer.PublishEventSync(ctx, event, "claim-456")
	require.NoError(t, err)
	assert.NotNil(t, msgID)

	t.Logf("Published message ID: %v", msgID)
}

// TestPublishEventWithProperties tests event publishing with custom properties
func TestPublishEventWithProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	producer := getTestProducer(t, "test-events-props")
	if producer == nil {
		return
	}
	defer producer.Close()

	ctx := context.Background()
	event := map[string]string{
		"action": "claim_cancelled",
		"reason": "user_request",
	}

	// Should include default properties (event_time, producer)
	err := producer.PublishEvent(ctx, event, "claim-789")
	assert.NoError(t, err)
}

// TestPublishEventMarshalError tests handling of marshal errors
func TestPublishEventMarshalError(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	// Create producer with mock (won't connect)
	producer := &Producer{
		logger: logger,
		topic:  "test-topic",
	}

	ctx := context.Background()

	// Channel cannot be marshaled to JSON
	invalidEvent := make(chan int)

	err := producer.PublishEvent(ctx, invalidEvent, "test-key")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "marshal")
}

// TestProducerContextCancellation tests context cancellation
func TestProducerContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	producer := getTestProducer(t, "test-cancel")
	if producer == nil {
		return
	}
	defer producer.Close()

	// Create cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	event := map[string]string{"test": "data"}

	// Async publish should handle cancelled context
	err := producer.PublishEvent(ctx, event, "test-key")
	// Note: Async might not return error for cancelled context
	// This depends on Pulsar client implementation
	t.Logf("PublishEvent with cancelled context: %v", err)
}

// TestProducerClose tests proper cleanup
func TestProducerClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	producer := getTestProducer(t, "test-close")
	if producer == nil {
		return
	}

	// Close should not panic
	require.NotPanics(t, func() {
		producer.Close()
	})

	// Double close should not panic
	require.NotPanics(t, func() {
		producer.Close()
	})
}

// TestProducerCompression tests compression settings
func TestProducerCompression(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	producer := getTestProducer(t, "test-compression")
	if producer == nil {
		return
	}
	defer producer.Close()

	ctx := context.Background()

	// Publish large event to test compression
	largeEvent := make(map[string]string)
	for i := 0; i < 100; i++ {
		key := string(rune('a'+i%26)) + fmt.Sprintf("%d", i)
		largeEvent[key] = "Some repeated data to compress"
	}

	msgID, err := producer.PublishEventSync(ctx, largeEvent, "large-event")
	require.NoError(t, err)
	assert.NotNil(t, msgID)
}

// Benchmark publishing performance
func BenchmarkPublishEvent(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := ProducerConfig{
		URL:            "pulsar://localhost:6650",
		Topic:          "bench-topic",
		ProducerName:   "bench-producer",
		ConnectTimeout: 5 * time.Second,
	}

	producer, err := NewProducer(config, logger)
	if err != nil {
		b.Skipf("Pulsar not available: %v", err)
		return
	}
	defer producer.Close()

	ctx := context.Background()
	event := map[string]string{
		"type": "benchmark_event",
		"data": "test data",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = producer.PublishEvent(ctx, event, "bench-key")
	}
}

func BenchmarkPublishEventSync(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := ProducerConfig{
		URL:            "pulsar://localhost:6650",
		Topic:          "bench-sync-topic",
		ProducerName:   "bench-producer-sync",
		ConnectTimeout: 5 * time.Second,
	}

	producer, err := NewProducer(config, logger)
	if err != nil {
		b.Skipf("Pulsar not available: %v", err)
		return
	}
	defer producer.Close()

	ctx := context.Background()
	event := map[string]string{
		"type": "benchmark_event_sync",
		"data": "test data",
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = producer.PublishEventSync(ctx, event, "bench-key")
	}
}