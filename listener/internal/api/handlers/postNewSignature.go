package handlers

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/api/validation"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type signatureRequest struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
	Network   string `json:"network"`
	Label     string `json:"label"`
}

// Posting a new singature consists in the following steps:
// 1. Parse request and decode against struct signatureRequest
// 2. Validate request format and decode payload
// 3. Validate active validators
// 4. Validate signatures
func PostNewSignature(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection) {
	logger.Debug("Received new POST '/newSignature' request")

	// Decode request
	var requests []signatureRequest
	err := json.NewDecoder(r.Body).Decode(&requests)
	if err != nil {
		logger.Error("Failed to decode request body: " + err.Error())
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate and decode payload
	var validRequests []types.SignatureRequestDecoded // Needed to store valid pubkeys for bulk validation later
	// For each element of the request slice, we validate the element format and decode its payload
	for _, req := range requests {
		if req.Network == "" || req.Label == "" || req.Signature == "" || req.Payload == "" {
			logger.Debug("Skipping invalid signature from request, missing fields")
			continue // Skipping invalid elements
		}
		if !validation.IsValidPayloadFormat(req.Signature) {
			logger.Debug("Skipping invalid signature from request, invalid signature format: " + req.Signature)
			continue // Skipping invalid signature format
		}
		decodedBytes, err := base64.StdEncoding.DecodeString(req.Payload)
		if err != nil {
			logger.Error("Failed to decode BASE64 payload from request: " + err.Error())
			continue // Skipping payloads that can't be decoded from BASE64
		}
		var decodedPayload types.DecodedPayload
		if err := json.Unmarshal(decodedBytes, &decodedPayload); err != nil {
			logger.Error("Failed to decode JSON payload from request: " + err.Error())
			continue // Skipping payloads that can't be decoded from JSON
		}
		// If the payload is valid, we append the request. Else, we skip it
		if decodedPayload.Platform == "dappnode" && decodedPayload.Timestamp != "" && decodedPayload.Pubkey != "" {
			validRequests = append(validRequests, types.SignatureRequestDecoded{
				DecodedPayload: decodedPayload,
				Payload:        req.Payload,
				Signature:      req.Signature,
				Network:        req.Network,
				Label:          req.Label,
			})

		} else {
			logger.Debug("Skipping invalid signature from request, invalid payload format")
		}
	}

	// Respond with an error if no valid pubkeys were found
	if len(validRequests) == 0 {
		logger.Error("No valid pubkeys found in request")
		respondError(w, http.StatusBadRequest, "No valid pubkeys found in request")
		return
	}

	// Validate active validators
	requestsWithActiveValidators, err := validation.GetActiveValidators(validRequests)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to validate active validators")
		return
	}
	// Respond with an error if no valid pubkeys were found
	if len(requestsWithActiveValidators) == 0 {
		respondError(w, http.StatusInternalServerError, "No active validators found in request")
		return
	}

	// Iterate over requestsWithActiveValidators and validate signatures
	for _, req := range requestsWithActiveValidators {
		isValidSignature, err := validation.IsValidSignature(req)
		if err != nil {
			logger.Error("Failed to validate signature: " + err.Error())
			continue
		}
		if isValidSignature {
			// Insert into MongoDB if signature is valid
			_, err := dbCollection.InsertOne(r.Context(), bson.M{
				"platform":  req.DecodedPayload.Platform,
				"timestamp": req.DecodedPayload.Timestamp,
				"pubkey":    req.DecodedPayload.Pubkey,
				"signature": req.Signature,
				"network":   req.Network,
				"label":     req.Label,
			})
			if err != nil {
				logger.Error("Failed to insert signature into MongoDB: " + err.Error())
				continue
			} else {
				logger.Debug("New Signature " + req.Signature + " inserted into MongoDB")
			}
		} else {
			logger.Debug("Invalid signature: " + req.Signature)
		}
	}

	respondOK(w, "Finished processing signatures")
}
