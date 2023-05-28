package docker

import (
	"sort"
	"testing"
)

func TestExtractSubscriptionsSucceedsWithoutLabels(t *testing.T) {
	// setup
	labels := map[string]string{}

	// execute
	subscriptions := extractSubscriptions("test", labels)

	// assert
	if len(subscriptions) != 0 {
		t.Errorf("expected 0 subscriptions, got %d", len(subscriptions))
	}
}

func TestExtractSubscriptionsExtractsValidSubscriptions(t *testing.T) {
	// setup
	labels := map[string]string{
		"pubsub.subscription.test.topic":    "test-topic",
		"pubsub.subscription.test.endpoint": "/messages",
	}

	// execute
	subscriptions := extractSubscriptions("test", labels)

	// assert
	if len(subscriptions) != 1 {
		t.Errorf("expected 1 subscription, got %d", len(subscriptions))
	}

	if subscriptions[0].Name != "test" {
		t.Errorf("expected subscription name to be 'test', got '%s'", subscriptions[0].Name)
	}

	if subscriptions[0].Topic != "test-topic" {
		t.Errorf("expected subscription topic to be 'test-topic', got '%s'", subscriptions[0].Topic)
	}

	if subscriptions[0].Endpoint != "/messages" {
		t.Errorf("expected subscription endpoint to be '/messages', got '%s'", subscriptions[0].Endpoint)
	}
}

func TestExtractSubscriptionsExtractsMultipleSubscriptions(t *testing.T) {
	// setup
	labels := map[string]string{
		"pubsub.subscription.test-1.topic":    "test-topic-1",
		"pubsub.subscription.test-1.endpoint": "/messages",
		"pubsub.subscription.test-2.topic":    "test-topic-2",
		"pubsub.subscription.test-2.endpoint": "/messages",
	}

	// execute
	subscriptions := extractSubscriptions("test", labels)

	// for assertion to succeed despite order of subscriptions
	// we have to sort the subscriptions by name
	sort.SliceStable(subscriptions, func(i, j int) bool {
		return subscriptions[i].Name < subscriptions[j].Name
	})

	// assert
	if len(subscriptions) != 2 {
		t.Errorf("expected 2 subscriptions, got %d", len(subscriptions))
	}

	if subscriptions[0].Name != "test-1" {
		t.Errorf("expected subscription name to be 'test-1', got '%s'", subscriptions[0].Name)
	}

	if subscriptions[1].Name != "test-2" {
		t.Errorf("expected subscription name to be 'test-2', got '%s'", subscriptions[0].Name)
	}
}

func TestExtractSubscriptionSkipsIncompleteSubscriptions(t *testing.T) {
	// setup
	labels := map[string]string{
		"pubsub.subscription.test.topic": "test-topic",
	}

	// execute
	subscriptions := extractSubscriptions("test", labels)

	// assert
	if len(subscriptions) != 0 {
		t.Errorf("expected 0 subscription, got %d", len(subscriptions))
	}
}

func TestExtractSubscriptionsSkipsUnknownLabels(t *testing.T) {
	// setup
	labels := map[string]string{
		"name":           "foobar",
		"my.other.label": "other-label",
	}

	// execute
	subscriptions := extractSubscriptions("test", labels)

	// assert
	if len(subscriptions) != 0 {
		t.Errorf("expected 0 subscription, got %d", len(subscriptions))
	}
}

func TestExtractSubscriptionsSkipsInvalidLabels(t *testing.T) {
	// setup
	labels := map[string]string{
		"pubsub.subscription.my_name.topic":    "invalid-name",
		"pubsub.subscription.my_name.endpoint": "invalid-name",
		"pubsub.subscription.test.foobar":      "invalid-field",
		"pubsub.subscription.x.y.z":            "invalid-key",
		"pubsub.subscription.x":                "invalid-key",
	}

	// execute
	subscriptions := extractSubscriptions("test", labels)

	// assert
	if len(subscriptions) != 0 {
		t.Errorf("expected 0 subscription, got %d", len(subscriptions))
	}
}

func TestExtractValidEventTypeSucceeds(t *testing.T) {
	// setup
	actions := []string{subscribeEventName, unsubscribeEventName}
	eventTypes := []EventType{EVENT_TYPE_SUBSCRIBE, EVENT_TYPE_UNSUBSCRIBE}
	extractedEventTypes := make([]EventType, 0)

	// execute
	for _, action := range actions {
		extractedEventTypes = append(extractedEventTypes, extractEventType(action))
	}

	// assert
	for i, eventType := range eventTypes {
		if extractedEventTypes[i] != eventType {
			t.Errorf("expected event type '%s', got '%s'", eventType, extractedEventTypes[i])
		}
	}
}

func TestExtractInvalidEventTypeFails(t *testing.T) {
	// setup
	action := "foobar"
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("expected extractEventType to panic")
		}
	}()

	// execute
	extractEventType(action)
}

func TestExtractServiceNameReturnsDockerComposeName(t *testing.T) {
	// setup
	labels := map[string]string{
		"com.docker.compose.service":          "service",
		"com.docker.compose.project":          "project",
		"com.docker.compose.container-number": "1",
	}

	// execute
	serviceName := extractServiceName("2", labels)

	// assert
	if serviceName != "project_service_1" {
		t.Errorf("expected service name to be 'project_service_1', got '%s'", serviceName)
	}
}

func TestExtractServiceNameReturnsCommonName(t *testing.T) {
	// setup
	labels := map[string]string{
		"org.opencontainers.image.title": "service",
	}

	// execute
	serviceName := extractServiceName("1", labels)

	// assert
	if serviceName != "service-1" {
		t.Errorf("expected service name to be 'service-1', got '%s'", serviceName)
	}
}
