package types

type DecodedPayload struct {
	Platform  string `json:"platform"`
	Timestamp string `json:"timestamp"`
	Pubkey    string `json:"pubkey"`
}
