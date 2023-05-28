package app

import (
	"context"

	"github.com/aplr/pubsub-emulator/docker"
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

	// TODO: events
	_, err = docker.Run(ctx)

	if err != nil {
		return err
	}

	return nil
}
