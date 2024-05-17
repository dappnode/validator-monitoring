package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
)

type activeValidator struct {
	Pubkey                     string `json:"pubkey"`
	WithdrawalCredentials      string `json:"withdrawal_credentials"`
	EffectiveBalance           string `json:"effective_balance"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch string `json:"activation_eligibility_epoch"`
	ActivationEpoch            string `json:"activation_epoch"`
	ExitEpoch                  string `json:"exit_epoch"`
	WithdrawableEpoch          string `json:"withdrawable_epoch"`
}

// https://ethereum.github.io/beacon-APIs/#/Beacon /eth/v1/beacon/states/{state_id}/validators
type activeValidatorsApiResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                []struct {
		Index     string          `json:"index"`
		Balance   string          `json:"balance"`
		Status    string          `json:"status"`
		Validator activeValidator `json:"validator"`
	} `json:"data"`
}

// GetValidatorsStatus checks the active status of validators from a specific beacon node.
// @returns validatorStatusMap error
func GetValidatorsStatus(pubkeys []string, beaconNodeUrl string) (map[string]types.Status, error) {
	if len(pubkeys) == 0 {
		logger.Warn("No public keys provided to retrieve active validators")
		return nil, fmt.Errorf("no public keys provided to retrieve active validators from beacon node")
	}

	// Use a map to store validator statuses
	statusMap := make(map[string]types.Status)

	// Serialize the request body to JSON
	// See https://ethereum.github.io/beacon-APIs/#/Beacon/postStateValidators
	// returns only active validators
	jsonData, err := json.Marshal(struct {
		Ids      []string `json:"ids"`
		Statuses []string `json:"statuses"`
	}{
		Ids:      pubkeys,
		Statuses: []string{"active_ongoing"}, // Only interested in currently active validators
	})
	if err != nil {
		logger.Error("Failed to serialize request data: " + err.Error())
		return nil, err
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	apiUrl := fmt.Sprintf("%s/eth/v1/beacon/states/head/validators", beaconNodeUrl)

	// Make API call
	resp, err := client.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("Failed to make request to beacon node: " + err.Error())
		return getMapUnknown(pubkeys), nil
	}
	defer resp.Body.Close()

	// Check if it's any server error 5xx
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		logger.Error("internal server error due to server issue, keeping signatures to be stored with status unknown: " + resp.Status)
		return getMapUnknown(pubkeys), nil
	}

	// Check the HTTP response status before reading the body and return nil if not ok
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response status from beacon node when retrieving active validators: " + resp.Status)
	}

	// Decode the API response directly into the ApiResponse struct
	var apiResponse activeValidatorsApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("error decoding response data from beacon node: " + err.Error())
	}

	for _, pubkey := range pubkeys {
		statusMap[pubkey] = types.Inactive
	}
	for _, validator := range apiResponse.Data {
		statusMap[validator.Validator.Pubkey] = types.Active
	}

	return statusMap, nil
}

func getMapUnknown(pubkeys []string) map[string]types.Status {
	statusMap := make(map[string]types.Status)
	for _, pubkey := range pubkeys {
		statusMap[pubkey] = types.Unknown
	}
	return statusMap
}
