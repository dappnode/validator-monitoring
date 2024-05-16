package validation

import (
	"testing"

	"github.com/dappnode/validator-monitoring/listener/internal/api/types"
)

func TestGetActiveValidators(t *testing.T) {
	// Setup the input data
	beaconNodeUrls := map[string]string{
		"holesky": "https://holeskyvals.53650f79ab75c6ff.dyndns.dappnode.io",
	}

	requestsDecoded := []types.SignatureRequestDecoded{
		{
			SignatureRequest: types.SignatureRequest{
				Network: "holesky",
				Pubkey:  "0xa685beb5a1f317f5a01ecd6dade42113aad945b2ab53fb1b356334ab441323e538feadd2889894b17f8fa2babe1989ca",
			},
			DecodedPayload: types.DecodedPayload{},
		},
		{
			SignatureRequest: types.SignatureRequest{
				Network: "holesky",
				Pubkey:  "0xab31efdd97f32087e96d3262f6fb84a4480411d391689be0dfc931fd8a5c16c3f51f10b127040b1cb65eb955f2b78a63"},
			DecodedPayload: types.DecodedPayload{},
		},
		{
			SignatureRequest: types.SignatureRequest{
				Network: "holesky",
				Pubkey:  "0xa24a030d7d8ca3c5e1f5824760d0f4157a7a89bcca6414377cca97e6e63445bef0e1b63761ee35a0fc46bb317e31b34b"},
			DecodedPayload: types.DecodedPayload{},
		},
	}

	result := GetActiveValidators(requestsDecoded, beaconNodeUrls["holesky"])

	// You may need to mock the server's response or adjust the expected values here according to your actual setup
	expectedNumValidators := 3 // This should match the number of mock validators that are "active"
	if len(result) != expectedNumValidators {
		t.Errorf("Expected %d active validators, got %d", expectedNumValidators, len(result))
	}
}
