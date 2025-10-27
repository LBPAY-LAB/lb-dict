package messaging_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lbpay-lab/core-dict/internal/infrastructure/messaging"
)

func TestDefaultProducerConfig(t *testing.T) {
	config := messaging.DefaultEntryProducerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, 100, config.BatchingMaxMessages)
	assert.Equal(t, 10*time.Millisecond, config.BatchingMaxPublishDelay)
	assert.Equal(t, 30*time.Second, config.SendTimeout)
	assert.Equal(t, 1000, config.MaxPendingMessages)
}

func TestDefaultConsumerConfig(t *testing.T) {
	config := messaging.DefaultEntryEventConsumerConfig()

	assert.NotNil(t, config)
	assert.Equal(t, "core-dict-entry-events-sub", config.SubscriptionName)
	assert.Equal(t, 3, config.MaxRedeliveryCount)
	assert.Equal(t, "dict.events.dlq", config.DLQTopic)
}
