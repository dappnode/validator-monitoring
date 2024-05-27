package config

import (
	"fmt"
	"os"
	"strconv"

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
	// Max number of entries allowed per BSON document
	MaxEntriesPerBson int
	JWTUsersFilePath  string
}

func GetConfig() (*Config, error) {
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
	maxEntriesPerBsonStr := os.Getenv("MAX_ENTRIES_PER_BSON")
	if maxEntriesPerBsonStr == "" {
		logger.Info("MAX_ENTRIES_PER_BSON is not set, using default 30")
		maxEntriesPerBsonStr = "30"
	}
	MaxEntriesPerBson, err := strconv.Atoi(maxEntriesPerBsonStr)
	if err != nil {
		return nil, fmt.Errorf("MAX_ENTRIES_PER_BSON is not a valid integer")
	}

	jwtUsersFilePath := os.Getenv("JWT_USERS_FILE_PATH")
	if jwtUsersFilePath == "" {
		return nil, fmt.Errorf("JWT_USERS_FILE_PATH is not set")
	}

	logger.Info("LOG_LEVEL: " + logLevel)
	logger.Info("API_PORT: " + apiPort)
	logger.Info("MONGO_DB_URI: " + mongoDBURI)
	logger.Info("BEACON_NODE_URL_MAINNET: " + beaconMainnet)
	logger.Info("BEACON_NODE_URL_HOLESKY: " + beaconHolesky)
	logger.Info("BEACON_NODE_URL_GNOSIS: " + beaconGnosis)
	logger.Info("BEACON_NODE_URL_LUKSO: " + beaconLukso)
	logger.Info("MAX_ENTRIES_PER_BSON: " + maxEntriesPerBsonStr)
	logger.Info("JWT_USERS_FILE_PATH: " + jwtUsersFilePath)

	beaconNodeURLs := map[types.Network]string{
		types.Mainnet: beaconMainnet,
		types.Holesky: beaconHolesky,
		types.Gnosis:  beaconGnosis,
		types.Lukso:   beaconLukso,
	}

	return &Config{
		Port:              apiPort,
		MongoDBURI:        mongoDBURI,
		LogLevel:          logLevel,
		BeaconNodeURLs:    beaconNodeURLs,
		MaxEntriesPerBson: MaxEntriesPerBson,
		JWTUsersFilePath:  jwtUsersFilePath,
	}, nil
}
