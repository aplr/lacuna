package app

import (
	"sort"
	"testing"

	"github.com/aplr/lacuna/docker"
)

func TestExtractSubscriptionsSucceedsWithoutLabels(t *testing.T) {
	// arrange
	container := docker.NewContainer("1", map[string]string{})

	// act
	subscriptions := extractSubscriptions(container)

	// assert
	if len(subscriptions) != 0 {
		t.Errorf("expected 0 subscriptions, got %d", len(subscriptions))
	}
}

func TestExtractSubscriptionsExtractsValidSubscriptions(t *testing.T) {
	// arrange
	container := docker.NewContainer("1", map[string]string{
		"lacuna.subscription.test.topic":    "test-topic",
		"lacuna.subscription.test.endpoint": "/messages",
	})

	// act
	subscriptions := extractSubscriptions(container)

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
	// arrange
	container := docker.NewContainer("1", map[string]string{
		"lacuna.subscription.test-1.topic":    "test-topic-1",
		"lacuna.subscription.test-1.endpoint": "/messages",
		"lacuna.subscription.test-2.topic":    "test-topic-2",
		"lacuna.subscription.test-2.endpoint": "/messages",
	})

	// act
	subscriptions := extractSubscriptions(container)

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
	// arrange
	container := docker.NewContainer("1", map[string]string{
		"lacuna.subscription.test.topic": "test-topic",
	})

	// act
	subscriptions := extractSubscriptions(container)

	// assert
	if len(subscriptions) != 0 {
		t.Errorf("expected 0 subscription, got %d", len(subscriptions))
	}
}

func TestExtractSubscriptionsSkipsUnknownLabels(t *testing.T) {
	// arrange
	container := docker.NewContainer("1", map[string]string{
		"name":           "foobar",
		"my.other.label": "other-label",
	})

	// act
	subscriptions := extractSubscriptions(container)

	// assert
	if len(subscriptions) != 0 {
		t.Errorf("expected 0 subscription, got %d", len(subscriptions))
	}
}

func TestExtractSubscriptionsSkipsInvalidLabels(t *testing.T) {
	// arrange
	container := docker.NewContainer("1", map[string]string{
		"lacuna.subscription.my_name.topic":    "invalid-name",
		"lacuna.subscription.my_name.endpoint": "invalid-name",
		"lacuna.subscription.test.foobar":      "invalid-field",
		"lacuna.subscription.x.y.z":            "invalid-key",
		"lacuna.subscription.x":                "invalid-key",
	})

	// act
	subscriptions := extractSubscriptions(container)

	// assert
	if len(subscriptions) != 0 {
		t.Errorf("expected 0 subscription, got %d", len(subscriptions))
	}
}
