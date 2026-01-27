package hyperliquid

import (
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

func (e *Exchange) Cancel(
	coin string,
	oid int64,
) (res *APIResponse[CancelOrderResponse], err error) {
	return e.BulkCancel([]CancelOrderRequest{
		{
			Coin:    coin,
			OrderID: oid,
		},
	})
}

func (e *Exchange) BulkCancel(
	requests []CancelOrderRequest,
) (res *APIResponse[CancelOrderResponse], err error) {
	cancels := slices.Map(requests, func(req CancelOrderRequest) CancelOrderWire {
		return CancelOrderWire{
			Asset:   e.info.NameToAsset(req.Coin),
			OrderID: req.OrderID,
		}
	})

	action := CancelAction{
		Type:    "cancel",
		Dex:     e.dex, // Include dex for HIP-3 builder-deployed perps
		Cancels: cancels,
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

func (e *Exchange) CancelByCloid(
	coin, cloid string,
) (res *APIResponse[CancelOrderResponse], err error) {
	return e.BulkCancelByCloids([]CancelOrderRequestByCloid{
		{
			Coin:  coin,
			Cloid: cloid,
		},
	})
}

func (e *Exchange) BulkCancelByCloids(
	requests []CancelOrderRequestByCloid,
) (res *APIResponse[CancelOrderResponse], err error) {
	cancels := slices.Map(requests, func(req CancelOrderRequestByCloid) CancelByCloidWire {
		return CancelByCloidWire{
			Asset:    e.info.NameToAsset(req.Coin),
			ClientID: req.Cloid,
		}
	})

	action := CancelByCloidAction{
		Type:    "cancelByCloid",
		Dex:     e.dex, // Include dex for HIP-3 builder-deployed perps
		Cancels: cancels,
	}

	if err = e.executeAction(action, &res); err != nil {
		return
	}
	return
}
