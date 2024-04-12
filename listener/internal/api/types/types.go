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
