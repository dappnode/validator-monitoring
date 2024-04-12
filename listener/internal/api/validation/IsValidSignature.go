package validation

import (
	"encoding/hex"
	"encoding/json"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/herumi/bls-eth-go-binary/bls"
)

func IsValidSignature(req types.SignatureRequestDecoded) (bool, error) {
	// Initialize the BLS system
	if err := bls.Init(bls.BLS12_381); err != nil {
		return false, err
	}

	// Decode the public key from hex
	pubkeyBytes, err := hex.DecodeString(req.DecodedPayload.Pubkey)
	if err != nil {
		return false, err
	}
	var pubkey bls.PublicKey
	if err := pubkey.Deserialize(pubkeyBytes); err != nil {
		return false, err
	}

	// Decode the signature from hex
	sigBytes, err := hex.DecodeString(req.Signature)
	if err != nil {
		return false, err
	}
	var sig bls.Sign
	if err := sig.Deserialize(sigBytes); err != nil {
		return false, err
	}

	// Serialize payload to string (assuming it's what was signed)
	payloadBytes, err := json.Marshal(req.DecodedPayload)
	if err != nil {
		return false, err
	}

	// Verify the signature
	if !sig.VerifyByte(&pubkey, payloadBytes) {
		return false, nil
	}

	return true, nil
}
