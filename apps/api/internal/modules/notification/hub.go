package notification

import "sync"

// Hub is an in-process pub/sub that fans out stream events to the live
// connections of each user. It needs no external broker; with several API
// instances each instance only reaches its own connections.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[chan StreamEvent]struct{}
}

// NewHub builds an empty Hub.
func NewHub() *Hub {
	return &Hub{subs: make(map[string]map[chan StreamEvent]struct{})}
}

// Subscribe registers a live connection for userID. The returned func must be
// called to release it.
func (h *Hub) Subscribe(userID string) (<-chan StreamEvent, func()) {
	ch := make(chan StreamEvent, 16)
	h.mu.Lock()
	if h.subs[userID] == nil {
		h.subs[userID] = make(map[chan StreamEvent]struct{})
	}
	h.subs[userID][ch] = struct{}{}
	h.mu.Unlock()

	var once sync.Once
	return ch, func() {
		once.Do(func() {
			h.mu.Lock()
			delete(h.subs[userID], ch)
			if len(h.subs[userID]) == 0 {
				delete(h.subs, userID)
			}
			h.mu.Unlock()
		})
	}
}

// Publish delivers an event to every connection of userID. Slow consumers are
// skipped rather than blocking the publisher.
func (h *Hub) Publish(userID string, ev StreamEvent) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[userID] {
		select {
		case ch <- ev:
		default:
		}
	}
}

// Connected reports whether userID has at least one live connection.
func (h *Hub) Connected(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.subs[userID]) > 0
}
