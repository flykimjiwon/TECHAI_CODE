package companion

import (
	"sync"
)

const (
	ringCapacity = 500 // max events kept in memory
	chanBuffer   = 64  // per-subscriber channel buffer
)

// Hub is a thread-safe event ring buffer with pub/sub.
// Emit is non-blocking so it never stalls the TUI goroutine.
type Hub struct {
	mu          sync.Mutex
	ring        []Event
	nextID      int64
	subscribers map[int]chan Event
	nextSubID   int
}

// NewHub creates a ready-to-use Hub.
func NewHub() *Hub {
	return &Hub{
		ring:        make([]Event, 0, ringCapacity),
		nextID:      1,
		subscribers: make(map[int]chan Event),
	}
}

// Emit adds an event to the ring buffer and fans it out to every
// subscriber. If a subscriber's channel is full the event is dropped
// (non-blocking) so the TUI is never stalled.
func (h *Hub) Emit(eventType string, data interface{}) {
	ev := NewEvent(eventType, data)

	h.mu.Lock()
	ev.ID = h.nextID
	h.nextID++

	// Append to ring; evict oldest when full.
	if len(h.ring) >= ringCapacity {
		copy(h.ring, h.ring[1:])
		h.ring[len(h.ring)-1] = ev
	} else {
		h.ring = append(h.ring, ev)
	}

	// Snapshot subscriber channels under lock so we can send outside.
	subs := make([]chan Event, 0, len(h.subscribers))
	for _, ch := range h.subscribers {
		subs = append(subs, ch)
	}
	h.mu.Unlock()

	// Non-blocking fan-out (mirrors orchestrator.go:343-355 pattern).
	for _, ch := range subs {
		select {
		case ch <- ev:
		default:
			// Drop: subscriber is too slow.
		}
	}
}

// Subscribe returns a channel that receives live events and a cancel
// function. The caller must call cancel when done to avoid leaks.
func (h *Hub) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, chanBuffer)

	h.mu.Lock()
	id := h.nextSubID
	h.nextSubID++
	h.subscribers[id] = ch
	h.mu.Unlock()

	cancel := func() {
		h.mu.Lock()
		delete(h.subscribers, id)
		h.mu.Unlock()
	}
	return ch, cancel
}

// EventsAfter returns all buffered events whose ID is strictly
// greater than afterID. Used for SSE catch-up on reconnect.
func (h *Hub) EventsAfter(afterID int64) []Event {
	h.mu.Lock()
	defer h.mu.Unlock()

	var out []Event
	for _, ev := range h.ring {
		if ev.ID > afterID {
			out = append(out, ev)
		}
	}
	return out
}
