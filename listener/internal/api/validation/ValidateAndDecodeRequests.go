package validation

import (
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
)

// ValidateAndDecodeRequests filters out Recieved Invalid Reques from the input array. The returned array contains only the valid requests, with the payload decoded.
func ValidateAndDecodeRequests(requests []types.SignatureRequest) ([]types.SignatureRequestDecoded, error) {
	var validRequests []types.SignatureRequestDecoded
	for _, req := range requests {
		if !isValidCodedRequest(&req) {
			logger.Debug("Skipping request due to invalid fields or format.")
			continue
		}
		decodedPayload, err := decodeAndValidatePayload(req.Payload)
		if err != nil {
			logger.Error("Failed to decode payload: " + err.Error())
			continue
		}
		validRequests = append(validRequests, types.SignatureRequestDecoded{
			DecodedPayload: decodedPayload,
			SignatureRequest: types.SignatureRequest{
				Payload:   req.Payload,
				Pubkey:    req.Pubkey,
				Signature: req.Signature,
				Network:   req.Network,
				Tag:       req.Tag,
			},
		})
	}
	return validRequests, nil
}

// isValidCodedRequest checks if the request has all the required fields, the correct signature format, and a valid BLS pubkey
// TODO: we should consider having an enum for Network and Tag fields and validate them as well.
func isValidCodedRequest(req *types.SignatureRequest) bool {
	// Check for any empty required fields
	if req.Network == "" || req.Tag == "" || req.Signature == "" || req.Payload == "" || req.Pubkey == "" {
		logger.Debug("Received Invalid Request: One or more required fields are empty.")
		return false
	}

	// Check if the signature format is correct (should start with '0x' and be 194 characters long)
	if len(req.Signature) != 194 || req.Signature[:2] != "0x" {
		logger.Debug("Received Invalid Request: Signature format is incorrect.")
		return false
	}

	// Validate BLS public key: should start with '0x' and be 98 characters long (96 hex characters + '0x')
	if len(req.Pubkey) != 98 || req.Pubkey[:2] != "0x" {
		logger.Debug("Received Invalid Request: Public key format is incorrect.")
		return false
	}

	// Decode the public key to make sure it's a valid hex and exactly 48 bytes long
	pubKeyBytes, err := hex.DecodeString(req.Pubkey[2:]) // Skip '0x' prefix
	if err != nil || len(pubKeyBytes) != 48 {
		logger.Debug("Received Invalid Request: Public key is not a valid BLS key.")
		return false
	}

	// TODO: verify also signature

	return true
}

// decodeAndValidatePayload decodes the base64 encoded payload and validates the format. It must be a valid JSON with the correct fields:
// - Platform: "dappnode"
// - Type: "PROOF_OF_VALIDATION"
// - Timestamp: a valid Unix timestamp within the last 30 days
func decodeAndValidatePayload(payload string) (types.DecodedPayload, error) {
	// Decode the base64 payload into bytes and unmarshal into DecodedPayload
	decodedBytes, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return types.DecodedPayload{}, errors.New("invalid base64 encoding")
	}

	var decodedPayload types.DecodedPayload
	if err := json.Unmarshal(decodedBytes, &decodedPayload); err != nil {
		return types.DecodedPayload{}, errors.New("error unmarshalling JSON")
	}

	// validate platform
	if decodedPayload.Platform != "dappnode" {
		return types.DecodedPayload{}, errors.New("invalid payload: must be from 'dappnode'")
	}

	// validate type
	if decodedPayload.Type != "PROOF_OF_VALIDATION" {
		return types.DecodedPayload{}, errors.New("invalid type: must be 'PROOF_OF_VALIDATION'")
	}

	// validate timestamp. Must be a valid Unix timestamp within the last 30 days
	timestampSecs, err := strconv.ParseInt(decodedPayload.Timestamp, 10, 64)
	if err != nil {
		return types.DecodedPayload{}, errors.New("timestamp is not a valid Unix timestamp")
	}

	timestampTime := time.Unix(timestampSecs, 0)
	if time.Since(timestampTime) > 30*24*time.Hour || decodedPayload.Timestamp == "" {
		return types.DecodedPayload{}, errors.New("invalid or old timestamp: must be within the last 30 days and not empty")
	}

	return decodedPayload, nil
}
