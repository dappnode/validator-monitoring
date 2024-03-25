package types

type SignatureRequest struct {
	Payload   string `json:"payload"`
	Signature string `json:"signature"`
	Network   string `json:"network"`
	Label     string `json:"label"`
}

type DecodedPayload struct {
	Platform  string `json:"platform"`
	Timestamp string `json:"timestamp"`
	Pubkey    string `json:"pubkey"`
}

type HttpErrorResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
