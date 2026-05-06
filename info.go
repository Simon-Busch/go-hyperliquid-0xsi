package hyperliquid

import (
	"encoding/json"
	"fmt"
)

const (
	// spotAssetIndexOffset is the offset added to spot asset indices
	spotAssetIndexOffset = 10000
	// builderPerpAssetBase is the base offset for builder-deployed perp asset ids.
	// See Asset IDs docs: asset = 100000 + perpDexIndex*10000 + indexInMeta.
	builderPerpAssetBase = 100000
	// outcomeAssetBase is the base offset for HIP-4 outcome (binary prediction)
	// market asset ids. See Asset IDs docs: asset = 100_000_000 + (10*outcome + side).
	outcomeAssetBase = 100000000
)

type Info struct {
	client         *Client
	coinToAsset    map[string]int
	nameToCoin     map[string]string
	assetToDecimal map[int]int
	perpDexName    string // For HIP-3 builder-deployed perps (e.g., "flx")
}

// postTimeRangeRequest makes a POST request with time range parameters
func (i *Info) postTimeRangeRequest(
	requestType, user string,
	startTime int64,
	endTime *int64,
	extraParams map[string]any,
) ([]byte, error) {
	payload := map[string]any{
		"type":      requestType,
		"startTime": startTime,
	}
	if user != "" {
		payload["user"] = user
	}
	if endTime != nil {
		payload["endTime"] = *endTime
	}
	for k, v := range extraParams {
		payload[k] = v
	}

	resp, err := i.client.post("/info", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch %s: %w", requestType, err)
	}
	return resp, nil
}

// NewInfo creates a new Info instance.
// perpDexName is optional - set to empty string for the default perp dex,
// or provide a builder dex name (e.g., "flx") for HIP-3 builder-deployed perps.
func NewInfo(baseURL string, skipWS bool, meta *Meta, spotMeta *SpotMeta, perpDexs *MixedArray, perpDexName string) *Info {
	info := &Info{
		client:         NewClient(baseURL),
		coinToAsset:    make(map[string]int),
		nameToCoin:     make(map[string]string),
		assetToDecimal: make(map[int]int),
		perpDexName:    perpDexName,
	}

	if meta == nil {
		var err error
		meta, err = info.Meta()
		if err != nil {
			panic(err)
		}
	}

	if spotMeta == nil {
		var err error
		spotMeta, err = info.SpotMeta()
		if err != nil {
			panic(err)
		}
	}

	// Map perp assets
	if info.perpDexName != "" {
		// Builder-deployed perp: compute full asset id as documented.
		if perpDexs == nil {
			var err error
			perpDexsNew, err := info.PerpDexs()
			perpDexs = &perpDexsNew
			if err != nil {
				panic(err)
			}
		}
		perpDexIndex := -1
		for i, mv := range *perpDexs {
			if mv.Type() != "object" {
				continue
			}
			var pd PerpDex
			if err := mv.Parse(&pd); err == nil && pd.Name == info.perpDexName {
				perpDexIndex = i
				break
			}
		}
		if perpDexIndex < 0 {
			panic(fmt.Errorf("unknown perp dex %q (not present in /info perpDexs)", info.perpDexName))
		}
		base := builderPerpAssetBase + perpDexIndex*10000
		for idxInMeta, assetInfo := range meta.Universe {
			assetID := base + idxInMeta
			info.coinToAsset[assetInfo.Name] = assetID
			info.nameToCoin[assetInfo.Name] = assetInfo.Name
			info.assetToDecimal[assetID] = assetInfo.SzDecimals
		}
	} else {
		// Default perp dex: asset id is just index in meta universe.
		for asset, assetInfo := range meta.Universe {
			info.coinToAsset[assetInfo.Name] = asset
			info.nameToCoin[assetInfo.Name] = assetInfo.Name
			info.assetToDecimal[asset] = assetInfo.SzDecimals
		}
	}

	// Map spot assets starting at 10000
	for _, spotInfo := range spotMeta.Universe {
		asset := spotInfo.Index + spotAssetIndexOffset
		info.coinToAsset[spotInfo.Name] = asset
		info.nameToCoin[spotInfo.Name] = spotInfo.Name
		info.assetToDecimal[asset] = spotMeta.Tokens[spotInfo.Tokens[0]].SzDecimals
	}

	// Map HIP-4 outcome assets starting at 100_000_000.
	// Failure here is non-fatal: outcomeMeta may be empty or missing on
	// some environments, and the SDK should still work for perp/spot users.
	if outcomeMeta, err := info.OutcomeMeta(); err == nil {
		for _, oc := range outcomeMeta.Outcomes {
			for sideIdx, spec := range oc.SideSpecs {
				enc := 10*oc.Outcome + sideIdx
				asset := outcomeAssetBase + enc
				// Canonical name is "#<enc>" — empirically verified as the
				// form the exchange echoes in l2Book/allMids responses.
				// "+<enc>" form does NOT work for L2 queries despite being
				// mentioned in the docs, so we don't register it.
				canonical := fmt.Sprintf("#%d", enc)
				friendly := fmt.Sprintf("%s:%s", oc.Name, spec.Name)
				info.coinToAsset[canonical] = asset
				info.coinToAsset[friendly] = asset
				info.nameToCoin[canonical] = canonical
				info.nameToCoin[friendly] = canonical
				// szDecimals=0: HIP-4 contracts are integer-quantised. The L2
				// wire format shows sizes like "18.0" but the exchange rejects
				// fractional sizes ("Order has invalid size"). outcomeMeta
				// does not expose szDecimals; revisit if exchange behavior
				// changes.
				info.assetToDecimal[asset] = 0
			}
		}
	}

	return info
}

