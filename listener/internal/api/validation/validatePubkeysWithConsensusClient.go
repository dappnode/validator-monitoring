package validation

// validatePubkeysWithConsensusClient simulates making a bulk request to a consensus client for validating pubkeys.
// It can return a map of validated pubkeys that exist as validators.
func ValidatePubkeysWithConsensusClient(pubkeys []string) map[string]bool {
	validatedPubkeys := make(map[string]bool)
	// make api call: GET /eth/v1/beacon/states/{state_id}/validators?id=validator_pubkey1,validator_pubkey2,validator_pubkey3

	for _, pubkey := range pubkeys {
		validatedPubkeys[pubkey] = true // Assuming all given pubkeys are valid
	}
	return validatedPubkeys
}
