package pubsub

import (
	"context"

	gcps "cloud.google.com/go/pubsub"
)

type PubSub interface {
	CreateSubscription(ctx context.Context, subscription Subscription) error
	DeleteSubscription(ctx context.Context, subscription Subscription) error
}

type pubSubImpl struct {
	PubSub

	client *gcps.Client
}

func NewPubSub(ctx context.Context, projectId string) (PubSub, error) {
	client, err := gcps.NewClient(ctx, projectId)

	if err != nil {
		return nil, err
	}

	return &pubSubImpl{
		client: client,
	}, nil
}

func (ps *pubSubImpl) CreateSubscription(ctx context.Context, subscription Subscription) error {
	topic := ps.client.Topic(subscription.Topic)

	ps.client.CreateSubscription(ctx, subscription.GetSubscriptionID(), gcps.SubscriptionConfig{
		Topic: topic,
		PushConfig: gcps.PushConfig{
			Endpoint: subscription.Endpoint,
		},
	})
	return nil
}

func (ps *pubSubImpl) DeleteSubscription(ctx context.Context, subscription Subscription) error {
	// TODO: verify, handle error
	ps.client.Subscription(subscription.GetSubscriptionID()).Delete(ctx)
	return nil
}
