package docker

import "testing"

func TestExtractServiceNameReturnsDockerComposeName(t *testing.T) {
	// arrange
	container := NewContainer("2", map[string]string{
		"com.docker.compose.service":          "service",
		"com.docker.compose.project":          "project",
		"com.docker.compose.container-number": "1",
	})

	// act
	serviceName := container.Name()

	// assert
	if serviceName != "project-service-1" {
		t.Errorf("expected service name to be 'project-service-1', got '%s'", serviceName)
	}
}

func TestExtractServiceNameReturnsCommonName(t *testing.T) {
	// arrange
	container := NewContainer("1", map[string]string{
		"org.opencontainers.image.title": "service",
	})

	// act
	serviceName := container.Name()

	// assert
	if serviceName != "service-1" {
		t.Errorf("expected service name to be 'service-1', got '%s'", serviceName)
	}
}

func TestExtractServiceNameReturnsId(t *testing.T) {
	// arrange
	container := NewContainer("1", map[string]string{})

	// act
	serviceName := container.Name()

	// assert
	if serviceName != "1" {
		t.Errorf("expected service name to be '1', got '%s'", serviceName)
	}
}
