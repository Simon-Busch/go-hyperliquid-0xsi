package hyperliquid

type Side string

const (
	SideAsk Side = "A"
	SideBid Side = "B"
)

type Grouping string

const (
	GroupingNA           Grouping = "na"
	GroupingNormalTpsl   Grouping = "normalTpsl"
	GroupingPositionTpls Grouping = "positionTpsl"
)

// Constants for default values
const (
	DefaultSlippage = 0.05 // 5%
)

// Order Time-in-Force constants
const (
	TifAlo = "Alo" // Add Liquidity Only
	TifIoc = "Ioc" // Immediate or Cancel
	TifGtc = "Gtc" // Good Till Cancel
)

type AssetInfo struct {
	Name          string `json:"name"`
	SzDecimals    int    `json:"szDecimals"`
	MaxLeverage   int    `json:"maxLeverage"`
	MarginTableId int    `json:"marginTableId"`
	OnlyIsolated  bool   `json:"onlyIsolated"`
	IsDelisted    bool   `json:"isDelisted"`
}

type Meta struct {
	Universe        []AssetInfo   `json:"universe"`
	MarginTables    []MarginTable `json:"marginTables"`
	CollateralToken int           `json:"collateralToken"`
}

type SpotAssetInfo struct {
	Name        string `json:"name"`
	Tokens      []int  `json:"tokens"`
	Index       int    `json:"index"`
	IsCanonical bool   `json:"isCanonical"`
}

type EvmContract struct {
	Address             string `json:"address"`
	EvmExtraWeiDecimals int    `json:"evm_extra_wei_decimals"`
}

type SpotTokenInfo struct {
	Name        string       `json:"name"`
	SzDecimals  int          `json:"szDecimals"`
	WeiDecimals int          `json:"weiDecimals"`
	Index       int          `json:"index"`
	TokenID     string       `json:"tokenId"`
	IsCanonical bool         `json:"isCanonical"`
	EvmContract *EvmContract `json:"evmContract"`
	FullName    *string      `json:"fullName"`
}

type SpotMeta struct {
	Universe []SpotAssetInfo `json:"universe"`
	Tokens   []SpotTokenInfo `json:"tokens"`
}

// OutcomeSideSpec describes one side (YES or NO) of a binary HIP-4 outcome.
type OutcomeSideSpec struct {
	Name string `json:"name"` // "Yes" or "No"
}

// OutcomeInfo describes a single binary prediction market.
//
// The Description field is a pipe-delimited string of key:value pairs,
// e.g. "class:priceBinary|underlying:BTC|expiry:20260507-0600|targetPrice:81287|period:1d".
type OutcomeInfo struct {
	Outcome     int               `json:"outcome"`     // numeric outcome ID
	Name        string            `json:"name"`        // e.g. "Recurring"
	Description string            `json:"description"` // structured metadata (see above)
	SideSpecs   []OutcomeSideSpec `json:"sideSpecs"`   // [YES, NO] in that order
}

// OutcomeMeta is the response to POST /info {"type":"outcomeMeta"}.
// Questions is currently always empty on mainnet; reserved for future use.
type OutcomeMeta struct {
	Outcomes  []OutcomeInfo `json:"outcomes"`
	Questions []any         `json:"questions"`
}

type SpotAssetCtx struct {
	DayNtlVlm         string  `json:"dayNtlVlm"`
	MarkPx            string  `json:"markPx"`
	MidPx             *string `json:"midPx"`
	PrevDayPx         string  `json:"prevDayPx"`
	CirculatingSupply string  `json:"circulatingSupply"`
	Coin              string  `json:"coin"`
}

// AssetCtx represents perpetual asset context data including mark price, funding, open interest, etc.
type AssetCtx struct {
	DayNtlVlm    string   `json:"dayNtlVlm"`
	Funding      string   `json:"funding"`
	ImpactPxs    []string `json:"impactPxs"`
	MarkPx       string   `json:"markPx"`
	MidPx        string   `json:"midPx"`
	OpenInterest string   `json:"openInterest"`
	OraclePx     string   `json:"oraclePx"`
	Premium      string   `json:"premium"`
	PrevDayPx    string   `json:"prevDayPx"`
}

// MarginTier represents a single margin tier
type MarginTier struct {
	LowerBound  string `json:"lowerBound"`
	MaxLeverage int    `json:"maxLeverage"`
}

// MarginTable represents a margin table with description and tiers
type MarginTable struct {
	ID          int
	Description string       `json:"description"`
	MarginTiers []MarginTier `json:"marginTiers"`
}

