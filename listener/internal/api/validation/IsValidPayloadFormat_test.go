package validation

import "testing"

// TestIsValidPayloadFormatValid tests IsValidPayloadFormat with a valid signature
func TestIsValidPayloadFormatValid(t *testing.T) {
	// Generate a valid signature
	validSignature := "0x" + "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2g3h4i5j6k7l8m9n0o1p2q3r4s5t6u7v8w9x0y1z2a3b4c5d6e7f8g9h"
	if len(validSignature) != 194 { // ensure the test signature length is correct
		t.Fatal("Test setup error: The valid signature does not meet the length requirement.")
	}

	// Call the function with a valid signature
	if !IsValidPayloadFormat(validSignature) {
		t.Errorf("IsValidPayloadFormat was incorrect, got: false, want: true.")
	}
}

// TestIsValidPayloadFormatInvalid tests IsValidPayloadFormat with invalid signatures
func TestIsValidPayloadFormatInvalid(t *testing.T) {
	// List of invalid signatures
	invalidSignatures := []string{
		"1x" + "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2g3h4i5j6k7l8m9n0o1p2q3r4s5t6u7v8w9x0y1z2a3b4c5d6e7f8g9h", // bad prefix
		"0x" + "a1b2c3d4e5", // too short
		"0x" + "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0u1v2w3x4y5z6a7b8c9d0e1f2g3h4i5j6k7l8m9n0o1p2q3r4s5t6u7v8w9x0y1z2a3b4c5d6e7f8g9ha", // too long
		"",   // empty string
		"0x", // only prefix
	}

	for _, sig := range invalidSignatures {
		if IsValidPayloadFormat(sig) {
			t.Errorf("IsValidPayloadFormat was incorrect, got: true, want: false for input %s", sig)
		}
	}
}
