package docker

import "github.com/aplr/pubsub-emulator/models"

type EventType string

const (
	EVENT_TYPE_SUBSCRIBE   EventType = "subscribe"
	EVENT_TYPE_UNSUBSCRIBE EventType = "unsubscribe"
)

type Event struct {
	Type         EventType           // subscribe to or unsubscribe from the topic
	Container    Container           // container that was started or stopped
	Subscription models.Subscription // subscription metadata
}
