package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
)

func GetActiveValidators(requestsDecoded []types.SignatureRequestDecoded, beaconNodeUrls map[string]string) []types.SignatureRequestDecoded {
	if len(requestsDecoded) == 0 {
		fmt.Println("no requests to process")
		return nil
	}

	var activeValidators []types.SignatureRequestDecoded

	// iterate over the networks available in the beaconNodeUrls map
	for network, url := range beaconNodeUrls {
		// prepare request body, get the list of ids from the requestsDecoded for the current network
		ids := make([]string, 0)
		for _, req := range requestsDecoded {
			if req.Network == network {
				ids = append(ids, req.DecodedPayload.Pubkey)
			}
		}
		// if there are no ids for the current network, log and skip it
		if len(ids) == 0 {
			fmt.Printf("no ids for network %s\n", network)
			continue
		}

		// serialize the request body to JSON
		// see https://ethereum.github.io/beacon-APIs/#/Beacon/postStateValidators
		jsonData, err := json.Marshal(struct {
			Ids      []string `json:"ids"`
			Statuses []string `json:"statuses"`
		}{
			Ids:      ids,
			Statuses: []string{"active_ongoing"},
		})
		if err != nil {
			fmt.Printf("error marshaling request data: %v\n", err)
			continue
		}

		// configure HTTP client with timeout
		// TODO: test timeout
		client := &http.Client{Timeout: 50 * time.Second}
		apiUrl := fmt.Sprintf("%s/eth/v1/beacon/states/head/validators", url)
		resp, err := client.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("error making API call to %s: %v\n", apiUrl, err)
			continue
		}
		defer resp.Body.Close() // close the response body when the function returns

		// check the HTTP response status before reading the body
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("unexpected response status: %d\n", resp.StatusCode)
			continue
		}

		// Decode the API response directly into the ApiResponse struct
		var apiResponse types.ActiveValidatorsApiResponse
		if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
			fmt.Printf("error decoding response data: %v\n", err)
			continue
		}

		// Map to store validator public keys for quick lookup
		validatorPubKeys := make(map[string]bool)
		for _, data := range apiResponse.Data {
			validatorPubKeys[data.Validator.Pubkey] = true
		}

		// Build the list of active validators for the current network
		// by filtering the requestsDecoded slice. This is done by checking
		// if the pubkey of each request is present in the validatorPubKeys map.
		for _, req := range requestsDecoded {
			if req.Network == network {
				if _, exists := validatorPubKeys[req.DecodedPayload.Pubkey]; exists {
					activeValidators = append(activeValidators, req)
				}
			}
		}
	}
	return activeValidators
}
