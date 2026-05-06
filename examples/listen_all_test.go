package examples

import (
	"context"
	"testing"
	"time"

	hyperliquid "github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/joho/godotenv"
)

// TestListenToAllChannels subscribes to multiple channels and logs everything
func TestListenToAllChannels(t *testing.T) {
	godotenv.Overload()

	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("=== LISTENING TO ALL CHANNELS WITH DEBUG LOGGING ===")
	t.Logf("Address: %s", testAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	counts := make(map[string]int)

	// Subscribe to orderUpdates
	_, err = wsClient.SubscribeToOrderUpdates(testAddress, func(msg hyperliquid.WSMessage) {
		counts["orderUpdates"]++
		t.Logf("🔵 orderUpdates #%d: %s", counts["orderUpdates"], string(msg.Data)[:min(100, len(msg.Data))])
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to orderUpdates: %v", err)
	}
	t.Log("✅ Subscribed to orderUpdates")

	// Subscribe to userEvents
	_, err = wsClient.SubscribeToUserEvents(testAddress, func(msg hyperliquid.WSMessage) {
		counts["userEvents"]++
		t.Logf("🟢 userEvents #%d: %s", counts["userEvents"], string(msg.Data)[:min(100, len(msg.Data))])
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to userEvents: %v", err)
	}
	t.Log("✅ Subscribed to userEvents")

	// Subscribe to userFills
	_, err = wsClient.SubscribeToUserFills(testAddress, func(msg hyperliquid.WSMessage) {
		counts["userFills"]++
		t.Logf("🟡 userFills #%d: %s", counts["userFills"], string(msg.Data)[:min(100, len(msg.Data))])
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to userFills: %v", err)
	}
	t.Log("✅ Subscribed to userFills")

	// Subscribe to webData2
	_, err = wsClient.SubscribeToWebData2(testAddress, func(msg hyperliquid.WSMessage) {
		counts["webData2"]++
		t.Logf("🔴 webData2 #%d: %s", counts["webData2"], string(msg.Data)[:min(100, len(msg.Data))])
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to webData2: %v", err)
	}
	t.Log("✅ Subscribed to webData2")

	t.Log("")
	t.Log("⚠️  ALL SUBSCRIPTIONS ACTIVE - NOW PLACE AN ORDER!")
	t.Log("Waiting for 120 seconds...")
	t.Log("")
	t.Log("💡 Check the logs above for debug messages showing:")
	t.Log("   - 'WebSocket message received - Channel: XXX' for all incoming messages")
	t.Log("   - 'Matched subscription' when a message matches a subscription")
	t.Log("   - 'No matching subscriptions' when a message doesn't match")

	time.Sleep(120 * time.Second)

	t.Log("")
	t.Log("=== FINAL COUNTS ===")
	t.Logf("orderUpdates: %d", counts["orderUpdates"])
	t.Logf("userEvents: %d", counts["userEvents"])
	t.Logf("userFills: %d", counts["userFills"])
	t.Logf("webData2: %d", counts["webData2"])
}
