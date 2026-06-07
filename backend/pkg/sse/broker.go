// Package sse provides a generic fan-out broker for Server-Sent Events.
package sse

import (
	"encoding/json"
	"sync"
)

// Event is a single SSE message with a named type and a JSON payload.
type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Broker distributes events to all currently subscribed clients.
type Broker struct {
	mu      sync.RWMutex
	clients map[chan Event]struct{}
}

// NewBroker allocates an empty Broker ready to use.
func NewBroker() *Broker {
	return &Broker{clients: make(map[chan Event]struct{})}
}

// Subscribe registers a new client channel. The caller must call the returned
// unsubscribe function (typically via defer) to release the channel.
func (b *Broker) Subscribe() (events <-chan Event, unsubscribe func()) {
	ch := make(chan Event, 64)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch, func() {
		b.mu.Lock()
		delete(b.clients, ch)
		close(ch)
		b.mu.Unlock()
	}
}

// Publish broadcasts an event to every subscribed client. Slow consumers are
// silently dropped (non-blocking send) to avoid back-pressure on the publisher.
func (b *Broker) Publish(e Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- e:
		default:
		}
	}
}
