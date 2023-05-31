package cmd

import (
	"fmt"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	verbosity int

	rootCmd = &cobra.Command{
		Use:   "lacuna",
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
	viper.SetEnvPrefix("lacuna")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Logging
	log.SetLevel(logVerbosityToLevel(verbosity))
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
