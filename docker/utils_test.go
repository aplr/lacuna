package docker

import (
	"testing"
)

func TestMapValidEventTypeSucceeds(t *testing.T) {
	// setup
	actions := []string{startEventName, stopEventName}
	eventTypes := []EventType{EVENT_TYPE_START, EVENT_TYPE_STOP}
	extractedEventTypes := make([]EventType, 0)

	// execute
	for _, action := range actions {
		extractedEventTypes = append(extractedEventTypes, mapEventType(action))
	}

	// assert
	for i, eventType := range eventTypes {
		if extractedEventTypes[i] != eventType {
			t.Errorf("expected event type '%s', got '%s'", eventType, extractedEventTypes[i])
		}
	}
}

func TestMapInvalidEventTypeFails(t *testing.T) {
	// setup
	action := "foobar"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected extractEventType to panic")
		}
	}()

	// execute
	mapEventType(action)
}
