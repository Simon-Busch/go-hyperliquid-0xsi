package hyperliquid

// Action structs with deterministic field ordering for consistent MessagePack serialization
// The order of fields in these structs is critical for signature generation

// CancelOrderWire represents cancel order item wire format
type CancelOrderWire struct {
	Asset   int    `json:"a" msgpack:"a"`
	OrderID string `json:"o" msgpack:"o"`
}

// CancelAction represents the cancel action
type CancelAction struct {
	Type    string            `json:"type"    msgpack:"type"`
	Cancels []CancelOrderWire `json:"cancels" msgpack:"cancels"`
}

// CancelByCloidWire represents cancel by cloid item wire format
type CancelByCloidWire struct {
	Asset   int    `json:"a"     msgpack:"a"`
	OrderID string `json:"cloid" msgpack:"cloid"`
}

// CancelByCloidAction represents the cancel by cloid action
type CancelByCloidAction struct {
	Type    string              `json:"type"    msgpack:"type"`
	Cancels []CancelByCloidWire `json:"cancels" msgpack:"cancels"`
}

// UsdClassTransferAction represents USD class transfer
type UsdClassTransferAction struct {
	Type   string `json:"type"   msgpack:"type"`
	Amount string `json:"amount" msgpack:"amount"`
	ToPerp bool   `json:"toPerp" msgpack:"toPerp"`
	Nonce  int64  `json:"nonce"  msgpack:"nonce"`
}

// SpotTransferAction represents spot transfer
type SpotTransferAction struct {
	Type        string `json:"type"        msgpack:"type"`
	Destination string `json:"destination" msgpack:"destination"`
	Amount      string `json:"amount"      msgpack:"amount"`
	Token       string `json:"token"       msgpack:"token"`
	Time        int64  `json:"time"        msgpack:"time"`
}

// UsdTransferAction represents USD transfer
type UsdTransferAction struct {
	Type        string `json:"type"        msgpack:"type"`
	Destination string `json:"destination" msgpack:"destination"`
	Amount      string `json:"amount"      msgpack:"amount"`
	Time        int64  `json:"time"        msgpack:"time"`
}

// SubAccountTransferAction represents sub-account transfer
type SubAccountTransferAction struct {
	Type           string `json:"type"           msgpack:"type"`
	SubAccountUser string `json:"subAccountUser" msgpack:"subAccountUser"`
	IsDeposit      bool   `json:"isDeposit"      msgpack:"isDeposit"`
	Usd            int    `json:"usd"            msgpack:"usd"`
}

// VaultUsdTransferAction represents vault USD transfer
type VaultUsdTransferAction struct {
	Type         string `json:"type"         msgpack:"type"`
	VaultAddress string `json:"vaultAddress" msgpack:"vaultAddress"`
	IsDeposit    bool   `json:"isDeposit"    msgpack:"isDeposit"`
	Usd          int    `json:"usd"          msgpack:"usd"`
}

// UpdateLeverageAction represents leverage update
type UpdateLeverageAction struct {
	Type     string         `json:"type"     msgpack:"type"`
	Asset    int            `json:"asset"    msgpack:"asset"`
	Leverage map[string]any `json:"leverage" msgpack:"leverage"`
}

// UpdateIsolatedMarginAction represents isolated margin update
type UpdateIsolatedMarginAction struct {
	Type  string  `json:"type"  msgpack:"type"`
	Asset int     `json:"asset" msgpack:"asset"`
	IsBuy bool    `json:"isBuy" msgpack:"isBuy"`
	Ntli  float64 `json:"ntli"  msgpack:"ntli"`
}

// OrderWire represents the wire format for orders with deterministic field ordering
type OrderWire struct {
	Asset      int            `json:"a"           msgpack:"a"`
	IsBuy      bool           `json:"b"           msgpack:"b"`
	LimitPx    string         `json:"p"           msgpack:"p"`
	Size       string         `json:"s"           msgpack:"s"`
	ReduceOnly bool           `json:"r"           msgpack:"r"`
	OrderType  map[string]any `json:"t"           msgpack:"t"`
	Cloid      *string        `json:"c,omitempty" msgpack:"c,omitempty"`
}

// OrderAction represents the order action with deterministic field ordering
type OrderAction struct {
	Type     string       `json:"type"              msgpack:"type"`
	Orders   []OrderWire  `json:"orders"            msgpack:"orders"`
	Grouping string       `json:"grouping"          msgpack:"grouping"`
	Builder  *BuilderInfo `json:"builder,omitempty" msgpack:"builder,omitempty"`
}

