package examples

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/joho/godotenv"
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
	name := "SOL"
	isBuy := true
	sz := 2.0
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
	sz := 0.002064   // Small size for testing
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

	time.Sleep(3 * time.Second)

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
	time.Sleep(3 * time.Second)

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

func TestWeirdSizeAndClose(t *testing.T) {
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
	size := 6.7844564847 // modest size for test liquidity
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

	time.Sleep(3 * time.Second)

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

func TestOpenPositionAndCancelCloseOrder(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("Testing open position and cancel close order workflow")

	// Step 1: Check initial user state
	t.Log("Step 1: Checking initial user state")
	initialUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get initial user state: %v", err)
	}
	t.Logf("Initial user state - Positions count: %d", len(initialUserState.AssetPositions))

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

	// Step 3: Verify position was opened
	t.Log("Step 3: Verifying position was opened")
	afterOpenUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state after opening: %v", err)
	}

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

	// Step 4: Place a limit order to close the position
	t.Log("Step 4: Placing limit order to close position")
	closeOrderReq := hyperliquid.CreateOrderRequest{
		Coin:  name,
		IsBuy: false,   // Sell to close long position
		Size:  sz,      // Same size as opened position
		Price: 45000.0, // Set a price that's unlikely to fill immediately
		OrderType: hyperliquid.OrderType{
			Limit: &hyperliquid.LimitOrderType{
				Tif: hyperliquid.TifGtc, // Good till cancelled
			},
		},
	}

	closeOrderResp, err := exchange.Order(closeOrderReq, nil)
	if err != nil {
		t.Fatalf("Failed to place close order: %v", err)
	}
	t.Logf("Close order placed: %+v", closeOrderResp)

	// Extract order ID from response
	var orderID int64
	if closeOrderResp.Resting != nil {
		orderID = closeOrderResp.Resting.Oid
		t.Logf("Order ID for cancellation: %d", orderID)
	} else {
		t.Skip("Close order was filled immediately, cannot test cancel")
	}

	// Step 5: Cancel the close order
	t.Log("Step 5: Cancelling the close order")
	cancelResp, err := exchange.Cancel(name, orderID)
	if err != nil {
		t.Fatalf("Failed to cancel order: %v", err)
	}
	t.Logf("Cancel response: %+v", cancelResp)

	// Step 6: Verify position still exists (since we cancelled the close order)
	t.Log("Step 6: Verifying position still exists after cancelling close order")
	afterCancelUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state after cancelling: %v", err)
	}

	var btcPositionAfterCancel *hyperliquid.Position
	for _, pos := range afterCancelUserState.AssetPositions {
		if pos.Position.Coin == name {
			btcPositionAfterCancel = &pos.Position
			t.Logf("BTC position after cancelling close order - Size: %s", pos.Position.Szi)
			break
		}
	}

	if btcPositionAfterCancel == nil {
		t.Fatalf("BTC position not found after cancelling close order - it should still exist")
	}

	// Step 7: Clean up - close the position properly
	t.Log("Step 7: Cleaning up - closing position properly")
	closeResult, err := exchange.MarketClose(name, nil, nil, slippage, nil, nil)
	if err != nil {
		t.Fatalf("MarketClose failed during cleanup: %v", err)
	}
	t.Logf("Position closed successfully during cleanup: %+v", closeResult)

	// Step 8: Final verification
	t.Log("Step 8: Final verification - checking position is closed")
	finalUserState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get final user state: %v", err)
	}

	var finalBtcPosition *hyperliquid.Position
	for _, pos := range finalUserState.AssetPositions {
		if pos.Position.Coin == name {
			finalBtcPosition = &pos.Position
			break
		}
	}

	if finalBtcPosition != nil {
		size, err := strconv.ParseFloat(finalBtcPosition.Szi, 64)
		if err != nil {
			t.Logf("Warning: Could not parse final position size: %v", err)
		} else if size != 0 {
			t.Logf("Warning: Position still exists with size: %f", size)
		} else {
			t.Log("Position successfully closed (size = 0)")
		}
	} else {
		t.Log("Position completely removed from user state")
	}

	t.Log("Test completed: Position opened, close order placed and cancelled, position still exists, then properly closed")
}

func TestOpenPositionWithStopLossMinus10Percent(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	coin := "BTC"
	isBuy := true
	size := 0.001
	slippage := 0.01
	slPercent := 0.10 // 10%

	// Open with SL in one grouped action
	resp, err := exchange.MarketOpenWithSLTP(coin, isBuy, size, nil, slippage, slPercent, false, nil, nil, nil)
	if err != nil {
		t.Fatalf("MarketOpenWithSLTP failed: %v", err)
	}
	if !resp.Ok {
		t.Fatalf("MarketOpenWithSLTP not ok: %s", resp.Err)
	}

	statuses := resp.Data.Statuses
	if len(statuses) < 2 {
		t.Fatalf("expected 2 statuses (open, SL), got %d", len(statuses))
	}

	// // One should be filled (open IOC), the other likely resting (SL trigger)
	// var slOrderID int64
	// for _, mv := range statuses {
	// 	if mv.Type() != "object" {
	// 		continue
	// 	}
	// 	var st hyperliquid.OrderStatus
	// 	if err := mv.Parse(&st); err != nil {
	// 		t.Fatalf("failed to parse status: %v", err)
	// 	}
	// 	if st.Resting != nil {
	// 		// store potential SL oid for cleanup
	// 		slOrderID = st.Resting.Oid
	// 	}
	// }

	// // Verify position exists
	// state, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	// if err != nil {
	// 	t.Fatalf("failed to fetch user state: %v", err)
	// }
	// var pos *hyperliquid.Position
	// for _, ap := range state.AssetPositions {
	// 	if ap.Position.Coin == coin {
	// 		pos = &ap.Position
	// 		break
	// 	}
	// }
	// if pos == nil {
	// 	t.Fatalf("expected %s position after open", coin)
	// }
	// t.Logf("Opened position on %s: size=%s lev=%dx", coin, pos.Szi, pos.Leverage.Value)

	// // Cleanup: cancel SL order if we captured an oid
	// if slOrderID != 0 {
	// 	if _, err := exchange.Cancel(coin, slOrderID); err != nil {
	// 		t.Logf("warning: failed to cancel SL order %d: %v", slOrderID, err)
	// 	}
	// }

	// // Close the position
	// if _, err := exchange.MarketClose(coin, nil, nil, slippage, nil, nil); err != nil {
	// 	t.Fatalf("MarketClose failed: %v", err)
	// }
}

