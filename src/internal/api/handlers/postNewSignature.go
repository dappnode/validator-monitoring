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

func PostNewSignature(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection) {
	logger.Debug("Received new POST '/newSignature' request")

	var sigs []signatureRequest
	err := json.NewDecoder(r.Body).Decode(&sigs)
	if err != nil {
		logger.Error("Failed to decode request body: " + err.Error())
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	var validPubkeys []string // Needed to store valid pubkeys for bulk validation later

	// For each element of the request slice, we validate the element format and decode its payload
	for _, req := range sigs {
		if req.Network == "" || req.Label == "" || req.Signature == "" || req.Payload == "" {
			logger.Debug("Skipping invalid signature from request, missing fields")
			continue // Skipping invalid elements
		}
		if !validation.ValidateSignature(req.Signature) {
			logger.Debug("Skipping invalid signature from request, invalid signature format: " + req.Signature)
			continue // Skipping invalid signatures
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
		// If the payload is valid, we append the pubkey to the validPubkeys slice. Else, we skip it
		if decodedPayload.Platform == "dappnode" && decodedPayload.Timestamp != "" && decodedPayload.Pubkey != "" {
			validPubkeys = append(validPubkeys, decodedPayload.Pubkey) // Collecting valid pubkeys
		} else {
			logger.Debug("Skipping invalid signature from request, invalid payload format")
		}
	}

	// Make a single API call to validate pubkeys in bulk
	validatedPubkeys := validation.ValidatePubkeysWithConsensusClient(validPubkeys)
	if len(validatedPubkeys) == 0 {
		respondError(w, http.StatusInternalServerError, "Failed to validate pubkeys with consensus client")
		return
	}

	// Now, iterate over the originally valid requests, check if the pubkey was validated, then verify signature and insert into DB
	// This means going over the requests again! TODO: find a better way?
	for _, req := range sigs {
		decodedBytes, _ := base64.StdEncoding.DecodeString(req.Payload)
		var decodedPayload types.DecodedPayload
		json.Unmarshal(decodedBytes, &decodedPayload)

		// Only try to validate message signatures if the pubkey is validated
		if _, exists := validatedPubkeys[decodedPayload.Pubkey]; exists {
			// If the pubkey is validated, we can proceed to validate the signature
			if validation.ValidateSignatureAgainstPayload(req.Signature, decodedPayload) {
				// Insert into MongoDB
				_, err := dbCollection.InsertOne(r.Context(), bson.M{
					"platform":  decodedPayload.Platform,
					"timestamp": decodedPayload.Timestamp,
					"pubkey":    decodedPayload.Pubkey,
					"signature": req.Signature,
					"network":   req.Network,
					"label":     req.Label,
				})
				if err != nil {
					continue // TODO: Log error or handle as needed
				} else {
					logger.Info("New Signature " + req.Signature + " inserted into MongoDB")
				}
			}
		}
	}

	respondOK(w, "Finished processing signatures")
}
