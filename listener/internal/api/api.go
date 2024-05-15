package api

import (
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
	// if somehow s.server is not nil, it means the server is already running, this should never happen
	if s.server != nil {
		logger.Fatal("HTTP server already started")
	}

	logger.Info("Server is running on port " + s.port)

	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: routes.SetupRouter(s.dbCollection, s.beaconNodeUrls, s.bypassValidatorFiltering),
	}

	if err := s.server.ListenAndServe(); err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}
