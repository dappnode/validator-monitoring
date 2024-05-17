package cron

import (
	"context"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/api/validation"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// RemoveUnknownSignatures deletes signatures that have status "unknown" from the MongoDB collection
func RemoveUnknownSignatures(collection *mongo.Collection, beaconNodeUrls map[string]string) {
	logger.Debug("Removing unknown signatures")
	filter := bson.M{
		"status": types.Unknown,
	}
	// Get all the unknown signatures
	unknownSignatures, err := collection.Find(context.Background(), filter)
	if err != nil {
		logger.Error("Failed to get unknown signatures: " + err.Error())
		return
	}

	// create array of signaturesRequestDecoded
	var unknownSignaturesArray []types.SignatureRequestDecoded
	// Iterate over the unknown signatures
	for unknownSignatures.Next(context.Background()) {
		var signature types.SignatureRequestDecoded
		err := unknownSignatures.Decode(&signature)
		if err != nil {
			logger.Error("Failed to decode unknown signature: " + err.Error())
			continue
		}
		unknownSignaturesArray = append(unknownSignaturesArray, signature)
	}
	// TODO: join unknown signatures by network and use the corresponding beaconNodeUrl
	// Call GetActiveValidators to determine weather or not the signature is active
	activeValidators, err := validation.GetActiveValidators(unknownSignaturesArray, beaconNodeUrl)
}