// PerpDexName returns the configured builder perp dex name (e.g. "flx"), or empty string for default dex.
func (i *Info) PerpDexName() string {
	return i.perpDexName
}

// Meta retrieves perpetuals metadata
// If dex is empty string, returns metadata for the first perp dex (default)
func (i *Info) Meta(dex ...string) (*Meta, error) {
	payload := map[string]any{
		"type": "meta",
	}
	if len(dex) > 0 && dex[0] != "" {
		payload["dex"] = dex[0]
	}

	resp, err := i.client.post("/info", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch meta: %w", err)
	}

	return parseMetaResponse(resp)
}

func parseMetaResponse(resp []byte) (*Meta, error) {
	var meta map[string]json.RawMessage
	if err := json.Unmarshal(resp, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta response: %w", err)
	}

	var universe []AssetInfo
	if err := json.Unmarshal(meta["universe"], &universe); err != nil {
		return nil, fmt.Errorf("failed to unmarshal universe: %w", err)
	}

	var collateralToken int
	if ct, ok := meta["collateralToken"]; ok {
		if err := json.Unmarshal(ct, &collateralToken); err != nil {
			return nil, fmt.Errorf("failed to unmarshal collateralToken: %w", err)
		}
	}

	var marginTables []MarginTable
	if mt, ok := meta["marginTables"]; ok {
		var marginTablesRaw [][]any
		if err := json.Unmarshal(mt, &marginTablesRaw); err != nil {
			return nil, fmt.Errorf("failed to unmarshal margin tables: %w", err)
		}

		marginTables = make([]MarginTable, len(marginTablesRaw))
		for idx, marginTable := range marginTablesRaw {
			if len(marginTable) < 2 {
				continue
			}
			id := int(marginTable[0].(float64))
			tableBytes, err := json.Marshal(marginTable[1])
			if err != nil {
				return nil, fmt.Errorf("failed to marshal margin table data: %w", err)
			}

			var marginTableData map[string]any
			if err := json.Unmarshal(tableBytes, &marginTableData); err != nil {
				return nil, fmt.Errorf("failed to unmarshal margin table data: %w", err)
			}

			marginTiersBytes, err := json.Marshal(marginTableData["marginTiers"])
			if err != nil {
				return nil, fmt.Errorf("failed to marshal margin tiers: %w", err)
			}

			var marginTiers []MarginTier
			if err := json.Unmarshal(marginTiersBytes, &marginTiers); err != nil {
				return nil, fmt.Errorf("failed to unmarshal margin tiers: %w", err)
			}

			desc := ""
			if d, ok := marginTableData["description"].(string); ok {
				desc = d
			}

			marginTables[idx] = MarginTable{
				ID:          id,
				Description: desc,
				MarginTiers: marginTiers,
			}
		}
	}

	return &Meta{
		Universe:        universe,
		MarginTables:    marginTables,
		CollateralToken: collateralToken,
	}, nil
}

