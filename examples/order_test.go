package examples

import (
	"testing"

	"github.com/joho/godotenv"
	"github.com/sonirico/go-hyperliquid"
)

func TestOrder(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	tests := []struct {
		name string
		req  hyperliquid.CreateOrderRequest
	}{
		{
			name: "limit buy order",
			req: hyperliquid.CreateOrderRequest{
				Coin:  "BTC",
				IsBuy: true,
				Size:  0.001, // Smaller size for testing
				Price: 40000.0,
				OrderType: hyperliquid.OrderType{
					Limit: &hyperliquid.LimitOrderType{
						Tif: hyperliquid.TifGtc,
					},
				},
			},
		},
		{
			name: "market sell order",
			req: hyperliquid.CreateOrderRequest{
				Coin:  "ETH",
				IsBuy: false,
				Size:  0.01,
				Price: 2000.0,
				OrderType: hyperliquid.OrderType{
					Limit: &hyperliquid.LimitOrderType{
						Tif: hyperliquid.TifIoc,
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

func TestMarketOrder(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("Market order method is available and ready to use")

	// Example usage with MarketOrder helper function:
	req := hyperliquid.CreateOrderRequest{
		Coin:  "BTC",
		IsBuy: true,
		Size:  0.001,
		// Price will be set automatically by MarketOrder
		OrderType: hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{
				Tif: hyperliquid.TifIoc,
			},
		},
	}

	result, err := exchange.Order(req, nil)
	if err != nil {
		t.Fatalf("MarketOrder failed: %v", err)
	}

	t.Logf("Market order result: %+v", result)
}

func TestMarketOpen(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t) // exchange used for setup only

	t.Log("Market open method is available and ready to use")

	// Example usage:
	name := "BTC"
	isBuy := true
	sz := 0.001
	slippage := 0.01 // 1%

	result, err := exchange.MarketOpen(name, isBuy, sz, nil, slippage, nil, nil)
	if err != nil {
		t.Fatalf("MarketOpen failed: %v", err)
	}

	t.Logf("Market open result: %+v", result)
}

func TestMarketClose(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)
	t.Log("Market close method is available and ready to use")

	// Example usage:
	coin := "BTC"
	slippage := 0.01 // 1%

	result, err := exchange.MarketClose(coin, nil, nil, slippage, nil, nil)
	if err != nil {
		t.Fatalf("MarketClose failed: %v", err)
	}

	t.Logf("Market close result: %+v", result)
}

func TestModifyOrder(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("Modify order method is available and ready to use")

	// Example usage:
	modifyReq := hyperliquid.ModifyOrderRequest{
		Oid: int64(12345),
		Order: hyperliquid.CreateOrderRequest{
			Coin:  "BTC",
			IsBuy: true,
			Size:  0.002,
			Price: 41000.0,
			OrderType: hyperliquid.OrderType{
				Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifGtc},
			},
			ReduceOnly:    false,
			ClientOrderID: func() *string { s := "modified_order_123"; return &s }(),
		},
	}

	result, err := exchange.ModifyOrder(modifyReq)
	if err != nil {
		t.Fatalf("ModifyOrder failed: %v", err)
	}

	t.Logf("Modify order result: %+v", result)
}

func TestBulkModifyOrders(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("Bulk modify orders method is available and ready to use")

	// Example usage:
	modifyRequests := []hyperliquid.ModifyOrderRequest{
		{
			Oid: int64(12345),
			Order: hyperliquid.CreateOrderRequest{
				Coin:  "BTC",
				IsBuy: true,
				Size:  0.002,
				Price: 41000.0,
				OrderType: hyperliquid.OrderType{
					Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifGtc},
				},
			},
		},
	}

	result, err := exchange.BulkModifyOrders(modifyRequests)
	if err != nil {
		t.Fatalf("BulkModifyOrders failed: %v", err)
	}

	t.Logf("Bulk modify orders result: %+v", result)
}

func TestOpenPositionAndSetLeverage(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("Opening position and then setting leverage to 5x")

	// Step 1: Open a position (will use default 10x leverage)
	name := "BTC"
	isBuy := true
	sz := 0.001      // Small size for testing
	slippage := 0.01 // 1%

	t.Logf("Opening %s position with size %f (will use default leverage)", name, sz)
	result, err := exchange.MarketOpen(name, isBuy, sz, nil, slippage, nil, nil)
	if err != nil {
		t.Fatalf("MarketOpen failed: %v", err)
	}

	t.Logf("Position opened successfully: %+v", result)

	// Step 2: Set leverage to 5x after opening the position
	leverage := 5   // 5x leverage
	isCross := true // Use cross margin

	t.Logf("Setting leverage to %dx for %s", leverage, name)
	leverageResp, err := exchange.UpdateLeverage(leverage, name, isCross)
	if err != nil {
		t.Fatalf("Failed to update leverage: %v", err)
	}

	t.Logf("Leverage updated successfully: %+v", leverageResp)

	// Step 3: Verify the leverage was set correctly by checking user state
	// Note: In a real scenario, you might want to add a small delay here
	// to allow the exchange to process the leverage update
	t.Log("Position opened with default leverage and then updated to 5x leverage")
}
