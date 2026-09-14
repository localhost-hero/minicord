package chat

import "sync"

// Hub fans out newly saved messages to every subscriber of a room.
type Hub struct {
	mu    sync.Mutex
	rooms map[string]map[chan Message]struct{}
}

// NewHub returns an empty, ready-to-use Hub.
func NewHub() *Hub {
	return &Hub{rooms: make(map[string]map[chan Message]struct{})}
}

// Subscribe registers a new listener for a room and returns the channel that
// will receive published messages. The caller must call Unsubscribe when done.
func (h *Hub) Subscribe(room string) chan Message {
	ch := make(chan Message, 16)

	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[room] == nil {
		h.rooms[room] = make(map[chan Message]struct{})
	}
	h.rooms[room][ch] = struct{}{}

	return ch
}

// Unsubscribe removes a listener and closes its channel.
func (h *Hub) Unsubscribe(room string, ch chan Message) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if listeners, ok := h.rooms[room]; ok {
		if _, subscribed := listeners[ch]; subscribed {
			delete(listeners, ch)
			close(ch)
		}
		if len(listeners) == 0 {
			delete(h.rooms, room)
		}
	}
}

// Publish delivers a message to every current subscriber of a room. Slow
// subscribers that cannot keep up have the message dropped rather than
// blocking the publisher.
func (h *Hub) Publish(room string, message Message) {
	h.mu.Lock()
	listeners := make([]chan Message, 0, len(h.rooms[room]))
	for ch := range h.rooms[room] {
		listeners = append(listeners, ch)
	}
	h.mu.Unlock()

	for _, ch := range listeners {
		select {
		case ch <- message:
		default:
		}
	}
}
