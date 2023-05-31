package app

import (
	"context"
	"errors"
	"testing"
	"time"

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

func TestNewDefaultAppCreatesNewApp(t *testing.T) {
	app := NewDefaultApp(context.Background())

	if app.docker == nil {
		t.Errorf("Expected docker to be non-nil, got %v", app.docker)
	}

	if app.pubsub == nil {
		t.Errorf("Expected pubsub to be non-nil, got %v", app.pubsub)
	}
}

func TestRunClosesOnContextCancel(t *testing.T) {
	// arrange
	docker := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			return make(chan docker.Event), make(chan error, 1)
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
	// arrange
	docker := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			errs := make(chan error, 1)
			go func() {
				errs <- errors.New("test error")
			}()
			return nil, errs
		},
	}
	pubsub := &mockPubSub{}

	app := NewApp(docker, pubsub)

	// act
	err := app.Run(context.Background())

	// assert
	if err == nil {
		t.Errorf("Expected error to be non-nil, got %v", err)
	}
}

func TestRunHandlesContainerStartEvent(t *testing.T) {
	// arrange
	events := make(chan docker.Event)
	subscriptions := make(chan pubsub.Subscription)
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			return events, make(chan error, 1)
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
		if err := app.Run(ctx); err != nil {
			t.Errorf("Expected error to be nil, got %v", err)
		}
	}()

	// act
	events <- docker.Event{
		Type: docker.EVENT_TYPE_START,
		Container: docker.NewContainer("1", map[string]string{
			"lacuna.subscription.test.topic":    "test",
			"lacuna.subscription.test.endpoint": "/messages",
		}),
	}

	// assert
	select {
	case <-ctx.Done():
		t.Errorf("Expected context to not be done")
	case subscription := <-subscriptions:
		if subscription.Topic != "test" {
			t.Errorf("Expected topic to be 'test', got %v", subscription.Topic)
		}
	}
}

func TestRunHandlesContainerStopEvent(t *testing.T) {
	// arrange
	events := make(chan docker.Event)
	subscriptions := make(chan pubsub.Subscription)
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			return events, make(chan error, 1)
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
		if err := app.Run(ctx); err != nil {
			t.Errorf("Expected error to be nil, got %v", err)
		}
	}()

	// act
	events <- docker.Event{
		Type: docker.EVENT_TYPE_STOP,
		Container: docker.NewContainer("1", map[string]string{
			"lacuna.subscription.test.topic":    "test",
			"lacuna.subscription.test.endpoint": "/messages",
		}),
	}

	// assert
	select {
	case <-ctx.Done():
		t.Errorf("Expected context to not be done")
	case subscription := <-subscriptions:
		if subscription.Topic != "test" {
			t.Errorf("Expected topic to be 'test', got %v", subscription.Topic)
		}
	}
}

func TestRunHandlesNoSubscriptions(t *testing.T) {
	// arrange
	events := make(chan docker.Event)
	subscriptions := make(chan pubsub.Subscription)
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			return events, make(chan error, 1)
		},
	}
	p := &mockPubSub{
		createSubscription: func(ctx context.Context, subscription pubsub.Subscription) error {
			subscriptions <- subscription
			return nil
		},
	}

	app := NewApp(d, p)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		if err := app.Run(ctx); err != nil {
			t.Errorf("Expected error to be nil, got %v", err)
		}
	}()

	// act
	events <- docker.Event{
		Type:      docker.EVENT_TYPE_START,
		Container: docker.NewContainer("1", map[string]string{}),
	}

	// assert
	select {
	case <-ctx.Done():
		// we expect a timeout to happen, as no events or errors should be sent
		return
	case <-subscriptions:
		t.Errorf("Expected no subscriptions to be created")
	}
}

func TestRunHandlesCreateSubascriptionError(t *testing.T) {
	// arrange
	events := make(chan docker.Event)
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			return events, make(chan error, 1)
		},
	}
	p := &mockPubSub{
		createSubscription: func(ctx context.Context, subscription pubsub.Subscription) error {
			return errors.New("create subscription failed")
		},
	}

	app := NewApp(d, p)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		if err := app.Run(ctx); err != nil {
			t.Errorf("Expected error to be nil, got %v", err)
		}
	}()

	// act
	events <- docker.Event{
		Type: docker.EVENT_TYPE_START,
		Container: docker.NewContainer("1", map[string]string{
			"lacuna.subscription.test.topic":    "test",
			"lacuna.subscription.test.endpoint": "/messages",
		}),
	}

	// assert
	<-ctx.Done()
}

func TestRunHandlesDeleteSubascriptionError(t *testing.T) {
	// arrange
	events := make(chan docker.Event)
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			return events, make(chan error, 1)
		},
	}
	p := &mockPubSub{
		deleteSubscription: func(ctx context.Context, subscription pubsub.Subscription) error {
			return errors.New("delete subscription failed")
		},
	}

	app := NewApp(d, p)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	go func() {
		if err := app.Run(ctx); err != nil {
			t.Errorf("Expected error to be nil, got %v", err)
		}
	}()

	// act
	events <- docker.Event{
		Type: docker.EVENT_TYPE_STOP,
		Container: docker.NewContainer("1", map[string]string{
			"lacuna.subscription.test.topic":    "test",
			"lacuna.subscription.test.endpoint": "/messages",
		}),
	}

	// assert
	<-ctx.Done()
}
