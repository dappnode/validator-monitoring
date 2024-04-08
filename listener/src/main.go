package main

import (
	"github.com/dappnode/validator-monitoring/listener/src/api"
	"github.com/dappnode/validator-monitoring/listener/src/config"
	"github.com/dappnode/validator-monitoring/listener/src/logger"
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
	)

	s.Start()
}
