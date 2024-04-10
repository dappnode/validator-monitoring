package handlers

import (
	"encoding/json"
	"net/http"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Endpoint that returns an aggregation of all signatures for a given pubkey and network in this format:
// [
//
//	{
//	    "label": "example_label",
//	    "network": "stader",
//	    "pubkey": "0xb48c495c19082d892f38227bced89f7199f4e9b642bf94c7f2f1ccf29c0e6a6f54d653002513aa7cdb56c88368797ec",
//	    "signatures": [
//	        {
//	            "platform": "dappnode",
//	            "signature": "0xa8b00e7746a523346c5165dfa80ffafe52317c6fe6cdcfabd41886f9c8209b829266c5000597142b58dddbcc9c23cfd81315180acf18bb000db50d08293bc539e06a7c751d3d9dec89fb441b3ba6aefdeeff9cfed72fb41171173f22e2993e74",
//	            "timestamp": "185921877"
//	        },
//	        {
//	            "platform": "dappnode",
//	            "signature": "0xa8b00e7746a523346c5165dfa80ffafe52317c6fe6cdcfabd41886f9c8209b829266c5000597142b58dddbcc9c23cfd81315180acf18bb000db50d08293bc539e06a7c751d3d9dec89fb441b3ba6aefdeeff9cfed72fb41171173f22e2993e74",
//	            "timestamp": "185921877"
//	        }
//	    ]
//	}
//
// ]
func PostSignaturesByValidatorAggr(w http.ResponseWriter, r *http.Request, dbCollection *mongo.Collection) {
	var req signaturesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Define the aggregation pipeline
	// We should probably add pubkey to each signatures array element, so a 3rd party can easily verify the signature?
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"pubkey":  req.Pubkey,
				"network": req.Network,
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{"pubkey": "$pubkey", "network": "$network", "label": "$label"},
				"signatures": bson.M{
					"$push": bson.M{
						"signature": "$signature",
						"timestamp": "$timestamp",
						"platform":  "$platform",
					},
				},
			},
		},
		{
			"$project": bson.M{
				"_id":        0,
				"pubkey":     "$_id.pubkey",
				"network":    "$_id.network",
				"label":      "$_id.label",
				"signatures": 1,
			},
		},
	}

	cursor, err := dbCollection.Aggregate(r.Context(), pipeline, options.Aggregate())
	if err != nil {
		respondError(w, http.StatusInternalServerError, "Error aggregating signatures from MongoDB")
		return
	}
	defer cursor.Close(r.Context())

	var results []bson.M
	if err := cursor.All(r.Context(), &results); err != nil {
		respondError(w, http.StatusInternalServerError, "Error reading aggregation results")
		return
	}

	// Respond with the aggregation results
	respondOK(w, results)
}
