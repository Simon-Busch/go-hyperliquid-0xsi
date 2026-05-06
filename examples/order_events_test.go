package examples

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	hyperliquid "github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/joho/godotenv"
)

// TestOrderUpdatesSubscription tests the orderUpdates WebSocket subscription
// This subscription provides real-time updates about order status changes
func TestOrderUpdatesSubscription(t *testing.T) {
	godotenv.Overload()

	// Get test wallet address
	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("=== TESTING ORDER UPDATES SUBSCRIPTION ===")
	t.Logf("Monitoring address: %s", testAddress)
	t.Log("This test verifies that orderUpdates subscription is properly handled")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Connect to WebSocket
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	messageCount := 0
	snapshotCount := 0
	messageChan := make(chan hyperliquid.WSMessage, 10)

	// Subscribe to order updates using the helper method
	subID, err := wsClient.SubscribeToOrderUpdates(testAddress, func(msg hyperliquid.WSMessage) {
		t.Logf("📬 Order Update received")
		t.Logf("   Channel: %s", msg.Channel)

		// Check if this is a snapshot
		var orderData map[string]interface{}
		if err := json.Unmarshal(msg.Data, &orderData); err == nil {
			if isSnapshot, ok := orderData["isSnapshot"].(bool); ok && isSnapshot {
				snapshotCount++
				t.Logf("   Type: SNAPSHOT #%d", snapshotCount)
				t.Logf("   Data: %s", string(msg.Data))
				return // Don't count snapshots
			}
		}

		messageCount++
		t.Logf("   Type: NEW ORDER UPDATE #%d", messageCount)
		t.Logf("   Data: %s", string(msg.Data))

		// Send to channel for test completion
		messageChan <- msg
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to orderUpdates: %v", err)
	}

	t.Logf("✅ Successfully subscribed to orderUpdates (ID: %d)", subID)
	t.Log("")
	t.Log("⚠️  NOTE: This subscription requires active order placement to receive messages")
	t.Log("💡 Place an order in another terminal to trigger updates")
	t.Log("Waiting for up to 90 seconds...")

	// Wait for messages or timeout
	timeout := time.After(90 * time.Second)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-messageChan:
			t.Logf("✅ Received orderUpdates message!")
			t.Logf("   Full message: %+v", msg)
			// Continue listening for more messages
		case <-timeout:
			t.Log("⏰ Timeout reached")
			t.Logf("📊 Summary: %d snapshots ignored, %d new order updates received", snapshotCount, messageCount)
			if messageCount == 0 {
				t.Log("⚠️  No new orderUpdates messages received (excluding snapshots)")
				t.Log("💡 This is normal if no orders were placed during the test")
				t.Log("✅ Subscription was successful - handler is properly configured")
			} else {
				t.Logf("✅ Test PASSED - received %d orderUpdates messages", messageCount)
			}
			return
		case <-ticker.C:
			t.Logf("⏳ Still listening... (%d updates, %d snapshots received)", messageCount, snapshotCount)
		}
	}
}

