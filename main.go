package main

import (
	"github.com/aplr/lacuna/cmd"
)

var Version string
var Buildtime string

func main() {
	version := "local"
	if Version != "" {
		version = Version
	}

	cmd.Execute(version)
}

func init() {
	// log.SetFormatter(&log.JSONFormatter{})
}
