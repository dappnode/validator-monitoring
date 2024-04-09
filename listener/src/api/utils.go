package api

import "github.com/dappnode/validator-monitoring/listener/src/types"

// A valid signature is a 0x prefixed hex string of 194 characters (including the prefix)
func validateSignature(signature string) bool {
	// validate the signature
	if len(signature) != 194 || signature[:2] != "0x" {
		return false
	}
	return true
}

// validatePubkeysWithConsensusClient simulates making a bulk request to a consensus client for validating pubkeys.
// It can return a map of validated pubkeys that exist as validators.
func validatePubkeysWithConsensusClient(pubkeys []string) map[string]bool {
	validatedPubkeys := make(map[string]bool)
	// make api call: GET /eth/v1/beacon/states/{state_id}/validators?id=validator_pubkey1,validator_pubkey2,validator_pubkey3

	for _, pubkey := range pubkeys {
		validatedPubkeys[pubkey] = true // Assuming all given pubkeys are valid
	}
	return validatedPubkeys
}

// Dummy implementation of validateSignatureAgainstPayload
func validateSignatureAgainstPayload(signature string, payload types.DecodedPayload) bool {
	// signature validation logic here
	return true
}
