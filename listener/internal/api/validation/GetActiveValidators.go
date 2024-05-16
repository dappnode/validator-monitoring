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

// GetActiveValidators checks the active status of validators from a specific beacon node.
func GetActiveValidators(requestsDecoded []types.SignatureRequestDecoded, beaconNodeUrl string) []types.SignatureRequestDecodedWithActive {
	if len(requestsDecoded) == 0 {
		logger.Warn("No requests to process")
		return nil
	}

	ids := make([]string, 0, len(requestsDecoded))
	for _, req := range requestsDecoded {
		ids = append(ids, req.Pubkey)
	}
	if len(ids) == 0 {
		logger.Warn("No valid public keys for network " + beaconNodeUrl + " to query")
		return nil
	}

	// Serialize the request body to JSON
	// See https://ethereum.github.io/beacon-APIs/#/Beacon/postStateValidators
	// returns only active validators
	jsonData, err := json.Marshal(struct {
		Ids      []string `json:"ids"`
		Statuses []string `json:"statuses"`
	}{
		Ids:      ids,
		Statuses: []string{"active_ongoing"}, // Only interested in currently active validators
	})
	if err != nil {
		logger.Error("Failed to serialize request data: " + err.Error())
		return nil
	}

	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}
	apiUrl := fmt.Sprintf("%s/eth/v1/beacon/states/head/validators", beaconNodeUrl)

	// Make API call
	resp, err := client.Post(apiUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		logger.Error("error making API call to " + apiUrl + ": " + err.Error())
		// if the api call fails return signatures with status unknown
		return GetSignatureRequestsDecodedWithUnknown(requestsDecoded)
	}
	defer resp.Body.Close()

	// check if its any server error 5xx
	// if its internal server error return unknown since we expect the cron to eventually resolve the status once the server is back up
	if resp.StatusCode >= 500 && resp.StatusCode < 600 {
		logger.Error("internal server error, returning signatures with status unknown: " + resp.Status)
		return GetSignatureRequestsDecodedWithUnknown(requestsDecoded)
	}

	// Check the HTTP response status before reading the body and return nil if not ok
	if resp.StatusCode != http.StatusOK {
		logger.Error("unexpected response status from beacon node: " + resp.Status)
		return nil
	}

	// Decode the API response directly into the ApiResponse struct
	var apiResponse activeValidatorsApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		logger.Error("error decoding response data: " + err.Error())
		return nil
	}

	// Use a map to quickly lookup active validators
	activeValidatorMap := make(map[string]bool)
	for _, validator := range apiResponse.Data {
		activeValidatorMap[validator.Validator.Pubkey] = true
	}

	// Filter the list of decoded requests to include only those that are active
	var activeValidators []types.SignatureRequestDecodedWithActive
	for _, req := range requestsDecoded {
		if _, isActive := activeValidatorMap[req.Pubkey]; isActive {
			activeValidators = append(activeValidators, types.SignatureRequestDecodedWithActive{
				SignatureRequestDecoded: req,
				Status:                  "active",
			})
		} else {
			// do not append inactive validators
			logger.Warn("Inactive validator: " + req.Pubkey)
		}
	}

	return activeValidators
}

// Append "unknown" status to all requests
func GetSignatureRequestsDecodedWithUnknown(requests []types.SignatureRequestDecoded) []types.SignatureRequestDecodedWithActive {
	var signatureRequestsDecodedWithActive []types.SignatureRequestDecodedWithActive
	for _, req := range requests {
		signatureRequestsDecodedWithActive = append(signatureRequestsDecodedWithActive, types.SignatureRequestDecodedWithActive{
			SignatureRequestDecoded: req,
			Status:                  "unknown",
		})
	}
	return signatureRequestsDecodedWithActive
}
