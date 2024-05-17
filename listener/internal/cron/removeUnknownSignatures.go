package cron

import (
	"context"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/api/validation"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RemoveUnknownSignatures deletes signatures that have status "unknown" from the MongoDB collection
func RemoveUnknownSignatures(collection *mongo.Collection, beaconNodeUrls map[types.Network]string) {
	logger.Debug("Removing unknown signatures")
	filter := bson.M{
		"status": types.Unknown, // Assuming types.Unknown translates to the string "Unknown"
	}
	// Define a projection to exclude the 'entries' field from the results
	projection := bson.M{
		"entries": 0, // 0 to exclude, 1 to include
	}

	// Get all the unknown signatures, excluding the 'entries' field
	opts := options.Find().SetProjection(projection)
	unknownSignatures, err := collection.Find(context.Background(), filter, opts)
	if err != nil {
		logger.Error("Failed to get unknown signatures: " + err.Error())
		return
	}

	// Create an array of SignatureRequestDecoded
	var unknownSignaturesArray []types.SignatureRequestDecoded

	// Iterate over the unknown signatures
	for unknownSignatures.Next(context.Background()) {
		var signature types.SignatureRequestDecoded
		err := unknownSignatures.Decode(&signature)
		if err != nil {
			logger.Error("Failed to decode unknown signature: " + err.Error())
			continue
		}

		// Since 'payload' and 'signature' may not exist in the retrieved data, we assign default values
		signature.Payload = "defaultPayload"     // Substitute with appropriate default if necessary
		signature.Signature = "defaultSignature" // Substitute with appropriate default if necessary

		// Assuming 'decodedPayload' might also be missing and needs default initialization
		if signature.DecodedPayload.Type == "" && signature.DecodedPayload.Platform == "" && signature.DecodedPayload.Timestamp == "" {
			signature.DecodedPayload = types.DecodedPayload{
				Type:      "defaultType",
				Platform:  "defaultPlatform",
				Timestamp: "defaultTimestamp", // Ensure this is a valid timestamp format or current time
			}
		}

		unknownSignaturesArray = append(unknownSignaturesArray, signature)
	}

	// Group unknownSignaturesArray by Network
	groupedByNetwork := make(map[types.Network][]types.SignatureRequestDecoded)
	for _, signature := range unknownSignaturesArray {
		groupedByNetwork[signature.Network] = append(groupedByNetwork[signature.Network], signature)
	}

	// Iterate over each group and process them with the corresponding beacon node URL
	for network, signatures := range groupedByNetwork {
		beaconNodeUrl, exists := beaconNodeUrls[network]
		if !exists {
			logger.Error("No beacon node URL found for network: " + string(network))
			continue
		}

		// Call the GetActiveValidators function
		activeValidators, err := validation.GetActiveValidators(signatures, beaconNodeUrl)
		if err != nil {
			logger.Error("Failed to get active validators for network " + string(network) + ": " + err.Error())
			continue
		}

		// Map active validators for quick lookup
		activeMap := make(map[string]bool)
		for _, validator := range activeValidators {
			activeMap[validator.Pubkey] = true
		}

		// Update database: set status to "active" for returned validators
		for _, validator := range activeValidators {
			filter := bson.M{
				"pubkey":  validator.Pubkey,
				"network": network,
			}
			update := bson.M{
				"$set": bson.M{
					"status": types.Active,
				},
			}
			_, err := collection.UpdateOne(context.Background(), filter, update)
			if err != nil {
				logger.Error("Failed to update validator status: " + err.Error())
			}
		}

		// Remove from database validators that were not returned as active
		for _, signature := range signatures {
			if !activeMap[signature.Pubkey] {
				filter := bson.M{
					"pubkey":  signature.Pubkey,
					"network": network,
				}
				_, err := collection.DeleteOne(context.Background(), filter)
				if err != nil {
					logger.Error("Failed to delete non-active validator: " + err.Error())
				}
			}
		}
	}
}