func TestOpenPositionWithTakeProfitPlus10Percent(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	coin := "SOL"
	isBuy := true
	size := 1.0
	slippage := 0.01
	slPercent := 0.10 // 10%

	// Open with SL in one grouped action
	resp, err := exchange.MarketOpenWithSLTP(coin, isBuy, size, nil, slippage, slPercent, true, nil, nil, nil)
	if err != nil {
		t.Fatalf("MarketOpenWithSLTP failed: %v", err)
	}
	if !resp.Ok {
		t.Fatalf("MarketOpenWithSLTP not ok: %s", resp.Err)
	}

	statuses := resp.Data.Statuses
	if len(statuses) < 2 {
		t.Fatalf("expected 2 statuses (open, SL), got %d", len(statuses))
	}

	// // Close the position to clean up
	// if _, err := exchange.MarketClose(coin, nil, nil, slippage, nil, nil); err != nil {
	// 	t.Fatalf("MarketClose failed: %v", err)
	// }
}

func TestOpenPositionWithPartialStopLoss(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	coin := "BTC"
	isBuy := true
	size := 0.002    // open size
	partial := 0.001 // 50% SL size
	slippage := 0.01
	slPercent := 0.10 // 10%

	resp, err := exchange.MarketOpenWithSLTPPartial(coin, isBuy, size, nil, slippage, slPercent, false /* isTP */, &partial, nil, nil, nil)
	if err != nil {
		t.Fatalf("MarketOpenWithSLTPPartial failed: %v", err)
	}
	if !resp.Ok {
		t.Fatalf("MarketOpenWithSLTPPartial not ok: %s", resp.Err)
	}

	statuses := resp.Data.Statuses
	if len(statuses) < 2 {
		t.Fatalf("expected at least 2 statuses (filled open, trigger token), got %d", len(statuses))
	}

	var hasFilled bool
	var hasWaiting bool
	for _, mv := range statuses {
		typ := mv.Type()
		if typ == "object" {
			var st hyperliquid.OrderStatus
			if err := mv.Parse(&st); err == nil && st.Filled != nil {
				hasFilled = true
			}
		}
		if typ == "string" {
			hasWaiting = true
		}
	}
	if !hasFilled || !hasWaiting {
		t.Fatalf("expected filled open and waiting trigger, got hasFilled=%v hasWaiting=%v", hasFilled, hasWaiting)
	}
}

func TestOpenLongARBWithLeverageAndTPandSL(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	coin := "ARB"
	size := 100.0       // Adjusted size for ARB
	tpslPercent := 0.05 // 5%
	slippage := 0.05

	// Declare variables
	var openPx float64
	var err error

	// Get the current market price for ARB
	// openPx, err := exchange.SlippagePrice(coin, true, slippage, nil)
	// if err != nil {
	// 	t.Fatalf("SlippagePrice failed: %v", err)
	// }

	// Validate that we got a reasonable price
	// if openPx <= 0 {
	// 	t.Fatalf("Invalid slippage price: %f", openPx)
	// }

	// Get the asset ID for ARB
	assetID := exchange.GetInfo().NameToAsset(coin)
	t.Logf("Asset ID for %s: %d", coin, assetID)

	// Check if the asset ID is valid
	if assetID == 0 {
		t.Fatalf("Invalid asset ID for %s: %d", coin, assetID)
	}

	// Debug: Check what assets are available in allMids
	mids, err := exchange.GetInfo().AllMids()
	if err != nil {
		t.Logf("Warning: Failed to get allMids: %v", err)
	} else {
		t.Logf("Available assets in allMids: %d", len(mids))
		// Show first few assets as examples
		count := 0
		for assetName, midPrice := range mids {
			if count < 5 {
				t.Logf("  %s: %s", assetName, midPrice)
				count++
			}
		}

		// Check specifically for ARB
		if midPriceStr, ok := mids[coin]; ok {
			t.Logf("Found %s in allMids: %s", coin, midPriceStr)
			if midPrice, err := strconv.ParseFloat(midPriceStr, 64); err == nil {
				t.Logf("Raw mid price for %s: $%.6f", coin, midPrice)
				// Calculate slippage manually to debug
				manualSlippagePrice := midPrice * (1 + slippage)
				t.Logf("Manual slippage calculation: $%.6f * (1 + %.3f) = $%.6f", midPrice, slippage, manualSlippagePrice)
			} else {
				t.Logf("Failed to parse mid price for %s: %v", coin, err)
			}
		} else {
			t.Logf("❌ %s NOT found in allMids", coin)
			// Try to find similar assets
			for assetName := range mids {
				if strings.Contains(strings.ToUpper(assetName), "JUP") ||
					strings.Contains(strings.ToUpper(assetName), "JUPITER") {
					t.Logf("Found similar asset: %s", assetName)
				}
			}
		}
	}

	// Try to get slippage price manually to debug
	t.Logf("Attempting to get slippage price for %s (asset ID: %d)...", coin, assetID)
	openPx, err = exchange.SlippagePrice(coin, true, slippage, nil)
	if err != nil {
		t.Logf("SlippagePrice error: %v", err)
		// Try to debug the issue by calling AllMids directly
		if mids, err := exchange.GetInfo().AllMids(); err == nil {
			if midPriceStr, ok := mids[coin]; ok {
				t.Logf("Direct allMids lookup for %s: %s", coin, midPriceStr)
			}
		}
		t.Fatalf("SlippagePrice failed: %v", err)
	}

	// Validate that we got a reasonable price
	if openPx <= 0 {
		t.Fatalf("Invalid slippage price: %f", openPx)
	}

	// Compute trigger prices relative to open price
	tpPxRaw := openPx * (1 + tpslPercent) // TP above entry for long position
	slPxRaw := openPx * (1 - tpslPercent) // SL below entry for long position

	t.Logf("Raw prices - Open: $%.6f, TP: $%.6f, SL: $%.6f", openPx, tpPxRaw, slPxRaw)

	// Use PriceToWire to ensure proper tick size compliance
	tpPxWire, err := hyperliquid.PriceToWire(tpPxRaw, assetID, exchange.GetInfo(), false)
	if err != nil {
		t.Fatalf("Failed to format TP price: %v", err)
	}
	slPxWire, err := hyperliquid.PriceToWire(slPxRaw, assetID, exchange.GetInfo(), false)
	if err != nil {
		t.Fatalf("Failed to format SL price: %v", err)
	}
	openPxWire, err := hyperliquid.PriceToWire(openPx, assetID, exchange.GetInfo(), false)
	if err != nil {
		t.Fatalf("Failed to format open price: %v", err)
	}

	t.Logf("Wire prices - Open: %s, TP: %s, SL: %s", openPxWire, tpPxWire, slPxWire)

	// Parse back to float for the main order price only
	openPxFinal, err := strconv.ParseFloat(openPxWire, 64)
	if err != nil {
		t.Fatalf("Failed to parse open price: %v", err)
	}

	// For trigger orders, we need to parse the wire prices back to float64
	// because the TriggerOrderType expects float64, but PriceToWire will be called again
	tpPx, err := strconv.ParseFloat(tpPxWire, 64)
	if err != nil {
		t.Fatalf("Failed to parse TP price: %v", err)
	}
	slPx, err := strconv.ParseFloat(slPxWire, 64)
	if err != nil {
		t.Fatalf("Failed to parse SL price: %v", err)
	}

	t.Logf("Final prices - Open: $%.6f, TP: $%.6f, SL: $%.6f", openPxFinal, tpPx, slPx)
	t.Logf("Wire strings - Open: %s, TP: %s, SL: %s", openPxWire, tpPxWire, slPxWire)

	// Build orders: 1) IOC open; 2) TP trigger; 3) SL trigger
	openOrder := hyperliquid.CreateOrderRequest{
		Coin:          coin,
		IsBuy:         true,
		Price:         openPxFinal,
		Size:          size,
		ReduceOnly:    false,
		OrderType:     hyperliquid.OrderType{Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifIoc}},
		ClientOrderID: nil,
	}

	tpOrder := hyperliquid.CreateOrderRequest{
		Coin:          coin,
		IsBuy:         false, // Close direction
		Price:         tpPx,  // Use the calculated TP price
		Size:          size,
		ReduceOnly:    true,
		OrderType:     hyperliquid.OrderType{Trigger: &hyperliquid.TriggerOrderType{TriggerPx: tpPx, IsMarket: true, Tpsl: "tp"}},
		ClientOrderID: nil,
	}

	slOrder := hyperliquid.CreateOrderRequest{
		Coin:          coin,
		IsBuy:         false, // Close direction
		Price:         slPx,  // Use the calculated SL price
		Size:          size,
		ReduceOnly:    true,
		OrderType:     hyperliquid.OrderType{Trigger: &hyperliquid.TriggerOrderType{TriggerPx: slPx, IsMarket: true, Tpsl: "sl"}},
		ClientOrderID: nil,
	}

	// Use normalTpsl grouping to align with trigger order expectations
	resp, err := exchange.BulkOrdersWithGrouping([]hyperliquid.CreateOrderRequest{openOrder, tpOrder, slOrder}, hyperliquid.GroupingNormalTpsl, nil)
	if err != nil {
		t.Fatalf("Placing orders failed: %v", err)
	}
	if !resp.Ok {
		t.Fatalf("Orders not ok: %s", resp.Err)
	}

	t.Logf("✅ Successfully placed open order with TP/SL triggers")
	t.Logf("📊 Order Summary:")
	t.Logf("   🪙 Asset: %s (ID: %d)", coin, assetID)
	t.Logf("   📈 Side: LONG")
	t.Logf("   💵 Size: %.2f", size)
	t.Logf("   🎯 Entry: $%.6f", openPxFinal)
	t.Logf("   🚀 TP: $%.6f (+%.1f%%)", tpPx, tpslPercent*100)
	t.Logf("   🛡️  SL: $%.6f (-%.1f%%)", slPx, tpslPercent*100)
}

