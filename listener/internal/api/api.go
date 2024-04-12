package api

import (
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/routes"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"github.com/dappnode/validator-monitoring/listener/internal/mongodb"
)

type httpApi struct {
	server        *http.Server
	port          string
	dbUri         string
	beaconNodeUrl string
}

// create a new api instance
func NewApi(port string, mongoDbUri string, beaconNodeUrl string) *httpApi {
	return &httpApi{
		port:          port,
		dbUri:         mongoDbUri,
		beaconNodeUrl: beaconNodeUrl,
	}
}

// start the api
func (s *httpApi) Start() {
	// if somehow s.server is not nil, it means the server is already running, this should never happen
	if s.server != nil {
		logger.Fatal("HTTP server already started")
	}

	logger.Info("Server is running on port " + s.port)
	var err error

	// connect to the MongoDB server
	dbClient, err := mongodb.ConnectMongoDB(s.dbUri)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB: " + err.Error())
	}

	// get the collection
	dbCollection := dbClient.Database("validatorMonitoring").Collection("signatures")
	if dbCollection == nil {
		logger.Fatal("Failed to connect to MongoDB collection")
	}

	// setup the http api
	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: routes.SetupRouter(dbCollection, s.beaconNodeUrl),
	}

	// start the api
	if err := s.server.ListenAndServe(); err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
