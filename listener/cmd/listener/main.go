package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/herumi/bls-eth-go-binary/bls"
	"github.com/robfig/cron"

	"github.com/dappnode/validator-monitoring/listener/internal/api"
	"github.com/dappnode/validator-monitoring/listener/internal/config"
	apiCron "github.com/dappnode/validator-monitoring/listener/internal/cron" // Renamed to avoid conflict with the cron/v3 package
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"github.com/dappnode/validator-monitoring/listener/internal/mongodb"
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
	)

	// Start the API server in a goroutine. Needs to be in a goroutine to allow for the cron job to run,
	// otherwise it blocks the main goroutine
	go func() {
		s.Start()
	}()

	// Set up the cron job
	c := cron.New()

	// The cron job runs once a day, see https://github.com/robfig/cron/blob/master/doc.go
	// to test it running once a minute, replace "@daily" for "* * * * *"
	c.AddFunc("@daily", func() {
		// there are 24 * 30 = 720 hours in 30 days
		apiCron.RemoveOldSignatures(dbCollection, 720)

	})
	c.AddFunc("@every 1m", func() {
		apiCron.UpdateSignaturesStatus(dbCollection, config.BeaconNodeURLs)
	})
	c.Start()

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // Block until a signal is received

	// Stop the cron job
	c.Stop()

	// Shutdown the HTTP server gracefully with a given timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		logger.Error("Failed to shut down server gracefully: " + fmt.Sprintln(err))
	}

	logger.Info("Listener stopped gracefully")
}
