package app

import (
	"context"

	"github.com/aplr/lacuna/docker"
)

var _ = docker.Docker(&mockDocker{})

type mockDocker struct {
	docker.Docker

	run func(ctx context.Context) (<-chan docker.Event, <-chan error)
}

func (d *mockDocker) Run(ctx context.Context) (<-chan docker.Event, <-chan error) {
	if d.run == nil {
		panic("no mock function provided")
	}

	return d.run(ctx)
}
