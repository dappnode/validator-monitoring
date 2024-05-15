package types

type SignatureRequest struct {
	Payload   string `json:"payload"`
	Pubkey    string `json:"pubkey"`
	Signature string `json:"signature"`
	Network   string `json:"network"`
	Tag       string `json:"tag"`
}

type SignatureRequestDecoded struct {
	DecodedPayload DecodedPayload `json:"decodedPayload"`
	Payload        string         `json:"payload"`
	Pubkey         string         `json:"pubkey"`
	Signature      string         `json:"signature"`
	Network        string         `json:"network"`
	Tag            string         `json:"tag"`
}

type DecodedPayload struct {
	Type      string `json:"type"`
	Platform  string `json:"platform"`
	Timestamp string `json:"timestamp"`
}
