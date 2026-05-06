package examples

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	hyperliquid "github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/joho/godotenv"
)

// WsOrder represents an order update from the WebSocket
type WsOrder struct {
	Order struct {
		Coin      string `json:"coin"`
		Side      string `json:"side"`
		LimitPx   string `json:"limitPx"`
		Sz        string `json:"sz"`
		Oid       int64  `json:"oid"`
		Timestamp int64  `json:"timestamp"`
		OrigSz    string `json:"origSz"`
		Cloid     string `json:"cloid,omitempty"`
	} `json:"order"`
	Status          string `json:"status"`
	StatusTimestamp int64  `json:"statusTimestamp"`
}

// TestOrderStatusMonitor provides a clean display of order status changes
func TestOrderStatusMonitor(t *testing.T) {
	godotenv.Overload()

	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("            ORDER STATUS MONITOR")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Address: %s", testAddress)
	t.Log("")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	orderCount := 0
	fillCount := 0

	// Subscribe to orderUpdates - shows order status changes
	_, err = wsClient.SubscribeToOrderUpdates(testAddress, func(msg hyperliquid.WSMessage) {
		var orders []WsOrder
		if err := json.Unmarshal(msg.Data, &orders); err != nil {
			t.Logf("⚠️  Failed to parse orderUpdates: %v", err)
			return
		}

		for _, order := range orders {
			orderCount++

			// Determine side symbol
			sideSymbol := "🟢 BUY "
			if order.Order.Side == "A" {
				sideSymbol = "🔴 SELL"
			}

			// Determine status emoji and text
			statusEmoji := "📝"
			statusText := order.Status
			switch order.Status {
			case "open":
				statusEmoji = "🟢"
				statusText = "OPEN"
			case "filled":
				statusEmoji = "✅"
				statusText = "FILLED"
				fillCount++
			case "canceled":
				statusEmoji = "❌"
				statusText = "CANCELLED"
			case "partially_filled":
				statusEmoji = "🟡"
				statusText = "PARTIAL"
			case "rejected":
				statusEmoji = "🚫"
				statusText = "REJECTED"
			}

			t.Log("───────────────────────────────────────────────────────────")
			t.Logf("%s ORDER #%d - %s", statusEmoji, orderCount, statusText)
			t.Logf("   %s %s %s @ $%s",
				sideSymbol,
				order.Order.Sz,
				order.Order.Coin,
				order.Order.LimitPx)
			t.Logf("   Order ID: %d", order.Order.Oid)
			if order.Order.Cloid != "" {
				t.Logf("   Client OID: %s", order.Order.Cloid)
			}
			t.Logf("   Original Size: %s", order.Order.OrigSz)
			t.Logf("   Time: %s", time.Unix(order.StatusTimestamp/1000, 0).Format("15:04:05"))
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to orderUpdates: %v", err)
	}

	// Subscribe to userFills - shows when orders are executed
	_, err = wsClient.SubscribeToUserFills(testAddress, func(msg hyperliquid.WSMessage) {
		var fillData struct {
			IsSnapshot bool `json:"isSnapshot,omitempty"`
			User       string `json:"user"`
			Fills      []struct {
				Coin      string `json:"coin"`
				Px        string `json:"px"`
				Sz        string `json:"sz"`
				Side      string `json:"side"`
				Time      int64  `json:"time"`
				ClosedPnl string `json:"closedPnl"`
				Hash      string `json:"hash"`
				Oid       int64  `json:"oid"`
				Crossed   bool   `json:"crossed"`
				Fee       string `json:"fee"`
				Tid       int64  `json:"tid"`
			} `json:"fills"`
		}

		if err := json.Unmarshal(msg.Data, &fillData); err != nil {
			return
		}

		// Skip snapshots
		if fillData.IsSnapshot {
			return
		}

		for _, fill := range fillData.Fills {
			sideSymbol := "🟢 BOUGHT"
			if fill.Side == "A" {
				sideSymbol = "🔴 SOLD "
			}

			t.Log("═══════════════════════════════════════════════════════════")
			t.Log("💰 FILL EXECUTED")
			t.Logf("   %s %s %s @ $%s",
				sideSymbol,
				fill.Sz,
				fill.Coin,
				fill.Px)
			t.Logf("   Order ID: %d", fill.Oid)
			t.Logf("   PnL: $%s | Fee: $%s", fill.ClosedPnl, fill.Fee)
			t.Logf("   Time: %s", time.Unix(fill.Time/1000, 0).Format("15:04:05"))
			t.Log("═══════════════════════════════════════════════════════════")
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to userFills: %v", err)
	}

	t.Log("")
	t.Log("✅ Monitoring started!")
	t.Log("")
	t.Log("💡 What to expect:")
	t.Log("   - 🟢 OPEN: When you place a limit order")
	t.Log("   - ✅ FILLED: When your order is completely filled")
	t.Log("   - 🟡 PARTIAL: When your order is partially filled")
	t.Log("   - ❌ CANCELLED: When you cancel an order")
	t.Log("   - 💰 FILL: When a trade executes (shows price & PnL)")
	t.Log("")
	t.Log("⚠️  Now place some orders to see them appear here!")
	t.Log("Monitoring for 5 minutes...")
	t.Log("")

	// Monitor for 5 minutes
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	for {
		select {
		case <-ctx.Done():
			goto summary
		case <-ticker.C:
			elapsed := time.Since(startTime)
			t.Logf("⏱️  Elapsed: %v | Orders: %d | Fills: %d",
				elapsed.Round(time.Second), orderCount, fillCount)

			if elapsed > 300*time.Second {
				goto summary
			}
		}
	}

summary:
	t.Log("")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("            MONITORING SUMMARY")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Total order status updates: %d", orderCount)
	t.Logf("Total fills executed: %d", fillCount)
	t.Log("")

	if orderCount > 0 {
		t.Log("✅ ORDER UPDATES WORKING!")
	} else {
		t.Log("⚠️  No orders detected during monitoring period")
	}
}

// TestOrderStatusQuick is a shorter 60-second version for quick testing
func TestOrderStatusQuick(t *testing.T) {
	godotenv.Overload()

	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("        QUICK ORDER STATUS MONITOR (60 seconds)")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Monitoring address: %s", testAddress)
	t.Logf("(Address in lowercase: %s)", strings.ToLower(testAddress))

	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()

	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	orderUpdates := make([]string, 0)

	// Subscribe to orderUpdates
	_, err = wsClient.SubscribeToOrderUpdates(testAddress, func(msg hyperliquid.WSMessage) {
		t.Logf("🔔 RAW orderUpdates message received: %s", string(msg.Data)[:min(200, len(msg.Data))])

		var orders []WsOrder
		if err := json.Unmarshal(msg.Data, &orders); err != nil {
			t.Logf("⚠️ Failed to parse orderUpdates: %v", err)
			return
		}

		t.Logf("📊 Parsed %d orders from message", len(orders))

		for _, order := range orders {
			side := "BUY"
			if order.Order.Side == "A" {
				side = "SELL"
			}

			summary := fmt.Sprintf("[%s] %s %s %s @ $%s - %s",
				time.Unix(order.StatusTimestamp/1000, 0).Format("15:04:05"),
				order.Status,
				side,
				order.Order.Sz,
				order.Order.Coin,
				order.Order.LimitPx)

			orderUpdates = append(orderUpdates, summary)
			t.Logf("📋 %s", summary)
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	t.Log("")
	t.Logf("✅ Successfully subscribed to orderUpdates for address: %s", testAddress)
	t.Log("✅ Monitoring... Place orders now!")
	t.Log("")

	// Give the subscription a moment to be fully registered
	time.Sleep(2 * time.Second)
	t.Log("⏰ Subscription registered, listening for order updates...")
	t.Log("")

	time.Sleep(58 * time.Second)

	t.Log("")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Captured %d order updates", len(orderUpdates))
	t.Log("═══════════════════════════════════════════════════════════")
}
