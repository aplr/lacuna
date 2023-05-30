package app

import (
	"context"

	"github.com/aplr/lacuna/docker"
	"github.com/aplr/lacuna/pubsub"
	log "github.com/sirupsen/logrus"
)

var (
	labelPrefix = "pubsub"
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

	pubsub, err := pubsub.NewPubSub(ctx, "project-id")

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
			app.handleContainerEvent(ctx, evt)
		}
	}

	return nil
}

func (app *App) handleContainerEvent(ctx context.Context, evt docker.Event) {
	app.log.WithField("event", evt).Debug("event received")

	subscriptions := extractSubscriptions(evt.Container)

	for _, subscription := range subscriptions {
		switch evt.Type {
		case docker.EVENT_TYPE_START:
			app.pubsub.CreateSubscription(ctx, subscription)
		case docker.EVENT_TYPE_STOP:
			app.pubsub.DeleteSubscription(ctx, subscription)
		}
	}
}
