package config

import (
	"fmt"
	"os"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
)

// Config is the struct that holds the configuration of the application
type Config struct {
	// Port is the port where the server will listen
	Port string
	// MongoDBURI is the URI of the MongoDB server. It includes the authentication credentials
	MongoDBURI string
	// LogLevel is the level of logging
	LogLevel string
	// BeaconNodeURLs is the URLs of the beacon nodes for different networks
	BeaconNodeURLs map[types.Network]string
}

func LoadConfig() (*Config, error) {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logger.Info("LOG_LEVEL is not set, using default INFO")
		logLevel = "INFO"
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		logger.Info("API_PORT is not set, using default 8080")
		apiPort = "8080"
	}

	mongoDBURI := os.Getenv("MONGO_DB_URI")
	if mongoDBURI == "" {
		return nil, fmt.Errorf("MONGO_DB_URI is not set")
	}

	// beacon node urls per network
	beaconMainnet := os.Getenv("BEACON_NODE_URL_MAINNET")
	if beaconMainnet == "" {
		return nil, fmt.Errorf("BEACON_NODE_URL_MAINNET is not set")
	}
	beaconHolesky := os.Getenv("BEACON_NODE_URL_HOLESKY")
	if beaconHolesky == "" {
		logger.Fatal("BEACON_NODE_URL_HOLESKY is not set")
	}
	beaconGnosis := os.Getenv("BEACON_NODE_URL_GNOSIS")
	if beaconGnosis == "" {
		return nil, fmt.Errorf("BEACON_NODE_URL_GNOSIS is not set")
	}
	beaconLukso := os.Getenv("BEACON_NODE_URL_LUKSO")
	if beaconLukso == "" {
		return nil, fmt.Errorf("BEACON_NODE_URL_LUKSO is not set")
	}

	// print all envs in a single line
	logger.Info("Loaded config: LOG_LEVEL=" + logLevel + " API_PORT=" + apiPort + " MONGO_DB_URI=" + mongoDBURI + " BEACON_NODE_URL_MAINNET=" + beaconMainnet + " BEACON_NODE_URL_HOLESKY=" + beaconHolesky + " BEACON_NODE_URL_GNOSIS=" + beaconGnosis + " BEACON_NODE_URL_LUKSO=" + beaconLukso)

	beaconNodeURLs := map[types.Network]string{
		types.Mainnet: beaconMainnet,
		types.Holesky: beaconHolesky,
		types.Gnosis:  beaconGnosis,
		types.Lukso:   beaconLukso,
	}

	return &Config{
		Port:           apiPort,
		MongoDBURI:     mongoDBURI,
		LogLevel:       logLevel,
		BeaconNodeURLs: beaconNodeURLs,
	}, nil
}
