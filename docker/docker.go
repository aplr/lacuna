package docker

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"

	log "github.com/sirupsen/logrus"
)

var (
	labelPrefix    = "lacuna"
	filterLabel    = labelPrefix + ".enabled=true"
	startEventName = "start" // TODO: evaluate event
	stopEventName  = "stop"  // TODO: evaluate event
)

type Docker interface {
	Run(ctx context.Context) (<-chan Event, error)
}

var _ = Docker(&dockerImpl{})

type dockerImpl struct {
	Docker

	log *log.Entry
	cli client.APIClient
}

func NewDocker() (Docker, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		return nil, err
	}

	return NewDockerWithClient(cli), nil
}

func NewDockerWithClient(cli client.APIClient) Docker {
	log := log.WithField("component", "docker")

	return &dockerImpl{
		cli: cli,
		log: log,
	}
}

func (docker *dockerImpl) Run(ctx context.Context) (<-chan Event, error) {
	out := make(chan Event)

	go func() {
		defer close(out)

		docker.handleInitialContainers(ctx, out)
		docker.listenForContainerChanges(ctx, out)
	}()

	return out, nil
}

func (docker *dockerImpl) handleInitialContainers(
	ctx context.Context,
	out chan Event,
) error {
	containers, _ := docker.cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "label", Value: filterLabel},
		),
	})

	for _, c := range containers {
		container := NewContainer(c.ID, c.Labels)
		docker.handleContainer(ctx, EVENT_TYPE_START, container, out)
	}

	return nil
}

func (docker *dockerImpl) listenForContainerChanges(ctx context.Context, out chan Event) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msgChannel, errChannel := docker.cli.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "type", Value: "container"},
			filters.KeyValuePair{Key: "label", Value: filterLabel},
			filters.KeyValuePair{Key: "event", Value: startEventName},
			filters.KeyValuePair{Key: "event", Value: stopEventName},
		),
	})

	for {
		select {
		case msg := <-msgChannel:
			docker.handleMessage(ctx, msg, out)
		case err := <-errChannel:
			docker.handleError(ctx, err)
			return
		}
	}
}

func (docker *dockerImpl) handleMessage(
	ctx context.Context,
	message events.Message,
	out chan Event,
) {
	eventType := mapEventType(message.Action)

	container := NewContainer(
		message.Actor.ID,
		message.Actor.Attributes,
	)

	docker.handleContainer(ctx, eventType, container, out)
}

func (docker *dockerImpl) handleContainer(
	ctx context.Context,
	eventType EventType,
	container Container,
	out chan Event,
) {
	docker.log.WithField("event", eventType).WithField("container", container.Name()).Debug("processing event")

	out <- Event{
		Type:      eventType,
		Container: container,
	}
}

func (docker *dockerImpl) handleError(ctx context.Context, err error) {
	// TODO: handle error
	docker.log.WithError(err).Error("error received")
}