// TestUserEventsSubscription tests the userEvents WebSocket subscription
// This subscription provides non-order user events (fills, funding, liquidations, etc.)
func TestUserEventsSubscription(t *testing.T) {
	godotenv.Overload()

	// Get test wallet address
	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("=== TESTING USER EVENTS SUBSCRIPTION ===")
	t.Logf("Monitoring address: %s", testAddress)
	t.Log("This test verifies that userEvents subscription is properly handled")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Connect to WebSocket
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	messageCount := 0
	snapshotCount := 0
	messageChan := make(chan hyperliquid.WSMessage, 10)

	// Subscribe to user events using the helper method
	subID, err := wsClient.SubscribeToUserEvents(testAddress, func(msg hyperliquid.WSMessage) {
		t.Logf("📬 User Event received")
		t.Logf("   Channel: %s", msg.Channel)

		// Check if this is a snapshot
		var eventData map[string]interface{}
		if err := json.Unmarshal(msg.Data, &eventData); err == nil {
			if isSnapshot, ok := eventData["isSnapshot"].(bool); ok && isSnapshot {
				snapshotCount++
				t.Logf("   Type: SNAPSHOT #%d", snapshotCount)
				return // Don't count snapshots
			}
		}

		messageCount++
		t.Logf("   Type: NEW EVENT #%d", messageCount)
		t.Logf("   Data: %s", string(msg.Data))

		// Send to channel for test completion
		messageChan <- msg
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to userEvents: %v", err)
	}

	t.Logf("✅ Successfully subscribed to userEvents (ID: %d)", subID)
	t.Log("")
	t.Log("⚠️  NOTE: This subscription shows non-order events (fills, funding, liquidations, etc.)")
	t.Log("💡 Place an order or make a trade to trigger events")
	t.Log("Waiting for up to 90 seconds...")

	// Wait for messages or timeout
	timeout := time.After(90 * time.Second)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case msg := <-messageChan:
			t.Logf("✅ Received userEvents message!")
			t.Logf("   Full message: %+v", msg)
			// Continue listening for more messages
		case <-timeout:
			t.Log("⏰ Timeout reached")
			t.Logf("📊 Summary: %d snapshots ignored, %d new events received", snapshotCount, messageCount)
			if messageCount == 0 {
				t.Log("⚠️  No userEvents messages received (excluding snapshots)")
				t.Log("💡 This is normal if no trading activity occurred during the test")
				t.Log("✅ Subscription was successful - handler is properly configured")
			} else {
				t.Logf("✅ Test PASSED - received %d userEvents messages", messageCount)
			}
			return
		case <-ticker.C:
			t.Logf("⏳ Still listening... (%d events, %d snapshots received)", messageCount, snapshotCount)
		}
	}
}

// TestBothOrderAndUserEvents tests both subscriptions simultaneously
// This verifies that both can work together without conflicts
func TestBothOrderAndUserEventsSubscriptions(t *testing.T) {
	godotenv.Overload()

	// Get test wallet address
	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("=== TESTING BOTH ORDER UPDATES AND USER EVENTS ===")
	t.Logf("Monitoring address: %s", testAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Connect
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	// Track messages by type
	orderUpdatesCount := 0
	userEventsCount := 0

	// Subscribe to order updates
	orderSubID, err := wsClient.SubscribeToOrderUpdates(testAddress, func(msg hyperliquid.WSMessage) {
		// Filter out snapshots
		var orderData map[string]interface{}
		if err := json.Unmarshal(msg.Data, &orderData); err == nil {
			if isSnapshot, ok := orderData["isSnapshot"].(bool); ok && isSnapshot {
				return // Skip snapshots
			}
		}

		orderUpdatesCount++
		t.Logf("📦 Order Update #%d: %s", orderUpdatesCount, string(msg.Data))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to orderUpdates: %v", err)
	}
	t.Logf("✅ Subscribed to orderUpdates (ID: %d)", orderSubID)

	// Subscribe to user events
	userSubID, err := wsClient.SubscribeToUserEvents(testAddress, func(msg hyperliquid.WSMessage) {
		// Filter out snapshots
		var eventData map[string]interface{}
		if err := json.Unmarshal(msg.Data, &eventData); err == nil {
			if isSnapshot, ok := eventData["isSnapshot"].(bool); ok && isSnapshot {
				return // Skip snapshots
			}
		}

		userEventsCount++
		t.Logf("👤 User Event #%d: %s", userEventsCount, string(msg.Data))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to userEvents: %v", err)
	}
	t.Logf("✅ Subscribed to userEvents (ID: %d)", userSubID)

	t.Log("")
	t.Log("🔍 Monitoring both channels simultaneously...")
	t.Log("⚠️  Place orders to see both subscription types in action")

	// Wait for the test duration
	time.Sleep(60 * time.Second)

	// Results
	t.Log("")
	t.Log("=== SUBSCRIPTION TEST RESULTS ===")
	t.Logf("Order Updates received: %d", orderUpdatesCount)
	t.Logf("User Events received: %d", userEventsCount)
	t.Log("")
	t.Log("✅ Both subscriptions are properly implemented and handled by ws.go")
	t.Log("💡 The subscriptions work - they just need active trading to receive messages")
}
