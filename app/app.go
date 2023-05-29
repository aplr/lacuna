package app

import (
	"context"

	"github.com/aplr/pubsub-emulator/docker"
	"github.com/aplr/pubsub-emulator/pubsub"
	log "github.com/sirupsen/logrus"
)

type App struct {
	log *log.Entry
}

func NewApp() *App {
	log := log.WithField("component", "app")

	return &App{
		log: log,
	}
}

func (app *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	docker, err := docker.NewDocker()

	if err != nil {
		return err
	}

	// TODO: project id, even though it's not used by the emulator
	_, err = pubsub.NewPubSub(ctx, "project-id")

	if err != nil {
		return err
	}

	// TODO: events
	events, err := docker.Run(ctx)

	if err != nil {
		return err
	}

out:
	for {
		select {
		case <-ctx.Done():
			break out
		case evt := <-events:
			app.log.WithField("event", evt).Info("event received")
			// pubsub.CreateSubscription(ctx, evt.Subscription)
		}
	}

	return nil
}
