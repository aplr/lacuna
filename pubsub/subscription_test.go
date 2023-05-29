package pubsub

import "testing"

func TestGetSubscriptionIdSuccess(t *testing.T) {
	subscription := Subscription{
		Service: "payment",
		Name:    "product-created",
	}

	subscriptionId := subscription.GetSubscriptionID()

	if subscriptionId != "payment:product-created" {
		t.Errorf("Expected subscriptionId to be 'payment:product-created, got %s", subscriptionId)
	}
}
