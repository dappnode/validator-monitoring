package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/api/validation"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Posting a new singature consists in the following steps:
// 1. Decode and validate
// 2. Get active validators
// 3. Validate signature and insert into MongoDB
func PostNewSignature(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection, beaconNodeUrls map[string]string) {
	logger.Debug("Received new POST '/newSignature' request")

	// Parse request body
	var requests []types.SignatureRequest
	err := json.NewDecoder(r.Body).Decode(&requests)
	if err != nil {
		logger.Error("Failed to decode request body: " + err.Error())
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}
	// Decode and validate incoming requests
	validRequests, err := validation.ValidateAndDecodeRequests(requests)
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

	activeValidators := validation.GetActiveValidators(validRequests, beaconNodeUrl)
	if len(activeValidators) == 0 {
		respondError(w, http.StatusInternalServerError, "No active validators found in network")
		return
	}

	validSignatures := []types.SignatureRequestDecodedWithActive{}
	for _, req := range activeValidators {
		isValidSignature, err := validation.VerifySignature(req)
		if err != nil {
			logger.Error("Failed to validate signature: " + err.Error())
			continue
		}
		if !isValidSignature {
			logger.Warn("Invalid signature: " + req.Signature)
			continue
		}
		validSignatures = append(validSignatures, req)
	}
	// Respond with an error if no valid signatures were found
	if len(validSignatures) == 0 {
		respondError(w, http.StatusBadRequest, "No valid signatures")
		return
	}

	// Iterate over all valid signatures and insert the signature
	for _, req := range validSignatures {
		filter := bson.M{
			"pubkey":  req.Pubkey,
			"tag":     req.Tag,
			"network": req.Network,
		}
		update := bson.M{
			"$push": bson.M{
				"entries": bson.M{
					"payload":   req.Payload,
					"signature": req.Signature,
					"decodedPayload": bson.M{
						"type":      req.DecodedPayload.Type,
						"platform":  req.DecodedPayload.Platform,
						"timestamp": req.DecodedPayload.Timestamp, // Needed to filter out old signatures
					},
					"status": req.Status,
				},
			},
		}
		options := options.Update().SetUpsert(true)
		_, err := dbCollection.UpdateOne(context.TODO(), filter, update, options)
		if err != nil {
			logger.Error("Failed to insert signature into MongoDB: " + err.Error())
			continue
		}
		logger.Debug("New Signature " + req.Signature + " inserted into MongoDB")
	}

	respondOK(w, "Finished processing signatures")
}
