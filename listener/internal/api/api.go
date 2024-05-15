package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/routes"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type httpApi struct {
	server                   *http.Server
	port                     string
	dbClient                 *mongo.Client
	dbCollection             *mongo.Collection
	beaconNodeUrls           map[string]string
	bypassValidatorFiltering bool
}

// create a new api instance
func NewApi(port string, dbClient *mongo.Client, dbCollection *mongo.Collection, beaconNodeUrls map[string]string, bypassValidatorFiltering bool) *httpApi {
	return &httpApi{
		port:                     port,
		dbClient:                 dbClient,
		dbCollection:             dbCollection,
		beaconNodeUrls:           beaconNodeUrls,
		bypassValidatorFiltering: bypassValidatorFiltering,
	}
}

func (s *httpApi) Start() {
	logger.Info("Server is running on port " + s.port)

	// if somehow s.server is not nil, it means the server is already running, this should never happen
	if s.server != nil {
		logger.Fatal("HTTP server already started")
	}

	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: routes.SetupRouter(s.dbCollection, s.beaconNodeUrls, s.bypassValidatorFiltering),
	}

	if err := s.server.ListenAndServe(); err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}

// Shutdown gracefully shuts down the server without interrupting any active connections
func (s *httpApi) Shutdown(ctx context.Context) error {
	if s.server == nil {
		logger.Error("Received shutdown request but server is not running, this should never happen")
		return nil // Server is not running
	}

	// Attempt to gracefully shut down the server
	err := s.server.Shutdown(ctx)
	if err != nil {
		logger.Error("Failed to shut down server gracefully: " + fmt.Sprintln(err))
		return err
	}

	logger.Info("Server has been shut down gracefully")
	return nil
}
