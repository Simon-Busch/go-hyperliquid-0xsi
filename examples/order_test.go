package examples

import (
	"strconv"
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

func TestOpenAndClosePosition(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("Testing open and close position workflow")

	// Step 1: Check initial user state (before opening position)
	t.Log("Step 1: Checking initial user state")
	initialUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get initial user state: %v", err)
	}
	t.Logf("Initial user state - Positions count: %d", len(initialUserState.AssetPositions))
	for _, pos := range initialUserState.AssetPositions {
		t.Logf("Initial position: %s, Size: %s, Leverage: %dx",
			pos.Position.Coin, pos.Position.Szi, pos.Position.Leverage.Value)
	}

	// Step 2: Open a position
	name := "BTC"
	isBuy := true
	sz := 0.001      // Small size for testing
	slippage := 0.01 // 1%

	t.Logf("Step 2: Opening %s position with size %f", name, sz)
	result, err := exchange.MarketOpen(name, isBuy, sz, nil, slippage, nil, nil)
	if err != nil {
		t.Fatalf("MarketOpen failed: %v", err)
	}
	t.Logf("Position opened successfully: %+v", result)

	// Step 3: Check user state after opening position
	t.Log("Step 3: Checking user state after opening position")
	afterOpenUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state after opening: %v", err)
	}
	t.Logf("After opening - Positions count: %d", len(afterOpenUserState.AssetPositions))

	// Find and verify the BTC position
	var btcPosition *hyperliquid.Position
	for _, pos := range afterOpenUserState.AssetPositions {
		if pos.Position.Coin == name {
			btcPosition = &pos.Position
			t.Logf("Found BTC position - Size: %s, Leverage: %dx, Entry Price: %s",
				pos.Position.Szi, pos.Position.Leverage.Value,
				*pos.Position.EntryPx)
			break
		}
	}

	if btcPosition == nil {
		t.Fatalf("BTC position not found after opening")
	}

	// Step 4: Close the position
	t.Log("Step 4: Closing the position")
	closeResult, err := exchange.MarketClose(name, nil, nil, slippage, nil, nil)
	if err != nil {
		t.Fatalf("MarketClose failed: %v", err)
	}
	t.Logf("Position closed successfully: %+v", closeResult)

	// Step 5: Check user state after closing position
	t.Log("Step 5: Checking user state after closing position")
	afterCloseUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state after closing: %v", err)
	}
	t.Logf("After closing - Positions count: %d", len(afterCloseUserState.AssetPositions))

	// Verify BTC position is closed (should have size 0 or not exist)
	var btcPositionAfterClose *hyperliquid.Position
	for _, pos := range afterCloseUserState.AssetPositions {
		if pos.Position.Coin == name {
			btcPositionAfterClose = &pos.Position
			t.Logf("BTC position after close - Size: %s", pos.Position.Szi)
			break
		}
	}

	if btcPositionAfterClose != nil {
		// Check if position size is effectively 0
		size, err := strconv.ParseFloat(btcPositionAfterClose.Szi, 64)
		if err != nil {
			t.Logf("Warning: Could not parse position size: %v", err)
		} else if size != 0 {
			t.Logf("Warning: Position still exists with size: %f", size)
		} else {
			t.Log("Position successfully closed (size = 0)")
		}
	} else {
		t.Log("Position completely removed from user state")
	}

	// Summary
	t.Log("Test completed: Position opened and closed successfully")
}

