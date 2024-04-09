package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"

	"github.com/dappnode/validator-monitoring/listener/src/logger"
	"github.com/dappnode/validator-monitoring/listener/src/types"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (s *httpApi) handleRoot(w http.ResponseWriter, r *http.Request) {
	s.respondOK(w, "Server is running")
}

func (s *httpApi) handleSignature(w http.ResponseWriter, r *http.Request) {
	logger.Debug("Received new POST '/signature' request")
	var req types.SignatureRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		logger.Error("Failed to decode request body" + err.Error())
		s.respondError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// validate that the request has the required fields and that the signature is valid
	// we expect network and label to be strings, signature to be a 0x prefixed hex string of 194 characters
	// and payload to be a 0x prefixed hex string.
	if req.Network == "" || req.Label == "" || req.Signature == "" || req.Payload == "" {
		logger.Debug("Invalid request, missing fields. Request had the following fields: " + req.Network + " " + req.Label + " " + req.Signature + " " + req.Payload)
		s.respondError(w, http.StatusBadRequest, "Invalid request, missing fields")
		return
	}

	// validate the signature
	if !validateSignature(req.Signature) {
		logger.Debug("Invalid signature. It should be a 0x prefixed hex string of 194 characters. Request had the following signature: " + req.Signature)
		s.respondError(w, http.StatusBadRequest, "Invalid signature")
		return
	}

	// Decode BASE64 payload
	decodedBytes, err := base64.StdEncoding.DecodeString(req.Payload)
	if err != nil {
		logger.Error("Failed to decode BASE64 payload" + err.Error())
		s.respondError(w, http.StatusBadRequest, "Invalid BASE64 payload")
		return
	}
	var decodedPayload types.DecodedPayload
	err = json.Unmarshal(decodedBytes, &decodedPayload)
	if err != nil {
		logger.Error("Failed to decode JSON payload" + err.Error())
		s.respondError(w, http.StatusBadRequest, "Invalid payload format")
		return
	}

	// Validate the decoded payload (maybe we should be more strict here)
	if decodedPayload.Platform != "dappnode" || decodedPayload.Timestamp == "" || decodedPayload.Pubkey == "" {
		logger.Debug("Invalid payload content. Request had the following payload: " + decodedPayload.Platform + " " + decodedPayload.Timestamp + " " + decodedPayload.Pubkey)
		s.respondError(w, http.StatusBadRequest, "Payload content is invalid")
		return
	}

	logger.Debug("Request's payload decoded successfully. Inserting decoded message into MongoDB")

	// Insert into MongoDB
	_, err = s.dbCollection.InsertOne(r.Context(), bson.M{
		"platform":  decodedPayload.Platform,
		"timestamp": decodedPayload.Timestamp,
		"pubkey":    decodedPayload.Pubkey,
		"signature": req.Signature,
		"network":   req.Network,
		"label":     req.Label,
	})
	if err != nil {
		logger.Error("Failed to insert message into MongoDB" + err.Error())
		s.respondError(w, http.StatusInternalServerError, "Failed to insert payload into MongoDB")
		return
	}
	logger.Info("A new message with pubkey " + decodedPayload.Pubkey + " was decoded and inserted into MongoDB successfully")
	s.respondOK(w, "Message validated and inserted into MongoDB")

}

func (s *httpApi) handleGetSignatures(w http.ResponseWriter, r *http.Request) {
	var req types.SignaturesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	filter := bson.M{"pubkey": req.Pubkey, "network": req.Network}
	cursor, err := s.dbCollection.Find(r.Context(), filter)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Error fetching signatures from MongoDB")
		return
	}
	defer cursor.Close(r.Context())

	var signatures []bson.M
	if err = cursor.All(r.Context(), &signatures); err != nil {
		s.respondError(w, http.StatusInternalServerError, "Error reading signatures from cursor")
		return
	}

	s.respondOK(w, signatures)
}

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
func (s *httpApi) handleGetSignaturesAggr(w http.ResponseWriter, r *http.Request) {
	var req types.SignaturesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "Invalid request body")
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

	cursor, err := s.dbCollection.Aggregate(r.Context(), pipeline, options.Aggregate())
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "Error aggregating signatures from MongoDB")
		return
	}
	defer cursor.Close(r.Context())

	var results []bson.M
	if err := cursor.All(r.Context(), &results); err != nil {
		s.respondError(w, http.StatusInternalServerError, "Error reading aggregation results")
		return
	}

	// Respond with the aggregation results
	s.respondOK(w, results)
}

// TODO: error handling
func (s *httpApi) respondError(w http.ResponseWriter, code int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(types.HttpErrorResp{Code: code, Message: message})
}

// TODO: error handling
func (s *httpApi) respondOK(w http.ResponseWriter, response any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

}
