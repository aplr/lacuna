package docker

import (
	"fmt"
)

func mapEventType(action string) EventType {
	switch action {
	case startEventName:
		return EVENT_TYPE_START
	case stopEventName:
		return EVENT_TYPE_STOP
	default:
		panic(fmt.Sprintf("unsupported event type: %s", action))
	}
}
