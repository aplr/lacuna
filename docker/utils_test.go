package docker

import (
	"testing"
)

func TestMapValidEventTypeSucceeds(t *testing.T) {
	// arrange
	actions := []string{startEventName, stopEventName}
	eventTypes := []EventType{EVENT_TYPE_START, EVENT_TYPE_STOP}
	extractedEventTypes := make([]EventType, 0)

	// act
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
	// arrange
	action := "foobar"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected extractEventType to panic")
		}
	}()

	// act
	mapEventType(action)
}
