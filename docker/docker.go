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
	startEventName = "start" // TODO: evaluate event
	stopEventName  = "stop"  // TODO: evaluate event
)

type Docker interface {
	Run(ctx context.Context) (<-chan Event, <-chan error)
}

var _ = Docker(&dockerImpl{})

type dockerImpl struct {
	Docker

	labelPrefix string
	log         *log.Entry
	cli         client.APIClient
}

func NewDocker(labelPrefix string) (Docker, error) {
	cli, err := client.NewClientWithOpts(
		client.FromEnv,
		client.WithAPIVersionNegotiation(),
	)

	if err != nil {
		return nil, err
	}

	return NewDockerWithClient(cli, labelPrefix), nil
}

func NewDockerWithClient(cli client.APIClient, labelPrefix string) Docker {
	log := log.WithField("component", "docker")

	return &dockerImpl{
		cli:         cli,
		log:         log,
		labelPrefix: labelPrefix,
	}
}

func (docker *dockerImpl) Run(ctx context.Context) (<-chan Event, <-chan error) {
	messages := make(chan Event)
	errs := make(chan error, 1)

	go func() {
		defer close(messages)
		defer close(errs)

		if err := docker.handleInitialContainers(ctx, messages); err != nil {
			errs <- err
			return
		}

		docker.listenForContainerChanges(ctx, messages, errs)
	}()

	return messages, errs
}

func (docker *dockerImpl) handleInitialContainers(
	ctx context.Context,
	out chan Event,
) error {
	containers, _ := docker.cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "label", Value: docker.filterLabel()},
		),
	})

	for _, c := range containers {
		container := NewContainer(c.ID, c.Labels)
		docker.handleContainer(ctx, EVENT_TYPE_START, container, out)
	}

	return nil
}

func (docker *dockerImpl) listenForContainerChanges(ctx context.Context, messages chan Event, errs chan error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msgChannel, errChannel := docker.cli.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "type", Value: "container"},
			filters.KeyValuePair{Key: "label", Value: docker.filterLabel()},
			filters.KeyValuePair{Key: "event", Value: startEventName},
			filters.KeyValuePair{Key: "event", Value: stopEventName},
		),
	})

	for {
		select {
		case <-ctx.Done():
			// Cancel the listener and return
			errs <- ctx.Err()
			return
		case msg := <-msgChannel:
			// Publish messages to channel
			docker.handleMessage(ctx, msg, messages)
		case err := <-errChannel:
			// Log errors and silently return from the listener
			docker.handleError(ctx, err, errs)
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

func (docker *dockerImpl) handleError(ctx context.Context, err error, errs chan error) {
	// TODO: handle error
	docker.log.WithError(err).Error("error received")
	errs <- err
}

func (docker *dockerImpl) filterLabel() string {
	return docker.labelPrefix + ".enabled=true"
}