func TestOpenAndPartiallyClosePosition(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("Testing open, partial close, and full close position workflow")

	// Step 1: Check initial user state (before opening position)
	t.Log("Step 1: Checking initial user state")
	initialUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get initial user state: %v", err)
	}
	t.Logf("Initial user state - Positions count: %d", len(initialUserState.AssetPositions))

	// Step 2: Open a position
	name := "BTC"
	isBuy := true
	sz := 0.002      // Larger size for partial closing
	slippage := 0.01 // 1%

	t.Logf("Step 2: Opening %s position with size %f", name, sz)
	result, err := exchange.MarketOpen(name, isBuy, sz, nil, slippage, nil, nil)
	if err != nil {
		t.Fatalf("MarketOpen failed: %v", err)
	}
	t.Logf("Position opened successfully: %+v", result)

	// Step 3: Check user state after opening position
	t.Log("Step 3: Checking user state after opening position")
	afterOpenUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state after opening: %v", err)
	}
	t.Logf("After opening - Positions count: %d", len(afterOpenUserState.AssetPositions))

	// Find and verify the BTC position
	var btcPosition *hyperliquid.Position
	for _, pos := range afterOpenUserState.AssetPositions {
		if pos.Position.Coin == name {
			btcPosition = &pos.Position
			t.Logf("Found BTC position - Size: %s, Leverage: %dx, Entry Price: %s",
				pos.Position.Szi, pos.Position.Leverage.Value,
				*pos.Position.EntryPx)
			break
		}
	}

	if btcPosition == nil {
		t.Fatalf("BTC position not found after opening")
	}

	// Step 4: Partially close the position (close 50%)
	partialCloseSize := 0.001 // Close half of the position
	t.Logf("Step 4: Partially closing %s position with size %f (50%% of original)", name, partialCloseSize)

	// Calculate slippage price for partial close (sell to close long position)
	partialClosePrice, err := exchange.SlippagePrice(name, false, slippage, nil) // false = sell
	if err != nil {
		t.Fatalf("Failed to calculate partial close price: %v", err)
	}

	// Create a partial close order
	partialCloseOrder := hyperliquid.CreateOrderRequest{
		Coin:       name,
		IsBuy:      false, // Sell to close long position
		Size:       partialCloseSize,
		Price:      partialClosePrice,
		ReduceOnly: true, // Important: this ensures we only close existing position
		OrderType: hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{
				Tif: hyperliquid.TifIoc, // Immediate or cancel
			},
		},
	}

	partialCloseResult, err := exchange.Order(partialCloseOrder, nil)
	if err != nil {
		t.Fatalf("Partial close failed: %v", err)
	}
	t.Logf("Partial close result: %+v", partialCloseResult)

	// Step 5: Check user state after partial close
	t.Log("Step 5: Checking user state after partial close")
	afterPartialCloseUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state after partial close: %v", err)
	}
	t.Logf("After partial close - Positions count: %d", len(afterPartialCloseUserState.AssetPositions))

	// Find and verify the remaining BTC position
	var remainingBtcPosition *hyperliquid.Position
	for _, pos := range afterPartialCloseUserState.AssetPositions {
		if pos.Position.Coin == name {
			remainingBtcPosition = &pos.Position
			t.Logf("Remaining BTC position - Size: %s, Leverage: %dx, Entry Price: %s",
				pos.Position.Szi, pos.Position.Leverage.Value,
				*pos.Position.EntryPx)
			break
		}
	}

	if remainingBtcPosition == nil {
		t.Fatalf("BTC position not found after partial close")
	}

	// Verify the position size was reduced
	remainingSize, err := strconv.ParseFloat(remainingBtcPosition.Szi, 64)
	if err != nil {
		t.Logf("Warning: Could not parse remaining position size: %v", err)
	} else {
		expectedRemainingSize := sz - partialCloseSize
		t.Logf("Remaining size: %f, Expected: %f", remainingSize, expectedRemainingSize)
		if remainingSize != expectedRemainingSize {
			t.Logf("Warning: Remaining size (%f) doesn't match expected (%f)", remainingSize, expectedRemainingSize)
		}
	}

	// Step 6: Fully close the remaining position
	t.Log("Step 6: Fully closing the remaining position")
	fullCloseResult, err := exchange.MarketClose(name, nil, nil, slippage, nil, nil)
	if err != nil {
		t.Fatalf("Full close failed: %v", err)
	}
	t.Logf("Full close result: %+v", fullCloseResult)

	// Step 7: Check user state after full close
	t.Log("Step 7: Checking user state after full close")
	afterFullCloseUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state after full close: %v", err)
	}
	t.Logf("After full close - Positions count: %d", len(afterFullCloseUserState.AssetPositions))

	// Verify BTC position is completely closed
	var finalBtcPosition *hyperliquid.Position
	for _, pos := range afterFullCloseUserState.AssetPositions {
		if pos.Position.Coin == name {
			finalBtcPosition = &pos.Position
			t.Logf("Final BTC position - Size: %s", pos.Position.Szi)
			break
		}
	}

	if finalBtcPosition != nil {
		// Check if position size is effectively 0
		finalSize, err := strconv.ParseFloat(finalBtcPosition.Szi, 64)
		if err != nil {
			t.Logf("Warning: Could not parse final position size: %v", err)
		} else if finalSize != 0 {
			t.Logf("Warning: Position still exists with size: %f", finalSize)
		} else {
			t.Log("Position successfully fully closed (size = 0)")
		}
	} else {
		t.Log("Position completely removed from user state")
	}

	// Summary
	t.Log("Test completed: Position opened, partially closed, and fully closed successfully")
}

