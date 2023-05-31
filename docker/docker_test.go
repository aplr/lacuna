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
			msgs := make(chan events.Message)
			errs := make(chan error)
			return msgs, errs
		},
	}

	docker := NewDockerWithClient(cli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, errs := docker.Run(ctx)

	select {
	case event := <-events:
		if event.Container.ID != "1" {
			t.Errorf("expected container id to be '1', got '%s'", event.Container.ID)
		}
	case err := <-errs:
		t.Errorf("Run() returned error: %v", err)
	}
}

func TestRunHandlesMessage(t *testing.T) {
	cli := &mockDocker{
		containerList: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
			return []types.Container{}, nil
		},
		events: func(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
			msgs := make(chan events.Message)
			errs := make(chan error)
			go func() {
				msgs <- events.Message{
					Action: "start",
					Actor:  events.Actor{ID: "1", Attributes: map[string]string{}},
				}
			}()
			return msgs, errs
		},
	}

	docker := NewDockerWithClient(cli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, errs := docker.Run(ctx)

	select {
	case event := <-events:
		if event.Container.ID != "1" {
			t.Errorf("expected container id to be '1', got '%s'", event.Container.ID)
		}
	case err := <-errs:
		t.Errorf("Run() returned error: %v", err)
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

	events, errs := docker.Run(ctx)

	cancel()

	select {
	case <-ctx.Done():
		return
	case evt := <-events:
		t.Errorf("Run() returned event: %v", evt)
	case err := <-errs:
		t.Errorf("Run() returned error: %v", err)
	}
}

func TestRunHandlesError(t *testing.T) {
	cli := &mockDocker{
		containerList: func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
			return []types.Container{}, nil
		},
		events: func(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
			msgs := make(chan events.Message)
			errs := make(chan error)
			go func() {
				errs <- errors.New("test error")
			}()
			return msgs, errs
		},
	}

	docker := NewDockerWithClient(cli)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, errs := docker.Run(ctx)

	select {
	case <-ctx.Done():
		t.Errorf("Run() returned context done")
		return
	case evt := <-events:
		t.Errorf("Run() returned event: %v", evt)
	case <-errs:
		return
	}
}