// ModifyAction represents a single order modification
type ModifyAction struct {
	Type  string    `json:"type"  msgpack:"type"`
	Oid   any       `json:"oid"   msgpack:"oid"`
	Order OrderWire `json:"order" msgpack:"order"`
}

// BatchModifyAction represents multiple order modifications
type BatchModifyAction struct {
	Type     string         `json:"type"     msgpack:"type"`
	Modifies []ModifyAction `json:"modifies" msgpack:"modifies"`
}

// PerpDexClassTransferAction represents perp dex class transfer
type PerpDexClassTransferAction struct {
	Type   string  `json:"type"   msgpack:"type"`
	Dex    string  `json:"dex"    msgpack:"dex"`
	Token  string  `json:"token"  msgpack:"token"`
	Amount float64 `json:"amount" msgpack:"amount"`
	ToPerp bool    `json:"toPerp" msgpack:"toPerp"`
}

// SubAccountSpotTransferAction represents sub-account spot transfer
type SubAccountSpotTransferAction struct {
	Type           string  `json:"type"           msgpack:"type"`
	SubAccountUser string  `json:"subAccountUser" msgpack:"subAccountUser"`
	IsDeposit      bool    `json:"isDeposit"      msgpack:"isDeposit"`
	Token          string  `json:"token"          msgpack:"token"`
	Amount         float64 `json:"amount"         msgpack:"amount"`
}

// ScheduleCancelAction represents schedule cancel action
type ScheduleCancelAction struct {
	Type string `json:"type"           msgpack:"type"`
	Time *int64 `json:"time,omitempty" msgpack:"time,omitempty"`
}

// SetReferrerAction represents set referrer action
type SetReferrerAction struct {
	Type string `json:"type" msgpack:"type"`
	Code string `json:"code" msgpack:"code"`
}

// CreateSubAccountAction represents create sub-account action
type CreateSubAccountAction struct {
	Type string `json:"type" msgpack:"type"`
	Name string `json:"name" msgpack:"name"`
}

// UseBigBlocksAction represents use big blocks action
type UseBigBlocksAction struct {
	Type           string `json:"type"           msgpack:"type"`
	UsingBigBlocks bool   `json:"usingBigBlocks" msgpack:"usingBigBlocks"`
}

// TokenDelegateAction represents token delegate action
type TokenDelegateAction struct {
	Type         string `json:"type"         msgpack:"type"`
	Validator    string `json:"validator"    msgpack:"validator"`
	Wei          int    `json:"wei"          msgpack:"wei"`
	IsUndelegate bool   `json:"isUndelegate" msgpack:"isUndelegate"`
	Nonce        int64  `json:"nonce"        msgpack:"nonce"`
}

// WithdrawFromBridgeAction represents withdraw from bridge action
type WithdrawFromBridgeAction struct {
	Type        string `json:"type"        msgpack:"type"`
	Destination string `json:"destination" msgpack:"destination"`
	Amount      string `json:"amount"      msgpack:"amount"`
	Time        int64  `json:"time"        msgpack:"time"`
}

// ApproveAgentAction represents approve agent action
type ApproveAgentAction struct {
	Type         string  `json:"type"                msgpack:"type"`
	AgentAddress string  `json:"agentAddress"        msgpack:"agentAddress"`
	AgentName    *string `json:"agentName,omitempty" msgpack:"agentName,omitempty"`
	Nonce        int64   `json:"nonce"               msgpack:"nonce"`
}

// ApproveBuilderFeeAction represents approve builder fee action
type ApproveBuilderFeeAction struct {
	Type       string `json:"type"       msgpack:"type"`
	Builder    string `json:"builder"    msgpack:"builder"`
	MaxFeeRate string `json:"maxFeeRate" msgpack:"maxFeeRate"`
	Nonce      int64  `json:"nonce"      msgpack:"nonce"`
}

// ConvertToMultiSigUserAction represents convert to multi-sig user action
type ConvertToMultiSigUserAction struct {
	Type    string `json:"type"    msgpack:"type"`
	Signers string `json:"signers" msgpack:"signers"`
	Nonce   int64  `json:"nonce"   msgpack:"nonce"`
}

// MultiSigAction represents multi-signature action
type MultiSigAction struct {
	Type       string         `json:"type"       msgpack:"type"`
	Action     map[string]any `json:"action"     msgpack:"action"`
	Signers    []string       `json:"signers"    msgpack:"signers"`
	Signatures []string       `json:"signatures" msgpack:"signatures"`
}
