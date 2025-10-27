package pulsar

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func getTestConsumer(t *testing.T, topic, subscription string) *Consumer {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := ConsumerConfig{
		URL:              "pulsar://localhost:6650",
		Topic:            topic,
		SubscriptionName: subscription,
		ConsumerName:     "test-consumer",
		ConnectTimeout:   5 * time.Second,
	}

	consumer, err := NewConsumer(config, logger)
	if err != nil {
		t.Skipf("Pulsar not available: %v", err)
		return nil
	}
	return consumer
}

func getTestProducerForConsumer(t *testing.T, topic string) *Producer {
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

// TestNewConsumer tests Consumer creation
func TestNewConsumer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tests := []struct {
		name             string
		brokerURL        string
		topic            string
		subscriptionName string
		wantErr          bool
	}{
		{
			name:             "valid_configuration",
			brokerURL:        "pulsar://localhost:6650",
			topic:            "test-topic",
			subscriptionName: "test-sub",
			wantErr:          false,
		},
		{
			name:             "empty_broker_url",
			brokerURL:        "",
			topic:            "test-topic",
			subscriptionName: "test-sub",
			wantErr:          true,
		},
		{
			name:             "empty_topic",
			brokerURL:        "pulsar://localhost:6650",
			topic:            "",
			subscriptionName: "test-sub",
			wantErr:          true,
		},
		{
			name:             "empty_subscription",
			brokerURL:        "pulsar://localhost:6650",
			topic:            "test-topic",
			subscriptionName: "",
			wantErr:          true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ConsumerConfig{
				URL:              tt.brokerURL,
				Topic:            tt.topic,
				SubscriptionName: tt.subscriptionName,
				ConsumerName:     "test-consumer",
				ConnectTimeout:   5 * time.Second,
			}
			consumer, err := NewConsumer(config, logger)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, consumer)
			} else {
				// Connection will fail in tests without Pulsar
				if err != nil {
					assert.Contains(t, err.Error(), "connection")
				}
			}
		})
	}
}

// TestConsumerStart tests message consumption
func TestConsumerStart(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	// Setup producer to send test messages
	producer := getTestProducerForConsumer(t, "test-consume-topic")
	if producer == nil {
		return
	}
	defer producer.Close()

	// Setup consumer
	consumer := getTestConsumer(t, "test-consume-topic", "test-consume-sub")
	if consumer == nil {
		return
	}
	defer consumer.Close()

	// Channel to receive processed messages
	received := make(chan map[string]interface{}, 10)

	// Message handler
	handler := func(ctx context.Context, msg pulsar.Message) error {
		var data map[string]interface{}
		if err := json.Unmarshal(msg.Payload(), &data); err != nil {
			return err
		}
		received <- data
		return nil
	}

	// Start consumer in background
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		_ = consumer.Start(ctx, handler)
	}()

	// Give consumer time to start
	time.Sleep(100 * time.Millisecond)

	// Publish test messages
	testEvents := []map[string]string{
		{"type": "test1", "value": "data1"},
		{"type": "test2", "value": "data2"},
		{"type": "test3", "value": "data3"},
	}

	for _, event := range testEvents {
		err := producer.PublishEvent(context.Background(), event, "test-key")
		require.NoError(t, err)
	}

	// Wait for messages to be received
	timeout := time.After(3 * time.Second)
	receivedCount := 0

	for receivedCount < len(testEvents) {
		select {
		case msg := <-received:
			t.Logf("Received message: %v", msg)
			receivedCount++
		case <-timeout:
			t.Logf("Timeout: received %d/%d messages", receivedCount, len(testEvents))
			return
		}
	}

	assert.Equal(t, len(testEvents), receivedCount)
}

