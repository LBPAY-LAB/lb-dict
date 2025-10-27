package adapters

import (
	"context"

	commonv1 "github.com/lbpay-lab/dict-contracts/gen/proto/common/v1"
	"github.com/lbpay-lab/conn-dict/internal/domain/events"
	"github.com/lbpay-lab/conn-dict/internal/infrastructure/pulsar"
	"github.com/sirupsen/logrus"
)

// EventPublisherAdapter adapts the infrastructure Producer to the use case interface
type EventPublisherAdapter struct {
	producer *pulsar.Producer
	logger   *logrus.Logger
}

// NewEventPublisherAdapter creates a new adapter
func NewEventPublisherAdapter(producer *pulsar.Producer, logger *logrus.Logger) *EventPublisherAdapter {
	return &EventPublisherAdapter{
		producer: producer,
		logger:   logger,
	}
}

// PublishEntryCreated publishes an entry created event
func (a *EventPublisherAdapter) PublishEntryCreated(ctx context.Context, entryID string, key *commonv1.DictKey, account *commonv1.Account) error {
	event := events.NewEntryCreatedEvent(
		entryID,
		key.KeyValue,
		mapKeyType(key.KeyType),
		account.Ispb,
		account.AccountNumber,
	)

	if err := a.producer.PublishEvent(ctx, event, entryID); err != nil {
		a.logger.WithError(err).WithField("entry_id", entryID).Error("Failed to publish EntryCreated event")
		return err
	}

	a.logger.WithField("entry_id", entryID).Debug("Published EntryCreated event")
	return nil
}

// PublishEntryUpdated publishes an entry updated event
func (a *EventPublisherAdapter) PublishEntryUpdated(ctx context.Context, entryID string, account *commonv1.Account) error {
	event := events.NewEntryUpdatedEvent(
		entryID,
		account.Ispb,
		account.AccountNumber,
	)

	if err := a.producer.PublishEvent(ctx, event, entryID); err != nil {
		a.logger.WithError(err).WithField("entry_id", entryID).Error("Failed to publish EntryUpdated event")
		return err
	}

	a.logger.WithField("entry_id", entryID).Debug("Published EntryUpdated event")
	return nil
}

// PublishEntryDeleted publishes an entry deleted event
func (a *EventPublisherAdapter) PublishEntryDeleted(ctx context.Context, entryID string) error {
	event := events.NewEntryDeletedEvent(entryID)

	if err := a.producer.PublishEvent(ctx, event, entryID); err != nil {
		a.logger.WithError(err).WithField("entry_id", entryID).Error("Failed to publish EntryDeleted event")
		return err
	}

	a.logger.WithField("entry_id", entryID).Debug("Published EntryDeleted event")
	return nil
}

// Helper to map proto KeyType to string
func mapKeyType(kt commonv1.KeyType) string {
	switch kt {
	case commonv1.KeyType_KEY_TYPE_CPF:
		return "CPF"
	case commonv1.KeyType_KEY_TYPE_CNPJ:
		return "CNPJ"
	case commonv1.KeyType_KEY_TYPE_EMAIL:
		return "EMAIL"
	case commonv1.KeyType_KEY_TYPE_PHONE:
		return "PHONE"
	case commonv1.KeyType_KEY_TYPE_EVP:
		return "EVP"
	default:
		return "UNKNOWN"
	}
}
