package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/src/logger"
	"github.com/dappnode/validator-monitoring/listener/src/types"
	"go.mongodb.org/mongo-driver/bson"
)

func (s *httpApi) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.respondOK(w, "Server is running")
}

func (s *httpApi) handleSignature(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Received new POST '/signature' request")
	var req types.SignatureRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error("Failed to decode request body" + err.Error())
		s.respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// validate that the request has the required fields and that the signature is valid
	// we expect network and label to be strings, signature to be a 0x prefixed hex string of 194 characters
	// and payload to be a 0x prefixed hex string.
	if req.Network == "" || req.Label == "" || req.Signature == "" || req.Payload == "" {
		logger.Debug("Invalid request, missing fields. Request had the following fields: " + req.Network + " " + req.Label + " " + req.Signature + " " + req.Payload)
		s.respondError(w, http.StatusBadRequest, "Invalid request, missing fields")
		return
	}

	// validate the signature
	if !validateSignature(req.Signature) {
		logger.Debug("Invalid signature. It should be a 0x prefixed hex string of 194 characters. Request had the following signature: " + req.Signature)
		s.respondError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	// Decode BASE64 payload
	decodedBytes, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		logger.Error("Failed to decode BASE64 payload" + err.Error())
		s.respondError(w, http.StatusBadRequest, "Invalid BASE64 payload")
		return
	}
	var decodedPayload types.DecodedPayload
	err = json.Unmarshal(decodedBytes, &decodedPayload)
	if err != nil {
		logger.Error("Failed to decode JSON payload" + err.Error())
		s.respondError(w, http.StatusBadRequest, "Invalid payload format")
		return
	}

	// Validate the decoded payload (maybe we should be more strict here)
	if decodedPayload.Platform != "dappnode" || decodedPayload.Timestamp == "" || decodedPayload.Pubkey == "" {
		logger.Debug("Invalid payload content. Request had the following payload: " + decodedPayload.Platform + " " + decodedPayload.Timestamp + " " + decodedPayload.Pubkey)
		s.respondError(w, http.StatusBadRequest, "Payload content is invalid")
		return
	}

	logger.Debug("Request's payload decoded successfully. Inserting decoded message into MongoDB")

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
		logger.Error("Failed to insert message into MongoDB" + err.Error())
		s.respondError(w, http.StatusInternalServerError, "Failed to insert payload into MongoDB")
		return
	}
	logger.Info("A new message with pubkey " + decodedPayload.Pubkey + " was decoded and inserted into MongoDB successfully")
	s.respondOK(w, "Message validated and inserted into MongoDB")

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
