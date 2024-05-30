package test

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/gavv/httpexpect/v2"
	"github.com/herumi/bls-eth-go-binary/bls"
)

// TestPostSignaturesIntegration tests the POST /signatures endpoint. It expects a "listener" service to be running at
// http://localhost:8080, with the proper mongoDB connected to it. The test sends a series of requests with different payloads,
// public keys, signatures, and tags to the endpoint and checks the response status code.
func TestPostSignaturesIntegration(t *testing.T) {
	// Initialize BLS for the test
	if err := bls.Init(bls.BLS12_381); err != nil {
		t.Fatalf("Failed to initialize BLS: %v", err)
	}

	// Create a new HTTPExpect instance
	e := httpexpect.Default(t, "http://localhost:8080")

	// Generate valid BLS keys and signature
	var secretKey bls.SecretKey
	secretKey.SetByCSPRNG()
	publicKey := secretKey.GetPublicKey()
	publicKeyHex := "0x" + publicKey.SerializeToHexStr()

	// Prepare timestamps and payloads
	currentTime := time.Now()
	validTimestamp := currentTime.AddDate(0, 0, -10).UnixMilli() // timestamp is 10 days ago

	validDecodedPayload := types.DecodedPayload{
		Type:      "PROOF_OF_VALIDATION",
		Platform:  "dappnode",
		Timestamp: strconv.FormatInt(validTimestamp, 10),
	}
	payloadBytes, _ := json.Marshal(validDecodedPayload)
	validPayload := base64.StdEncoding.EncodeToString(payloadBytes)
	signature := secretKey.SignByte(payloadBytes)
	validSignature := "0x" + signature.SerializeToHexStr()
	invalidDecodedPayload := types.DecodedPayload{
		Type:      "INVALID_TYPE",
		Platform:  "dappnode",
		Timestamp: strconv.FormatInt(validTimestamp, 10),
	}
	invalidPayloadBytes, _ := json.Marshal(invalidDecodedPayload)
	invalidPayload := base64.StdEncoding.EncodeToString(invalidPayloadBytes)

	// Define test cases
	tests := []struct {
		description  string
		payload      string
		pubkey       string
		signature    string
		tag          types.Tag
		expectedCode int
	}{
		{
			description:  "Valid request",
			payload:      validPayload,
			pubkey:       publicKeyHex,
			signature:    validSignature,
			tag:          types.Solo,
			expectedCode: http.StatusOK,
		},
		{
			description:  "Invalid payload format",
			payload:      invalidPayload,
			pubkey:       publicKeyHex,
			signature:    validSignature,
			tag:          types.Solo,
			expectedCode: http.StatusBadRequest,
		},
		{
			description:  "Valid signature format arbitrary bytes signed, shouldnt pass the crypto verification",
			payload:      validPayload,
			pubkey:       publicKeyHex,
			signature:    "0x8bc341f083e34d27b8df9f48b0bfcdaa7ed009146969cee0d0d4e03afd383242e1767627d5e2ef50cce410dd02ed88280bb91309f96e5ad1ad31b204f1ed5e64a43cdf3c32603450b477a40df366f3ae145014cade0f22d588786f4f07bc7c7d",
			tag:          types.Solo,
			expectedCode: http.StatusBadRequest,
		},
		{
			description:  "Invalid BLS public key",
			payload:      validPayload,
			pubkey:       "0xinvalidKey",
			signature:    validSignature,
			tag:          types.Solo,
			expectedCode: http.StatusBadRequest,
		},
		{
			description:  "Invalid JSON format",
			payload:      `{bad json}`,
			pubkey:       publicKeyHex,
			signature:    validSignature,
			tag:          types.Solo,
			expectedCode: http.StatusBadRequest,
		},
		{
			description:  "Invalid tag",
			payload:      validPayload,
			pubkey:       publicKeyHex,
			signature:    validSignature,
			tag:          "invalidTag",
			expectedCode: http.StatusOK,
		},
	}

	// Execute tests
	for _, tc := range tests {
		t.Run(tc.description, func(t *testing.T) {
			e.POST("/signatures").
				WithQuery("network", "mainnet").
				WithJSON([]types.SignatureRequest{{
					Payload:   tc.payload,
					Pubkey:    tc.pubkey,
					Signature: tc.signature,
					Tag:       tc.tag,
				}}).
				Expect().
				Status(tc.expectedCode)
		})
	}
}
