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
	Name       string `json:"name"`
	SzDecimals int    `json:"szDecimals"`
}

type Meta struct {
	Universe []AssetInfo `json:"universe"`
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

type SpotAssetCtx struct {
	DayNtlVlm         string  `json:"dayNtlVlm"`
	MarkPx            string  `json:"markPx"`
	MidPx             *string `json:"midPx"`
	PrevDayPx         string  `json:"prevDayPx"`
	CirculatingSupply string  `json:"circulatingSupply"`
	Coin              string  `json:"coin"`
}

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
	TriggerPx float64 `json:"triggerPx"`
	IsMarket  bool    `json:"isMarket"`
	Tpsl      string  `json:"tpsl"` // "tp" or "sl"
}

type BuilderInfo struct {
	Builder string `json:"b"`
	Fee     int    `json:"f"`
}

type OrderWire struct {
	Asset      int     `json:"a"`
	IsBuy      bool    `json:"b"`
	OrderType  string  `json:"t,omitempty"`
	LimitPx    float64 `json:"p"`
	Size       float64 `json:"s"`
	ReduceOnly bool    `json:"r"`
	TriggerPx  float64 `json:"tp,omitempty"`
	IsMarket   bool    `json:"im,omitempty"`
	Tpsl       string  `json:"tpsl,omitempty"`
	Tif        string  `json:"tif,omitempty"`
	Cloid      string  `json:"c,omitempty"`
}

type CancelRequest struct {
	Coin string `json:"coin"`
	Oid  int64  `json:"oid"`
}

type CancelByCloidRequest struct {
	Coin  string `json:"coin"`
	Cloid string `json:"cloid"`
}

type ScheduleCancelAction struct {
	Type string `json:"type"`
	Time *int64 `json:"time,omitempty"`
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
	ReferralCode string   `json:"referralCode"`
	Referrer     string   `json:"referrer"`
	Referred     []string `json:"referred"`
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