// MetaAndAssetCtxsResponse represents the response from the metaAndAssetCtxs endpoint
// The API returns an array with two elements: [meta, assetCtxs]
type MetaAndAssetCtxsResponse struct {
	Meta      Meta      `json:"universe"`
	AssetCtxs []AssetCtx `json:"assetCtxs"`
}

// MetaAndAssetCtxsRawResponse represents the raw array response from the API
type MetaAndAssetCtxsRawResponse [2]interface{}

// WsMsg represents a WebSocket message with a channel and data payload.
type WsMsg struct {
	Channel string         `json:"channel"`
	Data    map[string]any `json:"data"`
}

type OrderType struct {
	Limit   *LimitOrderType   `json:"limit,omitempty"`
	Trigger *TriggerOrderType `json:"trigger,omitempty"`
}

type LimitOrderType struct {
	Tif string `json:"tif"` // TifAlo, TifIoc, TifGtc
}

type TriggerOrderType struct {
	TriggerPx float64 `json:"triggerPx"` // Keep as float64 for internal use, convert to string in wire format
	IsMarket  bool    `json:"isMarket"`
	Tpsl      string  `json:"tpsl"` // "tp" or "sl"
}

type BuilderInfo struct {
	Builder string `json:"b" msgpack:"b"`
	Fee     int    `json:"f" msgpack:"f"`
}

// Wire format types for order types (used in actions.go)
type OrderTypeWire struct {
	Limit   *LimitOrderTypeWire   `json:"limit,omitempty" msgpack:"limit,omitempty"`
	Trigger *TriggerOrderTypeWire `json:"trigger,omitempty" msgpack:"trigger,omitempty"`
}

type LimitOrderTypeWire struct {
	Tif string `json:"tif" msgpack:"tif"` // TifAlo, TifIoc, TifGtc
}

type TriggerOrderTypeWire struct {
	IsMarket  bool   `json:"isMarket" msgpack:"isMarket"`
	TriggerPx string `json:"triggerPx" msgpack:"triggerPx"`
	Tpsl      string `json:"tpsl" msgpack:"tpsl"` // "tp" or "sl"
}

type CancelRequest struct {
	Coin string `json:"coin"`
	Oid  int64  `json:"oid"`
}

type CancelByCloidRequest struct {
	Coin  string `json:"coin"`
	Cloid string `json:"cloid"`
}

type Cloid struct {
	Value string
}

func (c Cloid) ToRaw() string {
	return c.Value
}

type PerpDexSchemaInput struct {
	FullName        string  `json:"fullName"`
	CollateralToken int     `json:"collateralToken"`
	OracleUpdater   *string `json:"oracleUpdater"`
}

type L2Book struct {
	Coin   string    `json:"coin"`
	Levels [][]Level `json:"levels"`
	Time   int64     `json:"time"`
}

type Level struct {
	N  int     `json:"n"`
	Px float64 `json:"px,string"`
	Sz float64 `json:"sz,string"`
}

type AssetPosition struct {
	Position Position `json:"position"`
	Type     string   `json:"type"`
}

type Position struct {
	Coin           string   `json:"coin"`
	EntryPx        *string  `json:"entryPx"`
	Leverage       Leverage `json:"leverage"`
	LiquidationPx  *string  `json:"liquidationPx"`
	MarginUsed     string   `json:"marginUsed"`
	PositionValue  string   `json:"positionValue"`
	ReturnOnEquity string   `json:"returnOnEquity"`
	Szi            string   `json:"szi"`
	UnrealizedPnl  string   `json:"unrealizedPnl"`
}

type Leverage struct {
	Type   string  `json:"type"`
	Value  int     `json:"value"`
	RawUsd *string `json:"rawUsd,omitempty"`
}

type UserState struct {
	AssetPositions     []AssetPosition `json:"assetPositions"`
	CrossMarginSummary MarginSummary   `json:"crossMarginSummary"`
	MarginSummary      MarginSummary   `json:"marginSummary"`
	Withdrawable       string          `json:"withdrawable"`
}

type MarginSummary struct {
	AccountValue    string `json:"accountValue"`
	TotalMarginUsed string `json:"totalMarginUsed"`
	TotalNtlPos     string `json:"totalNtlPos"`
	TotalRawUsd     string `json:"totalRawUsd"`
}

type OpenOrder struct {
	Coin      string  `json:"coin"`
	LimitPx   float64 `json:"limitPx,string"`
	Oid       int64   `json:"oid"`
	Side      string  `json:"side"`
	Size      float64 `json:"sz,string"`
	Timestamp int64   `json:"timestamp"`
}

