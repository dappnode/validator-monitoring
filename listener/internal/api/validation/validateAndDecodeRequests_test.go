package validation

import (
	"encoding/base64"
	"strconv"
	"testing"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
)

// Helper function to repeat a string
func repeatString(s string, count int) string {
	var result string
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

func TestValidateAndDecodeRequests(t *testing.T) {
	// Setup current time for timestamp tests
	currentTime := time.Now()
	validTimestamp := currentTime.AddDate(0, 0, -10).Unix() // 10 days ago, within valid range
	oldTimestamp := currentTime.AddDate(0, -2, 0).Unix()    // 2 months ago, too old

	// Create a base64 encoded valid payload with a Unix timestamp and correct type
	validEncodedPayload := base64.StdEncoding.EncodeToString([]byte(`{"type":"PROOF_OF_VALIDATION","platform":"dappnode","timestamp":"` + strconv.FormatInt(validTimestamp, 10) + `"}`))
	oldEncodedPayload := base64.StdEncoding.EncodeToString([]byte(`{"type":"PROOF_OF_VALIDATION","platform":"dappnode","timestamp":"` + strconv.FormatInt(oldTimestamp, 10) + `"}`))
	invalidTypePayload := base64.StdEncoding.EncodeToString([]byte(`{"type":"INVALID_TYPE","platform":"dappnode","timestamp":"` + strconv.FormatInt(validTimestamp, 10) + `"}`))

	// Example of a valid BLS public key for Eth2
	validBlsPubkey := "0xa06251962339450df57631d128fa54e4d54e2d17015571f1bcccd9b45c6ea971245f209cc9be087d5440bec19495a99a"
	invalidBlsPubkey := "0x123456" // Example of an invalid BLS public key (too short)

	// Mock requests
	requests := []types.SignatureRequest{
		{ // Valid request
			Payload:   validEncodedPayload,
			Pubkey:    validBlsPubkey,
			Signature: "0x" + repeatString("a", 192), // valid signature
			Tag:       "tag1",
		},
		{ // Missing fields
			Payload:   "",
			Pubkey:    "",
			Signature: "",
			Tag:       "",
		},
		{ // Invalid signature format
			Payload:   validEncodedPayload,
			Pubkey:    validBlsPubkey,
			Signature: "bad_signature",
			Tag:       "tag2",
		},
		{ // Old timestamp
			Payload:   oldEncodedPayload,
			Pubkey:    validBlsPubkey,
			Signature: "0x" + repeatString("a", 192),
			Tag:       "tag3",
		},
		{ // Invalid type
			Payload:   invalidTypePayload,
			Pubkey:    validBlsPubkey,
			Signature: "0x" + repeatString("a", 192),
			Tag:       "tag4",
		},
		{ // Invalid BLS public key
			Payload:   validEncodedPayload,
			Pubkey:    invalidBlsPubkey,
			Signature: "0x" + repeatString("a", 192),
			Tag:       "tag5",
		},
	}

	// Expected results is a slice of structs with the expected number of valid requests and whether we expect an error.
	expectedResults := []struct {
		expectError bool
		expectedLen int
	}{
		{false, 1}, // Expect one valid decoded request
		{false, 0}, // No valid request, all fields missing
		{false, 0}, // Signature format invalid
		{false, 0}, // Old timestamp
		{false, 0}, // Invalid type
		{false, 0}, // Invalid BLS public key
	}

	// Run tests. We expect the number of valid requests to match the expected results
	for i, req := range requests {
		decodedRequests, _ := ValidateAndDecodeRequests([]types.SignatureRequest{req})
		if len(decodedRequests) != expectedResults[i].expectedLen {
			t.Errorf("Test %d failed, expected %d valid requests, got %d", i+1, expectedResults[i].expectedLen, len(decodedRequests))
		}
	}
}
