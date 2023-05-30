package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

type mockDocker struct {
	client.APIClient

	containerList func(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error)
	events        func(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error)
}

func (d *mockDocker) ContainerList(ctx context.Context, options types.ContainerListOptions) ([]types.Container, error) {
	if d.containerList == nil {
		panic("no mock function provided")
	}

	return d.containerList(ctx, options)
}

func (d *mockDocker) Events(ctx context.Context, options types.EventsOptions) (<-chan events.Message, <-chan error) {
	if d.events == nil {
		panic("no mock function provided")
	}

	return d.events(ctx, options)
}

var _ client.APIClient = client.APIClient(&mockDocker{})
