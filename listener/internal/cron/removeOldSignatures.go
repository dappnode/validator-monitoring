package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// RemoveOldSignatures deletes signatures older than 30 days (24h * 30) from the MongoDB collection
func RemoveOldSignatures(collection *mongo.Collection) {
	logger.Debug("Removing signatures older than 30 days")
	thirtyDaysAgoMillis := time.Now().Add(-720 * time.Hour).UnixMilli()
	filter := bson.M{
		"entries.decodedPayload.timestamp": bson.M{
			"$lt": fmt.Sprintf("%d", thirtyDaysAgoMillis), // Compare timestamps as strings
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
