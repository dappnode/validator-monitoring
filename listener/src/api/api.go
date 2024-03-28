package api

import (
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/src/logger"
	"github.com/dappnode/validator-monitoring/listener/src/mongodb"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

type httpApi struct {
	server       *http.Server
	port         string
	dbUri        string
	dbClient     *mongo.Client
	dbCollection *mongo.Collection
}

// create a new api instance
func NewApi(port string, mongoDbUri string) *httpApi {
	return &httpApi{
		port:  port,
		dbUri: mongoDbUri,
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
	s.dbClient, err = mongodb.ConnectMongoDB(s.dbUri)
	if err != nil {
		logger.Fatal("Failed to connect to MongoDB: " + err.Error())
	}

	// get the collection
	s.dbCollection = s.dbClient.Database("validatorMonitoring").Collection("signatures")
	if s.dbCollection == nil {
		logger.Fatal("Failed to connect to MongoDB collection")
	}

	// setup the http api
	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: s.SetupRouter(),
	}

	// start the api
	if err := s.server.ListenAndServe(); err != nil {
		logger.Fatal("Failed to start server: " + err.Error())
	}
}

// initialize and setup the router. This is where all the endpoints are defined
// and the middlewares are applied
func (s *httpApi) SetupRouter() *mux.Router {
	r := mux.NewRouter()

	// Endpoints
	r.HandleFunc("/", s.handleRoot).Methods(http.MethodGet)
	r.HandleFunc("/newSignature", s.handleSignature).Methods(http.MethodPost)
	r.HandleFunc("/signaturesByValidator", s.handleGetSignatures).Methods(http.MethodPost)
	r.HandleFunc("/signaturesByValidatorAggr", s.handleGetSignaturesAggr).Methods(http.MethodPost)

	// Middlewares
	// r.Use(corsmiddleware()))
	return r
}
