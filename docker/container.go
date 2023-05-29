package docker

import "strings"

type Container struct {
	ID     string
	Labels map[string]string
}

func NewContainer(ID string, Labels map[string]string) Container {
	return Container{ID: ID, Labels: Labels}
}

func (container *Container) Name() string {
	if _, ok := container.Labels["com.docker.compose.project"]; ok {
		return strings.Join([]string{
			container.Labels["com.docker.compose.project"],
			container.Labels["com.docker.compose.service"],
			container.Labels["com.docker.compose.container-number"],
		}, "-")
	}

	return strings.Join([]string{
		container.Labels["org.opencontainers.image.title"],
		container.ID,
	}, "-")
}
