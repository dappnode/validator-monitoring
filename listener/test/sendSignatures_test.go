package test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/gavv/httpexpect/v2"
	"github.com/herumi/bls-eth-go-binary/bls"
)

func readJWT(filePath string) (string, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// generateSignature generates a BLS signature and base64 encoded payload using a secret key.
func generateSignature(payload types.DecodedPayload, secretKey *bls.SecretKey) (signatureHex, payloadBase64 string) {
	payloadBytes, _ := json.Marshal(payload)
	signature := secretKey.SignByte(payloadBytes)
	signatureHex = "0x" + signature.SerializeToHexStr()
	payloadBase64 = base64.StdEncoding.EncodeToString(payloadBytes)
	return
}

func setupSecretKey() *bls.SecretKey {
	var secretKey bls.SecretKey
	secretKey.SetByCSPRNG()
	return &secretKey
}

func TestPostSignaturesIntegration(t *testing.T) {
	// BLS initialization
	if err := bls.Init(bls.BLS12_381); err != nil {
		t.Fatalf("Failed to initialize BLS: %v", err)
	}
	e := httpexpect.Default(t, "http://localhost:8080")

	// Setup test data
	currentTime := time.Now().UnixMilli()
	secretKey := setupSecretKey()
	secretKey2 := setupSecretKey()
	// Get the current working directory
	dir, err := os.Getwd()
	if err != nil {
		// Handle the error if getting the working directory fails
		fmt.Println("Error:", err)
		return
	}

	// Print the current working directory
	fmt.Println("Current directory:", dir)
	// Read JWT token from file
	jwtToken, err := readJWT("data/token.jwt")
	if err != nil {
		t.Fatalf("Failed to read JWT: %v", err)
	}

	testCases := []struct {
		description  string
		payload      types.DecodedPayload
		secretKey    *bls.SecretKey
		expectedCode int
		tag          types.Tag
		invalidKey   bool   // Use invalid key if true
		customSig    string // Use custom signature if not empty
	}{
		{
			description: "Valid request solo tag",
			payload: types.DecodedPayload{
				Type:      "PROOF_OF_VALIDATION",
				Platform:  "dappnode",
				Timestamp: strconv.FormatInt(currentTime-10*24*60*60*1000, 10), // 10 days ago
			},
			secretKey:    secretKey,
			expectedCode: http.StatusOK,
			tag:          types.Solo,
		},
		{
			description: "Valid request solo tag with new timestamp",
			payload: types.DecodedPayload{
				Type:      "PROOF_OF_VALIDATION",
				Platform:  "dappnode",
				Timestamp: strconv.FormatInt(currentTime-5*24*60*60*1000, 10), // 5 days ago
			},
			secretKey:    secretKey, // Reuse the key for a different payload
			expectedCode: http.StatusOK,
			tag:          types.Solo,
		},
		{
			description: "Valid request solo tag with new timestamp and pubkey",
			payload: types.DecodedPayload{
				Type:      "PROOF_OF_VALIDATION",
				Platform:  "dappnode",
				Timestamp: strconv.FormatInt(currentTime-5*24*60*60*1000, 10), // 5 days ago
			},
			secretKey:    secretKey2, // Reuse the key for a different payload
			expectedCode: http.StatusOK,
			tag:          types.Solo,
		},
		{
			description: "Valid request ssv tag",
			payload: types.DecodedPayload{
				Type:      "PROOF_OF_VALIDATION",
				Platform:  "dappnode",
				Timestamp: strconv.FormatInt(currentTime-10*24*60*60*1000, 10),
			},
			secretKey:    secretKey,
			expectedCode: http.StatusOK,
			tag:          types.Ssv,
		},
		{
			description: "Invalid payload format",
			payload: types.DecodedPayload{
				Type:      "INVALID_TYPE",
				Platform:  "dappnode",
				Timestamp: strconv.FormatInt(currentTime-10*24*60*60*1000, 10),
			},
			secretKey:    secretKey,
			expectedCode: http.StatusBadRequest,
			tag:          types.Solo,
		},
		{
			description: "Valid signature format arbitrary bytes signed, shouldn't pass the crypto verification",
			payload: types.DecodedPayload{
				Type:      "PROOF_OF_VALIDATION",
				Platform:  "dappnode",
				Timestamp: strconv.FormatInt(currentTime-10*24*60*60*1000, 10),
			},
			secretKey:    secretKey,
			expectedCode: http.StatusBadRequest,
			tag:          types.Solo,
			customSig:    "0x8bc341f083e34d27b8df9f48b0bfcdaa7ed009146969cee0d0d4e03afd383242e1767627d5e2ef50cce410dd02ed88280bb91309f96e5ad1ad31b204f1ed5e64a43cdf3c32603450b477a40df366f3ae145014cade0f22d588786f4f07bc7c7d",
		},
		{
			description: "Invalid BLS public key",
			payload: types.DecodedPayload{
				Type:      "PROOF_OF_VALIDATION",
				Platform:  "dappnode",
				Timestamp: strconv.FormatInt(currentTime-10*24*60*60*1000, 10),
			},
			secretKey:    secretKey,
			expectedCode: http.StatusBadRequest,
			tag:          types.Solo,
			invalidKey:   true, // Use an invalid key in test
		},
		{
			description: "Invalid JSON format",
			payload: types.DecodedPayload{
				Type:      "INVALID_JSON",
				Platform:  "dappnode",
				Timestamp: `{bad json}`,
			},
			secretKey:    secretKey,
			expectedCode: http.StatusBadRequest,
			tag:          types.Solo,
		},
		{
			description: "Invalid tag",
			payload: types.DecodedPayload{
				Type:      "PROOF_OF_VALIDATION",
				Platform:  "dappnode",
				Timestamp: strconv.FormatInt(currentTime-10*24*60*60*1000, 10),
			},
			secretKey:    secretKey,
			expectedCode: http.StatusBadRequest,
			tag:          "invalidTag",
		},
	}

	// wait some seconds to ensure the server is ready
	time.Sleep(10 * time.Second)

	// Execute test cases
	for _, tc := range testCases {
		sigHex, payloadBase64 := generateSignature(tc.payload, tc.secretKey)
		if tc.customSig != "" {
			sigHex = tc.customSig
		}
		pubKeyHex := "0x" + tc.secretKey.GetPublicKey().SerializeToHexStr()
		if tc.invalidKey {
			pubKeyHex = "0xinvalidKey"
		}

		t.Run(tc.description, func(t *testing.T) {
			e.POST("/signatures").
				WithQuery("network", "mainnet").
				WithJSON([]types.SignatureRequest{{
					Payload:   payloadBase64,
					Pubkey:    pubKeyHex,
					Signature: sigHex,
					Tag:       tc.tag,
				}}).
				Expect().
				Status(tc.expectedCode)
		})
	}

	// wait some seconds to ensure all the signatures are processed
	time.Sleep(15 * time.Second)

	t.Run("Fetch Signatures with JWT", func(t *testing.T) {
		response := e.GET("/signatures").
			WithHeader("Authorization", "Bearer "+jwtToken).
			Expect().
			Status(http.StatusOK).
			JSON().Array()

		// // Convert the response to a pretty JSON format. Useful for debugging
		// jsonResponse, err := json.MarshalIndent(response.Raw(), "", "    ")
		// if err != nil {
		// 	t.Fatalf("Failed to marshal JSON: %v", err)
		// }
		// fmt.Printf("Response: %s\n", string(jsonResponse))

		// Verify that the response contains two elements
		response.Length().IsEqual(2)

		// Verify that the first element has pubkey corresponding to secretKey
		response.Value(0).Object().Value("pubkey").String().IsEqual("0x" + secretKey.GetPublicKey().SerializeToHexStr())

		// Verify that the second element has pubkey corresponding to secretKey2
		response.Value(1).Object().Value("pubkey").String().IsEqual("0x" + secretKey2.GetPublicKey().SerializeToHexStr())
	})
}
