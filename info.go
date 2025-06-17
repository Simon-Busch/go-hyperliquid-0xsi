package hyperliquid

import (
	"encoding/json"
	"fmt"
)

const (
	// spotAssetIndexOffset is the offset added to spot asset indices
	spotAssetIndexOffset = 10000
)

type Info struct {
	client         *Client
	coinToAsset    map[string]int
	nameToCoin     map[string]string
	assetToDecimal map[int]int
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

func NewInfo(baseURL string, skipWS bool, meta *Meta, spotMeta *SpotMeta) *Info {
	info := &Info{
		client:         NewClient(baseURL),
		coinToAsset:    make(map[string]int),
		nameToCoin:     make(map[string]string),
		assetToDecimal: make(map[int]int),
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
	for asset, assetInfo := range meta.Universe {
		info.coinToAsset[assetInfo.Name] = asset
		info.nameToCoin[assetInfo.Name] = assetInfo.Name
		info.assetToDecimal[asset] = assetInfo.SzDecimals
	}

	// Map spot assets starting at 10000
	for _, spotInfo := range spotMeta.Universe {
		asset := spotInfo.Index + spotAssetIndexOffset
		info.coinToAsset[spotInfo.Name] = asset
		info.nameToCoin[spotInfo.Name] = spotInfo.Name
		info.assetToDecimal[asset] = spotMeta.Tokens[spotInfo.Tokens[0]].SzDecimals
	}

	return info
}

func (i *Info) Meta() (*Meta, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "meta",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch meta: %w", err)
	}

	var meta Meta
	if err := json.Unmarshal(resp, &meta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta response: %w", err)
	}

	return &meta, nil
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

func (i *Info) NameToAsset(name string) int {
	coin := i.nameToCoin[name]
	return i.coinToAsset[coin]
}

func (i *Info) UserState(address string) (*UserState, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "clearinghouseState",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user state: %w", err)
	}

	var result UserState
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user state: %w", err)
	}
	return &result, nil
}

func (i *Info) SpotUserState(address string) (*UserState, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "spotClearinghouseState",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch spot user state: %w", err)
	}

	var result UserState
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal spot user state: %w", err)
	}
	return &result, nil
}

func (i *Info) OpenOrders(address string) ([]OpenOrder, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "openOrders",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch open orders: %w", err)
	}

	var result []OpenOrder
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal open orders: %w", err)
	}
	return result, nil
}

func (i *Info) FrontendOpenOrders(address string) ([]OpenOrder, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "frontendOpenOrders",
		"user": address,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch frontend open orders: %w", err)
	}

	var result []OpenOrder
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal frontend open orders: %w", err)
	}
	return result, nil
}

func (i *Info) AllMids() (map[string]string, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "allMids",
	})
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

func (i *Info) MetaAndAssetCtxs() (map[string]any, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "metaAndAssetCtxs",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch meta and asset contexts: %w", err)
	}

	var result map[string]any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal meta and asset contexts: %w", err)
	}
	return result, nil
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

func (i *Info) QueryOrderByOid(user string, oid int64) (*OpenOrder, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "orderStatus",
		"user": user,
		"oid":  oid,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order status: %w", err)
	}

	var result OpenOrder
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal order status: %w", err)
	}
	return &result, nil
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
func (i *Info) PerpDexs() ([]string, error) {
	resp, err := i.client.post("/info", map[string]any{
		"type": "perpDexs",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch perp dexs: %w", err)
	}

	var result []string
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal perp dexs: %w", err)
	}
	return result, nil
}
