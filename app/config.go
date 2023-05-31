package app

import (
	"io/fs"

	"github.com/aplr/lacuna/pubsub"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Config struct {
	LabelPrefix string         `mapstructure:"label_prefix"`
	PubSub      *pubsub.Config `mapstructure:"pubsub"`
}

func init() {
	viper.BindEnv("label_prefix")
	viper.SetDefault("label_prefix", "lacuna")
}

func GetConfig() (*Config, error) {
	log := log.WithField("component", "config")

	var config Config

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.WithError(err).Debug("no config file found, using environment variables")
		} else if _, ok := err.(*fs.PathError); ok {
			log.WithError(err).Debug("specified config file not found, using environment variables")
		} else {
			// Config file was found but another error was produced
			log.WithError(err).Debug("fatal error config file")
		}
	} else {
		log.Infof("config loaded from %s.", viper.ConfigFileUsed())
	}

	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