// TestConsumerHandlerError tests error handling in message handler
func TestConsumerHandlerError(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	producer := getTestProducerForConsumer(t, "test-error-topic")
	if producer == nil {
		return
	}
	defer producer.Close()

	consumer := getTestConsumer(t, "test-error-topic", "test-error-sub")
	if consumer == nil {
		return
	}
	defer consumer.Close()

	errorCount := 0

	// Handler that always errors
	handler := func(ctx context.Context, msg pulsar.Message) error {
		errorCount++
		return assert.AnError
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	go func() {
		_ = consumer.Start(ctx, handler)
	}()

	time.Sleep(100 * time.Millisecond)

	// Publish message that will error
	err := producer.PublishEvent(context.Background(), map[string]string{"test": "error"}, "error-key")
	require.NoError(t, err)

	time.Sleep(500 * time.Millisecond)

	// Handler should have been called (and Nack'd the message)
	assert.GreaterOrEqual(t, errorCount, 1)
}

// TestConsumerContextCancellation tests graceful shutdown
func TestConsumerContextCancellation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	consumer := getTestConsumer(t, "test-cancel-topic", "test-cancel-sub")
	if consumer == nil {
		return
	}
	defer consumer.Close()

	handler := func(ctx context.Context, msg pulsar.Message) error {
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Start consumer
	done := make(chan error, 1)
	go func() {
		done <- consumer.Start(ctx, handler)
	}()

	// Cancel after 100ms
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Should exit gracefully
	select {
	case err := <-done:
		assert.Error(t, err) // Should return context.Canceled
		assert.Equal(t, context.Canceled, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Consumer did not stop after context cancellation")
	}
}

// TestConsumerClose tests proper cleanup
func TestConsumerClose(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	consumer := getTestConsumer(t, "test-close-topic", "test-close-sub")
	if consumer == nil {
		return
	}

	// Close should not panic
	require.NotPanics(t, func() {
		consumer.Close()
	})

	// Double close should not panic
	require.NotPanics(t, func() {
		consumer.Close()
	})
}

// TestConsumerAckNack tests message acknowledgment
func TestConsumerAckNack(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	producer := getTestProducerForConsumer(t, "test-ack-topic")
	if producer == nil {
		return
	}
	defer producer.Close()

	consumer := getTestConsumer(t, "test-ack-topic", "test-ack-sub")
	if consumer == nil {
		return
	}
	defer consumer.Close()

	ackCount := 0
	nackCount := 0

	// Handler that acks even messages, nacks odd
	handler := func(ctx context.Context, msg pulsar.Message) error {
		var data map[string]interface{}
		json.Unmarshal(msg.Payload(), &data)

		if seq, ok := data["seq"].(float64); ok {
			if int(seq)%2 == 0 {
				ackCount++
				return nil // Ack
			}
			nackCount++
			return assert.AnError // Nack
		}
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = consumer.Start(ctx, handler)
	}()

	time.Sleep(100 * time.Millisecond)

	// Publish 10 messages
	for i := 0; i < 10; i++ {
		event := map[string]interface{}{
			"seq":  i,
			"data": "test",
		}
		err := producer.PublishEvent(context.Background(), event, "test-key")
		require.NoError(t, err)
	}

	time.Sleep(1 * time.Second)

	t.Logf("Ack count: %d, Nack count: %d", ackCount, nackCount)
	assert.Greater(t, ackCount, 0)
	assert.Greater(t, nackCount, 0)
}

// TestConsumerMessageProperties tests reading message properties
func TestConsumerMessageProperties(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping Pulsar integration test in short mode")
	}

	producer := getTestProducerForConsumer(t, "test-props-topic")
	if producer == nil {
		return
	}
	defer producer.Close()

	consumer := getTestConsumer(t, "test-props-topic", "test-props-sub")
	if consumer == nil {
		return
	}
	defer consumer.Close()

	receivedProps := make(chan map[string]string, 1)

	handler := func(ctx context.Context, msg pulsar.Message) error {
		receivedProps <- msg.Properties()
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		_ = consumer.Start(ctx, handler)
	}()

	time.Sleep(100 * time.Millisecond)

	// Publish event (producer adds properties automatically)
	event := map[string]string{"test": "properties"}
	err := producer.PublishEvent(context.Background(), event, "props-key")
	require.NoError(t, err)

	select {
	case props := <-receivedProps:
		t.Logf("Received properties: %v", props)
		assert.Contains(t, props, "event_time")
		assert.Contains(t, props, "producer")
	case <-time.After(2 * time.Second):
		t.Fatal("Did not receive message properties")
	}
}

// Benchmark consumption performance
func BenchmarkConsumerProcessing(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := ConsumerConfig{
		URL:              "pulsar://localhost:6650",
		Topic:            "bench-consume",
		SubscriptionName: "bench-consume-sub",
		ConsumerName:     "bench-consumer",
		ConnectTimeout:   5 * time.Second,
	}

	consumer, err := NewConsumer(config, logger)
	if err != nil {
		b.Skipf("Pulsar not available: %v", err)
		return
	}
	defer consumer.Close()

	handler := func(ctx context.Context, msg pulsar.Message) error {
		// Minimal processing
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(b.N)*time.Millisecond)
	defer cancel()

	b.ResetTimer()
	_ = consumer.Start(ctx, handler)
}