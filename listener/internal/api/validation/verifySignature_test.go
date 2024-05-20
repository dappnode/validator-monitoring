package validation

import (
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/herumi/bls-eth-go-binary/bls"
)

func TestVerifySignature(t *testing.T) {
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
	}

	// Serialize the message
	messageBytes, err := json.Marshal(decodedPayload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Sign the message
	signature := secretKey.SignByte(messageBytes)

	// Prepare the request
	req := types.SignatureRequestDecodedWithStatus{
		SignatureRequestDecoded: types.SignatureRequestDecoded{
			DecodedPayload: decodedPayload,
			SignatureRequest: types.SignatureRequest{
				Pubkey:    publicKey.SerializeToHexStr(),
				Payload:   base64.StdEncoding.EncodeToString(messageBytes),
				Signature: signature.SerializeToHexStr(),
				Tag:       "solo"},
		},
		Status: types.Active,
	}

	// Validate the signature
	isValid, err := VerifySignature(req)
	if err != nil {
		t.Errorf("IsValidSignature returned an error: %v", err)
	}
	if !isValid {
		t.Errorf("IsValidSignature returned false, expected true")
	}
}

// TestVerifySignatureError tests the IsValidSignature function for expected errors
func TestVerifySignatureError(t *testing.T) {
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
	}

	// Serialize the payload
	payloadBytes, err := json.Marshal(decodedPayload)
	if err != nil {
		t.Fatalf("Failed to marshal payload: %v", err)
	}

	// Create the SignatureRequestDecoded with a bad signature to ensure it fails
	req := types.SignatureRequestDecodedWithStatus{
		SignatureRequestDecoded: types.SignatureRequestDecoded{
			DecodedPayload: decodedPayload,
			SignatureRequest: types.SignatureRequest{
				Pubkey:    badPublicKey,
				Payload:   base64.StdEncoding.EncodeToString(payloadBytes),
				Signature: "clearlyInvalidSignature",
				Tag:       "solo",
			},
		},
		Status: types.Active,
	}

	// Validate the signature
	isValid, err := VerifySignature(req)
	if err == nil {
		t.Errorf("Expected an error for invalid signature data, but got none")
	}
	if isValid {
		t.Errorf("Expected isValid to be false for invalid data, got true")
	}
}
