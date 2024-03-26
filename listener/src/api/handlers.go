package api

import (
	"encoding/json"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/src/types"
)

func (s *httpApi) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.respondOK(w, "Server is running")
}

func (s *httpApi) handleSignature(w http.ResponseWriter, r *http.Request) {
	var req types.SignatureRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// validate that the request has the required fields and that the signature is valid
	// we expect network and label to be strings, signature to be a 0x prefixed hex string of 194 characters
	// and payload to be a 0x prefixed hex string.
	if req.Network == "" || req.Label == "" || req.Signature == "" || req.Payload == "" {
		s.respondError(w, http.StatusBadRequest, "Invalid request, missing fields")
		return
	}

	// validate the signature
	if !validateSignature(req.Payload, req.Signature) {
		s.respondError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

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