func (i *Info) SpotMeta() (*SpotMeta, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "spotMeta",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spot meta: %w", err)
	}

	var spotMeta SpotMeta
	if err := json.Unmarshal(resp, &spotMeta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spot meta response: %w", err)
	}

	return &spotMeta, nil
}

// OutcomeMeta retrieves metadata for HIP-4 binary outcome (prediction) markets.
// Returns the list of active outcomes, each with its sides (YES/NO).
// May return an empty slice if no markets are live on the target environment.
func (i *Info) OutcomeMeta() (*OutcomeMeta, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "outcomeMeta",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch outcome meta: %w", err)
	}

	var outcomeMeta OutcomeMeta
	if err := json.Unmarshal(resp, &outcomeMeta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal outcome meta response: %w", err)
	}

	return &outcomeMeta, nil
}

func (i *Info) NameToAsset(name string) int {
	coin := i.nameToCoin[name]
	return i.coinToAsset[coin]
}

// UserState retrieves user's perpetuals account summary
// If dex is empty string, returns state for the first perp dex (default)
func (i *Info) UserState(address string, dex ...string) (*UserState, error) {
	payload := map[string]any{
		"type": "clearinghouseState",
		"user": address,
	}
	if len(dex) > 0 && dex[0] != "" {
		payload["dex"] = dex[0]
	}

	resp, err := i.client.post("/info", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user state: %w", err)
	}

	var result UserState
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user state: %w", err)
	}
	return &result, nil
}

// SpotBalance represents a single spot token balance entry returned by the
// spotClearinghouseState endpoint.
type SpotBalance struct {
	Coin     string `json:"coin"`
	Token    int    `json:"token"`
	Hold     string `json:"hold"`
	Total    string `json:"total"`
	EntryNtl string `json:"entryNtl"`
}

// SpotClearinghouseState is the response model for the spot balances endpoint.
type SpotClearinghouseState struct {
	Balances []SpotBalance `json:"balances"`
}

func (i *Info) SpotUserState(address string) (*SpotClearinghouseState, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "spotClearinghouseState",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spot user state: %w", err)
	}

	var result SpotClearinghouseState
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spot user state: %w", err)
	}
	return &result, nil
}

// OpenOrders retrieves user's open orders
// If dex is empty string, returns orders for the first perp dex (default)
// Note: Spot open orders are only included with the first perp dex
func (i *Info) OpenOrders(address string, dex ...string) ([]OpenOrder, error) {
	payload := map[string]any{
		"type": "openOrders",
		"user": address,
	}
	if len(dex) > 0 && dex[0] != "" {
		payload["dex"] = dex[0]
	}

	resp, err := i.client.post("/info", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open orders: %w", err)
	}

	var result []OpenOrder
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal open orders: %w", err)
	}
	return result, nil
}

// FrontendOpenOrders retrieves user's open orders with frontend info
// If dex is empty string, returns orders for the first perp dex (default)
// Note: Spot open orders are only included with the first perp dex
func (i *Info) FrontendOpenOrders(address string, dex ...string) ([]FrontendOpenOrder, error) {
	payload := map[string]any{
		"type": "frontendOpenOrders",
		"user": address,
	}
	if len(dex) > 0 && dex[0] != "" {
		payload["dex"] = dex[0]
	}

	resp, err := i.client.post("/info", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch frontend open orders: %w", err)
	}

	var result []FrontendOpenOrder
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal frontend open orders: %w", err)
	}
	return result, nil
}

