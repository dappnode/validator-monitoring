package main

import (
	"github.com/dappnode/validator-monitoring/listener/internal/api"
	"github.com/dappnode/validator-monitoring/listener/internal/config"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
)

func main() {
	logger.Info("Starting listener")
	// Load config
	config, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config: " + err.Error())
	}
	logger.SetLogLevelFromString(config.LogLevel)

	s := api.NewApi(
		config.Port,
		config.MongoDBURI,
		config.BeaconNodeURL,
	)

	s.Start()
}
