package docker

type Container struct {
	ID     string
	Name   string
	Labels map[string]string
}

func NewContainer(ID string, Name string, Labels map[string]string) Container {
	return Container{ID: ID, Name: Name, Labels: Labels}
}
