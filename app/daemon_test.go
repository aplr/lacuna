package app

import (
	"context"
	"errors"
	"syscall"
	"testing"

	"github.com/aplr/lacuna/docker"
)

func TestRunExitsWhenContextCancelled(t *testing.T) {
	// arrange
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			return make(chan docker.Event), make(chan error, 1)
		},
	}
	p := &mockPubSub{}

	app, err := NewApp(d, p)

	if err != nil {
		t.Errorf("Expected err to be nil, got %v", err)
	}

	daemon := NewDaemonWithApp(app)

	ctx, cancel := context.WithCancel(context.Background())

	// act
	cancel()

	daemon.Run(ctx)

	<-ctx.Done()
}

func TestRunExitsOnSigint(t *testing.T) {
	// arrange
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			return make(chan docker.Event), make(chan error, 1)
		},
	}
	p := &mockPubSub{}

	app, err := NewApp(d, p)

	if err != nil {
		t.Errorf("Expected err to be nil, got %v", err)
	}

	daemon := NewDaemonWithApp(app)

	done := make(chan bool)

	go func() {
		daemon.Run(context.Background())
		done <- true
	}()

	// act
	go syscall.Kill(syscall.Getpid(), syscall.SIGINT)

	// assert
	<-done
}

func TestRunExitsOnAppError(t *testing.T) {
	// arrange
	errs := make(chan error, 1)
	d := &mockDocker{
		run: func(ctx context.Context) (<-chan docker.Event, <-chan error) {
			return make(chan docker.Event), errs
		},
	}
	p := &mockPubSub{}

	app, err := NewApp(d, p)

	if err != nil {
		t.Errorf("Expected err to be nil, got %v", err)
	}

	daemon := NewDaemonWithApp(app)

	done := make(chan bool)

	go func() {
		defer func() { done <- true }()
		daemon.Run(context.Background())
	}()

	// act
	go func() { errs <- errors.New("test error") }()

	// assert
	<-done
}
