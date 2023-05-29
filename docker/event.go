package docker

type EventType string

const (
	EVENT_TYPE_START EventType = "start"
	EVENT_TYPE_STOP  EventType = "stop"
)

type Event struct {
	Type      EventType // start or stop
	Container Container // container that was started or stopped
}
