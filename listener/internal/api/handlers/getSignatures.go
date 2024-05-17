package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// GetSignatures fetches signatures from MongoDB based on the network, tag, and timestamp criteria.
func GetSignatures(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection) {
	logger.Debug("Received new POST '/getSignatures' request")

	var params types.GetSignatureParams
	err := json.NewDecoder(r.Body).Decode(&params)
	if err != nil {
		logger.Error("Failed to decode request body: " + err.Error())
		respondError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Calculate the cutoff time in Unix milliseconds as a string
	cutoffTime := time.Now().Add(-time.Duration(params.Hours) * time.Hour).UnixMilli()
	cutoffTimeStr := fmt.Sprintf("%d", cutoffTime)

	filter := bson.M{
		"network": params.Network,
		"tag":     params.Tag,
		"entries.decodedPayload.timestamp": bson.M{
			"$gt": cutoffTimeStr,
		},
	}

	projection := bson.M{
		"_id": 0, // Exclude the _id field
	}

	findOptions := options.Find().SetProjection(projection)

	var results []bson.M
	cursor, err := dbCollection.Find(context.Background(), filter, findOptions)
	if err != nil {
		logger.Error("Failed to fetch signatures from MongoDB: " + err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to fetch signatures")
		return
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &results); err != nil {
		logger.Error("Failed to decode signatures: " + err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to decode signatures")
		return
	}

	// Set the content type to application/json
	w.Header().Set("Content-Type", "application/json")

	// Write the JSON response directly to the ResponseWriter
	err = json.NewEncoder(w).Encode(results)
	if err != nil {
		logger.Error("Failed to encode response: " + err.Error())
		respondError(w, http.StatusInternalServerError, "Failed to encode response")
		return
	}
}
