package repo

//go:generate go-enum --values --names --nocase

// EventType
/*
ENUM(
artwork_create
artwork_update
artwork_delete
)
*/
type EventType string // topic of the event

type EventBus[P any] interface {
	Publish(eventType EventType, payload P)
	Subscribe(eventType EventType, handler func(payload P), filter func(payload P) bool) (unsubscribe func())
}
