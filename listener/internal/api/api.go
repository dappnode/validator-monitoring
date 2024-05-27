package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/routes"
	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/mongo"
)

type httpApi struct {
	server            *http.Server
	port              string
	dbClient          *mongo.Client
	dbCollection      *mongo.Collection
	beaconNodeUrls    map[types.Network]string
	maxEntriesPerBson int
	jwtUsersFilePath  string
}

// create a new api instance
func NewApi(port string, dbClient *mongo.Client, dbCollection *mongo.Collection, beaconNodeUrls map[types.Network]string, maxEntriesPerBson int, jwtUsersFilePath string) *httpApi {
	return &httpApi{
		port:              port,
		dbClient:          dbClient,
		dbCollection:      dbCollection,
		beaconNodeUrls:    beaconNodeUrls,
		maxEntriesPerBson: maxEntriesPerBson,
		jwtUsersFilePath:  jwtUsersFilePath,
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
		Handler: routes.SetupRouter(s.dbCollection, s.beaconNodeUrls, s.maxEntriesPerBson, s.jwtUsersFilePath),
	}

	// ListenAndServe returns ErrServerClosed to indicate that the server has been shut down when the server is closed gracefully. We need to
	// handle this error to avoid treating it as a fatal error. See https://pkg.go.dev/net/http#Server.ListenAndServe
	err := s.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal("Failed to start server: " + err.Error())
	} else if err == http.ErrServerClosed {
		logger.Info("Server closed gracefully")
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