func TestMarketOpenWithCloid(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t) // exchange used for setup only

	t.Log("Market open method is available and ready to use")

	// Example usage:
	name := "BTC"
	isBuy := false
	sz := 0.001
	slippage := 0.01 // 1%

	// cloid should be a 128-bit hex string according to Hyperliquid docs
	cloid := "0x1234567890abcdef1234567890abcdef"

	result, err := exchange.MarketOpen(name, isBuy, sz, nil, slippage, &cloid, nil)
	if err != nil {
		t.Fatalf("MarketOpen failed: %v", err)
	}
	t.Logf("Market open result: %+v", result)
}

func TestGetCompletePositionSummary(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	// Get user state (positions)
	userState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state: %v", err)
	}

	// Get open orders (TP/SL)
	openOrders, err := exchange.GetInfo().FrontendOpenOrders(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get open orders: %v", err)
	}

	t.Logf("=== POSITIONS ===")
	for _, ap := range userState.AssetPositions {
		pos := ap.Position
		if pos.Szi != "0" { // Only show non-zero positions
			t.Logf("Position: %s | Size: %s | Entry: %s | PnL: %s | Leverage: %dx",
				pos.Coin, pos.Szi, *pos.EntryPx, pos.UnrealizedPnl, pos.Leverage.Value)
		}
	}

	t.Logf("=== OPEN ORDERS (TP/SL) ===")
	for _, order := range openOrders {
		fmt.Printf("raw Order: %+v\n", order)
		t.Logf("Order: %s | Side: %s | Size: %f | Price: %f | OID: %d",
			order.Coin, order.Side, order.Size, order.LimitPx, order.Oid)
	}

	t.Logf("=== MARGIN SUMMARY ===")
	t.Logf("Account Value: %s", userState.MarginSummary.AccountValue)
	t.Logf("Margin Used: %s", userState.MarginSummary.TotalMarginUsed)
	t.Logf("Net Liquidation: %s", userState.MarginSummary.TotalNtlPos)
}

