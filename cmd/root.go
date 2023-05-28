package cmd

import (
	"fmt"
	"io/fs"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	verbosity int

	rootCmd = &cobra.Command{
		Use:   "pubsub",
		Short: "",
	}
)

func Execute(version string) {
	rootCmd.Version = version

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().CountVarP(&verbosity, "verbose", "v", "log level")
}

func initConfig() {
	// Environment
	// viper.SetEnvPrefix("LICTORES")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.SetConfigName("default")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Logging
	log.SetLevel(logVerbosityToLevel(verbosity))

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// Config file not found; ignore error if desired
			log.Debug("No config file found, using environment variables.")
		} else if _, ok := err.(*fs.PathError); ok {
			log.Debug("Specified config file not found, using environment variables.")
		} else {
			// Config file was found but another error was produced
			panic(fmt.Errorf("fatal error config file: %w", err))
		}
	} else {
		log.Infof("Config loaded from %s.", viper.ConfigFileUsed())
	}
}

func logVerbosityToLevel(count int) log.Level {
	if count > 1 {
		return log.TraceLevel
	} else if count > 0 {
		return log.DebugLevel
	} else {
		return log.InfoLevel
	}
}
