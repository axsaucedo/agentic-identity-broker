package approval

import (
	"sync"
	"time"
)

// ApprovalSyncBroadcaster manages long-poll subscriber channels and delivers
// coalesced wake-up notifications when approval state changes.
type ApprovalSyncBroadcaster struct {
	mu              sync.Mutex
	subscribers     map[chan struct{}]struct{}
	coalesceWindow  time.Duration
	coalesceTimer   *time.Timer
	coalescePending bool
}

// NewApprovalSyncBroadcaster creates a broadcaster with the given coalesce window.
// A zero duration disables coalescing (immediate broadcast).
func NewApprovalSyncBroadcaster(coalesceWindow time.Duration) *ApprovalSyncBroadcaster {
	return &ApprovalSyncBroadcaster{
		subscribers:    make(map[chan struct{}]struct{}),
		coalesceWindow: coalesceWindow,
	}
}

// Subscribe returns a channel that receives a signal when approval state changes.
// The channel is buffered with size 1 to avoid blocking the broadcaster.
func (b *ApprovalSyncBroadcaster) Subscribe() chan struct{} {
	ch := make(chan struct{}, 1)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

// Unsubscribe removes a subscriber channel and closes it.
func (b *ApprovalSyncBroadcaster) Unsubscribe(ch chan struct{}) {
	b.mu.Lock()
	delete(b.subscribers, ch)
	b.mu.Unlock()
	close(ch)
}

// Broadcast notifies all subscribers that approval state has changed.
// If a coalesce window is configured, the actual notification is delayed
// until the window expires, absorbing any additional broadcasts during that time.
func (b *ApprovalSyncBroadcaster) Broadcast() {
	if b.coalesceWindow <= 0 {
		b.notifyAll()
		return
	}

	b.mu.Lock()
	if b.coalescePending {
		// Already waiting to fire — this broadcast is absorbed
		b.mu.Unlock()
		return
	}
	b.coalescePending = true
	b.coalesceTimer = time.AfterFunc(b.coalesceWindow, func() {
		b.mu.Lock()
		b.coalescePending = false
		b.mu.Unlock()
		b.notifyAll()
	})
	b.mu.Unlock()
}

// notifyAll sends a signal to all current subscribers.
func (b *ApprovalSyncBroadcaster) notifyAll() {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subscribers {
		select {
		case ch <- struct{}{}:
		default:
			// Channel already has a pending signal — skip to avoid blocking
		}
	}
}
