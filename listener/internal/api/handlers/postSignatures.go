package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/api/validation"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func PostSignatures(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection, beaconNodeUrls map[types.Network]string, maxEntriesPerBson int) {
	logger.Debug("Received new POST '/signatures' request")
	var requests []types.SignatureRequest

	// Check if dbCollection is nil, just in case
	if dbCollection == nil {
		logger.Error("Database collection is nil")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Check if beaconNodeUrls is nil, just in case
	if beaconNodeUrls == nil {
		logger.Error("Beacon node URLs is nil")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Get network from query parameter
	networkVar := r.URL.Query().Get("network")
	if networkVar == "" {
		respondError(w, http.StatusBadRequest, "Missing network query parameter")
		return
	}
	network := types.Network(networkVar)
	beaconNodeUrl, ok := beaconNodeUrls[network]
	if !ok {
		respondError(w, http.StatusBadRequest, "Invalid network")
		return
	}

	// Parse and validate request body
	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		logger.Error("Failed to decode request body: " + err.Error())
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	if len(requests) == 0 {
		logger.Error("No valid requests in payload")
		respondError(w, http.StatusBadRequest, "No requests in payload")
		return
	}

	// Process each request and validate
	requestsValidatedAndDecoded, err := validation.ValidateAndDecodeRequests(requests)
	if err != nil {
		logger.Error("Failed to validate and decode requests: " + err.Error())
		respondError(w, http.StatusBadRequest, "No valid requests")
		return
	}
	if len(requestsValidatedAndDecoded) == 0 {
		logger.Error("All signature requests are invalid after decoding")
		respondError(w, http.StatusBadRequest, "No valid requests")
		return
	}

	// Get active validators and process signatures
	pubkeys := getPubkeys(requestsValidatedAndDecoded)
	validatorsStatusMap, err := validation.GetValidatorsStatus(pubkeys, beaconNodeUrl)
	if err != nil {
		logger.Error("Failed to get active validators: " + err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to get active validators: "+err.Error())
		return
	}

	validSignatures := filterAndVerifySignatures(requestsValidatedAndDecoded, validatorsStatusMap)
	if len(validSignatures) == 0 {
		respondError(w, http.StatusBadRequest, "No valid signatures")
		return
	}

	// Insert valid signatures into MongoDB
	if err := insertSignaturesIntoDB(validSignatures, network, dbCollection, maxEntriesPerBson); err != nil {
		logger.Error("Failed to insert signatures into MongoDB: " + err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to insert signatures into MongoDB: "+err.Error())
		return
	}

	respondOK(w, "Finished processing signatures")
}

func getPubkeys(requests []types.SignatureRequestDecoded) []string {
	pubkeys := make([]string, len(requests))
	for i, req := range requests {
		pubkeys[i] = req.Pubkey
	}
	return pubkeys
}

func filterAndVerifySignatures(requests []types.SignatureRequestDecoded, validatorsStatusMap map[string]types.Status) []types.SignatureRequestDecodedWithStatus {
	validSignatures := []types.SignatureRequestDecodedWithStatus{}
	for _, req := range requests {
		status, ok := validatorsStatusMap[req.Pubkey]
		if !ok {
			logger.Warn("Validator not found: " + req.Pubkey)
			continue
		}
		if status == types.Inactive {
			logger.Warn("Inactive validator: " + req.Pubkey)
			continue
		}
		reqWithStatus := types.SignatureRequestDecodedWithStatus{
			SignatureRequestDecoded: req,
			Status:                  status,
		}
		if isValid, err := validation.VerifySignature(reqWithStatus); err == nil && isValid {
			validSignatures = append(validSignatures, reqWithStatus)
		} else {
			logger.Warn("Invalid signature: " + req.Signature)
		}
	}
	return validSignatures
}

func insertSignaturesIntoDB(signatures []types.SignatureRequestDecodedWithStatus, network types.Network, dbCollection *mongo.Collection, maxEntriesPerBson int) error {
	for _, req := range signatures {
		filter := bson.M{
			"pubkey":  req.Pubkey,
			"tag":     req.Tag,
			"network": network,
		}

		// Check the number of entries
		var result struct {
			Entries []bson.M `bson:"entries"`
		}
		err := dbCollection.FindOne(context.Background(), filter).Decode(&result)
		if err != nil && err != mongo.ErrNoDocuments {
			return err
		}

		// mongo DB has a limit of 16MB per document
		// if this limit is reached the following exception is thrown: `write exception: write errors: [Resulting document after update is larger than 16777216]`
		if len(result.Entries) >= maxEntriesPerBson {
			return errors.New("Max number of entries reached for pubkey " + req.Pubkey + ". Max entries per pubkey: " + fmt.Sprint(maxEntriesPerBson))
		}

		// Create a base update document with $push operation
		update := bson.M{
			"$push": bson.M{
				"entries": bson.M{
					"payload":   req.Payload,
					"signature": req.Signature,
					"decodedPayload": bson.M{
						"type":      req.DecodedPayload.Type,
						"platform":  req.DecodedPayload.Platform,
						"timestamp": req.DecodedPayload.Timestamp,
					},
				},
			},
		}

		// Only update status unknown -> active
		if req.Status == "active" {
			update["$set"] = bson.M{"status": req.Status}
		} else {
			update["$setOnInsert"] = bson.M{"status": req.Status}
		}

		options := options.Update().SetUpsert(true)
		if _, err := dbCollection.UpdateOne(context.Background(), filter, update, options); err != nil {
			return err
		}

		logger.Debug("New Signature " + req.Signature + " inserted into MongoDB")
	}
	return nil
}