// FrontendOpenOrder represents the detailed order information returned by frontendOpenOrders
type FrontendOpenOrder struct {
	Coin             string  `json:"coin"`
	IsPositionTpsl   bool    `json:"isPositionTpsl"`
	IsTrigger        bool    `json:"isTrigger"`
	LimitPx          float64 `json:"limitPx,string"`
	Oid              int64   `json:"oid"`
	OrderType        string  `json:"orderType"`
	OrigSz           float64 `json:"origSz,string"`
	ReduceOnly       bool    `json:"reduceOnly"`
	Side             string  `json:"side"`
	Size             float64 `json:"sz,string"`
	Timestamp        int64   `json:"timestamp"`
	TriggerCondition string  `json:"triggerCondition"`
	TriggerPx        float64 `json:"triggerPx,string"`
}

type Fill struct {
	ClosedPnl     string `json:"closedPnl"`
	Coin          string `json:"coin"`
	Crossed       bool   `json:"crossed"`
	Dir           string `json:"dir"`
	Hash          string `json:"hash"`
	Oid           int64  `json:"oid"`
	Price         string `json:"px"`
	Side          string `json:"side"`
	StartPosition string `json:"startPosition"`
	Size          string `json:"sz"`
	Time          int64  `json:"time"`
	Fee           string `json:"fee"`
	FeeToken      string `json:"feeToken"`
}

type FundingHistory struct {
	Coin        string `json:"coin"`
	FundingRate string `json:"fundingRate"`
	Premium     string `json:"premium"`
	Time        int64  `json:"time"`
}

type UserFundingHistory struct {
	User      string `json:"user"`
	Type      string `json:"type"`
	StartTime int64  `json:"startTime"`
	EndTime   int64  `json:"endTime"`
}

type Candle struct {
	Timestamp int64  `json:"T"`
	Close     string `json:"c"`
	High      string `json:"h"`
	Interval  string `json:"i"`
	Low       string `json:"l"`
	Number    int    `json:"n"`
	Open      string `json:"o"`
	Symbol    string `json:"s"`
	Time      int64  `json:"t"`
	Volume    string `json:"v"`
}

type UserFees struct {
	ActiveReferralDiscount string       `json:"activeReferralDiscount"`
	DailyUserVolume        []UserVolume `json:"dailyUserVlm"`
	FeeSchedule            FeeSchedule  `json:"feeSchedule"`
	UserAddRate            string       `json:"userAddRate"`
	UserCrossRate          string       `json:"userCrossRate"`
}

type UserVolume struct {
	Date      string `json:"date"`
	Exchange  string `json:"exchange"`
	UserAdd   string `json:"userAdd"`
	UserCross string `json:"userCross"`
}

type FeeSchedule struct {
	Add              string `json:"add"`
	Cross            string `json:"cross"`
	ReferralDiscount string `json:"referralDiscount"`
	Tiers            Tiers  `json:"tiers"`
}

type Tiers struct {
	MM  []MMTier  `json:"mm"`
	VIP []VIPTier `json:"vip"`
}

type MMTier struct {
	Add                 string `json:"add"`
	MakerFractionCutoff string `json:"makerFractionCutoff"`
}

type VIPTier struct {
	Add       string `json:"add"`
	Cross     string `json:"cross"`
	NtlCutoff string `json:"ntlCutoff"`
}

type StakingSummary struct {
	Delegated              string `json:"delegated"`
	Undelegated            string `json:"undelegated"`
	TotalPendingWithdrawal string `json:"totalPendingWithdrawal"`
	NPendingWithdrawals    int    `json:"nPendingWithdrawals"`
}

type StakingDelegation struct {
	Validator            string `json:"validator"`
	Amount               string `json:"amount"`
	LockedUntilTimestamp int64  `json:"lockedUntilTimestamp"`
}

type StakingReward struct {
	Time        int64  `json:"time"`
	Source      string `json:"source"`
	TotalAmount string `json:"totalAmount"`
}

type ReferralState struct {
	ReferredBy       *ReferredBy    `json:"referredBy,omitempty"`
	CumVlm           string         `json:"cumVlm"`
	UnclaimedRewards string         `json:"unclaimedRewards"`
	ClaimedRewards   string         `json:"claimedRewards"`
	BuilderRewards   string         `json:"builderRewards"`
	ReferrerState    *ReferrerState `json:"referrerState,omitempty"`
	RewardHistory    []interface{}  `json:"rewardHistory"`
}

type ReferredBy struct {
	Referrer string `json:"referrer"`
	Code     string `json:"code"`
}

type ReferrerState struct {
	Stage string        `json:"stage"`
	Data  *ReferrerData `json:"data,omitempty"`
}

type ReferrerData struct {
	Code           string           `json:"code"`
	ReferralStates []ReferralMember `json:"referralStates"`
}

