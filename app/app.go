package app

import (
	"context"
	"time"

	"github.com/aplr/lacuna/docker"
	"github.com/aplr/lacuna/pubsub"
	log "github.com/sirupsen/logrus"
)

var (
	labelPrefix = "lacuna"
)

type App struct {
	log    *log.Entry
	docker docker.Docker
	pubsub pubsub.PubSub
}

func NewApp(docker docker.Docker, pubsub pubsub.PubSub) *App {
	log := log.WithField("component", "app")

	return &App{
		log:    log,
		docker: docker,
		pubsub: pubsub,
	}
}

func NewDefaultApp(ctx context.Context) *App {
	docker, err := docker.NewDocker()

	if err != nil {
		log.Fatal(err)
	}

	pubsub, err := pubsub.NewPubSub(ctx, "pubsub")

	if err != nil {
		log.Fatal(err)
	}

	return NewApp(docker, pubsub)
}

func (app *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	events, err := app.docker.Run(ctx)

	if err != nil {
		return err
	}

out:
	for {
		select {
		case <-ctx.Done():
			break out
		case evt := <-events:
			go app.handleContainerEvent(ctx, evt)
		}
	}

	return nil
}

func (app *App) handleContainerEvent(ctx context.Context, evt docker.Event) {
	log := app.log.WithField("event_type", evt.Type).WithField("container", evt.Container.Name())

	subscriptions := extractSubscriptions(evt.Container)

	if (len(subscriptions)) == 0 {
		log.Warn("no subscriptions found")
		return
	}

	log.Debugf("processing %d subscriptions", len(subscriptions))

	for _, subscription := range subscriptions {
		app.processSubscription(ctx, subscription, evt.Type)
	}
}

func (app *App) processSubscription(ctx context.Context, subscription pubsub.Subscription, eventType docker.EventType) {
	ctx, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
	defer cancel()

	switch eventType {
	case docker.EVENT_TYPE_START:
		if err := app.pubsub.CreateSubscription(ctx, subscription); err != nil {
			app.log.WithError(err).Error("failed to create subscription")
		}
	case docker.EVENT_TYPE_STOP:
		if err := app.pubsub.DeleteSubscription(ctx, subscription); err != nil {
			app.log.WithError(err).Error("failed to delete subscription")
		}
	}
}