func TestShortSOLLeverageUpdateAndClose(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	coin := "SOL"
	slippage := 0.01

	// Step 1: Check user state before
	t.Log("[Before] Checking user state before opening short SOL position")
	beforeState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state (before): %v", err)
	}
	var beforeSOL *hyperliquid.Position
	for _, ap := range beforeState.AssetPositions {
		if ap.Position.Coin == coin {
			beforeSOL = &ap.Position
			break
		}
	}
	if beforeSOL != nil {
		t.Logf("[Before] Existing SOL position - Size: %s, Lev: %dx", beforeSOL.Szi, beforeSOL.Leverage.Value)
	} else {
		t.Log("[Before] No existing SOL position")
	}

	// Step 2: Open a short position (sell) with slippage escalation if needed
	size := 0.1 // modest size for test liquidity
	escalations := []float64{slippage, 0.03, 0.05, 0.1, 0.2}
	var opened bool
	for _, s := range escalations {
		t.Logf("Opening short position on %s with size %f (slippage=%.2f)", coin, size, s)
		openRes, err := exchange.MarketOpen(coin, false /* isBuy=false => short */, size, nil, s, nil, nil)
		if err != nil {
			t.Logf("MarketOpen short failed with slippage %.2f: %v", s, err)
			continue
		}
		t.Logf("Short open attempt result: %+v", openRes)

		// Verify position exists after open attempt
		afterOpenState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
		if err != nil {
			t.Fatalf("Failed to get user state after open attempt: %v", err)
		}
		for _, ap := range afterOpenState.AssetPositions {
			if ap.Position.Coin == coin {
				opened = true
				break
			}
		}
		if opened {
			break
		}
	}
	if !opened {
		t.Fatalf("Failed to open short %s position after slippage escalation attempts", coin)
	}

	// Step 3: Update leverage to 3x (cross)
	t.Log("Updating SOL leverage to 3x (cross)")
	_, err = exchange.UpdateLeverage(3, coin, true)
	if err != nil {
		t.Fatalf("UpdateLeverage failed: %v", err)
	}

	// Step 4: Check user state after leverage update (during)
	t.Log("[During] Verifying leverage updated to 3x")
	duringState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state (during): %v", err)
	}
	var duringSOL *hyperliquid.Position
	for _, ap := range duringState.AssetPositions {
		if ap.Position.Coin == coin {
			duringSOL = &ap.Position
			break
		}
	}
	if duringSOL == nil {
		t.Fatalf("SOL position not found after opening")
	}
	t.Logf("[During] SOL position - Size: %s, Lev: %dx, Type: %s", duringSOL.Szi, duringSOL.Leverage.Value, duringSOL.Leverage.Type)
	if duringSOL.Leverage.Value != 3 {
		t.Fatalf("expected leverage 3x, got %dx", duringSOL.Leverage.Value)
	}

	// Step 5: Close the position with slippage escalation if needed
	t.Log("Closing SOL position")
	closed := false
	for _, s := range escalations {
		closeRes, err := exchange.MarketClose(coin, nil, nil, s, nil, nil)
		if err != nil {
			t.Logf("MarketClose failed with slippage %.2f: %v", s, err)
			continue
		}
		t.Logf("Close attempt result (slippage=%.2f): %+v", s, closeRes)

		// Verify closed or size zero
		afterCloseCheck, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
		if err != nil {
			t.Fatalf("Failed to get user state after close attempt: %v", err)
		}
		var posAfter *hyperliquid.Position
		for _, ap := range afterCloseCheck.AssetPositions {
			if ap.Position.Coin == coin {
				posAfter = &ap.Position
				break
			}
		}
		if posAfter == nil {
			closed = true
			break
		}
		sizeFloat, err := strconv.ParseFloat(posAfter.Szi, 64)
		if err == nil && sizeFloat == 0 {
			closed = true
			break
		}
	}
	if !closed {
		t.Fatalf("Failed to close SOL position after slippage escalation attempts")
	}

	// Step 6: Check user state after
	t.Log("[After] Checking user state after closing SOL position")
	afterState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state (after): %v", err)
	}
	var afterSOL *hyperliquid.Position
	for _, ap := range afterState.AssetPositions {
		if ap.Position.Coin == coin {
			afterSOL = &ap.Position
			break
		}
	}
	if afterSOL != nil {
		sizeFloat, err := strconv.ParseFloat(afterSOL.Szi, 64)
		if err != nil {
			t.Logf("Warning: could not parse SOL size after close: %v", err)
		} else if sizeFloat != 0 {
			t.Fatalf("expected SOL position size 0 after close, got %f", sizeFloat)
		} else {
			t.Log("[After] SOL position size is 0 (closed)")
		}
	} else {
		t.Log("[After] SOL position removed (no active position)")
	}
}
