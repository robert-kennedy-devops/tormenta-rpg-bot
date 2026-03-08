package eventbus

import (
	"sync"
)

// ─── Handler ──────────────────────────────────────────────────────────────────

// Handler is a function that processes a published event.
type Handler func(e Event)

// ─── Bus ──────────────────────────────────────────────────────────────────────

// Bus is a thread-safe publish/subscribe event bus.
// Events are dispatched asynchronously to all registered handlers.
type Bus struct {
	mu       sync.RWMutex
	handlers map[Kind][]Handler
	queue    chan Event
	workers  int
	done     chan struct{}
}

// DefaultWorkers is the number of consumer goroutines.
const DefaultWorkers = 8

// New creates a new Bus with the specified queue depth and worker count.
func New(queueDepth, workers int) *Bus {
	if workers <= 0 {
		workers = DefaultWorkers
	}
	b := &Bus{
		handlers: make(map[Kind][]Handler),
		queue:    make(chan Event, queueDepth),
		workers:  workers,
		done:     make(chan struct{}),
	}
	b.start()
	return b
}

// Global is the singleton event bus used by all game systems.
var Global = New(10_000, DefaultWorkers)

// start launches the consumer goroutine pool.
func (b *Bus) start() {
	for i := 0; i < b.workers; i++ {
		go b.consume()
	}
}

func (b *Bus) consume() {
	for {
		select {
		case e := <-b.queue:
			b.dispatch(e)
		case <-b.done:
			return
		}
	}
}

func (b *Bus) dispatch(e Event) {
	b.mu.RLock()
	handlers := make([]Handler, len(b.handlers[e.Kind]))
	copy(handlers, b.handlers[e.Kind])
	wildcards := make([]Handler, len(b.handlers["*"]))
	copy(wildcards, b.handlers["*"])
	b.mu.RUnlock()

	for _, h := range handlers {
		safeCall(h, e)
	}
	for _, h := range wildcards {
		safeCall(h, e)
	}
}

// Publish enqueues an event for async delivery.
// Non-blocking: if the queue is full the event is dropped (log internally if needed).
func (b *Bus) Publish(e Event) {
	select {
	case b.queue <- e:
	default:
		// Queue full — drop event (metrics should track this in production)
	}
}

// Subscribe registers a handler for a specific event kind.
// Use kind "*" to receive all events.
func (b *Bus) Subscribe(kind Kind, h Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[kind] = append(b.handlers[kind], h)
}

// Unsubscribe removes ALL handlers for a given kind.
func (b *Bus) Unsubscribe(kind Kind) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.handlers, kind)
}

// Stop shuts down the consumer pool.  In-flight events may be dropped.
func (b *Bus) Stop() {
	close(b.done)
}

// safeCall invokes h(e) and recovers from any panic to avoid crashing the worker.
func safeCall(h Handler, e Event) {
	defer func() { recover() }() //nolint:errcheck
	h(e)
}

// ─── Convenience publish helpers ──────────────────────────────────────────────

// Pub is shorthand for Global.Publish.
func Pub(e Event) { Global.Publish(e) }

// Sub is shorthand for Global.Subscribe.
func Sub(kind Kind, h Handler) { Global.Subscribe(kind, h) }
