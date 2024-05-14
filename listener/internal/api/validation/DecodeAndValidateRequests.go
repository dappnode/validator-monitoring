package validation

import (
	"encoding/base64"
	"encoding/json"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
	"github.com/dappnode/validator-monitoring/listener/internal/logger"
)

func DecodeAndValidateRequests(requests []types.SignatureRequest) ([]types.SignatureRequestDecoded, error) {
	var validRequests []types.SignatureRequestDecoded
	for _, req := range requests {
		if req.Network == "" || req.Tag == "" || req.Signature == "" || req.Payload == "" || req.Pubkey == "" {
			logger.Debug("Skipping invalid signature from request, missing required fields")
			continue
		}
		if !IsValidPayloadFormat(req.Signature) {
			logger.Debug("Skipping invalid signature from request, invalid signature format: " + req.Signature)
			continue
		}
		decodedBytes, err := base64.StdEncoding.DecodeString(req.Payload)
		if err != nil {
			logger.Error("Failed to decode BASE64 payload from request: " + err.Error())
			continue
		}
		var decodedPayload types.DecodedPayload
		if err := json.Unmarshal(decodedBytes, &decodedPayload); err != nil {
			logger.Error("Failed to decode JSON payload from request: " + err.Error())
			continue
		}
		if decodedPayload.Platform == "dappnode" && decodedPayload.Timestamp != "" {
			validRequests = append(validRequests, types.SignatureRequestDecoded{
				DecodedPayload: decodedPayload,
				Payload:        req.Payload,
				Pubkey:         req.Pubkey,
				Signature:      req.Signature,
				Network:        req.Network,
				Tag:            req.Tag,
			})
		} else {
			logger.Debug("Skipping invalid signature from request, invalid payload format")
		}
	}

	return validRequests, nil
}
