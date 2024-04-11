package config

import (
	"os"

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
}

func LoadConfig() (*Config, error) {

	mongoDBURI := os.Getenv("MONGO_DB_URI")
	if mongoDBURI == "" {
		logger.Fatal("MONGO_DB_URI is not set")
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logger.Info("LOG_LEVEL is not set, using default INFO")
		logLevel = "INFO"
	}

	apiPort := os.Getenv("API_PORT")
	if apiPort == "" {
		logger.Fatal("API_PORT is not set")
	}

	return &Config{
		Port:       apiPort,
		MongoDBURI: mongoDBURI,
		LogLevel:   logLevel,
	}, nil
}