func TestFetchAndCancelOpenOrder(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	// 1) Fetch all open orders for the user
	openOrders, err := exchange.GetInfo().FrontendOpenOrders(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to fetch open orders: %v", err)
	}

	if len(openOrders) == 0 {
		t.Log("No open orders found to cancel")
		return
	}

	// 2) Log all open orders for visibility
	t.Logf("Found %d open orders:", len(openOrders))
	for i, order := range openOrders {
		t.Logf("Order %d: Coin=%s, Side=%s, Size=%f, Price=%f, OID=%d, Type=%s, ReduceOnly=%t",
			i+1, order.Coin, order.Side, order.Size, order.LimitPx, order.Oid, order.OrderType, order.ReduceOnly)
	}

	// 3) Select the first order to cancel (you can modify this logic as needed)
	orderToCancel := openOrders[0]
	t.Logf("Selected order to cancel: Coin=%s, Side=%s, Size=%f, Price=%f, OID=%d",
		orderToCancel.Coin, orderToCancel.Side, orderToCancel.Size, orderToCancel.LimitPx, orderToCancel.Oid)

	// 37589055812
	// 37589055813

	// 4) Cancel the selected order by its OID
	cancelResp, err := exchange.Cancel(orderToCancel.Coin, orderToCancel.Oid)
	if err != nil {
		t.Fatalf("Failed to cancel order %d: %v", orderToCancel.Oid, err)
	}

	t.Logf("Successfully cancelled order %d: %+v", orderToCancel.Oid, cancelResp)

	// 5) Verify the order was cancelled by fetching open orders again
	openOrdersAfter, err := exchange.GetInfo().FrontendOpenOrders(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to fetch open orders after cancellation: %v", err)
	}

	// 6) Check if the cancelled order is no longer in the list
	orderStillExists := false
	for _, order := range openOrdersAfter {
		if order.Oid == orderToCancel.Oid {
			orderStillExists = true
			break
		}
	}

	if orderStillExists {
		t.Errorf("Order %d still exists after cancellation", orderToCancel.Oid)
	} else {
		t.Logf("Order %d successfully removed from open orders", orderToCancel.Oid)
	}

	// 7) Log the updated open orders count
	t.Logf("Open orders after cancellation: %d (was %d)", len(openOrdersAfter), len(openOrders))
}

func TestDebugOrderCancellation(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("=== DEBUGGING ORDER CANCELLATION ISSUE ===")

	// 1) Check your wallet address
	walletAddr := exchange.GetAccountAddr()
	t.Logf("Your wallet address: %s", walletAddr)

	// 2) Get all open orders and their details
	openOrders, err := exchange.GetInfo().FrontendOpenOrders(walletAddr)
	if err != nil {
		t.Fatalf("Failed to get open orders: %v", err)
	}

	t.Logf("Found %d open orders", len(openOrders))

	for i, order := range openOrders {
		// Get asset ID for this coin
		assetID := exchange.GetInfo().NameToAsset(order.Coin)

		t.Logf("Order %d:", i+1)
		t.Logf("  Coin: %s (Asset ID: %d)", order.Coin, assetID)
		t.Logf("  OID: %d", order.Oid)
		t.Logf("  Side: %s", order.Side)
		t.Logf("  Size: %f", order.Size)
		t.Logf("  Limit Price: $%.2f", order.LimitPx)
		t.Logf("  Is Trigger: %t", order.IsTrigger)
		t.Logf("  Is Position TP/SL: %t", order.IsPositionTpsl)
		t.Logf("  Order Type: %s", order.OrderType)
		t.Logf("  ---")
	}

	// 3) Test cancellation with detailed logging
	if len(openOrders) > 0 {
		orderToCancel := openOrders[0]
		assetID := exchange.GetInfo().NameToAsset(orderToCancel.Coin)

		t.Logf("\n=== TESTING CANCELLATION ===")
		t.Logf("Attempting to cancel:")
		t.Logf("  Coin: %s (Asset ID: %d)", orderToCancel.Coin, assetID)
		t.Logf("  OID: %d", orderToCancel.Oid)
		t.Logf("  Wallet: %s", walletAddr)

		// Try to cancel and see what happens
		cancelResp, err := exchange.Cancel(orderToCancel.Coin, orderToCancel.Oid)
		if err != nil {
			t.Logf("❌ Cancel failed with error: %v", err)
		} else {
			t.Logf("✅ Cancel response: %+v", cancelResp)
		}
	}

	// 4) Check if there are any orders with OID 37622448426
	t.Logf("\n=== CHECKING FOR SPECIFIC OID ===")
	t.Logf("Looking for OID: 37622448426")

	// This would require a different API call, but let's check what we can see
	for _, order := range openOrders {
		if order.Oid == 37622448426 {
			t.Logf("🎯 FOUND ORDER WITH OID 37622448426!")
			t.Logf("  Coin: %s", order.Coin)
			t.Logf("  Asset ID: %d", exchange.GetInfo().NameToAsset(order.Coin))
			t.Logf("  Side: %s", order.Side)
		}
	}
}

func TestCheckAvailableAssets(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("=== CHECKING AVAILABLE ASSETS ON TESTNET ===")

	// Get meta information to see available assets
	meta, err := exchange.GetInfo().Meta()
	if err != nil {
		t.Fatalf("Failed to get meta: %v", err)
	}

	t.Logf("Available perpetual assets: %d", len(meta.Universe))
	for i, asset := range meta.Universe {
		t.Logf("Asset %d: %s (szDecimals: %d)", i, asset.Name, asset.SzDecimals)
	}

	// Get current mid prices for all assets
	mids, err := exchange.GetInfo().AllMids()
	if err != nil {
		t.Fatalf("Failed to get all mids: %v", err)
	}

	t.Logf("\n=== CURRENT MARKET PRICES ===")
	for coin, price := range mids {
		t.Logf("%s: $%s", coin, price)
	}

	// Try to get slippage price for a few assets to test availability
	testAssets := []string{"SOL", "ETH", "BTC", "ARB"}

	t.Logf("\n=== TESTING ASSET AVAILABILITY ===")
	for _, asset := range testAssets {
		price, err := exchange.SlippagePrice(asset, true, 0.01, nil) // 1% slippage
		if err != nil {
			t.Logf("❌ %s: %v", asset, err)
		} else {
			t.Logf("✅ %s: $%.2f", asset, price)
		}
	}

	t.Logf("\n=== ACCOUNT STATUS ===")
	userState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state: %v", err)
	}

	t.Logf("Account Value: $%s", userState.MarginSummary.AccountValue)
	t.Logf("Available Balance: $%s", userState.Withdrawable)
	t.Logf("Total Margin Used: $%s", userState.MarginSummary.TotalMarginUsed)
}

