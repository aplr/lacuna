package pubsub

import "github.com/spf13/viper"

type Config struct {
	ProjectID string `mapstructure:"project_id"`
}

func init() {
	viper.BindEnv("pubsub_project_id")
	viper.SetDefault("pubsub.project_id", "pubsub")
}
