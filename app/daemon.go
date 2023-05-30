package app

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	log "github.com/sirupsen/logrus"
)

type Daemon struct {
	app *App
}

func NewDaemon(ctx context.Context) *Daemon {
	app := NewDefaultApp(ctx)

	return &Daemon{
		app: app,
	}
}

func (d *Daemon) Run(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		err := d.app.Run(ctx)

		if err != nil {
			log.Fatal(err)
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

	wg.Wait()
}
