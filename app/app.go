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
	config *Config
	docker docker.Docker
	pubsub pubsub.PubSub
}

func NewApp(docker docker.Docker, pubsub pubsub.PubSub) (*App, error) {
	log := log.WithField("component", "app")

	config, err := GetConfig()

	if err != nil {
		return nil, err
	}

	return &App{
		log:    log,
		config: config,
		docker: docker,
		pubsub: pubsub,
	}, nil
}

func NewDefaultApp(ctx context.Context) (*App, error) {
	app, err := NewApp(nil, nil)

	if err != nil {
		log.Fatal(err)
	}

	docker, err := docker.NewDocker(app.config.LabelPrefix)

	if err != nil {
		log.Fatal(err)
	}

	app.docker = docker

	pubsub, err := pubsub.NewPubSub(ctx, app.config.PubSub)

	if err != nil {
		log.Fatal(err)
	}

	app.pubsub = pubsub

	return app, nil
}

func (app *App) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	events, errs := app.docker.Run(ctx)

out:
	for {
		select {
		case <-ctx.Done():
			break out
		case err := <-errs:
			return err
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
		if err := app.processSubscription(ctx, subscription, evt); err != nil {
			// don't propagate errors, just log them
			log.WithError(err).Error("failed to process subscription")
		}
	}
}

func (app *App) processSubscription(ctx context.Context, subscription pubsub.Subscription, evt docker.Event) error {
	log := app.log.WithField("container", evt.Container.Name()).WithField("subscription", subscription.Name).WithField("topic", subscription.Topic)

	ctx, cancel := context.WithTimeout(ctx, 5000*time.Millisecond)
	defer cancel()

	switch evt.Type {
	case docker.EVENT_TYPE_START:
		if err := app.pubsub.CreateSubscription(ctx, subscription); err != nil {
			return err
		}
		log.Info("subscription created")
	case docker.EVENT_TYPE_STOP:
		if err := app.pubsub.DeleteSubscription(ctx, subscription); err != nil {
			return err
		}
		log.Info("subscription removed")
	}

	return nil
}
