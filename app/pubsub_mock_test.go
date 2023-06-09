package app

import (
	"context"

	"github.com/aplr/lacuna/pubsub"
)

var _ = pubsub.PubSub(&mockPubSub{})

type mockPubSub struct {
	pubsub.PubSub

	createSubscription func(ctx context.Context, subscription pubsub.Subscription) error
	deleteSubscription func(ctx context.Context, subscription pubsub.Subscription) error
}

func (ps *mockPubSub) CreateSubscription(ctx context.Context, subscription pubsub.Subscription) error {
	if ps.createSubscription == nil {
		panic("no mock function provided")
	}

	return ps.createSubscription(ctx, subscription)
}

func (ps *mockPubSub) DeleteSubscription(ctx context.Context, subscription pubsub.Subscription) error {
	if ps.deleteSubscription == nil {
		panic("no mock function provided")
	}

	return ps.deleteSubscription(ctx, subscription)
}
