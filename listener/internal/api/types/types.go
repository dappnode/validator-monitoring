package types

type SignatureRequestDecoded struct {
	DecodedPayload DecodedPayload `json:"decodedPayload"`
	Payload        string         `json:"payload"`
	Signature      string         `json:"signature"`
	Network        string         `json:"network"`
	Label          string         `json:"label"`
}

type DecodedPayload struct {
	Platform  string `json:"platform"`
	Timestamp string `json:"timestamp"`
	Pubkey    string `json:"pubkey"`
}

type ActiveValidator struct {
	Pubkey                     string `json:"pubkey"`
	WithdrawalCredentials      string `json:"withdrawal_credentials"`
	EffectiveBalance           string `json:"effective_balance"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch string `json:"activation_eligibility_epoch"`
	ActivationEpoch            string `json:"activation_epoch"`
	ExitEpoch                  string `json:"exit_epoch"`
	WithdrawableEpoch          string `json:"withdrawable_epoch"`
}

type ActiveValidatorsApiResponse struct {
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
	Data                []struct {
		Index     string          `json:"index"`
		Balance   string          `json:"balance"`
		Status    string          `json:"status"`
		Validator ActiveValidator `json:"validator"`
	} `json:"data"`
}
