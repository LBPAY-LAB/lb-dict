package testhelpers

import (
	"sync"
	"time"
)

// PulsarMock simulates Apache Pulsar for testing
type PulsarMock struct {
	mu            sync.RWMutex
	publishedMsgs map[string][]MockMessage
	subscribers   map[string][]chan MockMessage
}

// MockMessage represents a Pulsar message
type MockMessage struct {
	Topic     string
	Key       string
	Payload   []byte
	Timestamp time.Time
}

// NewPulsarMock creates a new Pulsar mock
func NewPulsarMock() *PulsarMock {
	return &PulsarMock{
		publishedMsgs: make(map[string][]MockMessage),
		subscribers:   make(map[string][]chan MockMessage),
	}
}

// Publish publishes a message to a topic
func (p *PulsarMock) Publish(topic string, key string, payload []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	msg := MockMessage{
		Topic:     topic,
		Key:       key,
		Payload:   payload,
		Timestamp: time.Now(),
	}

	p.publishedMsgs[topic] = append(p.publishedMsgs[topic], msg)

	// Notify subscribers
	if subs, ok := p.subscribers[topic]; ok {
		for _, ch := range subs {
			select {
			case ch <- msg:
			default:
			}
		}
	}

	return nil
}

// Subscribe subscribes to a topic
func (p *PulsarMock) Subscribe(topic string) <-chan MockMessage {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan MockMessage, 100)
	p.subscribers[topic] = append(p.subscribers[topic], ch)
	return ch
}

// GetMessages returns all messages published to a topic
func (p *PulsarMock) GetMessages(topic string) []MockMessage {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return append([]MockMessage{}, p.publishedMsgs[topic]...)
}

// ReceivedEvent checks if an event was received
func (p *PulsarMock) ReceivedEvent(topic string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.publishedMsgs[topic]) > 0
}

// WaitForEvent waits for an event with timeout
func (p *PulsarMock) WaitForEvent(topic string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if p.ReceivedEvent(topic) {
			return true
		}
		time.Sleep(50 * time.Millisecond)
	}
	return false
}

// GetMessageCount returns the number of messages for a topic
func (p *PulsarMock) GetMessageCount(topic string) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.publishedMsgs[topic])
}

// Reset clears all messages
func (p *PulsarMock) Reset() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.publishedMsgs = make(map[string][]MockMessage)
}

// Stop stops the mock
func (p *PulsarMock) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, subs := range p.subscribers {
		for _, ch := range subs {
			close(ch)
		}
	}
	p.subscribers = make(map[string][]chan MockMessage)
}
