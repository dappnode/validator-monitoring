package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
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
		ids := make([]string, 0, len(requestsDecoded))
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
		client := &http.Client{Timeout: 10 * time.Second}
		url := fmt.Sprintf("%s/eth/v1/beacon/states/head/validators", url)
		resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("error making API call to %s: %v\n", url, err)
			continue
		}

		// check the HTTP response status before reading the body
		if resp.StatusCode != http.StatusOK {
			fmt.Printf("unexpected response status: %d\n", resp.StatusCode)
			continue
		}

		// read and log the response body
		responseData, err := io.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading response data: %v\n", err)
			continue
		}

		// assuming the server returns a list of active validators in the format expected
		if err := json.Unmarshal(responseData, &activeValidators); err != nil {
			fmt.Printf("error unmarshaling response data: %v\n", err)
			continue
		}

		// append the active validators to the list
		activeValidators = append(activeValidators, activeValidators...)

		// close the response body
		resp.Body.Close()
	}

	return activeValidators
}