// AllMids retrieves mids for all coins
// If dex is empty string, returns mids for the first perp dex (default)
// Note: Spot mids are only included with the first perp dex
func (i *Info) AllMids(dex ...string) (map[string]string, error) {
	payload := map[string]any{
		"type": "allMids",
	}
	if len(dex) > 0 && dex[0] != "" {
		payload["dex"] = dex[0]
	}

	resp, err := i.client.post("/info", payload)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch all mids: %w", err)
	}

	var result map[string]string
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal all mids: %w", err)
	}
	return result, nil
}

func (i *Info) UserFills(address string) ([]Fill, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "userFills",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user fills: %w", err)
	}

	var result []Fill
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user fills: %w", err)
	}
	return result, nil
}

func (i *Info) UserFillsByTime(address string, startTime int64, endTime *int64) ([]Fill, error) {
	resp, err := i.postTimeRangeRequest("userFillsByTime", address, startTime, endTime, nil)
	if err != nil {
		return nil, err
	}

	var result []Fill
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user fills by time: %w", err)
	}
	return result, nil
}

func (i *Info) MetaAndAssetCtxs() (*MetaAndAssetCtxsResponse, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "metaAndAssetCtxs",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch meta and asset contexts: %w", err)
	}

	// The API returns an array with two elements: [meta, assetCtxs]
	var rawResponse [2]interface{}
	if err := json.Unmarshal(resp, &rawResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta and asset contexts array: %w", err)
	}

	// Parse the meta object (first element)
	metaBytes, err := json.Marshal(rawResponse[0])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal meta object: %w", err)
	}

	var meta Meta
	if err := json.Unmarshal(metaBytes, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta: %w", err)
	}

	// Parse the asset contexts array (second element)
	assetCtxsBytes, err := json.Marshal(rawResponse[1])
	if err != nil {
		return nil, fmt.Errorf("failed to marshal asset contexts array: %w", err)
	}

	var assetCtxs []AssetCtx
	if err := json.Unmarshal(assetCtxsBytes, &assetCtxs); err != nil {
		return nil, fmt.Errorf("failed to unmarshal asset contexts: %w", err)
	}

	return &MetaAndAssetCtxsResponse{
		Meta:      meta,
		AssetCtxs: assetCtxs,
	}, nil
}

func (i *Info) SpotMetaAndAssetCtxs() (map[string]any, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "spotMetaAndAssetCtxs",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spot meta and asset contexts: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spot meta and asset contexts: %w", err)
	}
	return result, nil
}

func (i *Info) FundingHistory(
	name string,
	startTime int64,
	endTime *int64,
) ([]FundingHistory, error) {
	coin := i.nameToCoin[name]
	resp, err := i.postTimeRangeRequest(
		"fundingHistory",
		"",
		startTime,
		endTime,
		map[string]any{"coin": coin},
	)
	if err != nil {
		return nil, err
	}

	var result []FundingHistory
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal funding history: %w", err)
	}
	return result, nil
}

func (i *Info) UserFundingHistory(
	user string,
	startTime int64,
	endTime *int64,
) ([]UserFundingHistory, error) {
	resp, err := i.postTimeRangeRequest("userFunding", user, startTime, endTime, nil)
	if err != nil {
		return nil, err
	}

	var result []UserFundingHistory
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user funding history: %w", err)
	}
	return result, nil
}

func (i *Info) L2Snapshot(name string) (*L2Book, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "l2Book",
		"coin": i.nameToCoin[name],
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch L2 snapshot: %w", err)
	}

	var result L2Book
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal L2 snapshot: %w", err)
	}
	return &result, nil
}

func (i *Info) CandlesSnapshot(name, interval string, startTime, endTime int64) ([]Candle, error) {
	req := map[string]any{
		"coin":      i.nameToCoin[name],
		"interval":  interval,
		"startTime": startTime,
		"endTime":   endTime,
	}

	resp, err := i.client.post("/info", map[string]any{
		"type": "candleSnapshot",
		"req":  req,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch candles snapshot: %w", err)
	}

	var result []Candle
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal candles snapshot: %w", err)
	}
	return result, nil
}

