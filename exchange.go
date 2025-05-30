package hyperliquid

import (
	"crypto/ecdsa"
	"encoding/json"
	"time"
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

func (e *Exchange) CancelAll(coin string) ([]OpenOrder, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "cancelAll",
		"coin": coin,
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

func (e *Exchange) UpdateLeverage(coin string, leverage int) (*UserState, error) {
	action := map[string]any{
		"type": "updateLeverage",
		"coin": coin,
		"leverage": map[string]any{
			"type":  "isolated",
			"value": leverage,
		},
	}

	var result UserState
	if err := e.executeAction(action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (e *Exchange) UpdateIsolatedMargin(coin string, margin float64) (*UserState, error) {
	action := map[string]any{
		"type":        "updateIsolatedMargin",
		"coin":        coin,
		"marginDelta": margin,
	}

	var result UserState
	if err := e.executeAction(action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (e *Exchange) WithdrawEth(amount float64, destination string) (*UserState, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "withdraw",
		"coin": "ETH",
		"amount": map[string]any{
			"fix":   amount,
			"asset": "ETH",
		},
		"destination": destination,
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

	var result UserState
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (e *Exchange) WithdrawUsdc(amount float64, destination string) (*UserState, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "usdTransfer",
		"amount": map[string]any{
			"fix":   amount,
			"asset": "USDC",
		},
		"destination": destination,
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

	var result UserState
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (e *Exchange) Transfer(amount float64, destination string) (*UserState, error) {
	action := map[string]any{
		"type":        "transfer",
		"destination": destination,
		"amount":      amount,
	}

	var result UserState
	if err := e.executeAction(action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ... Additional methods for other operations like cancels, transfers etc.

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
