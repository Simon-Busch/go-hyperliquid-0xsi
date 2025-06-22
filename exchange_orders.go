package hyperliquid

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// floatToWire converts a float64 to a wire-compatible string format
func floatToWire(x float64) (string, error) {
	// Format to 8 decimal places
	rounded := fmt.Sprintf("%.8f", x)

	// Check if rounding causes significant error
	parsed, err := strconv.ParseFloat(rounded, 64)
	if err != nil {
		return "", err
	}

	if math.Abs(parsed-x) >= 1e-12 {
		return "", fmt.Errorf("float_to_wire causes rounding: %f", x)
	}

	// Handle -0 case
	if rounded == "-0.00000000" {
		rounded = "0.00000000"
	}

	// Remove trailing zeros and decimal point if not needed
	result := strings.TrimRight(rounded, "0")
	result = strings.TrimRight(result, ".")

	return result, nil
}

type CreateOrderRequest struct {
	Coin          string
	IsBuy         bool
	Price         float64
	Size          float64
	ReduceOnly    bool
	OrderType     OrderType
	ClientOrderID *string
}

type createOrderRequest struct {
	Asset         int       `json:"a"`
	IsBuy         bool      `json:"b"`
	Price         string    `json:"p"`
	Size          string    `json:"s"`
	ReduceOnly    bool      `json:"r"`
	OrderType     OrderType `json:"t"`
	ClientOrderID *string   `json:"c,omitempty"`
}

type OrderStatusResting struct {
	Oid      int64  `json:"oid"`
	ClientID string `json:"cid"`
	Status   string `json:"status"`
}

type OrderStatusFilled struct {
	TotalSz string `json:"totalSz"`
	AvgPx   string `json:"avgPx"`
	Oid     int    `json:"oid"`
}

type OrderStatus struct {
	Resting *OrderStatusResting `json:"resting,omitempty"`
	Filled  *OrderStatusFilled  `json:"filled,omitempty"`
	Error   *error              `json:"error,omitempty"`
}

type OrderResponse struct {
	Statuses []OrderStatus
}

func newOrderRequest(e *Exchange, p CreateOrderRequest) (createOrderRequest, error) {
	priceWire, err := floatToWire(p.Price)
	if err != nil {
		return createOrderRequest{}, fmt.Errorf("failed to wire price: %w", err)
	}

	sizeWire, err := floatToWire(p.Size)
	if err != nil {
		return createOrderRequest{}, fmt.Errorf("failed to wire size: %w", err)
	}

	return createOrderRequest{
		Asset:         e.info.NameToAsset(p.Coin),
		IsBuy:         p.IsBuy,
		Price:         priceWire,
		Size:          sizeWire,
		ReduceOnly:    p.ReduceOnly,
		OrderType:     p.OrderType,
		ClientOrderID: p.ClientOrderID,
	}, nil
}

func newCreateOrderAction(
	e *Exchange,
	orders []CreateOrderRequest,
	info *BuilderInfo,
) (map[string]any, error) {
	orderRequests := make([]createOrderRequest, len(orders))
	for i, order := range orders {
		req, err := newOrderRequest(e, order)
		if err != nil {
			return nil, fmt.Errorf("failed to create order request %d: %w", i, err)
		}
		orderRequests[i] = req
	}

	res := map[string]any{
		"type":     "order",
		"orders":   orderRequests,
		"grouping": GroupingNA,
	}

	if info != nil {
		res["builder"] = *info
	}

	return res, nil
}

func (e *Exchange) Order(
	req CreateOrderRequest,
	builder *BuilderInfo,
) (result OrderStatus, err error) {
	resp, err := e.BulkOrders([]CreateOrderRequest{req}, builder)
	if err != nil {
		return
	}

	if !resp.Ok {
		err = fmt.Errorf("failed to create order: %s", resp.Err)
		return
	}

	data := resp.Data
	if len(data.Statuses) == 0 {
		err = fmt.Errorf("no status for order: %s", resp.Err)
		return
	}

	return data.Statuses[0], nil
}

func (e *Exchange) BulkOrders(
	orders []CreateOrderRequest,
	builder *BuilderInfo,
) (result *APIResponse[OrderResponse], err error) {
	action, err := newCreateOrderAction(e, orders, builder)
	if err != nil {
		return nil, err
	}
	err = e.executeAction(action, &result)
	return
}

type cancelOrderItem struct {
	Asset   int    `json:"a"`
	OrderID string `json:"o"`
}

type CancelOrderResponse struct {
	Statuses MixedArray
}

func (e *Exchange) Cancel(coin string, oid int64) (res *APIResponse[CancelResponse], err error) {
	action := map[string]any{
		"type": "cancel",
		"cancels": []cancelOrderItem{
			{
				Asset:   e.info.NameToAsset(coin),
				OrderID: strconv.FormatInt(oid, 10),
			},
		},
	}

	if err = e.executeAction(action, &res); err != nil {
		return
	}
	return
}

type cancelOrderItemByCloid struct {
	Asset   int    `json:"a"`
	OrderID string `json:"cloid"`
}

