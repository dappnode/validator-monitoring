package validation

import "github.com/dappnode/validator-monitoring/listener/internal/api/types"

// ValidatePubkeysWithConsensusClient checks if the given pubkeys from the requests are from active validators
// or not by making SINGLE API call to the consensus client. It returns an array of the active validators pubkeys.
func GetActiveValidators(requestsDecoded []types.SignatureRequestDecoded) ([]types.SignatureRequestDecoded, error) {
	requestsActiveValidators := requestsDecoded
	// make api call: GET /eth/v1/beacon/states/{state_id}/validators?id=validator_pubkey1,validator_pubkey2,validator_pubkey3

	return requestsActiveValidators, nil
}
