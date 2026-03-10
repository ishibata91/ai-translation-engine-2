package progress

import (
	"context"
	"sync"
)

// Hub is a ProgressNotifier that broadcasts events to multiple subscribers (e.g., WebSockets).
type Hub struct {
	mu        sync.RWMutex
	listeners map[chan ProgressEvent]bool
}

// NewHub initiates a new Progress notification hub.
func NewHub() *Hub {
	return &Hub{
		listeners: make(map[chan ProgressEvent]bool),
	}
}

// OnProgress implements ProgressNotifier. It broadcasts the event to all registered listeners.
func (h *Hub) OnProgress(ctx context.Context, event ProgressEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	for ch := range h.listeners {
		select {
		case ch <- event:
		default:
			// If listener is slow, skip to avoid blocking the whole system.
		}
	}
}

// Subscribe adds a new channel to receive progress events.
func (h *Hub) Subscribe(ch chan ProgressEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.listeners[ch] = true
}

// Unsubscribe removes a channel from listeners.
func (h *Hub) Unsubscribe(ch chan ProgressEvent) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.listeners, ch)
}
