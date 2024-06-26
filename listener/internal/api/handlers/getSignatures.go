package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/middleware"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSignatures(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection) {
	logger.Debug("Received new GET '/signatures' request")
	// Get tags from the context
	tags, ok := r.Context().Value(middleware.TagsKey).([]string)
	// middlewware already checks that tags is not empty. If something fails here, it is
	// because middleware didnt pass context correctly
	if !ok || len(tags) == 0 {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Query MongoDB for documents with tags matching the context tags
	var results []bson.M
	filter := bson.M{
		"tag": bson.M{"$in": tags},
	}
	cursor, err := dbCollection.Find(context.Background(), filter)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to query MongoDB: %v", err), http.StatusInternalServerError)
		return
	}
	defer cursor.Close(context.Background())

	if err := cursor.All(context.Background(), &results); err != nil {
		http.Error(w, fmt.Sprintf("Failed to read cursor: %v", err), http.StatusInternalServerError)
		return
	}

	// Return the results as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, fmt.Sprintf("Failed to encode results: %v", err), http.StatusInternalServerError)
	}
}
