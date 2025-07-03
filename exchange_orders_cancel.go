package hyperliquid

import (
	"strconv"

	"github.com/sonirico/vago/slices"
)

type (
	CancelOrderRequest struct {
		Coin    string
		OrderID int64
	}

	CancelOrderResponse struct {
		Statuses MixedArray
	}
)

func (e *Exchange) Cancel(coin string, oid int64) (res *APIResponse[CancelOrderResponse], err error) {
	return e.BulkCancel([]CancelOrderRequest{
		{
			Coin:    coin,
			OrderID: oid,
		},
	})
}

func (e *Exchange) BulkCancel(requests []CancelOrderRequest) (res *APIResponse[CancelOrderResponse], err error) {
	type cancelOrderItem struct {
		Asset   int    `json:"a"`
		OrderID string `json:"o"`
	}

	action := map[string]any{
		"type": "cancel",
		"cancels": slices.Map(requests, func(req CancelOrderRequest) cancelOrderItem {
			return cancelOrderItem{
				Asset:   e.info.NameToAsset(req.Coin),
				OrderID: strconv.FormatInt(req.OrderID, 10),
			}
		}),
	}

	if err = e.executeAction(action, &res); err != nil {
		return
	}
	return
}

type CancelOrderRequestByCloid struct {
	Coin  string
	Cloid string
}

func (e *Exchange) CancelByCloid(coin, cloid string) (res *APIResponse[CancelOrderResponse], err error) {
	return e.BulkCancelByCloids([]CancelOrderRequestByCloid{
		{
			Coin:  coin,
			Cloid: cloid,
		},
	})
}

func (e *Exchange) BulkCancelByCloids(requests []CancelOrderRequestByCloid) (res *APIResponse[CancelOrderResponse], err error) {
	type cancelOrderItemByCloid struct {
		Asset   int    `json:"a"`
		OrderID string `json:"cloid"`
	}

	action := map[string]any{
		"type": "cancelByCloid",
		"cancels": slices.Map(requests, func(req CancelOrderRequestByCloid) cancelOrderItemByCloid {
			return cancelOrderItemByCloid{
				Asset:   e.info.NameToAsset(req.Coin),
				OrderID: req.Cloid,
			}
		}),
	}

	if err = e.executeAction(action, &res); err != nil {
		return
	}
	return
}
