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
	Pubkey    string `json:"pubkey"`
	Signature string `json:"signature"`
	Network   string `json:"network"`
	Tag       string `json:"tag"`
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
		if req.Network == "" || req.Tag == "" || req.Signature == "" || req.Payload == "" || req.Pubkey == "" {
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
		if decodedPayload.Platform == "dappnode" && decodedPayload.Timestamp != "" {
			validRequests = append(validRequests, types.SignatureRequestDecoded{
				DecodedPayload: decodedPayload,
				Payload:        req.Payload,
				Pubkey:         req.Pubkey,
				Signature:      req.Signature,
				Network:        req.Network,
				Tag:            req.Tag,
			})
		} else {
			logger.Debug("Skipping invalid signature from request, invalid payload format")
		}
	}

	// print req pubkey
	for _, req := range validRequests {
		logger.Debug("req.Pubkey: " + req.Pubkey)
		logger.Debug("req.Signature: " + req.Signature)
		logger.Debug("req.Network: " + req.Network)
		logger.Debug("req.Tag: " + req.Tag)
		logger.Debug("req.DecodedPayload.Type: " + req.DecodedPayload.Type)
		logger.Debug("req.DecodedPayload.Platform: " + req.DecodedPayload.Platform)
		logger.Debug("req.DecodedPayload.Timestamp: " + req.DecodedPayload.Timestamp)
	}

	return validRequests, nil
}

func validateAndInsertSignature(req types.SignatureRequestDecoded, dbCollection *mongo.Collection) {
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
		"pubkey":    req.Pubkey,
		"signature": req.Signature,
		"network":   req.Network,
		"tag":       req.Tag,
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
func PostNewSignature(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection, beaconNodeUrls map[string]string, bypassValidatorFiltering bool) {
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
		respondError(w, http.StatusBadRequest, "No valid requests")
		return
	}

	// Get active validators from the network, get the network from the first item in the array
	beaconNodeUrl, ok := beaconNodeUrls[validRequests[0].Network]
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid network")
		return
	}

	// if bypassValidatorFiltering is true, we skip the active validators check
	activeValidators := []types.SignatureRequestDecoded{}
	if bypassValidatorFiltering {
		logger.Debug("Bypassing active validators check")
		activeValidators = validRequests
	} else {
		activeValidators = validation.GetActiveValidators(validRequests, beaconNodeUrl)
	}

	if len(activeValidators) == 0 {
		respondError(w, http.StatusInternalServerError, "No active validators found in network")
		return
	}

	// Iterate over all active validators and validate and insert the signature
	dbMutex := new(sync.Mutex) // Mutex for database operations
	for _, req := range activeValidators {
		dbMutex.Lock()
		validateAndInsertSignature(req, dbCollection) // Do we really need to lock the db insertions?
		dbMutex.Unlock()
	}

	respondOK(w, "Finished processing signatures")
}
