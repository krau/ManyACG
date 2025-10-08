package eventbus

import (
	"github.com/duke-git/lancet/v2/eventbus"
	"github.com/krau/ManyACG/internal/repo"
)

type EventBus[P any] struct {
	bus *eventbus.EventBus[P]
}

// Publish implements repo.EventBus.
func (e *EventBus[P]) Publish(eventType repo.EventType, payload P) {
	e.bus.Publish(eventbus.Event[P]{
		Topic:   string(eventType),
		Payload: payload,
	})
}

// Subscribe implements repo.EventBus.
func (e *EventBus[P]) Subscribe(eventType repo.EventType, handler func(payload P), filter func(payload P) bool) (unsubscribe func()) {
	e.bus.Subscribe(string(eventType), handler, true, 1, filter)
	unsub := func() {
		e.bus.Unsubscribe(string(eventType), handler)
	}
	return unsub
}

func New[P any]() *EventBus[P] {
	return &EventBus[P]{
		bus: eventbus.NewEventBus[P](),
	}
}
