package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/src/types"
	"go.mongodb.org/mongo-driver/bson"
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
	if !validateSignature(req.Signature) {
		s.respondError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	// Decode BASE64 payload
	decodedBytes, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid BASE64 payload")
		return
	}
	var decodedPayload types.DecodedPayload
	err = json.Unmarshal(decodedBytes, &decodedPayload)
	if err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid payload format")
		return
	}

	// Validate the decoded payload (maybe we should be more strict here)
	if decodedPayload.Platform != "dappnode" || decodedPayload.Timestamp == "" || decodedPayload.Pubkey == "" {
		s.respondError(w, http.StatusBadRequest, "Payload content is invalid")
		return
	}

	// Insert into MongoDB
	_, err = s.dbCollection.InsertOne(r.Context(), bson.M{
		"platform":  decodedPayload.Platform,
		"timestamp": decodedPayload.Timestamp,
		"pubkey":    decodedPayload.Pubkey,
		"signature": req.Signature,
		"network":   req.Network,
		"label":     req.Label,
	})
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Failed to insert payload into MongoDB")
		return
	}

	s.respondOK(w, "Payload validated and inserted into MongoDB")

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
