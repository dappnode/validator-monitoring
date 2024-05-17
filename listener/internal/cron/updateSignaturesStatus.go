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

func UpdateSignaturesStatus(collection *mongo.Collection, beaconNodeUrls map[types.Network]string) {
	logger.Debug("Updating statuses and removing inactive signatures")

	// Step 1: Query the MongoDB collection to retrieve all documents with status "unknown"
	filter := bson.M{"status": types.Unknown}
	projection := bson.M{
		"pubkey":  1,
		"tag":     1,
		"network": 1,
	}
	cursor, err := collection.Find(context.Background(), filter, options.Find().SetProjection(projection))
	if err != nil {
		logger.Error("Failed to query MongoDB collection: " + err.Error())
		return
	}
	defer cursor.Close(context.Background())

	// Step 2: Extract pubkeys, tags, and networks from the documents
	type PubkeyTagNetwork struct {
		Pubkey  string
		Tag     types.Tag
		Network types.Network
	}
	var pubkeyTagNetworkPairs []PubkeyTagNetwork
	for cursor.Next(context.Background()) {
		var signature PubkeyTagNetwork
		if err := cursor.Decode(&signature); err != nil {
			logger.Error("Failed to decode MongoDB document: " + err.Error())
			continue

		}
		pubkeyTagNetworkPairs = append(pubkeyTagNetworkPairs, signature)
	}

	if err := cursor.Err(); err != nil {
		logger.Error("Failed to iterate over MongoDB cursor: " + err.Error())
		return
	}

	// Step 3: Query GetValidatorsStatus using these pubkeys and the corresponding beacon node URL
	pubkeyStatusMap := make(map[string]types.Status)
	for network, url := range beaconNodeUrls {
		var networkPubkeys []string
		for _, pair := range pubkeyTagNetworkPairs {
			if pair.Network == network {
				networkPubkeys = append(networkPubkeys, pair.Pubkey)
			}
		}
		if len(networkPubkeys) > 0 {
			statusMap, err := validation.GetValidatorsStatus(networkPubkeys, url)
			if err != nil {
				logger.Error("Failed to get active validators: " + err.Error())
				continue
			}
			for pubkey, status := range statusMap {
				pubkeyStatusMap[pubkey] = status
			}
		}
	}

	// Step 4: Update or remove documents based on the validator status
	for _, pair := range pubkeyTagNetworkPairs {
		status, exists := pubkeyStatusMap[pair.Pubkey]
		if !exists {
			continue
		}

		if status == types.Active {
			// Update the status to "active"
			update := bson.M{
				"$set": bson.M{"status": types.Active},
			}
			_, err := collection.UpdateOne(context.Background(), bson.M{"pubkey": pair.Pubkey, "tag": pair.Tag, "network": pair.Network, "status": types.Unknown}, update)
			if err != nil {
				logger.Error("Failed to update signature: " + err.Error())
				continue
			}
			logger.Debug("Updated signature with pubkey " + pair.Pubkey + " to active")
		} else if status == types.Inactive {
			// Remove the signature
			_, err := collection.DeleteOne(context.Background(), bson.M{"pubkey": pair.Pubkey, "tag": pair.Tag, "network": pair.Network, "status": types.Unknown})
			if err != nil {
				logger.Error("Failed to remove signature: " + err.Error())
				continue
			}
			logger.Debug("Removed signature with pubkey " + pair.Pubkey + " due to inactive validator status")
		}
	}
}
