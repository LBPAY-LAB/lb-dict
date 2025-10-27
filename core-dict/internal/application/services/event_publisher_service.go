package services

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

// EventPublisherService publica eventos de domínio para Apache Pulsar
type EventPublisherService struct {
	pulsarProducer PulsarProducer
	topicMapper    TopicMapper
}

// PulsarProducer interface para publicar mensagens no Pulsar
type PulsarProducer interface {
	Send(ctx context.Context, topic string, message []byte) error
	SendAsync(ctx context.Context, topic string, message []byte) error
	Close() error
}

// TopicMapper mapeia tipo de evento para tópico Pulsar
type TopicMapper interface {
	GetTopic(eventType string) string
}

// DefaultTopicMapper implementação padrão do topic mapper
type DefaultTopicMapper struct {
	topicPrefix string
}

// NewDefaultTopicMapper cria novo topic mapper com prefixo
func NewDefaultTopicMapper(topicPrefix string) *DefaultTopicMapper {
	return &DefaultTopicMapper{
		topicPrefix: topicPrefix,
	}
}

// GetTopic retorna tópico Pulsar baseado no tipo de evento
func (m *DefaultTopicMapper) GetTopic(eventType string) string {
	// Mapeia eventos para tópicos específicos
	topicMap := map[string]string{
		"EntryCreated":      "persistent://lbpay/core-dict/entry-created",
		"EntryUpdated":      "persistent://lbpay/core-dict/entry-updated",
		"EntryDeleted":      "persistent://lbpay/core-dict/entry-deleted",
		"EntryBlocked":      "persistent://lbpay/core-dict/entry-blocked",
		"EntryUnblocked":    "persistent://lbpay/core-dict/entry-unblocked",
		"ClaimReceived":     "persistent://lbpay/core-dict/claim-received",
		"ClaimConfirmed":    "persistent://lbpay/core-dict/claim-confirmed",
		"ClaimCancelled":    "persistent://lbpay/core-dict/claim-cancelled",
		"ClaimCompleted":    "persistent://lbpay/core-dict/claim-completed",
		"InfractionCreated": "persistent://lbpay/core-dict/infraction-created",
	}

	if topic, ok := topicMap[eventType]; ok {
		return topic
	}

	// Tópico padrão para eventos desconhecidos
	return m.topicPrefix + "/domain-events"
}

// NewEventPublisherService cria nova instância
func NewEventPublisherService(producer PulsarProducer, topicMapper TopicMapper) *EventPublisherService {
	return &EventPublisherService{
		pulsarProducer: producer,
		topicMapper:    topicMapper,
	}
}

// DomainEvent representa um evento de domínio
type DomainEvent struct {
	EventID       string                 `json:"event_id"`
	EventType     string                 `json:"event_type"`
	AggregateID   string                 `json:"aggregate_id"`
	AggregateType string                 `json:"aggregate_type"`
	OccurredAt    time.Time              `json:"occurred_at"`
	Payload       map[string]interface{} `json:"payload"`
}

// Publish publica evento de domínio no Pulsar (síncrono)
func (s *EventPublisherService) Publish(ctx context.Context, event DomainEvent) error {
	// 1. Validar evento
	if err := s.validateEvent(event); err != nil {
		return errors.New("invalid event: " + err.Error())
	}

	// 2. Serializar para JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return errors.New("failed to serialize event: " + err.Error())
	}

	// 3. Determinar tópico
	topic := s.topicMapper.GetTopic(event.EventType)

	// 4. Publicar no Pulsar (síncrono)
	if err := s.pulsarProducer.Send(ctx, topic, eventJSON); err != nil {
		return errors.New("failed to publish event: " + err.Error())
	}

	return nil
}

// PublishAsync publica evento de domínio no Pulsar (assíncrono)
func (s *EventPublisherService) PublishAsync(ctx context.Context, event DomainEvent) error {
	// 1. Validar evento
	if err := s.validateEvent(event); err != nil {
		return errors.New("invalid event: " + err.Error())
	}

	// 2. Serializar para JSON
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return errors.New("failed to serialize event: " + err.Error())
	}

	// 3. Determinar tópico
	topic := s.topicMapper.GetTopic(event.EventType)

	// 4. Publicar no Pulsar (assíncrono - fire and forget)
	if err := s.pulsarProducer.SendAsync(ctx, topic, eventJSON); err != nil {
		return errors.New("failed to publish async event: " + err.Error())
	}

	return nil
}

// PublishBatch publica múltiplos eventos em batch (otimização)
func (s *EventPublisherService) PublishBatch(ctx context.Context, events []DomainEvent) error {
	// Publicar todos os eventos em sequência
	// TODO: Otimizar para batch real do Pulsar
	for _, event := range events {
		if err := s.Publish(ctx, event); err != nil {
			return errors.New("failed to publish batch event " + event.EventID + ": " + err.Error())
		}
	}
	return nil
}

// validateEvent valida campos obrigatórios do evento
func (s *EventPublisherService) validateEvent(event DomainEvent) error {
	if event.EventType == "" {
		return errors.New("event_type is required")
	}
	if event.AggregateID == "" {
		return errors.New("aggregate_id is required")
	}
	if event.AggregateType == "" {
		return errors.New("aggregate_type is required")
	}
	if event.OccurredAt.IsZero() {
		return errors.New("occurred_at is required")
	}
	return nil
}

// Close fecha conexão com Pulsar
func (s *EventPublisherService) Close() error {
	return s.pulsarProducer.Close()
}
