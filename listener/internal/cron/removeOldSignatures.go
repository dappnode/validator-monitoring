package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// RemoveOldSignatures deletes signatures older than a specified number of hours from the MongoDB collection
func RemoveOldSignatures(collection *mongo.Collection, hours int) {
	logger.Debug(fmt.Sprintf("Removing signatures older than %d hours", hours))
	targetTimeMillis := time.Now().Add(time.Duration(-hours) * time.Hour).UnixMilli() // Calculate time in the past based on hours
	filter := bson.M{
		"entries.decodedPayload.timestamp": bson.M{
			"$lt": fmt.Sprintf("%d", targetTimeMillis), // Compare timestamps as strings
		},
	}
	// DeleteMany returns the number of documents deleted, it is useless for us since we're never
	// deleting a document, but an entry on its "entries" array
	_, err := collection.DeleteMany(context.Background(), filter)
	if err != nil {
		logger.Error("Failed to delete old signatures: " + err.Error())
	} else {
		logger.Debug("Deleted old signatures")
	}
}
