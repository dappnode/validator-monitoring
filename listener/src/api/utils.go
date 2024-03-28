package api

// A valid signature is a 0x prefixed hex string of 194 characters (including the prefix)
func validateSignature(signature string) bool {

	// validate the signature
	if len(signature) != 194 || signature[:2] != "0x" {
		return false
	}
	return true
}
