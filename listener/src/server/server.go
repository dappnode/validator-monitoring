package server

import (
	"encoding/json"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/src/logger"
	"github.com/dappnode/validator-monitoring/listener/src/types"
	"github.com/gorilla/mux"
)

type httpApi struct {
	server *http.Server
	port   string
}

// create a new api instance
func NewApi(port string) *httpApi {
	return &httpApi{
		port: port,
	}
}

// start the api
func (s *httpApi) Start() {
	logger.Info("Server is running on port " + s.port)
	if s.server != nil {
		logger.Fatal("HTTP server already started")
	}

	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: s.SetupRouter(),
	}

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
	// r.HandleFunc("/signature", s.handleSignature).Methods(http.MethodPost)

	// Middlewares
	// r.Use(corsmiddleware()))
	return r
}

func (s *httpApi) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.respondOK(w, "Server is running")
}

// TODO: error handling
func (s *httpApi) respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(types.HttpErrorResp{Code: code, Message: message})
}

// TODO: error handling
func (s *httpApi) respondOK(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}
