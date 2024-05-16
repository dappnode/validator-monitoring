package validation

import (
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
	"github.com/herumi/bls-eth-go-binary/bls"
)

func VerifySignature(req types.SignatureRequestDecodedWithActive) (bool, error) {
	// Decode the public key from hex, remove the 0x prefix ONLY if exists from req.Pubkey
	req.Pubkey = strings.TrimPrefix(req.Pubkey, "0x")
	req.Pubkey = strings.TrimSpace(req.Pubkey)
	pubkeyBytes, err := hex.DecodeString(req.Pubkey)
	if err != nil {
		logger.Error("Failed to decode public key from hex: " + err.Error())
		return false, err
	}
	var pubkeyDes bls.PublicKey
	if err := pubkeyDes.Deserialize(pubkeyBytes); err != nil {
		logger.Error("Failed to deserialize public key: " + err.Error())
		return false, err
	}

	// Decode the signature from hex, remove the 0x prefix ONLY if exists from req.Signature
	req.Signature = strings.TrimPrefix(req.Signature, "0x")
	req.Signature = strings.TrimSpace(req.Signature)
	sigBytes, err := hex.DecodeString(req.Signature)
	if err != nil {
		logger.Error("Failed to decode signature from hex: " + err.Error())
		return false, err
	}
	var sig bls.Sign
	if err := sig.Deserialize(sigBytes); err != nil {
		logger.Error("Failed to deserialize signature: " + err.Error())
		return false, err
	}

	// Serialize payload to string (assuming it's what was signed)
	payloadBytes, err := json.Marshal(req.DecodedPayload)
	if err != nil {
		logger.Error("Failed to serialize payload to string: " + err.Error())
		return false, err
	}
	// Verify the signature
	if !sig.VerifyByte(&pubkeyDes, payloadBytes) {
		logger.Debug("Failed to verify signature")
		return false, nil
	}

	return true, nil
}
