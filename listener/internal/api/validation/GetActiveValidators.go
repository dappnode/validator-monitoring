package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
)

// GetActiveValidators checks the active status of validators from a specific beacon node.
// If bypass is true, it simply returns all decoded requests.
func GetActiveValidators(requestsDecoded []types.SignatureRequestDecoded, beaconNodeUrl string, bypass bool) []types.SignatureRequestDecoded {

	if len(requestsDecoded) == 0 {
		fmt.Println("no requests to process")
		return nil
	}

	ids := make([]string, 0, len(requestsDecoded))
	for _, req := range requestsDecoded {
		ids = append(ids, req.DecodedPayload.Pubkey)
	}

	if len(ids) == 0 {
		fmt.Println("no valid public keys for network ", beaconNodeUrl, " to query")
		return nil
	}

	// Serialize the request body to JSON
	// See https://ethereum.github.io/beacon-APIs/#/Beacon/postStateValidators
	jsonData, err := json.Marshal(struct {
		Ids      []string `json:"ids"`
		Statuses []string `json:"statuses"`
	}{
		Ids:      ids,
		Statuses: []string{"active_ongoing"}, // Only interested in currently active validators
	})
	if err != nil {
		fmt.Printf("error marshaling request data: %v\n", err)
		return nil
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 50 * time.Second}
	apiUrl := fmt.Sprintf("%s/eth/v1/beacon/states/head/validators", beaconNodeUrl)

	// Make API call
	resp, err := client.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("error making API call to %s: %v\n", apiUrl, err)
		return nil
	}
	defer resp.Body.Close()

	// Check the HTTP response status before reading the body
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("unexpected response status: %d\n", resp.StatusCode)
		return nil
	}

	// Decode the API response directly into the ApiResponse struct
	var apiResponse types.ActiveValidatorsApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		fmt.Printf("error decoding response data: %v\n", err)
		return nil
	}

	// Use a map to quickly lookup active validators
	activeValidatorMap := make(map[string]bool)
	for _, validator := range apiResponse.Data {
		activeValidatorMap[validator.Validator.Pubkey] = true
	}

	// Filter the list of decoded requests to include only those that are active
	var activeValidators []types.SignatureRequestDecoded
	for _, req := range requestsDecoded {
		if _, isActive := activeValidatorMap[req.DecodedPayload.Pubkey]; isActive {
			activeValidators = append(activeValidators, req)
		}
	}

	if bypass {
		return requestsDecoded // do not return the filtered list
	} else {
		return activeValidators // return the filtered list (default behaviour)
	}
}
