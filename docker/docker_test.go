package docker

import (
	"context"
	"errors"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

// TODO: could not work as docker might not be installed in the test execution environment
func TestNewDockerReturnsDefaultClient(t *testing.T) {
	_, err := NewDocker()

	if err != nil {
		t.Errorf("NewDocker() returned error: %v", err)
	}
}

func TestNewDockerWithClientReturnsClient(t *testing.T) {
	client, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		t.Errorf("client.NewClientWithOpts() returned error: %v", err)
	}

	cli := NewDockerWithClient(client)

	if cli == nil {
		t.Errorf("NewDockerWithClient() returned nil")
	}
}

func TestRunReturnsInitialContainers(t *testing.T) {
	cli := &mockDocker{
		containerList: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
			return []types.Container{{ID: "1"}}, nil
		},
		events: func(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
			msg := make(chan events.Message)
			err := make(chan error)
			return msg, err
		},
	}

	docker := NewDockerWithClient(cli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := docker.Run(ctx)

	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	event := <-events

	if event.Container.ID != "1" {
		t.Errorf("expected container id to be '1', got '%s'", event.Container.ID)
	}
}

func TestRunHandlesMessage(t *testing.T) {
	cli := &mockDocker{
		containerList: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
			return []types.Container{}, nil
		},
		events: func(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
			msg := make(chan events.Message)
			err := make(chan error)
			go func() {
				msg <- events.Message{
					Action: "start",
					Actor:  events.Actor{ID: "1", Attributes: map[string]string{}},
				}
			}()
			return msg, err
		},
	}

	docker := NewDockerWithClient(cli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := docker.Run(ctx)

	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	event := <-events

	if event.Container.ID != "1" {
		t.Errorf("expected container id to be '1', got '%s'", event.Container.ID)
	}
}

func TestRunHandlesContextClose(t *testing.T) {
	cli := &mockDocker{
		containerList: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
			return []types.Container{}, nil
		},
		events: func(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
			msg := make(chan events.Message)
			err := make(chan error)
			return msg, err
		},
	}

	docker := NewDockerWithClient(cli)

	ctx, cancel := context.WithCancel(context.Background())

	events, err := docker.Run(ctx)

	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	cancel()

	select {
	case <-ctx.Done():
		return
	case <-events:
		return
	}
}

func TestRunHandlesError(t *testing.T) {
	cli := &mockDocker{
		containerList: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
			return []types.Container{}, nil
		},
		events: func(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
			msg := make(chan events.Message)
			err := make(chan error)
			go func() {
				err <- errors.New("test error")
			}()
			return msg, err
		},
	}

	docker := NewDockerWithClient(cli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := docker.Run(ctx)

	if err != nil {
		t.Errorf("Run() returned error: %v", err)
	}

	<-events
}
