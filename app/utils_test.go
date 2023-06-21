package app

import (
	"sort"
	"testing"
	"time"

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

func TestExtractSubscriptionsExtractsValidSubscriptionOptions(t *testing.T) {
	// arrange
	container := docker.NewContainer("1", map[string]string{
		"lacuna.subscription.test.topic":                             "test-topic",
		"lacuna.subscription.test.endpoint":                          "/messages",
		"lacuna.subscription.test.ack-deadline":                      "10s",
		"lacuna.subscription.test.retain-acked-messages":             "true",
		"lacuna.subscription.test.retention-duration":                "24h",
		"lacuna.subscription.test.enable-ordering":                   "true",
		"lacuna.subscription.test.expiration-ttl":                    "5s",
		"lacuna.subscription.test.filter":                            "foo=bar",
		"lacuna.subscription.test.deliver-exactly-once":              "true",
		"lacuna.subscription.test.dead-letter-topic":                 "dead-letter-topic",
		"lacuna.subscription.test.max-dead-letter-delivery-attempts": "10",
		"lacuna.subscription.test.retry-minimum-backoff":             "10s",
		"lacuna.subscription.test.retry-maximum-backoff":             "10s",
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
		t.Errorf("expected topic to be 'test-topic', got '%s'", subscriptions[0].Topic)
	}

	if subscriptions[0].Endpoint != "/messages" {
		t.Errorf("expected endpoint to be '/messages', got '%s'", subscriptions[0].Endpoint)
	}

	if subscriptions[0].AckDeadline != 10*time.Second {
		t.Errorf("expected ack-deadline to be 10, got '%d'", subscriptions[0].AckDeadline)
	}

	if subscriptions[0].RetainAckedMessages != true {
		t.Errorf("expected retain-acked-messages to be true, got '%t'", subscriptions[0].RetainAckedMessages)
	}

	if subscriptions[0].RetentionDuration != 24*time.Hour {
		t.Errorf("expected retention-duration to be 24h, got '%d'", subscriptions[0].RetentionDuration)
	}

	if subscriptions[0].EnableOrdering != true {
		t.Errorf("expected enable-ordering to be true, got '%t'", subscriptions[0].EnableOrdering)
	}

	if subscriptions[0].ExpirationTTL != 5*time.Second {
		t.Errorf("expected expiration-ttl to be 5s, got '%d'", subscriptions[0].ExpirationTTL)
	}

	if subscriptions[0].Filter != "foo=bar" {
		t.Errorf("expected filter to be 'foo=bar', got '%s'", subscriptions[0].Filter)
	}

	if subscriptions[0].DeliverExactlyOnce != true {
		t.Errorf("expected deliver-exactly-once to be true, got '%t'", subscriptions[0].DeliverExactlyOnce)
	}

	if subscriptions[0].DeadLetterTopic != "dead-letter-topic" {
		t.Errorf("expected dead-letter-topic to be 'dead-letter-topic', got '%s'", subscriptions[0].DeadLetterTopic)
	}

	if subscriptions[0].MaxDeadLetterDeliveryAttempts != 10 {
		t.Errorf("expected max-dead-letter-delivery-attempts to be 10, got '%d'", subscriptions[0].MaxDeadLetterDeliveryAttempts)
	}

	if *subscriptions[0].RetryMinimumBackoff != 10*time.Second {
		t.Errorf("expected retry-minimum-backoff to be 10s, got '%d'", subscriptions[0].RetryMinimumBackoff)
	}

	if *subscriptions[0].RetryMaximumBackoff != 10*time.Second {
		t.Errorf("expected retry-maximum-backoff to be 10s, got '%d'", subscriptions[0].RetryMaximumBackoff)
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

func TestExtractSubscriptionsSkipsInvalidValues(t *testing.T) {
	// arrange
	container := docker.NewContainer("1", map[string]string{
		"lacuna.subscription.test.topic":                             "test",
		"lacuna.subscription.test.endpoint":                          "/messages",
		"lacuna.subscription.test.ack-deadline":                      "invalid",
		"lacuna.subscription.test.retain-acked-messages":             "invalid",
		"lacuna.subscription.test.retention-duration":                "invalid",
		"lacuna.subscription.test.enable-ordering":                   "invalid",
		"lacuna.subscription.test.expiration-ttl":                    "invalid",
		"lacuna.subscription.test.deliver-exactly-once":              "invalid",
		"lacuna.subscription.test.max-dead-letter-delivery-attempts": "invalid",
		"lacuna.subscription.test.retry-minimum-backoff":             "invalid",
		"lacuna.subscription.test.retry-maximum-backoff":             "invalid",
	})

	// act
	subscriptions := extractSubscriptions(container)

	// assert
	if len(subscriptions) != 1 {
		t.Errorf("expected 1 subscription, got %d", len(subscriptions))
	}

	// TODO: check if subscription has default values for invalid fields
}