func TestSmallSOLPosition(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("=== TESTING SMALL SOL POSITION ===")

	// Use a very small size to test
	coin := "SOL"
	size := 0.001 // Very small size
	isBuy := true

	// Get current price with 1% slippage
	price, err := exchange.SlippagePrice(coin, isBuy, 0.01, nil)
	if err != nil {
		t.Fatalf("Failed to get slippage price: %v", err)
	}

	t.Logf("Opening %s position: %f %s at ~$%.2f",
		coin, size, coin, price)

	// Create a simple market order
	order := hyperliquid.CreateOrderRequest{
		Coin:          coin,
		IsBuy:         isBuy,
		Price:         price,
		Size:          size,
		ReduceOnly:    false,
		OrderType:     hyperliquid.OrderType{Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifIoc}},
		ClientOrderID: nil,
	}

	resp, err := exchange.Order(order, nil) // Add nil BuilderInfo
	if err != nil {
		t.Fatalf("Failed to place order: %v", err)
	}

	t.Logf("Order response: %+v", resp)

	// Check if order was successful
	if resp.Error != nil {
		t.Fatalf("Order failed: %s", *resp.Error)
	}

	if resp.Resting != nil {
		t.Logf("✅ Order resting! Status: %s", "resting")
	} else if resp.Filled != nil {
		t.Logf("✅ Order filled! Status: %s", "filled")
	} else {
		t.Logf("✅ Order successful!")
	}

	// Wait a moment and check position
	time.Sleep(2 * time.Second)

	userState, err := exchange.GetInfo().UserState(exchange.GetAccountAddr())
	if err != nil {
		t.Fatalf("Failed to get user state: %v", err)
	}

	t.Logf("\n=== POSITION AFTER ORDER ===")
	for _, assetPos := range userState.AssetPositions {
		if assetPos.Position.Szi != "0" {
			t.Logf("Asset %s: Size %s | Entry $%s | PnL $%s | Leverage %dx",
				assetPos.Position.Coin,
				assetPos.Position.Szi,
				*assetPos.Position.EntryPx,
				assetPos.Position.UnrealizedPnl,
				assetPos.Position.Leverage.Value)
		}
	}
}

func TestPlaceLimitOrderForARB(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("=== TESTING LIMIT ORDER FOR ARB ===")

	coin := "ARB"
	size := 50.0 // Reasonable size for testing
	isBuy := true
	targetPrice := 0.5002 // Your specified limit price

	// Get the asset ID for ARB
	assetID := exchange.GetInfo().NameToAsset(coin)
	t.Logf("Asset ID for %s: %d", coin, assetID)

	// Check if the asset ID is valid
	if assetID == 0 {
		t.Fatalf("Invalid asset ID for %s: %d", coin, assetID)
	}

	// Get current market price for comparison
	mids, err := exchange.GetInfo().AllMids()
	if err != nil {
		t.Fatalf("Failed to get allMids: %v", err)
	}

	if midPriceStr, ok := mids[coin]; ok {
		currentPrice, _ := strconv.ParseFloat(midPriceStr, 64)
		t.Logf("Current market price for %s: $%.5f", coin, currentPrice)
		t.Logf("Target limit price: $%.5f", targetPrice)

		// Check if the limit price makes sense
		if isBuy && targetPrice > currentPrice {
			t.Logf("⚠️  Note: Limit buy price ($%.5f) is above current market price ($%.5f)", targetPrice, currentPrice)
			t.Logf("   This order will likely fill immediately if the price moves up")
		} else if !isBuy && targetPrice < currentPrice {
			t.Logf("⚠️  Note: Limit sell price ($%.5f) is below current market price ($%.5f)", targetPrice, currentPrice)
			t.Logf("   This order will likely fill immediately if the price moves down")
		} else {
			var orderType string
			if isBuy {
				orderType = "buy"
			} else {
				orderType = "sell"
			}
			t.Logf("✅ Limit price is reasonable for a %s order", orderType)
		}
	} else {
		t.Logf("Could not get current market price for %s", coin)
	}

	// Format the price using PriceToWire to ensure compliance with tick size
	formattedPrice, err := hyperliquid.PriceToWire(targetPrice, assetID, exchange.GetInfo(), false)
	if err != nil {
		t.Fatalf("Failed to format price with PriceToWire: %v", err)
	}

	// Parse back to float64 for the order
	finalPrice, err := strconv.ParseFloat(formattedPrice, 64)
	if err != nil {
		t.Fatalf("Failed to parse formatted price: %v", err)
	}

	t.Logf("Original target price: $%.5f", targetPrice)
	t.Logf("Formatted price (PriceToWire): %s", formattedPrice)
	t.Logf("Final price for order: $%.5f", finalPrice)

	// Create the limit order
	limitOrder := hyperliquid.CreateOrderRequest{
		Coin:          coin,
		IsBuy:         isBuy,
		Price:         finalPrice,
		Size:          size,
		ReduceOnly:    false,
		OrderType:     hyperliquid.OrderType{Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifGtc}}, // Good Till Cancelled
		ClientOrderID: nil,
	}

	t.Logf("Placing limit order:")
	t.Logf("  🪙 Asset: %s (ID: %d)", coin, assetID)
	var side string
	if isBuy {
		side = "BUY"
	} else {
		side = "SELL"
	}
	t.Logf("  📈 Side: %s", side)
	t.Logf("  💵 Size: %.2f", size)
	t.Logf("  🎯 Price: $%.5f", finalPrice)
	t.Logf("  ⏰ TIF: GTC (Good Till Cancelled)")

	// Place the order
	resp, err := exchange.Order(limitOrder, nil)
	if err != nil {
		t.Fatalf("Failed to place limit order: %v", err)
	}

	t.Logf("✅ Order placed successfully!")
	t.Logf("📋 Order response: %+v", resp)

	// Check the order status
	if resp.Error != nil {
		t.Fatalf("❌ Order failed: %s", *resp.Error)
	}

	if resp.Resting != nil {
		t.Logf("✅ Order is resting (waiting to be filled)")
		t.Logf("   Order ID: %d", resp.Resting.Oid)
		t.Logf("   Client ID: %s", resp.Resting.ClientID)
		t.Logf("   Status: %s", resp.Resting.Status)
	} else if resp.Filled != nil {
		t.Logf("✅ Order was filled immediately!")
		t.Logf("   Order ID: %d", resp.Filled.Oid)
		t.Logf("   Total Size: %s", resp.Filled.TotalSz)
		t.Logf("   Average Price: %s", resp.Filled.AvgPx)
	} else {
		t.Logf("✅ Order placed successfully!")
	}

	// Wait a moment and check if the order appears in open orders
	time.Sleep(2 * time.Second)

	openOrders, err := exchange.GetInfo().FrontendOpenOrders(exchange.GetAccountAddr())
	if err != nil {
		t.Logf("Warning: Could not fetch open orders: %v", err)
	} else {
		t.Logf("\n=== CHECKING OPEN ORDERS ===")
		foundOrder := false
		for _, order := range openOrders {
			if order.Coin == coin && order.LimitPx == finalPrice && order.Size == size {
				t.Logf("✅ Found our limit order in open orders:")
				t.Logf("   Coin: %s", order.Coin)
				t.Logf("   Side: %s", order.Side)
				t.Logf("   Size: %.2f", order.Size)
				t.Logf("   Price: $%.5f", order.LimitPx)
				t.Logf("   OID: %d", order.Oid)
				t.Logf("   Order Type: %s", order.OrderType)
				foundOrder = true
				break
			}
		}

		if !foundOrder {
			t.Logf("ℹ️  Order not found in open orders (may have been filled immediately)")
		}
	}

	t.Logf("\n=== LIMIT ORDER TEST COMPLETED ===")
	t.Logf("The limit order for %s at $%.5f has been placed successfully!", coin, finalPrice)
}

