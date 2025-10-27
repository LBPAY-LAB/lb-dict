package pulsar

import (
	"context"
	"testing"
	"time"

	"github.com/lbpay-lab/conn-bridge/internal/domain/entities"
	"github.com/stretchr/testify/assert"
)

func TestNewDictPublisher(t *testing.T) {
	tests := []struct {
		name        string
		config      *DictPublisherConfig
		expectError bool
	}{
		{
			name:        "nil config",
			config:      nil,
			expectError: true,
		},
		{
			name: "empty broker URL",
			config: &DictPublisherConfig{
				BrokerURL: "",
			},
			expectError: true,
		},
		{
			name: "valid config with defaults",
			config: &DictPublisherConfig{
				BrokerURL: "pulsar://localhost:6650",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			publisher, err := NewDictPublisher(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, publisher)
			} else {
				// Note: This will fail if Pulsar is not running
				// In a real test environment, you would use a mock or test container
				if err != nil {
					t.Skipf("Skipping test - Pulsar not available: %v", err)
				} else {
					assert.NoError(t, err)
					assert.NotNil(t, publisher)
					assert.NotNil(t, publisher.client)
					assert.NotNil(t, publisher.producer)
					defer publisher.Close()
				}
			}
		})
	}
}

func TestDefaultDictPublisherConfig(t *testing.T) {
	brokerURL := "pulsar://test:6650"
	config := DefaultDictPublisherConfig(brokerURL)

	assert.NotNil(t, config)
	assert.Equal(t, brokerURL, config.BrokerURL)
	assert.Equal(t, TopicDictResOut, config.Topic)
	assert.Equal(t, ProducerName, config.ProducerName)
	assert.True(t, config.BatchingEnabled)
	assert.Equal(t, uint(100), config.MaxMessages)
	assert.Equal(t, 10*time.Millisecond, config.BatchingMaxDelay)
}

func TestPublishEntryCreated_NilEntry(t *testing.T) {
	// Create a mock publisher (won't connect to real Pulsar)
	publisher := &DictPublisher{
		closed: false,
		logger: nil,
	}

	ctx := context.Background()
	err := publisher.PublishEntryCreated(ctx, nil, "trace-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entry cannot be nil")
}

func TestPublishEntryUpdated_NilEntry(t *testing.T) {
	publisher := &DictPublisher{
		closed: false,
		logger: nil,
	}

	ctx := context.Background()
	err := publisher.PublishEntryUpdated(ctx, nil, "trace-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "entry cannot be nil")
}

func TestPublishEntryDeleted_EmptyKeyID(t *testing.T) {
	publisher := &DictPublisher{
		closed: false,
		logger: nil,
	}

	ctx := context.Background()
	err := publisher.PublishEntryDeleted(ctx, "", "trace-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "keyID cannot be empty")
}

func TestPublishError_NilError(t *testing.T) {
	publisher := &DictPublisher{
		closed: false,
		logger: nil,
	}

	ctx := context.Background()
	err := publisher.PublishError(ctx, nil, "test context", "trace-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error cannot be nil")
}

func TestPublishEvent_ClosedPublisher(t *testing.T) {
	publisher := &DictPublisher{
		closed: true,
		logger: nil,
	}

	ctx := context.Background()
	entry := &entities.DictEntry{
		Key:  "test@example.com",
		Type: entities.KeyTypeEmail,
	}

	err := publisher.PublishEntryCreated(ctx, entry, "trace-123")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "publisher is closed")
}

func TestClose_AlreadyClosed(t *testing.T) {
	publisher := &DictPublisher{
		closed: true,
		logger: nil,
	}

	err := publisher.Close()
	assert.NoError(t, err)
}
