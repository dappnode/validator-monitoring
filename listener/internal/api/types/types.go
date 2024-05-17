package types

// In sync with brain
type Network string // "mainnet" | "holesky" | "gnosis" | "lukso"

const (
	Mainnet Network = "mainnet"
	Holesky Network = "holesky"
	Gnosis  Network = "gnosis"
	Lukso   Network = "lukso"
)

// In sync with brain
// @see https://github.com/dappnode/StakingBrain/blob/0aaeefa8aec1b21ba2f2882cb444747419a3ff5d/packages/common/src/types/db/types.ts#L27
type Tag string //  "obol" | "diva" | "ssv" | "rocketpool" | "stakewise" | "stakehouse" | "solo" | "stader"

const (
	Obol       Tag = "obol"
	Diva       Tag = "diva"
	Ssv        Tag = "ssv"
	Rocketpool Tag = "rocketpool"
	Stakewise  Tag = "stakewise"
	Stakehouse Tag = "stakehouse"
	Solo       Tag = "solo"
	Stader     Tag = "stader"
)

type SignatureRequest struct {
	Payload   string  `json:"payload"`
	Pubkey    string  `json:"pubkey"`
	Signature string  `json:"signature"`
	Network   Network `json:"network"`
	Tag       string  `json:"tag"`
}

type DecodedPayload struct {
	Type      string `json:"type"`
	Platform  string `json:"platform"`
	Timestamp string `json:"timestamp"`
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

type SignatureRequestDecodedWithStatus struct {
	SignatureRequestDecoded
	Status Status `json:"status"` // "unknown" | "active" | "inactive"
}
