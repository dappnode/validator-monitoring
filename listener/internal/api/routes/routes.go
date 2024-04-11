package routes

import (
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/handlers"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/mongo"
)

func SetupRouter(dbCollection *mongo.Collection) *mux.Router {
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/", handlers.GetHealthCheck).Methods(http.MethodGet)
	// closure function to inject dbCollection into the handler
	r.HandleFunc("/newSignature", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostNewSignature(w, r, dbCollection)
	}).Methods(http.MethodPost)
	r.HandleFunc("/signaturesByValidator", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostSignaturesByValidator(w, r, dbCollection)
	}).Methods(http.MethodPost)
	r.HandleFunc("/signaturesByValidatorAggr", func(w http.ResponseWriter, r *http.Request) {
		handlers.PostSignaturesByValidatorAggr(w, r, dbCollection)
	}).Methods(http.MethodPost)

	// Middlewares
	// r.Use(corsmiddleware()))

	return r
}
