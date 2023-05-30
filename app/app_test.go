package app

import (
	"context"
	"errors"
	"testing"

	"github.com/aplr/lacuna/docker"
	"github.com/aplr/lacuna/pubsub"
)

func TestNewAppCreatesNewApp(t *testing.T) {
	docker := &mockDocker{}
	pubsub := &mockPubSub{}

	app := NewApp(docker, pubsub)

	if app.docker != docker {
		t.Errorf("Expected docker to be %v, got %v", docker, app.docker)
	}

	if app.pubsub != pubsub {
		t.Errorf("Expected pubsub to be %v, got %v", pubsub, app.pubsub)
	}
}

func TestRunClosesOnContextCancel(t *testing.T) {
	// setup
	docker := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, error) {
			return make(chan docker.Event), nil
		},
	}
	pubsub := &mockPubSub{}

	app := NewApp(docker, pubsub)

	ctx, cancel := context.WithCancel(context.Background())

	// act
	cancel()
	err := app.Run(ctx)

	// assert
	if err != nil {
		t.Errorf("Expected error to be nil, got %v", err)
	}
}

func TestRunPropagatesErrorFromDocker(t *testing.T) {
	// setup
	docker := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, error) {
			return nil, errors.New("failed to run docker")
		},
	}
	pubsub := &mockPubSub{}

	app := NewApp(docker, pubsub)

	ctx := context.Background()

	// act
	err := app.Run(ctx)

	// assert
	if err == nil {
		t.Errorf("Expected error to be non-nil, got %v", err)
	}
}

func TestRunHandlesContainerStartEvent(t *testing.T) {
	// setup
	events := make(chan docker.Event)
	subscriptions := make(chan pubsub.Subscription)
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, error) {
			return events, nil
		},
	}
	p := &mockPubSub{
		createSubscription: func(ctx context.Context, subscription pubsub.Subscription) error {
			subscriptions <- subscription
			return nil
		},
	}

	app := NewApp(d, p)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := app.Run(ctx)

		if err != nil {
			t.Errorf("Expected error to be nil, got %v", err)
		}
	}()

	// act
	events <- docker.Event{
		Type: docker.EVENT_TYPE_START,
		Container: docker.NewContainer("1", map[string]string{
			"pubsub.subscription.test.topic":    "test",
			"pubsub.subscription.test.endpoint": "/messages",
		}),
	}

	subscription := <-subscriptions

	// assert
	if subscription.Topic != "test" {
		t.Errorf("Expected topic to be 'test', got %v", subscription.Topic)
	}
}

func TestRunHandlesContainerStopEvent(t *testing.T) {
	// setup
	events := make(chan docker.Event)
	subscriptions := make(chan pubsub.Subscription)
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, error) {
			return events, nil
		},
	}
	p := &mockPubSub{
		deleteSubscription: func(ctx context.Context, subscription pubsub.Subscription) error {
			subscriptions <- subscription
			return nil
		},
	}

	app := NewApp(d, p)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		err := app.Run(ctx)

		if err != nil {
			t.Errorf("Expected error to be nil, got %v", err)
		}
	}()

	// act
	events <- docker.Event{
		Type: docker.EVENT_TYPE_STOP,
		Container: docker.NewContainer("1", map[string]string{
			"pubsub.subscription.test.topic":    "test",
			"pubsub.subscription.test.endpoint": "/messages",
		}),
	}

	subscription := <-subscriptions

	// assert
	if subscription.Topic != "test" {
		t.Errorf("Expected topic to be 'test', got %v", subscription.Topic)
	}
}
