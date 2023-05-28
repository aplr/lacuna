package pubsub

import (
	"context"

	gcps "cloud.google.com/go/pubsub"
	"github.com/aplr/pubsub-emulator/models"
)

type PubSub struct {
	client *gcps.Client
}

func NewPubSub(ctx context.Context, projectId string) (*PubSub, error) {
	client, err := gcps.NewClient(ctx, projectId)

	if err != nil {
		return nil, err
	}

	return &PubSub{
		client: client,
	}, nil
}

func (ps *PubSub) CreateSubscription(ctx context.Context, subscription models.Subscription) {
	topic := ps.client.Topic(subscription.Topic)

	ps.client.CreateSubscription(ctx, subscription.GetSubscriptionID(), gcps.SubscriptionConfig{
		Topic: topic,
		PushConfig: gcps.PushConfig{
			Endpoint: subscription.Endpoint,
		},
	})
}
