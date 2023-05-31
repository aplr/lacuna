package app

type Config struct {
	LabelPrefix string
	PubSub      PubSubConfig
}

type PubSubConfig struct {
	Host string
	Port int
}
