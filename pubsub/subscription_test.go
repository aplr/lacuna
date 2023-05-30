package pubsub

import "testing"

func TestGetSubscriptionId(t *testing.T) {
	subscription := Subscription{
		Service: "payment",
		Name:    "product-created",
	}

	subscriptionId := subscription.GetSubscriptionID()

	if subscriptionId != "payment_product-created" {
		t.Errorf("Expected subscriptionId to be 'payment_product-created, got %s", subscriptionId)
	}
}
