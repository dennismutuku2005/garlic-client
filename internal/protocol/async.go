package protocol

import (
	"strconv"
	"sync"
	"sync/atomic"
)

// ReplyChan is a channel that delivers replies.
type ReplyChan chan *Reply

// AsyncManager coordinates routing RouterOS replies to their corresponding command callers using tags.
type AsyncManager struct {
	mu        sync.RWMutex
	listeners map[string]ReplyChan
	tagSeq    uint64
	closed    bool
}

// NewAsyncManager creates a new async manager.
func NewAsyncManager() *AsyncManager {
	return &AsyncManager{
		listeners: make(map[string]ReplyChan),
	}
}

// NextTag generates a unique tag for a command.
func (m *AsyncManager) NextTag() string {
	seq := atomic.AddUint64(&m.tagSeq, 1)
	return "t" + strconv.FormatUint(seq, 10)
}

// Register registers a listener channel for the given tag.
func (m *AsyncManager) Register(tag string) (<-chan *Reply, ReplyChan) {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(ReplyChan, 100) // Buffer to avoid blocking the read loop under normal load
	if m.closed {
		close(ch)
		return ch, ch
	}

	m.listeners[tag] = ch
	return ch, ch
}

// Unregister closes and removes the listener for the given tag.
func (m *AsyncManager) Unregister(tag string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if ch, ok := m.listeners[tag]; ok {
		close(ch)
		delete(m.listeners, tag)
	}
}

// Distribute forwards a reply to its registered listener.
// If the listener's channel is full, it drops the reply and unregisters the listener
// to prevent blocking the background readLoop (which would lock up the entire client).
func (m *AsyncManager) Distribute(rep *Reply) {
	m.mu.RLock()
	ch, ok := m.listeners[rep.Tag]
	m.mu.RUnlock()

	if ok {
		select {
		case ch <- rep:
			if rep.Type == "!done" {
				m.Unregister(rep.Tag)
			}
		default:
			// Channel is full or blocked. Unregister it immediately to save resource leaks and blockages.
			m.Unregister(rep.Tag)
		}
	}
}

// CloseAll closes all listeners (e.g. when connection dies) and prevents new registrations.
func (m *AsyncManager) CloseAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.closed = true
	for tag, ch := range m.listeners {
		close(ch)
		delete(m.listeners, tag)
	}
}
