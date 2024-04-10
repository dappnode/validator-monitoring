package handlers

import (
	"encoding/json"
	"net/http"
)

// TODO: error handling
func respondOK(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
