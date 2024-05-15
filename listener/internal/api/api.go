package api

import (
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/routes"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"github.com/dappnode/validator-monitoring/listener/internal/mongodb"
)

type httpApi struct {
	server                   *http.Server
	port                     string
	dbUri                    string
	beaconNodeUrls           map[string]string
	bypassValidatorFiltering bool
}

// create a new api instance
func NewApi(port string, mongoDbUri string, beaconNodeUrls map[string]string, bypassValidatorFiltering bool) *httpApi {
	return &httpApi{
		port:                     port,
		dbUri:                    mongoDbUri,
		beaconNodeUrls:           beaconNodeUrls,
		bypassValidatorFiltering: bypassValidatorFiltering,
	}
}

// start the api
func (s *httpApi) Start() {
	logger.Info("Server is running on port " + s.port)

	// if somehow s.server is not nil, it means the server is already running, this should never happen
	if s.server != nil {
		logger.Fatal("HTTP server already started")
	}

	// get mongo db client
	dbClient, err := mongodb.GetMongoDbClient(s.dbUri)
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
		Handler: routes.SetupRouter(dbCollection, s.beaconNodeUrls, s.bypassValidatorFiltering),
	}

	// start the api
	if err := s.server.ListenAndServe(); err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