func (i *Info) UserFees(address string) (*UserFees, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "userFees",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user fees: %w", err)
	}

	var result UserFees
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user fees: %w", err)
	}
	return &result, nil
}

func (i *Info) UserStakingSummary(address string) (*StakingSummary, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "delegatorSummary",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch staking summary: %w", err)
	}

	var result StakingSummary
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal staking summary: %w", err)
	}
	return &result, nil
}

func (i *Info) UserStakingDelegations(address string) ([]StakingDelegation, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "delegations",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch staking delegations: %w", err)
	}

	var result []StakingDelegation
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal staking delegations: %w", err)
	}
	return result, nil
}

func (i *Info) UserStakingRewards(address string) ([]StakingReward, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "delegatorRewards",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch staking rewards: %w", err)
	}

	var result []StakingReward
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal staking rewards: %w", err)
	}
	return result, nil
}

// OrderStatusResponse represents the actual response from the orderStatus endpoint
type OrderStatusResponse struct {
	Status string `json:"status"`
	Order  struct {
		Order struct {
			Coin             string  `json:"coin"`
			Side             string  `json:"side"`
			LimitPx          string  `json:"limitPx"`
			Sz               string  `json:"sz"`
			Oid              int64   `json:"oid"`
			Timestamp        int64   `json:"timestamp"`
			TriggerCondition string  `json:"triggerCondition"`
			IsTrigger        bool    `json:"isTrigger"`
			TriggerPx        string  `json:"triggerPx"`
			Children         []any   `json:"children"`
			IsPositionTpsl   bool    `json:"isPositionTpsl"`
			ReduceOnly       bool    `json:"reduceOnly"`
			OrderType        string  `json:"orderType"`
			OrigSz           string  `json:"origSz"`
			Tif              *string `json:"tif"`
			Cloid            *string `json:"cloid"`
		} `json:"order"`
		Status          string `json:"status"`
		StatusTimestamp int64  `json:"statusTimestamp"`
	} `json:"order"`
}

func (i *Info) QueryOrderByOid(user string, oid int64) (*OrderStatusResponse, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "orderStatus",
		"user": user,
		"oid":  oid,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order status: %w", err)
	}

	var result OrderStatusResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order status: %w", err)
	}

	return &result, nil
}

// QueryFillByOid finds a specific fill by OID from user fills
// Since there's no direct fill query endpoint, we filter userFills by OID
func (i *Info) QueryFillByOid(user string, oid int64) (*Fill, error) {
	fills, err := i.UserFills(user)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user fills: %w", err)
	}

	for _, fill := range fills {
		if fill.Oid == oid {
			return &fill, nil
		}
	}

	return nil, fmt.Errorf("fill with OID %d not found for user %s", oid, user)
}

func (i *Info) QueryOrderByCloid(user, cloid string) (*OpenOrder, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "orderStatus",
		"user": user,
		"oid":  cloid,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order status by cloid: %w", err)
	}

	var result OpenOrder
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order status: %w", err)
	}
	return &result, nil
}

func (i *Info) QueryReferralState(user string) (*ReferralState, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "referral",
		"user": user,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch referral state: %w", err)
	}

	var result ReferralState
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal referral state: %w", err)
	}
	return &result, nil
}

func (i *Info) QuerySubAccounts(user string) ([]SubAccount, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "subAccounts",
		"user": user,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sub accounts: %w", err)
	}

	var result []SubAccount
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sub accounts: %w", err)
	}
	return result, nil
}

func (i *Info) QueryUserToMultiSigSigners(multiSigUser string) ([]MultiSigSigner, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "userToMultiSigSigners",
		"user": multiSigUser,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch multi-sig signers: %w", err)
	}

	var result []MultiSigSigner
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal multi-sig signers: %w", err)
	}
	return result, nil
}

// PerpDexs returns the list of available perpetual dexes
// Returns an array where each element can be nil (for the default dex) or a PerpDex object
// The first element is always null (representing the default dex)
func (i *Info) PerpDexs() (MixedArray, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "perpDexs",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch perp dexs: %w", err)
	}

	var result MixedArray
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal perp dexs: %w", err)
	}
	return result, nil
}
