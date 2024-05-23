package routes

import (
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/handlers"
	"github.com/dappnode/validator-monitoring/listener/internal/api/middleware"
	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRouter(dbCollection *mongo.Collection, beaconNodeUrls map[types.Network]string, maxEntriesPerBson int) *mux.Router {
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/", handlers.GetHealthCheck).Methods(http.MethodGet)
	// closure function to inject dbCollection into the handler
	r.HandleFunc("/signatures", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostSignatures(w, r, dbCollection, beaconNodeUrls, maxEntriesPerBson)
	}).Methods(http.MethodPost)

	//TODO: implement handler for GET signatures
	r.Handle("/signatures", middleware.JWTMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Assuming you have a handler for GET signatures
		handlers.PostSignatures(w, r, dbCollection, beaconNodeUrls, maxEntriesPerBson)
	}))).Methods(http.MethodGet)

	return r
}
