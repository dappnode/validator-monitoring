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
	SignatureRequest
}

type SignatureRequestDecodedWithActive struct {
	SignatureRequestDecoded
	Status string `json:"status"` // "unknown" | "active" | "inactive"
}

type DecodedPayload struct {
	Type      string `json:"type"`
	Platform  string `json:"platform"`
	Timestamp string `json:"timestamp"`
}
