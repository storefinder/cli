package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/storefinder/cli/commands"
)

func init() {
	log.Info("Initializing the Storelocator CLI")
}

func main() {
	commands.Execute()
}