func (e *Exchange) CancelByCloid(coin, cloid string) (res *APIResponse[CancelResponse], err error) {
	action := map[string]any{
		"type": "cancelByCloid",
		"cancels": []cancelOrderItemByCloid{
			{
				Asset:   e.info.NameToAsset(coin),
				OrderID: cloid,
			},
		},
	}

	if err = e.executeAction(action, &res); err != nil {
		return
	}
	return
}

type ModifyOrderRequest struct {
	Oid   any // can be int64 or Cloid
	Order CreateOrderRequest
}

type modifyOrderRequest struct {
	Asset         int       `json:"a"`
	IsBuy         bool      `json:"b"`
	Price         string    `json:"p"`
	Size          string    `json:"s"`
	ReduceOnly    bool      `json:"r"`
	OrderType     OrderType `json:"t"`
	ClientOrderID *string   `json:"c,omitempty"`
}

func newModifyOrderAction(
	e *Exchange,
	modifyRequest ModifyOrderRequest,
) (map[string]any, error) {
	priceWire, err := floatToWire(modifyRequest.Order.Price)
	if err != nil {
		return nil, fmt.Errorf("failed to wire price: %w", err)
	}

	sizeWire, err := floatToWire(modifyRequest.Order.Size)
	if err != nil {
		return nil, fmt.Errorf("failed to wire size: %w", err)
	}

	return map[string]any{
		"type": "modify",
		"oid":  modifyRequest.Oid,
		"order": modifyOrderRequest{
			Asset:         e.info.NameToAsset(modifyRequest.Order.Coin),
			IsBuy:         modifyRequest.Order.IsBuy,
			Price:         priceWire,
			Size:          sizeWire,
			ReduceOnly:    modifyRequest.Order.ReduceOnly,
			OrderType:     modifyRequest.Order.OrderType,
			ClientOrderID: modifyRequest.Order.ClientOrderID,
		},
	}, nil
}

func newModifyOrdersAction(
	e *Exchange,
	modifyRequests []ModifyOrderRequest,
) (map[string]any, error) {
	modifies := make([]map[string]any, len(modifyRequests))
	for i, req := range modifyRequests {
		modify, err := newModifyOrderAction(e, req)
		if err != nil {
			return nil, fmt.Errorf("failed to create modify request %d: %w", i, err)
		}
		modifies[i] = modify
	}

	return map[string]any{
		"type":     "batchModify",
		"modifies": modifies,
	}, nil
}

// ModifyOrder modifies an existing order
func (e *Exchange) ModifyOrder(
	req ModifyOrderRequest,
) (result OrderStatus, err error) {
	resp := APIResponse[OrderResponse]{}
	action, err := newModifyOrderAction(e, req)
	if err != nil {
		return result, fmt.Errorf("failed to create modify action: %w", err)
	}

	err = e.executeAction(action, &resp)
	if err != nil {
		err = fmt.Errorf("failed to modify order: %w", err)
		return
	}

	if !resp.Ok {
		err = fmt.Errorf("failed to modify order: %s", resp.Err)
		return
	}

	data := resp.Data
	if len(data.Statuses) == 0 {
		err = fmt.Errorf("no status for modified order: %s", resp.Err)
		return
	}

	return data.Statuses[0], nil
}

// BulkModifyOrders modifies multiple orders
func (e *Exchange) BulkModifyOrders(
	modifyRequests []ModifyOrderRequest,
) ([]OrderStatus, error) {
	resp := APIResponse[OrderResponse]{}
	action, err := newModifyOrdersAction(e, modifyRequests)
	if err != nil {
		return nil, fmt.Errorf("failed to create bulk modify action: %w", err)
	}

	err = e.executeAction(action, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to modify orders: %w", err)
	}

	if !resp.Ok {
		return nil, fmt.Errorf("failed to modify orders: %s", resp.Err)
	}

	data := resp.Data
	if len(data.Statuses) == 0 {
		return nil, fmt.Errorf("no status for modified order: %s", resp.Err)
	}

	return data.Statuses, nil
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
) (res OrderStatus, err error) {
	slippagePrice, err := e.SlippagePrice(name, isBuy, slippage, px)
	if err != nil {
		return
	}

	orderType := OrderType{
		Limit: &LimitOrderType{Tif: TifIoc},
	}

	return e.Order(CreateOrderRequest{
		Coin:          name,
		IsBuy:         isBuy,
		Size:          sz,
		Price:         slippagePrice,
		OrderType:     orderType,
		ReduceOnly:    false,
		ClientOrderID: cloid,
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
) (OrderStatus, error) {
	address := e.accountAddr
	if address == "" {
		address = e.vault
	}

	userState, err := e.info.UserState(address)
	if err != nil {
		return OrderStatus{}, err
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
			return OrderStatus{}, err
		}

		orderType := OrderType{
			Limit: &LimitOrderType{Tif: TifIoc},
		}

		return e.Order(CreateOrderRequest{
			Coin:          coin,
			IsBuy:         isBuy,
			Size:          size,
			Price:         slippagePrice,
			OrderType:     orderType,
			ReduceOnly:    true,
			ClientOrderID: cloid,
		}, builder)
	}

	return OrderStatus{}, fmt.Errorf("position not found for coin: %s", coin)
}
