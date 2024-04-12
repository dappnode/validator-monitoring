package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"sync"

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

// decodeAndValidateRequests decodes and validates incoming HTTP requests.
func decodeAndValidateRequests(r *http.Request) ([]types.SignatureRequestDecoded, error) {
	var requests []signatureRequest
	err := json.NewDecoder(r.Body).Decode(&requests)
	if err != nil {
		return nil, err
	}

	var validRequests []types.SignatureRequestDecoded
	for _, req := range requests {
		if req.Network == "" || req.Label == "" || req.Signature == "" || req.Payload == "" {
			logger.Debug("Skipping invalid signature from request, missing fields")
			continue
		}
		if !validation.IsValidPayloadFormat(req.Signature) {
			logger.Debug("Skipping invalid signature from request, invalid signature format: " + req.Signature)
			continue
		}
		decodedBytes, err := base64.StdEncoding.DecodeString(req.Payload)
		if err != nil {
			logger.Error("Failed to decode BASE64 payload from request: " + err.Error())
			continue
		}
		var decodedPayload types.DecodedPayload
		if err := json.Unmarshal(decodedBytes, &decodedPayload); err != nil {
			logger.Error("Failed to decode JSON payload from request: " + err.Error())
			continue
		}
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

	return validRequests, nil
}

func validateAndInsertSignature(req types.SignatureRequestDecoded, dbCollection *mongo.Collection, wg *sync.WaitGroup) {
	defer wg.Done()

	isValidSignature, err := validation.IsValidSignature(req)
	if err != nil {
		logger.Error("Failed to validate signature: " + err.Error())
		return
	}
	if !isValidSignature {
		logger.Debug("Invalid signature: " + req.Signature)
		return
	}

	// Insert into MongoDB if signature is valid
	_, err = dbCollection.InsertOne(context.TODO(), bson.M{
		"platform":  req.DecodedPayload.Platform,
		"timestamp": req.DecodedPayload.Timestamp,
		"pubkey":    req.DecodedPayload.Pubkey,
		"signature": req.Signature,
		"network":   req.Network,
		"label":     req.Label,
	})
	if err != nil {
		logger.Error("Failed to insert signature into MongoDB: " + err.Error())
		return
	}

	logger.Debug("New Signature " + req.Signature + " inserted into MongoDB")
}

// Posting a new singature consists in the following steps:
// 1. Decode and validate
// 2. Get active validators
// 3. Validate signature and insert into MongoDB
func PostNewSignature(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection) {
	logger.Debug("Received new POST '/newSignature' request")

	// Decode and validate incoming requests
	validRequests, err := decodeAndValidateRequests(r)
	if err != nil {
		logger.Error("Failed to decode request body: " + err.Error())
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	// Respond with an error if no valid requests were found
	if len(validRequests) == 0 {
		logger.Error("No valid requests")
		respondError(w, http.StatusBadRequest, "No valid requests")
		return
	}

	// Get active validators
	requestsWithActiveValidators, err := validation.GetActiveValidators(validRequests)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Failed to validate active validators")
		return
	}
	// Respond with an error if no active validators were found
	if len(requestsWithActiveValidators) == 0 {
		respondError(w, http.StatusInternalServerError, "No active validators found in request")
		return
	}

	var wg sync.WaitGroup
	// Insert into MongoDB if signature is valid
	for _, req := range requestsWithActiveValidators {
		// create a goroutine for each request
		wg.Add(1)
		go validateAndInsertSignature(req, dbCollection, &wg)
	}
	// Wait for all goroutines to finish
	wg.Wait()
	respondOK(w, "Finished processing signatures")
}