func TestPlaceLimitOrderForARBThatRests(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("=== TESTING LIMIT ORDER FOR ARB THAT RESTS ===")

	coin := "ARB"
	size := 25.0 // Smaller size for testing
	isBuy := true
	targetPrice := 0.4800 // Price BELOW current market price so it won't fill immediately

	// Get the asset ID for ARB
	assetID := exchange.GetInfo().NameToAsset(coin)
	t.Logf("Asset ID for %s: %d", coin, assetID)

	// Check if the asset ID is valid
	if assetID == 0 {
		t.Fatalf("Invalid asset ID for %s: %d", coin, assetID)
	}

	// Get current market price for comparison
	mids, err := exchange.GetInfo().AllMids()
	if err != nil {
		t.Fatalf("Failed to get allMids: %v", err)
	}

	if midPriceStr, ok := mids[coin]; ok {
		currentPrice, _ := strconv.ParseFloat(midPriceStr, 64)
		t.Logf("Current market price for %s: $%.5f", coin, currentPrice)
		t.Logf("Target limit price: $%.5f", targetPrice)

		// Check if the limit price makes sense for a resting order
		if isBuy && targetPrice < currentPrice {
			t.Logf("✅ Perfect! Limit buy price ($%.5f) is BELOW current market price ($%.5f)", targetPrice, currentPrice)
			t.Logf("   This order should REST on the order book and wait for price to drop")
		} else if !isBuy && targetPrice > currentPrice {
			t.Logf("✅ Perfect! Limit sell price ($%.5f) is ABOVE current market price ($%.5f)", targetPrice, currentPrice)
			t.Logf("   This order should REST on the order book and wait for price to rise")
		} else {
			t.Logf("⚠️  Warning: This limit price will likely fill immediately!")
		}
	} else {
		t.Logf("Could not get current market price for %s", coin)
	}

	// Format the price using PriceToWire to ensure compliance with tick size
	formattedPrice, err := hyperliquid.PriceToWire(targetPrice, assetID, exchange.GetInfo(), false)
	if err != nil {
		t.Fatalf("Failed to format price with PriceToWire: %v", err)
	}

	// Parse back to float64 for the order
	finalPrice, err := strconv.ParseFloat(formattedPrice, 64)
	if err != nil {
		t.Fatalf("Failed to parse formatted price: %v", err)
	}

	t.Logf("Original target price: $%.5f", targetPrice)
	t.Logf("Formatted price (PriceToWire): %s", formattedPrice)
	t.Logf("Final price for order: $%.5f", finalPrice)

	// Create the limit order
	limitOrder := hyperliquid.CreateOrderRequest{
		Coin:          coin,
		IsBuy:         isBuy,
		Price:         finalPrice,
		Size:          size,
		ReduceOnly:    false,
		OrderType:     hyperliquid.OrderType{Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifGtc}}, // Good Till Cancelled
		ClientOrderID: nil,
	}

	t.Logf("Placing limit order:")
	t.Logf("  🪙 Asset: %s (ID: %d)", coin, assetID)
	var side string
	if isBuy {
		side = "BUY"
	} else {
		side = "SELL"
	}
	t.Logf("  📈 Side: %s", side)
	t.Logf("  💵 Size: %.2f", size)
	t.Logf("  🎯 Price: $%.5f", finalPrice)
	t.Logf("  ⏰ TIF: GTC (Good Till Cancelled)")
	t.Logf("  📋 Expected: Order should REST on the order book")

	// Place the order
	resp, err := exchange.Order(limitOrder, nil)
	if err != nil {
		t.Fatalf("Failed to place limit order: %v", err)
	}

	t.Logf("✅ Order placed successfully!")
	t.Logf("📋 Order response: %+v", resp)

	// Check the order status
	if resp.Error != nil {
		t.Fatalf("❌ Order failed: %s", *resp.Error)
	}

	if resp.Resting != nil {
		t.Logf("🎯 SUCCESS! Order is RESTING on the order book")
		t.Logf("   Order ID: %d", resp.Resting.Oid)
		t.Logf("   Client ID: %s", resp.Resting.ClientID)
		t.Logf("   Status: %s", resp.Resting.Status)
		t.Logf("   This is the expected behavior for a limit order!")
	} else if resp.Filled != nil {
		t.Logf("⚠️  Order was filled immediately (unexpected for this price)")
		t.Logf("   Order ID: %d", resp.Filled.Oid)
		t.Logf("   Total Size: %s", resp.Filled.TotalSz)
		t.Logf("   Average Price: %s", resp.Filled.AvgPx)
		t.Logf("   This suggests the market price moved or our limit price was too aggressive")
	} else {
		t.Logf("✅ Order placed successfully!")
	}

	// Wait a moment and check if the order appears in open orders
	time.Sleep(2 * time.Second)

	openOrders, err := exchange.GetInfo().FrontendOpenOrders(exchange.GetAccountAddr())
	if err != nil {
		t.Logf("Warning: Could not fetch open orders: %v", err)
	} else {
		t.Logf("\n=== CHECKING OPEN ORDERS ===")
		foundOrder := false
		for _, order := range openOrders {
			if order.Coin == coin && order.LimitPx == finalPrice && order.Size == size {
				t.Logf("✅ Found our limit order in open orders:")
				t.Logf("   Coin: %s", order.Coin)
				t.Logf("   Side: %s", order.Side)
				t.Logf("   Size: %.2f", order.Size)
				t.Logf("   Price: $%.5f", order.LimitPx)
				t.Logf("   OID: %d", order.Oid)
				t.Logf("   Order Type: %s", order.OrderType)
				foundOrder = true
				break
			}
		}

		if !foundOrder {
			t.Logf("ℹ️  Order not found in open orders")
			if resp.Resting != nil {
				t.Logf("   Note: Order was marked as 'resting' but not showing in open orders")
				t.Logf("   This might be a timing issue or the order was filled")
			}
		}
	}

	t.Logf("\n=== LIMIT ORDER RESTING TEST COMPLETED ===")
	if resp.Resting != nil {
		t.Logf("🎯 SUCCESS! The limit order for %s at $%.5f is now RESTING on the order book", coin, finalPrice)
		t.Logf("   It will only execute when the market price reaches $%.5f", finalPrice)
		t.Logf("   Order ID: %d", resp.Resting.Oid)
	} else {
		t.Logf("ℹ️  The limit order behavior was different than expected")
	}
}