type ReferralMember struct {
	CumVlm                       string `json:"cumVlm"`
	CumRewardedFeesSinceReferred string `json:"cumRewardedFeesSinceReferred"`
	CumFeesRewardedToReferrer    string `json:"cumFeesRewardedToReferrer"`
	TimeJoined                   int64  `json:"timeJoined"`
	User                         string `json:"user"`
}

type SubAccount struct {
	Name        string   `json:"name"`
	User        string   `json:"user"`
	Permissions []string `json:"permissions"`
}

type MultiSigSigner struct {
	User      string `json:"user"`
	Threshold int    `json:"threshold"`
}

type Trade struct {
	Coin  string   `json:"coin"`
	Side  string   `json:"side"`
	Px    string   `json:"px"`
	Sz    string   `json:"sz"`
	Time  int64    `json:"time"`
	Hash  string   `json:"hash"`
	Tid   int64    `json:"tid"`
	Users []string `json:"users"`
}

type BulkOrderResponse struct {
	Status string        `json:"status"`
	Data   []OrderStatus `json:"data,omitempty"`
	Error  string        `json:"error,omitempty"`
}

type CancelResponse struct {
	Status string     `json:"status"`
	Data   *OpenOrder `json:"data,omitempty"`
	Error  string     `json:"error,omitempty"`
}

type BulkCancelResponse struct {
	Status string      `json:"status"`
	Data   []OpenOrder `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type ModifyResponse struct {
	Status string        `json:"status"`
	Data   []OrderStatus `json:"data,omitempty"`
	Error  string        `json:"error,omitempty"`
}

type TransferResponse struct {
	Status string `json:"status"`
	TxHash string `json:"txHash,omitempty"`
	Error  string `json:"error,omitempty"`
}

type ApprovalResponse struct {
	Status string `json:"status"`
	TxHash string `json:"txHash,omitempty"`
	Error  string `json:"error,omitempty"`
}

type CreateSubAccountResponse struct {
	Status string      `json:"status"`
	Data   *SubAccount `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type SetReferrerResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type ScheduleCancelResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type AgentApprovalResponse struct {
	Status string `json:"status"`
	TxHash string `json:"txHash,omitempty"`
	Error  string `json:"error,omitempty"`
}

type MultiSigConversionResponse struct {
	Status string `json:"status"`
	TxHash string `json:"txHash,omitempty"`
	Error  string `json:"error,omitempty"`
}

type SpotDeployResponse struct {
	Status string `json:"status"`
	TxHash string `json:"txHash,omitempty"`
	Error  string `json:"error,omitempty"`
}

type ValidatorResponse struct {
	Status string `json:"status"`
	TxHash string `json:"txHash,omitempty"`
	Error  string `json:"error,omitempty"`
}

type MultiSigResponse struct {
	Status string `json:"status"`
	TxHash string `json:"txHash,omitempty"`
	Error  string `json:"error,omitempty"`
}

type PerpDeployResponse struct {
	Status string `json:"status"`
	Data   struct {
		Statuses []TxStatus `json:"statuses"`
	} `json:"data"`
}

type TxStatus struct {
	Coin   string `json:"coin"`
	Status string `json:"status"`
}

// PerpDex represents a perpetual DEX
type PerpDex struct {
	Name                     string     `json:"name"`
	FullName                 string     `json:"fullName"`
	Deployer                 string     `json:"deployer"`
	OracleUpdater            *string    `json:"oracleUpdater"`
	FeeRecipient             *string    `json:"feeRecipient"`
	AssetToStreamingOiCap    [][]string `json:"assetToStreamingOiCap"`    // Array of [coin, cap] tuples
	AssetToFundingMultiplier [][]string `json:"assetToFundingMultiplier"` // Array of [coin, multiplier] tuples
}

// PerpDexLimits represents limits for a builder-deployed perp DEX
type PerpDexLimits struct {
	TotalOiCap     string     `json:"totalOiCap"`
	OiSzCapPerPerp string     `json:"oiSzCapPerPerp"`
	MaxTransferNtl string     `json:"maxTransferNtl"`
	CoinToOiCap    [][]string `json:"coinToOiCap"` // Array of [coin, cap] tuples
}

// PerpDexStatus represents status for a builder-deployed perp DEX
type PerpDexStatus struct {
	TotalNetDeposit string `json:"totalNetDeposit"`
}

// PerpDeployAuctionStatus represents the status of a perp deploy auction
type PerpDeployAuctionStatus struct {
	StartTimeSeconds int64   `json:"startTimeSeconds"`
	DurationSeconds  int64   `json:"durationSeconds"`
	StartGas         string  `json:"startGas"`
	CurrentGas       string  `json:"currentGas"`
	EndGas           *string `json:"endGas"`
}
