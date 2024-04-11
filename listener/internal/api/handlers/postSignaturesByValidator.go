package handlers

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type signaturesRequest struct {
	Pubkey  string `json:"pubkey"`
	Network string `json:"network"`
}

func PostSignaturesByValidator(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection) {
	var req signaturesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	filter := bson.M{"pubkey": req.Pubkey, "network": req.Network}
	cursor, err := dbCollection.Find(r.Context(), filter)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error fetching signatures from MongoDB")
		return
	}
	defer cursor.Close(r.Context())

	var signatures []bson.M
	if err = cursor.All(r.Context(), &signatures); err != nil {
		respondError(w, http.StatusInternalServerError, "Error reading signatures from cursor")
		return
	}

	respondOK(w, signatures)
}
