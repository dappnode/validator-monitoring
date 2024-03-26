package api

// A valid signature is a 0x prefixed hex string of 192 characters (without the prefix)
// A valid payload is a 0x prefixed hex string.
func validateSignature(payload string, signature string) bool {
	// validate the payload
	if len(payload) < 2 || payload[:2] != "0x" {
		return false
	}

	// validate the signature
	if len(signature) != 194 || signature[:2] != "0x" {
		return false
	}
	return true
}
