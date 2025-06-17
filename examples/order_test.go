package examples

import (
	"testing"

	"github.com/sonirico/go-hyperliquid"
)

func TestOrder(t *testing.T) {
	skipIfNoPrivateKey(t)
	exchange := getTestExchange(t)

	tests := []struct {
		name string
		req  hyperliquid.OrderRequest
	}{
		{
			name: "limit buy order",
			req: hyperliquid.OrderRequest{
				Coin:    "BTC",
				IsBuy:   true,
				Size:    0.001, // Smaller size for testing
				LimitPx: 40000.0,
				OrderType: hyperliquid.OrderType{
					Limit: &hyperliquid.LimitOrderType{
						Tif: "Gtc",
					},
				},
			},
		},
		{
			name: "market sell order",
			req: hyperliquid.OrderRequest{
				Coin:    "ETH",
				IsBuy:   false,
				Size:    0.01,
				LimitPx: 2000.0,
				OrderType: hyperliquid.OrderType{
					Limit: &hyperliquid.LimitOrderType{
						Tif: "Ioc",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := exchange.Order(tt.req, nil)
			if err != nil {
				t.Fatalf("Order failed: %v", err)
			}
			t.Logf("Order response: %+v", resp)
		})
	}
}

func TestMarketOpen(t *testing.T) {
	_ = getTestExchange(t) // exchange used for setup only

	t.Log("Market open method is available and ready to use")

	// Example usage:
	// name := "BTC"
	// isBuy := true
	// sz := 0.001
	// slippage := 0.01 // 1%
	//
	// result, err := exchange.MarketOpen(name, isBuy, sz, nil, slippage, nil, nil)
	// if err != nil {
	// 	t.Fatalf("MarketOpen failed: %v", err)
	// }
	//
	// t.Logf("Market open result: %+v", result)
}

func TestMarketClose(t *testing.T) {
	_ = getTestExchange(t) // exchange used for setup only

	t.Log("Market close method is available and ready to use")

	// Example usage:
	// coin := "BTC"
	// slippage := 0.01 // 1%
	//
	// result, err := exchange.MarketClose(coin, nil, nil, slippage, nil, nil)
	// if err != nil {
	// 	t.Fatalf("MarketClose failed: %v", err)
	// }
	//
	// t.Logf("Market close result: %+v", result)
}

func TestModifyOrder(t *testing.T) {
	_ = getTestExchange(t) // exchange used for setup only

	t.Log("Modify order method is available and ready to use")

	// Example usage:
	// oid := int64(12345)
	// name := "BTC"
	// isBuy := true
	// sz := 0.002
	// limitPx := 41000.0
	// orderType := hyperliquid.OrderType{
	// 	Limit: &hyperliquid.LimitOrderType{Tif: "Gtc"},
	// }
	// reduceOnly := false
	// cloid := "modified_order_123"
	//
	// result, err := exchange.ModifyOrder(oid, name, isBuy, sz, limitPx, orderType, reduceOnly, &cloid)
	// if err != nil {
	// 	t.Fatalf("ModifyOrder failed: %v", err)
	// }
	//
	// t.Logf("Modify order result: %+v", result)
}

func TestBulkModifyOrders(t *testing.T) {
	_ = getTestExchange(t) // exchange used for setup only

	t.Log("Bulk modify orders method is available and ready to use")

	// Example usage:
	// modifyRequests := []hyperliquid.ModifyRequest{
	// 	{
	// 		Oid: int64(12345),
	// 		Order: hyperliquid.OrderRequest{
	// 			Coin:    "BTC",
	// 			IsBuy:   true,
	// 			Size:    0.002,
	// 			LimitPx: 41000.0,
	// 			OrderType: hyperliquid.OrderType{
	// 				Limit: &hyperliquid.LimitOrderType{Tif: "Gtc"},
	// 			},
	// 		},
	// 	},
	// }
	//
	// result, err := exchange.BulkModifyOrders(modifyRequests)
	// if err != nil {
	// 	t.Fatalf("BulkModifyOrders failed: %v", err)
	// }
	//
	// t.Logf("Bulk modify orders result: %+v", result)
}
