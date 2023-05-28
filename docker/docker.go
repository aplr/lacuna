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
	labelPrefix          = "pubsub"
	filterLabel          = labelPrefix + ".enabled=true"
	subscribeEventName   = "start" // TODO: evaluate event
	unsubscribeEventName = "stop"  // TODO: evaluate event
)

type Docker struct {
	log        *log.Entry
	cli        *client.Client
	containers map[string]Container
}

func NewDocker() (*Docker, error) {
	log := log.WithField("component", "docker")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		return nil, err
	}

	return &Docker{
		cli:        cli,
		log:        log,
		containers: map[string]Container{},
	}, nil
}

func (docker *Docker) Run(ctx context.Context) (chan Event, error) {
	out := make(chan Event)
	defer close(out)

	go docker.listenForContainerChanges(ctx, out)

	err := docker.handleInitialContainers(ctx, out)

	if err != nil {
		return nil, err
	}

	return out, nil
}

func (docker *Docker) handleInitialContainers(ctx context.Context, out chan Event) error {
	containers, err := docker.cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "label", Value: filterLabel},
		),
	})

	if err != nil {
		return err
	}

	for _, c := range containers {
		// TODO: verify if container name is valid
		container := NewContainer(c.ID, c.Names[0], c.Labels)
		docker.handleContainer(ctx, EVENT_TYPE_SUBSCRIBE, container, out)
	}

	return nil
}

func (docker *Docker) listenForContainerChanges(ctx context.Context, out chan Event) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	msgChannel, errChannel := docker.cli.Events(ctx, types.EventsOptions{
		Filters: filters.NewArgs(
			filters.KeyValuePair{Key: "type", Value: "container"},
			filters.KeyValuePair{Key: "label", Value: filterLabel},
			filters.KeyValuePair{Key: "event", Value: subscribeEventName},
			filters.KeyValuePair{Key: "event", Value: unsubscribeEventName},
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

func (docker *Docker) handleMessage(ctx context.Context, message events.Message, out chan Event) {
	eventType := extractEventType(message.Action)

	container := NewContainer(
		message.Actor.ID,
		// TODO: verify if container name is valid
		message.Actor.Attributes["name"],
		message.Actor.Attributes,
	)

	docker.handleContainer(ctx, eventType, container, out)

	docker.log.WithField("type", "message").Debug(message.Actor.Attributes["image"], " ", message.Type, " ", message.Action)
}

func (docker *Docker) handleContainer(ctx context.Context, eventType EventType, container Container, out chan Event) {
	subscriptions := extractSubscriptions(container.Name, container.Labels)

	for _, subscription := range subscriptions {
		out <- Event{
			Type:         eventType,
			Container:    container,
			Subscription: subscription,
		}
	}
}

func (docker *Docker) handleError(ctx context.Context, err error) {
	// TODO: handle error
	docker.log.WithField("type", "error").Error(err)
}
