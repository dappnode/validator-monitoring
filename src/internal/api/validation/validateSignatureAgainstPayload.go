package validation

import "github.com/dappnode/validator-monitoring/listener/internal/api/types"

// Dummy implementation of validateSignatureAgainstPayload
func ValidateSignatureAgainstPayload(signature string, payload types.DecodedPayload) bool {
	// signature validation logic here
	return true
}
