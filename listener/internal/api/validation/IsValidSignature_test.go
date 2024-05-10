package validation

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/herumi/bls-eth-go-binary/bls"
)

func TestIsValidSignature(t *testing.T) {
	// Initialize BLS
	if err := bls.Init(bls.BLS12_381); err != nil {
		t.Fatalf("Failed to initialize BLS: %v", err)
	}

	// Generate a new key pair
	var secretKey bls.SecretKey
	secretKey.SetByCSPRNG()
	publicKey := secretKey.GetPublicKey()

	// Prepare the message
	decodedPayload := types.DecodedPayload{
		Platform:  "dappnode",
		Timestamp: "2024-04-01T15:00:00Z",
		Pubkey:    publicKey.SerializeToHexStr(),
	}

	// Serialize the message
	messageBytes, err := json.Marshal(decodedPayload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Sign the message
	signature := secretKey.SignByte(messageBytes)

	// Prepare the request
	req := types.SignatureRequestDecoded{
		DecodedPayload: decodedPayload,
		Payload:        base64.StdEncoding.EncodeToString(messageBytes),
		Signature:      signature.SerializeToHexStr(),
		Network:        "mainnet",
		Tag:            "solo",
	}

	// Validate the signature
	isValid, err := IsValidSignature(req)
	if err != nil {
		t.Errorf("IsValidSignature returned an error: %v", err)
	}
	if !isValid {
		t.Errorf("IsValidSignature returned false, expected true")
	}
}

// TestIsValidSignatureError tests the IsValidSignature function for expected errors
func TestIsValidSignatureError(t *testing.T) {
	// Initialize BLS just as in normal tests
	if err := bls.Init(bls.BLS12_381); err != nil {
		t.Fatalf("Failed to initialize BLS: %v", err)
	}

	// Example of an intentionally bad public key (corrupted or incomplete)
	badPublicKey := "invalidPublicKeyData"

	// Setup DecodedPayload with invalid data
	decodedPayload := types.DecodedPayload{
		Platform:  "dappnode",
		Timestamp: "2024-04-01T15:00:00Z",
		Pubkey:    badPublicKey,
	}

	// Serialize the payload
	payloadBytes, err := json.Marshal(decodedPayload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create the SignatureRequestDecoded with a bad signature to ensure it fails
	req := types.SignatureRequestDecoded{
		DecodedPayload: decodedPayload,
		Payload:        base64.StdEncoding.EncodeToString(payloadBytes),
		Signature:      "clearlyInvalidSignature", // Intentionally invalid
		Network:        "mainnet",
		Tag:            "solo",
	}

	// Validate the signature
	isValid, err := IsValidSignature(req)
	if err == nil {
		t.Errorf("Expected an error for invalid signature data, but got none")
	}
	if isValid {
		t.Errorf("Expected isValid to be false for invalid data, got true")
	}
}