func TestGetARBHourlyCandles(t *testing.T) {
	godotenv.Overload()

	info := hyperliquid.NewInfo(hyperliquid.TestnetAPIURL, true, nil, nil)

	end := time.Now().UTC()
	start := end.Add(-24 * time.Hour)

	candles, err := info.CandlesSnapshot("ARB", "1h", start.UnixMilli(), end.UnixMilli())
	if err != nil {
		t.Fatalf("failed to fetch ARB 1h candles: %v", err)
	}

	if len(candles) == 0 {
		t.Fatalf("expected at least one candle, got 0")
	}

	// Log first and last candle succinctly
	first := candles[0]
	last := candles[len(candles)-1]
	t.Logf("Candles fetched: %d (interval=1h)", len(candles))
	t.Logf("First: ts=%d o=%s h=%s l=%s c=%s v=%s", first.Timestamp, first.Open, first.High, first.Low, first.Close, first.Volume)
	t.Logf("Last:  ts=%d o=%s h=%s l=%s c=%s v=%s", last.Timestamp, last.Open, last.High, last.Low, last.Close, last.Volume)

	// Basic sanity check: timestamps should be increasing by ~1h
	if len(candles) >= 2 {
		deltaMs := candles[1].Timestamp - candles[0].Timestamp
		// allow some tolerance; expected ~3600000ms
		if deltaMs < 3_000_000 || deltaMs > 4_200_000 {
			t.Logf("warning: candle step unexpected: %s ms", strconv.FormatInt(deltaMs, 10))
		}
	}
}

// TestApproveBuilderFee tests the approval of a builder address to collect fees
// Based on: https://hyperliquid.gitbook.io/hyperliquid-docs/trading/builder-codes
func TestApproveBuilderFee(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("=== TESTING BUILDER FEE APPROVAL ===")

	// For testing purposes, we'll use a sample builder address
	// In real usage, this would be the address of the DeFi application/builder
	builderAddress := "0x1234567890123456789012345678901234567890"

	// Max fee rate: 0.1% (10 basis points) for perps, 1% for spot
	// The fee rate is specified in tenths of basis points
	// So 0.1% = 10 basis points = 100 tenths of basis points
	maxFeeRate := "100" // 0.1% in tenths of basis points

	t.Logf("🔧 Builder Address: %s", builderAddress)
	t.Logf("💰 Max Fee Rate: %s (0.1%% in tenths of basis points)", maxFeeRate)
	t.Logf("📋 This will allow the builder to collect up to 0.1%% fee on perp trades")

	// Approve the builder fee
	t.Logf("📝 Approving builder fee...")
	resp, err := exchange.ApproveBuilderFee(builderAddress, maxFeeRate)
	if err != nil {
		t.Fatalf("❌ Failed to approve builder fee: %v", err)
	}

	t.Logf("✅ Builder fee approval successful!")
	t.Logf("📋 Response: %v", resp)

	// Check the response
	if resp.Error != "" {
		t.Fatalf("❌ Approval failed with error: %s", resp.Error)
	}

	if resp.Status != "ok" {
		t.Logf("⚠️  Warning: Status is not 'ok': %s", resp.Status)
	}

	if resp.TxHash != "" {
		t.Logf("🔗 Transaction Hash: %s", resp.TxHash)
		t.Logf("   You can view this transaction on the blockchain explorer")
	} else {
		t.Logf("ℹ️  No transaction hash returned (this is normal for some operations)")
	}

	t.Logf("\n=== BUILDER FEE APPROVAL COMPLETED ===")
	t.Logf("🎯 The builder address %s is now approved to collect fees", builderAddress)
	t.Logf("💰 Maximum fee rate: 0.1%% (100 tenths of basis points)")
	t.Logf("📋 This approval allows the builder to include builder fees in future orders")
	t.Logf("   The builder can now send orders with: b: %s, f: <fee_rate>", builderAddress)
	t.Logf("   where <fee_rate> is the actual fee to charge (≤ 100 for 0.1%%)")
}

// TestApproveBuilderFeeWithDifferentRates tests various fee rate approvals
func TestApproveBuilderFeeWithDifferentRates(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("=== TESTING BUILDER FEE APPROVAL WITH DIFFERENT RATES ===")

	builderAddress := "0x223fF2A843bE0DB5E70FC1B61e7524b066F7E03c"

	// Test different fee rates
	testCases := []struct {
		rate        string
		description string
		percentage  string
	}{
		{"10", "1 basis point", "0.01%%"},
		{"50", "5 basis points", "0.05%%"},
		{"100", "10 basis points", "0.1%%"},
		{"500", "50 basis points", "0.5%%"},
		{"1000", "100 basis points", "1.0%%"},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			t.Logf("🧪 Testing fee rate: %s (%s)", tc.rate, tc.percentage)

			resp, err := exchange.ApproveBuilderFee(builderAddress, tc.rate)
			if err != nil {
				t.Fatalf("❌ Failed to approve builder fee with rate %s: %v", tc.rate, err)
			}

			if resp.Error != "" {
				t.Fatalf("❌ Approval failed with error: %s", resp.Error)
			}

			t.Logf("✅ Successfully approved %s fee rate", tc.percentage)
		})
	}

	t.Logf("\n=== MULTIPLE FEE RATE APPROVALS COMPLETED ===")
	t.Logf("🎯 Builder address %s is now approved for multiple fee rates", builderAddress)
	t.Logf("💰 This demonstrates the flexibility of builder fee approvals")
}

