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

func GetActiveValidators(requestsDecoded []types.SignatureRequestDecoded, beaconNodeUrl string) ([]types.SignatureRequestDecoded, error) {
	if len(requestsDecoded) == 0 {
		return nil, fmt.Errorf("no validators to check")
	}

	// Prepare the request body
	ids := make([]string, 0, len(requestsDecoded))
	for _, req := range requestsDecoded {
		ids = append(ids, req.DecodedPayload.Pubkey)
	}
	requestBody := struct {
		Ids      []string `json:"ids"`
		Statuses []string `json:"statuses"`
	}{
		Ids:      ids,
		Statuses: []string{"active_ongoing"},
	}

	// Serialize the request body to JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request body: %w", err)
	}

	// Configure HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	url := fmt.Sprintf("%s/eth/v1/beacon/states/head/validators", beaconNodeUrl)
	resp, err := client.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("error making API call to %s: %w", url, err)
	}
	defer resp.Body.Close()

	// Check the HTTP response status before reading the body
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API call to %s returned status %d", url, resp.StatusCode)
	}

	// Read and log the response body
	responseData, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response data: %w", err)
	}

	// Assuming the server returns a list of active validators in the format you expect
	var activeValidators []types.SignatureRequestDecoded
	if err := json.Unmarshal(responseData, &activeValidators); err != nil {
		return nil, fmt.Errorf("error unmarshaling response data: %w", err)
	}

	return activeValidators, nil
}
