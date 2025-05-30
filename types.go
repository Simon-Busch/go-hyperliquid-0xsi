package hyperliquid

//go:generate easyjson -all types.go

type Side string

const (
	SideAsk Side = "A"
	SideBid Side = "B"
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

type OrderRequest struct {
	Coin       string    `json:"coin"`
	IsBuy      bool      `json:"is_buy"`
	Size       float64   `json:"sz"`
	LimitPx    float64   `json:"limit_px"`
	OrderType  OrderType `json:"order_type"`
	ReduceOnly bool      `json:"reduce_only"`
	Cloid      *string   `json:"cloid,omitempty"`
}

type OrderType struct {
	Limit   *LimitOrderType   `json:"limit,omitempty"`
	Trigger *TriggerOrderType `json:"trigger,omitempty"`
}

type LimitOrderType struct {
	Tif string `json:"tif"` // "Alo", "Ioc", "Gtc"
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