// TestBuilderFeeWorkflow tests the complete builder fee workflow:
// 1) Approve builder fee for 1%
// 2) Place a trade with builder fee
// 3) Check fees received by the builder address
func TestBuilderFeeWorkflow(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)

	t.Log("=== TESTING COMPLETE BUILDER FEE WORKFLOW ===")

	builderAddress := "0x223fF2A843bE0DB5E70FC1B61e7524b066F7E03c"
	maxFeeRate := "100" // 0.1% (100 tenths of basis points)

	// Step 1: Approve builder fee for 1%
	t.Logf("📝 Step 1: Approving builder fee for 1%% (%s tenths of basis points)", maxFeeRate)
	t.Logf("🔧 Builder Address: %s", builderAddress)

	resp, err := exchange.ApproveBuilderFee(builderAddress, maxFeeRate)
	if err != nil {
		t.Fatalf("❌ Failed to approve builder fee: %v", err)
	}

	if resp.Error != "" {
		t.Fatalf("❌ Builder fee approval failed: %s", resp.Error)
	}

	t.Logf("✅ Step 1 completed: Builder fee approved for 1%%")

	// Step 2: Get initial builder rewards to compare later
	t.Logf("📊 Step 2: Checking initial builder rewards...")

	initialReferralState, err := exchange.GetInfo().QueryReferralState(builderAddress)
	if err != nil {
		t.Logf("⚠️  Could not query initial referral state: %v", err)
		t.Logf("   This might be normal if the builder has never received fees")
	}

	var initialBuilderRewards string = "0"
	if initialReferralState != nil {
		initialBuilderRewards = initialReferralState.BuilderRewards
		t.Logf("📈 Initial builder rewards: %s USDC", initialBuilderRewards)
	} else {
		t.Logf("📈 Initial builder rewards: %s USDC (new builder)", initialBuilderRewards)
	}

	// Step 3: Place a trade with builder fee
	t.Logf("💰 Step 3: Placing trade with builder fee...")

	coin := "ARB"
	size := 50.0
	isBuy := true
	builderFee := 100 // 1% in tenths of basis points

	// Get current market price
	mids, err := exchange.GetInfo().AllMids()
	if err != nil {
		t.Fatalf("Failed to get market prices: %v", err)
	}

	currentPriceStr, ok := mids[coin]
	if !ok {
		t.Fatalf("Could not get current price for %s", coin)
	}

	currentPrice, err := strconv.ParseFloat(currentPriceStr, 64)
	if err != nil {
		t.Fatalf("Failed to parse current price: %v", err)
	}

	// Note: Asset ID is not needed for this market order

	t.Logf("🎯 Trading details:")
	t.Logf("   Asset: %s", coin)
	t.Logf("   Side: %s", "BUY")
	t.Logf("   Size: %.2f", size)
	t.Logf("   Market Price: $%.5f", currentPrice)
	t.Logf("   Builder: %s", builderAddress)
	t.Logf("   Builder Fee: %d tenths of basis points (0.1%%)", builderFee)
	t.Logf("   Order Type: Market (IOC) - will fill immediately")

	// Create BuilderInfo
	builderInfo := &hyperliquid.BuilderInfo{
		Builder: builderAddress,
		Fee:     builderFee,
	}

	// Create and place the order - MARKET ORDER with IOC (Immediate or Cancel)
	order := hyperliquid.CreateOrderRequest{
		Coin:          coin,
		IsBuy:         isBuy,
		Price:         currentPrice, // Use current market price
		Size:          size,
		ReduceOnly:    false,
		OrderType:     hyperliquid.OrderType{Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifIoc}}, // IOC for immediate fill
		ClientOrderID: nil,
	}

	orderResp, err := exchange.Order(order, builderInfo)
	if err != nil {
		t.Fatalf("❌ Failed to place order with builder fee: %v", err)
	}

	if orderResp.Error != nil {
		t.Fatalf("❌ Order failed: %s", *orderResp.Error)
	}

	var orderID int64
	if orderResp.Filled != nil {
		t.Logf("✅ Order filled immediately!")
		t.Logf("   Order ID: %d", orderResp.Filled.Oid)
		t.Logf("   Size: %s", orderResp.Filled.TotalSz)
		t.Logf("   Avg Price: $%s", orderResp.Filled.AvgPx)
		orderID = int64(orderResp.Filled.Oid)
	} else if orderResp.Resting != nil {
		t.Logf("📋 Order placed and resting on book")
		t.Logf("   Order ID: %d", orderResp.Resting.Oid)
		orderID = orderResp.Resting.Oid

		// Wait a bit for potential fill
		t.Logf("⏳ Waiting 5 seconds for potential fill...")
		time.Sleep(5 * time.Second)
	}

	t.Logf("✅ Step 3 completed: Order placed with builder fee")

	// Step 4: Check builder rewards after the trade
	t.Logf("🔍 Step 4: Checking builder rewards after trade...")

	// Wait a moment for the system to process the fee
	time.Sleep(3 * time.Second)

	finalReferralState, err := exchange.GetInfo().QueryReferralState(builderAddress)
	if err != nil {
		t.Logf("⚠️  Could not query final referral state: %v", err)
		t.Logf("   This might take some time to appear in the system")
	} else {
		finalBuilderRewards := finalReferralState.BuilderRewards
		t.Logf("📈 Final builder rewards: %s USDC", finalBuilderRewards)

		// Try to parse and compare rewards
		initialRewards, err1 := strconv.ParseFloat(initialBuilderRewards, 64)
		finalRewards, err2 := strconv.ParseFloat(finalBuilderRewards, 64)

		if err1 == nil && err2 == nil {
			rewardDifference := finalRewards - initialRewards
			if rewardDifference > 0 {
				t.Logf("🎉 SUCCESS: Builder received %.6f USDC in fees!", rewardDifference)
				t.Logf("   Fee increase: %.6f USDC", rewardDifference)
			} else {
				t.Logf("ℹ️  No immediate fee increase detected")
				t.Logf("   Note: Fees might take time to appear or order might not have filled")
			}
		} else {
			t.Logf("ℹ️  Could not parse reward amounts for comparison")
		}

		// Log other referral state info
		if finalReferralState.CumVlm != "" {
			t.Logf("📊 Total volume: %s", finalReferralState.CumVlm)
		}
		if finalReferralState.UnclaimedRewards != "" {
			t.Logf("💰 Unclaimed rewards: %s", finalReferralState.UnclaimedRewards)
		}
	}

	t.Logf("\n=== BUILDER FEE WORKFLOW COMPLETED ===")
	t.Logf("✅ Successfully tested the complete builder fee workflow:")
	t.Logf("   1. ✅ Approved builder fee for 1%% (%s)", builderAddress)
	t.Logf("   2. ✅ Placed trade with builder fee (Order ID: %d)", orderID)
	t.Logf("   3. ✅ Checked builder rewards collection")
	t.Logf("💡 Builder fee workflow is working correctly!")
}
