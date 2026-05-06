package examples

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	hyperliquid "github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/joho/godotenv"
)

// TestDebugOrderUpdates - Very verbose test to debug orderUpdates
func TestDebugOrderUpdates(t *testing.T) {
	godotenv.Overload()

	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("            DEBUG ORDER UPDATES TEST")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Monitoring address: %s", testAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	messageCount := 0
	parseErrorCount := 0

	// Subscribe to orderUpdates with very verbose logging
	_, err = wsClient.SubscribeToOrderUpdates(testAddress, func(msg hyperliquid.WSMessage) {
		messageCount++

		// Log the raw message
		rawData := string(msg.Data)
		if len(rawData) > 500 {
			rawData = rawData[:500] + "..."
		}
		t.Logf("\n━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
		t.Logf("🔔 Message #%d received", messageCount)
		t.Logf("   Channel: %s", msg.Channel)
		t.Logf("   Raw Data: %s", rawData)
		t.Logf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")

		// Try to parse as array
		var ordersArray []interface{}
		if err := json.Unmarshal(msg.Data, &ordersArray); err == nil {
			t.Logf("   ✅ Parsed as array, length: %d", len(ordersArray))
			for i, order := range ordersArray {
				orderJSON, _ := json.MarshalIndent(order, "      ", "  ")
				t.Logf("   Order #%d:\n%s", i+1, string(orderJSON))
			}
		} else {
			parseErrorCount++
			t.Logf("   ⚠️ Failed to parse as array: %v", err)

			// Try to parse as object
			var orderObj interface{}
			if err := json.Unmarshal(msg.Data, &orderObj); err == nil {
				t.Logf("   ℹ️  Parsed as object:")
				orderJSON, _ := json.MarshalIndent(orderObj, "      ", "  ")
				t.Logf("%s", string(orderJSON))
			} else {
				t.Logf("   ❌ Failed to parse as object: %v", err)
			}
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	t.Log("")
	t.Log("✅ Subscription successful!")
	t.Log("⏰ Waiting 5 seconds for subscription to register...")
	time.Sleep(5 * time.Second)
	t.Log("")
	t.Log("🚀 NOW PLACE YOUR ORDERS!")
	t.Log("   Test will run for 60 seconds...")
	t.Log("")

	// Monitor for 60 seconds
	time.Sleep(60 * time.Second)

	t.Log("")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("            RESULTS")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Total messages received: %d", messageCount)
	t.Logf("Parse errors: %d", parseErrorCount)

	if messageCount == 0 {
		t.Log("")
		t.Log("⚠️  NO MESSAGES RECEIVED!")
		t.Log("   Possible reasons:")
		t.Logf("   1. Wrong wallet address (using: %s)", testAddress)
		t.Log("   2. No orders placed during test window")
		t.Log("   3. WebSocket subscription issue")
	}
}
