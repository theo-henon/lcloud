package plugin

import (
	"context"
	"sync"
)

type EventHandler func(ctx context.Context, event Event)

type EventBus struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[string][]EventHandler),
	}
}

func (b *EventBus) Subscribe(eventType string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[eventType] = append(b.handlers[eventType], handler)
}

func (b *EventBus) Publish(ctx context.Context, event Event) {
	b.mu.RLock()
	handlers := append([]EventHandler(nil), b.handlers[event.Type]...)
	allHandlers := append([]EventHandler(nil), b.handlers["*"]...)
	b.mu.RUnlock()

	dispatch := append(handlers, allHandlers...)
	for _, handler := range dispatch {
		h := handler
		go h(ctx, event)
	}
}
