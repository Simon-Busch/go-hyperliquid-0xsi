package examples

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	hyperliquid "github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/joho/godotenv"
)

// TestDebugAllMessages - Listen to ALL WebSocket messages to verify connectivity
func TestDebugAllMessages(t *testing.T) {
	godotenv.Overload()

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("         DEBUG ALL MESSAGES TEST")
	t.Log("═══════════════════════════════════════════════════════════")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	var messageCount atomic.Int32

	// Subscribe to L2 book for BTC (we know this works!)
	_, err = wsClient.SubscribeToOrderbook("BTC", func(msg hyperliquid.WSMessage) {
		count := messageCount.Add(1)
		if count <= 5 {
			t.Logf("📊 L2Book message #%d received for BTC", count)
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to L2Book: %v", err)
	}

	// Also subscribe to orderUpdates for the test address
	testAddress := accountAddress(t)
	var orderUpdateCount atomic.Int32

	_, err = wsClient.SubscribeToOrderUpdates(testAddress, func(msg hyperliquid.WSMessage) {
		count := orderUpdateCount.Add(1)
		t.Logf("🔔 OrderUpdate #%d received!", count)
		t.Logf("   Data: %s", string(msg.Data)[:min(200, len(msg.Data))])
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to orderUpdates: %v", err)
	}

	t.Log("")
	t.Log("✅ Subscribed to:")
	t.Log("   - BTC L2 Book (should receive messages immediately)")
	t.Logf("   - Order updates for %s", testAddress)
	t.Log("")
	t.Log("⏰ Monitoring for 20 seconds...")
	t.Log("")

	time.Sleep(20 * time.Second)

	l2Count := messageCount.Load()
	orderCount := orderUpdateCount.Load()

	t.Log("")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("            RESULTS")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("L2 Book messages (BTC): %d", l2Count)
	t.Logf("Order update messages: %d", orderCount)
	t.Log("")

	if l2Count > 0 {
		t.Log("✅ WebSocket is WORKING (received L2Book data)")
	} else {
		t.Log("❌ WebSocket NOT working (no L2Book data)")
	}

	if orderCount == 0 {
		t.Log("")
		t.Log("⚠️  NO ORDER UPDATES RECEIVED")
		t.Log("   This means:")
		t.Log("   1. Either no orders were placed during the test")
		t.Log("   2. Or the wallet address doesn't have active orders")
		t.Log("   3. Or Hyperliquid API requires authentication for orderUpdates")
		t.Log("")
		t.Log("💡 TIP: Try placing an order RIGHT NOW and run the test again")
	} else {
		t.Logf("✅ Received %d order updates!", orderCount)
	}
}
