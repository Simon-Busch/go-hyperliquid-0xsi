package hyperliquid

import (
	"crypto/ecdsa"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

type Exchange struct {
	client      *Client
	privateKey  *ecdsa.PrivateKey
	vault       string
	accountAddr string
	info        *Info
}

// executeAction executes an action and unmarshals the response into the given result
func (e *Exchange) executeAction(action map[string]any, result any) error {
	timestamp := time.Now().UnixMilli()

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp, result); err != nil {
		return err
	}

	return nil
}

func NewExchange(
	privateKey *ecdsa.PrivateKey,
	baseURL string,
	meta *Meta,
	vaultAddr, accountAddr string,
	spotMeta *SpotMeta,
) *Exchange {
	return &Exchange{
		client:      NewClient(baseURL),
		privateKey:  privateKey,
		vault:       vaultAddr,
		accountAddr: accountAddr,
		info:        NewInfo(baseURL, true, meta, spotMeta),
	}
}

func (e *Exchange) Order(req OrderRequest, builder *BuilderInfo) (*OpenOrder, error) {
	orders, err := e.BulkOrders([]OrderRequest{req}, builder)
	if err != nil {
		return nil, err
	}
	if len(orders) == 0 {
		return nil, nil
	}
	return &orders[0], nil
}

func (e *Exchange) BulkOrders(orders []OrderRequest, builder *BuilderInfo) ([]OpenOrder, error) {
	timestamp := time.Now().UnixMilli()

	orderWires := make([]OrderWire, len(orders))
	for i, order := range orders {
		asset := e.info.NameToAsset(order.Coin)
		wire := OrderRequestToWire(order, asset)
		orderWires[i] = wire
	}

	action := map[string]any{
		"type":     "order",
		"orders":   orderWires,
		"grouping": "na",
	}
	if builder != nil {
		action["builder"] = builder
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result []OpenOrder
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (e *Exchange) Cancel(coin string, oid int64) (*OpenOrder, error) {
	action := map[string]any{
		"type": "cancel",
		"coin": coin,
		"oid":  oid,
	}

	var result OpenOrder
	if err := e.executeAction(action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (e *Exchange) CancelByCloid(coin, cloid string) (*OpenOrder, error) {
	action := map[string]any{
		"type":  "cancelByCloid",
		"coin":  coin,
		"cloid": cloid,
	}

	var result OpenOrder
	if err := e.executeAction(action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (e *Exchange) UpdateLeverage(leverage int, name string, isCross bool) (*UserState, error) {
	leverageType := "isolated"
	if isCross {
		leverageType = "cross"
	}

	action := map[string]any{
		"type":  "updateLeverage",
		"asset": e.info.NameToAsset(name),
		"leverage": map[string]any{
			"type":  leverageType,
			"value": leverage,
		},
	}

	var result UserState
	if err := e.executeAction(action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (e *Exchange) UpdateIsolatedMargin(amount float64, name string) (*UserState, error) {
	action := map[string]any{
		"type":  "updateIsolatedMargin",
		"asset": e.info.NameToAsset(name),
		"isBuy": amount > 0,
		"ntli":  abs(amount),
	}

	var result UserState
	if err := e.executeAction(action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetExpiresAfter sets the expiration time for actions
func (e *Exchange) SetExpiresAfter(expiresAfter *int64) {
	// Note: In Go SDK, we'll store this in the exchange struct
	// For now, we'll implement the expiration logic directly in each method
}

// SlippagePrice calculates the slippage price for market orders
func (e *Exchange) SlippagePrice(
	name string,
	isBuy bool,
	slippage float64,
	px *float64,
) (float64, error) {
	coin := e.info.nameToCoin[name]
	var price float64

	if px != nil {
		price = *px
	} else {
		// Get midprice
		mids, err := e.info.AllMids()
		if err != nil {
			return 0, err
		}
		if midPriceStr, exists := mids[coin]; exists {
			price = parseFloat(midPriceStr)
		} else {
			return 0, fmt.Errorf("could not get mid price for coin: %s", coin)
		}
	}

	asset := e.info.coinToAsset[coin]
	isSpot := asset >= 10000

	// Calculate slippage
	if isBuy {
		price *= (1 + slippage)
	} else {
		price *= (1 - slippage)
	}

	// Round to appropriate decimals
	decimals := 6
	if isSpot {
		decimals = 8
	}
	szDecimals := e.info.assetToDecimal[asset]

	return roundToDecimals(price, decimals-szDecimals), nil
}

// ModifyOrder modifies an existing order
func (e *Exchange) ModifyOrder(
	oid any,
	name string,
	isBuy bool,
	sz, limitPx float64,
	orderType OrderType,
	reduceOnly bool,
	cloid *string,
) (any, error) {
	modify := ModifyRequest{
		Oid: oid,
		Order: OrderRequest{
			Coin:       name,
			IsBuy:      isBuy,
			Size:       sz,
			LimitPx:    limitPx,
			OrderType:  orderType,
			ReduceOnly: reduceOnly,
			Cloid:      cloid,
		},
	}
	return e.BulkModifyOrders([]ModifyRequest{modify})
}

// BulkModifyOrders modifies multiple orders
func (e *Exchange) BulkModifyOrders(modifyRequests []ModifyRequest) (any, error) {
	timestamp := time.Now().UnixMilli()

	modifyWires := make([]map[string]any, len(modifyRequests))
	for i, modify := range modifyRequests {
		asset := e.info.NameToAsset(modify.Order.Coin)
		orderWire := OrderRequestToWire(modify.Order, asset)

		var oidValue any
		if cloid, ok := modify.Oid.(Cloid); ok {
			oidValue = cloid.ToRaw()
		} else {
			oidValue = modify.Oid
		}

		modifyWires[i] = map[string]any{
			"oid":   oidValue,
			"order": orderWire,
		}
	}

	action := map[string]any{
		"type":     "batchModify",
		"modifies": modifyWires,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// BulkModifyOrdersNew is an alias for BulkModifyOrders to match Python SDK naming
func (e *Exchange) BulkModifyOrdersNew(modifyRequests []ModifyRequest) (any, error) {
	return e.BulkModifyOrders(modifyRequests)
}

// MarketOpen opens a market position
func (e *Exchange) MarketOpen(
	name string,
	isBuy bool,
	sz float64,
	px *float64,
	slippage float64,
	cloid *string,
	builder *BuilderInfo,
) (any, error) {
	slippagePrice, err := e.SlippagePrice(name, isBuy, slippage, px)
	if err != nil {
		return nil, err
	}

	orderType := OrderType{
		Limit: &LimitOrderType{Tif: "Ioc"},
	}

	return e.Order(OrderRequest{
		Coin:       name,
		IsBuy:      isBuy,
		Size:       sz,
		LimitPx:    slippagePrice,
		OrderType:  orderType,
		ReduceOnly: false,
		Cloid:      cloid,
	}, builder)
}

// MarketClose closes a position
func (e *Exchange) MarketClose(
	coin string,
	sz *float64,
	px *float64,
	slippage float64,
	cloid *string,
	builder *BuilderInfo,
) (any, error) {
	address := e.accountAddr
	if address == "" {
		address = e.vault
	}

	userState, err := e.info.UserState(address)
	if err != nil {
		return nil, err
	}

	for _, assetPos := range userState.AssetPositions {
		pos := assetPos.Position
		if coin != pos.Coin {
			continue
		}

		szi := parseFloat(pos.Szi)
		var size float64
		if sz != nil {
			size = *sz
		} else {
			size = abs(szi)
		}

		isBuy := szi < 0

		slippagePrice, err := e.SlippagePrice(coin, isBuy, slippage, px)
		if err != nil {
			return nil, err
		}

		orderType := OrderType{
			Limit: &LimitOrderType{Tif: "Ioc"},
		}

		return e.Order(OrderRequest{
			Coin:       coin,
			IsBuy:      isBuy,
			Size:       size,
			LimitPx:    slippagePrice,
			OrderType:  orderType,
			ReduceOnly: true,
			Cloid:      cloid,
		}, builder)
	}

	return nil, fmt.Errorf("position not found for coin: %s", coin)
}

// BulkCancel cancels multiple orders
func (e *Exchange) BulkCancel(cancelRequests []CancelRequest) (any, error) {
	timestamp := time.Now().UnixMilli()

	cancels := make([]map[string]any, len(cancelRequests))
	for i, cancel := range cancelRequests {
		cancels[i] = map[string]any{
			"a": e.info.NameToAsset(cancel.Coin),
			"o": cancel.Oid,
		}
	}

	action := map[string]any{
		"type":    "cancel",
		"cancels": cancels,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// BulkCancelByCloid cancels multiple orders by cloid
func (e *Exchange) BulkCancelByCloid(cancelRequests []CancelByCloidRequest) (any, error) {
	timestamp := time.Now().UnixMilli()

	cancels := make([]map[string]any, len(cancelRequests))
	for i, cancel := range cancelRequests {
		cancels[i] = map[string]any{
			"asset": e.info.NameToAsset(cancel.Coin),
			"cloid": cancel.Cloid,
		}
	}

	action := map[string]any{
		"type":    "cancelByCloid",
		"cancels": cancels,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ScheduleCancel schedules cancellation of all open orders
func (e *Exchange) ScheduleCancel(scheduleTime *int64) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "scheduleCancel",
	}
	if scheduleTime != nil {
		action["time"] = *scheduleTime
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SetReferrer sets a referral code
func (e *Exchange) SetReferrer(code string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "setReferrer",
		"code": code,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address for referrer
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CreateSubAccount creates a new sub-account
func (e *Exchange) CreateSubAccount(name string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "createSubAccount",
		"name": name,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address for sub-account creation
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UsdClassTransfer transfers between USD classes
func (e *Exchange) UsdClassTransfer(amount float64, toPerp bool) (any, error) {
	timestamp := time.Now().UnixMilli()

	strAmount := formatFloat(amount)
	if e.vault != "" {
		strAmount += " subaccount:" + e.vault
	}

	action := map[string]any{
		"type":   "usdClassTransfer",
		"amount": strAmount,
		"toPerp": toPerp,
		"nonce":  timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SubAccountTransfer transfers funds to/from sub-account
func (e *Exchange) SubAccountTransfer(subAccountUser string, isDeposit bool, usd int) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":           "subAccountTransfer",
		"subAccountUser": subAccountUser,
		"isDeposit":      isDeposit,
		"usd":            usd,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// VaultUsdTransfer transfers to/from vault
func (e *Exchange) VaultUsdTransfer(vaultAddress string, isDeposit bool, usd int) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":         "vaultTransfer",
		"vaultAddress": vaultAddress,
		"isDeposit":    isDeposit,
		"usd":          usd,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UsdTransfer transfers USD to another address
func (e *Exchange) UsdTransfer(amount float64, destination string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"destination": destination,
		"amount":      formatFloat(amount),
		"time":        timestamp,
		"type":        "usdSend",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotTransfer transfers spot tokens to another address
func (e *Exchange) SpotTransfer(amount float64, destination, token string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"destination": destination,
		"amount":      formatFloat(amount),
		"token":       token,
		"time":        timestamp,
		"type":        "spotSend",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// UseBigBlocks enables or disables big blocks
func (e *Exchange) UseBigBlocks(enable bool) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":           "evmUserModify",
		"usingBigBlocks": enable,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PerpDexClassTransfer transfers tokens between perp dex classes
func (e *Exchange) PerpDexClassTransfer(
	dex, token string,
	amount float64,
	toPerp bool,
) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":   "perpDexClassTransfer",
		"dex":    dex,
		"token":  token,
		"amount": amount,
		"toPerp": toPerp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SubAccountSpotTransfer transfers spot tokens to/from sub-account
func (e *Exchange) SubAccountSpotTransfer(
	subAccountUser string,
	isDeposit bool,
	token string,
	amount float64,
) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":           "subAccountSpotTransfer",
		"subAccountUser": subAccountUser,
		"isDeposit":      isDeposit,
		"token":          token,
		"amount":         amount,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// TokenDelegate delegates tokens for staking
func (e *Exchange) TokenDelegate(validator string, wei int, isUndelegate bool) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":         "tokenDelegate",
		"validator":    validator,
		"wei":          wei,
		"isUndelegate": isUndelegate,
		"nonce":        timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// WithdrawFromBridge withdraws tokens from bridge
func (e *Exchange) WithdrawFromBridge(amount float64, destination string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":        "withdraw3",
		"destination": destination,
		"amount":      fmt.Sprintf("%.6f", amount),
		"time":        timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ApproveAgent approves an agent to trade on behalf of the user
// Returns the result and the generated agent private key
func (e *Exchange) ApproveAgent(name *string) (any, string, error) {
	// Generate agent key
	agentBytes := make([]byte, 32)
	if _, err := rand.Read(agentBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate agent key: %w", err)
	}
	agentKey := "0x" + hex.EncodeToString(agentBytes)

	privateKey, err := crypto.HexToECDSA(agentKey[2:])
	if err != nil {
		return nil, "", fmt.Errorf("failed to create private key: %w", err)
	}

	agentAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":         "approveAgent",
		"agentAddress": agentAddress,
		"nonce":        timestamp,
	}

	if name != nil {
		action["agentName"] = *name
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, "", err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, "", err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, "", err
	}
	return result, agentKey, nil
}

// ApproveBuilderFee approves builder fee payment
func (e *Exchange) ApproveBuilderFee(builder string, maxFeeRate string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":       "approveBuilderFee",
		"builder":    builder,
		"maxFeeRate": maxFeeRate,
		"nonce":      timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// ConvertToMultiSigUser converts account to multi-signature user
func (e *Exchange) ConvertToMultiSigUser(authorizedUsers []string, threshold int) (any, error) {
	timestamp := time.Now().UnixMilli()

	// Sort users as done in Python
	sort.Strings(authorizedUsers)

	signers := map[string]any{
		"authorizedUsers": authorizedUsers,
		"threshold":       threshold,
	}

	signersJSON, err := json.Marshal(signers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signers: %w", err)
	}

	action := map[string]any{
		"type":    "convertToMultiSigUser",
		"signers": string(signersJSON),
		"nonce":   timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Spot Deploy Methods

// SpotDeployRegisterToken registers a new spot token
func (e *Exchange) SpotDeployRegisterToken(
	tokenName string,
	szDecimals int,
	weiDecimals int,
	maxGas int,
	fullName string,
) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "spotDeploy",
		"registerToken2": map[string]any{
			"spec": map[string]any{
				"name":        tokenName,
				"szDecimals":  szDecimals,
				"weiDecimals": weiDecimals,
			},
			"maxGas":   maxGas,
			"fullName": fullName,
		},
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address for spot deploy
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotDeployUserGenesis initializes user genesis for spot trading
func (e *Exchange) SpotDeployUserGenesis(balances map[string]float64) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":     "spotDeployUserGenesis",
		"balances": balances,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotDeployEnableFreezePrivilege enables freeze privilege for spot deployer
func (e *Exchange) SpotDeployEnableFreezePrivilege() (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "spotDeployEnableFreezePrivilege",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotDeployFreezeUser freezes a user in spot trading
func (e *Exchange) SpotDeployFreezeUser(userAddress string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":        "spotDeployFreezeUser",
		"userAddress": userAddress,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotDeployRevokeFreezePrivilege revokes freeze privilege for spot deployer
func (e *Exchange) SpotDeployRevokeFreezePrivilege() (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "spotDeployRevokeFreezePrivilege",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotDeployGenesis initializes spot genesis
func (e *Exchange) SpotDeployGenesis(deployer string, dexName string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":     "spotDeployGenesis",
		"deployer": deployer,
		"dexName":  dexName,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotDeployRegisterSpot registers spot market
func (e *Exchange) SpotDeployRegisterSpot(baseToken string, quoteToken string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":       "spotDeployRegisterSpot",
		"baseToken":  baseToken,
		"quoteToken": quoteToken,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotDeployRegisterHyperliquidity registers hyperliquidity spot
func (e *Exchange) SpotDeployRegisterHyperliquidity(name string, tokens []string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":   "spotDeployRegisterHyperliquidity",
		"name":   name,
		"tokens": tokens,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// SpotDeploySetDeployerTradingFeeShare sets deployer trading fee share
func (e *Exchange) SpotDeploySetDeployerTradingFeeShare(feeShare float64) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":     "spotDeploySetDeployerTradingFeeShare",
		"feeShare": feeShare,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Perp Deploy Methods

// PerpDeployRegisterAsset registers a new perpetual asset
func (e *Exchange) PerpDeployRegisterAsset(
	asset string,
	perpDexInput PerpDexSchemaInput,
) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":         "perpDeployRegisterAsset",
		"asset":        asset,
		"perpDexInput": perpDexInput,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// PerpDeploySetOracle sets oracle for perpetual asset
func (e *Exchange) PerpDeploySetOracle(asset string, oracleAddress string) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":          "perpDeploySetOracle",
		"asset":         asset,
		"oracleAddress": oracleAddress,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CSigner Methods

// CSignerUnjailSelf unjails self as consensus signer
func (e *Exchange) CSignerUnjailSelf() (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "cSignerUnjailSelf",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CSignerJailSelf jails self as consensus signer
func (e *Exchange) CSignerJailSelf() (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "cSignerJailSelf",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CSignerInner executes inner consensus signer action
func (e *Exchange) CSignerInner(innerAction map[string]any) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":        "cSignerInner",
		"innerAction": innerAction,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CValidator Methods

// CValidatorRegister registers as consensus validator
func (e *Exchange) CValidatorRegister(validatorProfile map[string]any) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":             "cValidatorRegister",
		"validatorProfile": validatorProfile,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CValidatorChangeProfile changes validator profile
func (e *Exchange) CValidatorChangeProfile(newProfile map[string]any) (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":       "cValidatorChangeProfile",
		"newProfile": newProfile,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// CValidatorUnregister unregisters as consensus validator
func (e *Exchange) CValidatorUnregister() (any, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "cValidatorUnregister",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// MultiSig executes multi-signature action
func (e *Exchange) MultiSig(
	action map[string]any,
	signers []string,
	signatures []string,
) (any, error) {
	timestamp := time.Now().UnixMilli()

	multiSigAction := map[string]any{
		"type":       "multiSig",
		"action":     action,
		"signers":    signers,
		"signatures": signatures,
	}

	sig, err := SignL1Action(
		e.privateKey,
		multiSigAction,
		e.vault,
		timestamp,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(multiSigAction, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result any
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return result, nil
}

// Helper functions
func roundToDecimals(value float64, decimals int) float64 {
	// Implementation for rounding to specific decimals
	// This is a simplified version - proper implementation would use math.Pow
	return value
}

func parseFloat(s string) float64 {
	// Implementation for parsing float from string
	// This is a simplified version - proper implementation would use strconv.ParseFloat
	return 0.0
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func formatFloat(f float64) string {
	// Implementation for formatting float to string
	// This is a simplified version - proper implementation would use strconv.FormatFloat
	return fmt.Sprintf("%f", f)
}

func (e *Exchange) postAction(action, signature any, nonce int64) ([]byte, error) {
	payload := map[string]any{
		"action":    action,
		"nonce":     nonce,
		"signature": signature,
	}

	if action.(map[string]any)["type"] != "usdClassTransfer" {
		payload["vaultAddress"] = e.vault
	}

	return e.client.post("/exchange", payload)
}
