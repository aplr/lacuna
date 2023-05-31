package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type Daemon struct {
	app *App
}

func NewDaemon(ctx context.Context) (*Daemon, error) {
	app, err := NewDefaultApp(ctx)

	if err != nil {
		return nil, err
	}

	return NewDaemonWithApp(app), nil
}

func NewDaemonWithApp(app *App) *Daemon {
	return &Daemon{
		app: app,
	}
}

func (d *Daemon) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	done := make(chan bool)

	go func() {
		defer func() { done <- true }()
		defer cancel()

		err := d.app.Run(ctx)

		if err != nil {
			log.WithError(err).Error("error running lacuna")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case <-quit:
		cancel()
	case <-ctx.Done():
	}

	log.Info("Shutting down lacuna...")

	<-done
}
