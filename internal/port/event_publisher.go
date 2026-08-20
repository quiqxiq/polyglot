package port

// EventPublisher abstracts real-time broadcasting.
type EventPublisher interface {
	PublishEvent(eventType string, data any)
}
