package types

type Network string // "mainnet" | "holesky" | "gnosis" | "lukso"

const (
	Mainnet Network = "mainnet"
	Holesky Network = "holesky"
	Gnosis  Network = "gnosis"
	Lukso   Network = "lukso"
)

type SignatureRequest struct {
	Payload   string  `json:"payload"`
	Pubkey    string  `json:"pubkey"`
	Signature string  `json:"signature"`
	Network   Network `json:"network"`
	Tag       string  `json:"tag"`
}

type SignatureRequestDecoded struct {
	DecodedPayload DecodedPayload `json:"decodedPayload"`
	SignatureRequest
}

type Status string

// create enum with status
const (
	Unknown  Status = "unknown"
	Active   Status = "active"
	Inactive Status = "inactive"
)

type SignatureRequestDecodedWithActive struct {
	SignatureRequestDecoded
	Status Status `json:"status"` // "unknown" | "active" | "inactive"
}

type DecodedPayload struct {
	Type      string `json:"type"`
	Platform  string `json:"platform"`
	Timestamp string `json:"timestamp"`
}
