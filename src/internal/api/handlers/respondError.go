package handlers

import (
	"encoding/json"
	"net/http"
)

type httpErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// TODO: error handling
func respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(httpErrorResp{Code: code, Message: message})
}
