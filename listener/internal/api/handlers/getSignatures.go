package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/internal/api/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func GetSignatures(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection) {
	// Get roles from the context
	roles, ok := r.Context().Value(middleware.RolesKey).([]string)
	if !ok || len(roles) == 0 {
		http.Error(w, "Roles not found in context", http.StatusUnauthorized)
		return
	}

	// Query MongoDB for documents with tags matching the roles
	var results []bson.M
	filter := bson.M{
		"tag": bson.M{"$in": roles},
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
