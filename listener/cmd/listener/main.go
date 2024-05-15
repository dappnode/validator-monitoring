package main

import (
	"github.com/dappnode/validator-monitoring/listener/internal/api"
	"github.com/dappnode/validator-monitoring/listener/internal/config"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"github.com/dappnode/validator-monitoring/listener/internal/mongodb"
	"github.com/herumi/bls-eth-go-binary/bls"
)

func main() {
	logger.Info("Starting listener")
	// Load config
	config, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("Failed to load config: " + err.Error())
	}
	logger.SetLogLevelFromString(config.LogLevel)

	// This is a configuration of the BLS library at the process level. Notice how bls.Init() does not return an initialized BLS object.
	// Any call to bls functions within the process will use this configuration. We initialize bls before starting the api.
	if err := bls.Init(bls.BLS12_381); err != nil {
		logger.Fatal("Failed to initialize BLS: " + err.Error())
	}
	if err := bls.SetETHmode(bls.EthModeDraft07); err != nil {
		logger.Fatal("Failed to set BLS ETH mode: " + err.Error())
	}

	// Connect to MongoDB client & get the collection
	dbClient, err := mongodb.GetMongoDbClient(config.MongoDBURI)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB: " + err.Error())
	}
	dbCollection := dbClient.Database("validatorMonitoring").Collection("signatures")
	if dbCollection == nil {
		logger.Fatal("Failed to connect to MongoDB collection")
	}

	s := api.NewApi(
		config.Port,
		dbClient,
		dbCollection,
		config.BeaconNodeURLs,
		config.BypassValidatorsFiltering,
	)

	s.Start()
}
