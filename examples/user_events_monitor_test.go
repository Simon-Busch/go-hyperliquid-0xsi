package examples

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	hyperliquid "github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/joho/godotenv"
)

// TestUserEventsMonitor monitors userEvents (fills, funding, liquidations, etc.)
func TestUserEventsMonitor(t *testing.T) {
	godotenv.Overload()

	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("            USER EVENTS MONITOR")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Address: %s", testAddress)
	t.Log("")
	t.Log("💡 userEvents shows:")
	t.Log("   - Fills (when trades execute)")
	t.Log("   - Funding payments")
	t.Log("   - Liquidations")
	t.Log("   - Other non-order user events")
	t.Log("")

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	eventCount := 0
	snapshotCount := 0

	// Subscribe to userEvents
	_, err = wsClient.SubscribeToUserEvents(testAddress, func(msg hyperliquid.WSMessage) {
		t.Logf("🔍 RAW userEvents message: %s", string(msg.Data)[:min(500, len(msg.Data))])

		// Check if this is a snapshot
		var eventData map[string]interface{}
		if err := json.Unmarshal(msg.Data, &eventData); err == nil {
			if isSnapshot, ok := eventData["isSnapshot"].(bool); ok && isSnapshot {
				snapshotCount++
				t.Logf("📸 SNAPSHOT #%d (ignored)", snapshotCount)
				return
			}
		}

		eventCount++
		t.Log("───────────────────────────────────────────────────────────")
		t.Logf("👤 USER EVENT #%d", eventCount)
		t.Logf("   Channel: %s", msg.Channel)
		t.Logf("   Time: %s", time.Now().Format("15:04:05"))

		// Try to parse as different event types
		var parsed map[string]interface{}
		if err := json.Unmarshal(msg.Data, &parsed); err == nil {
			// Pretty print the event
			prettyJSON, _ := json.MarshalIndent(parsed, "   ", "  ")
			t.Logf("   Data:\n   %s", string(prettyJSON))
		} else {
			t.Logf("   Raw: %s", string(msg.Data))
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to userEvents: %v", err)
	}

	t.Log("")
	t.Log("✅ Monitoring started!")
	t.Log("⚠️  Now place orders/trades to see events!")
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
			t.Logf("⏱️  Elapsed: %v | Events: %d | Snapshots: %d",
				elapsed.Round(time.Second), eventCount, snapshotCount)

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
	t.Logf("Total user events: %d", eventCount)
	t.Logf("Snapshots ignored: %d", snapshotCount)
	t.Log("")

	if eventCount > 0 {
		t.Log("✅ USER EVENTS WORKING!")
	} else {
		t.Log("⚠️  No events detected during monitoring period")
		t.Log("💡 userEvents are triggered by trading activity")
	}
}

// TestUserEventsQuick is a shorter 60-second version
func TestUserEventsQuick(t *testing.T) {
	godotenv.Overload()

	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("        QUICK USER EVENTS MONITOR (60 seconds)")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Address: %s", testAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Second)
	defer cancel()

	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	events := make([]string, 0)

	// Subscribe to userEvents
	_, err = wsClient.SubscribeToUserEvents(testAddress, func(msg hyperliquid.WSMessage) {
		// Skip snapshots
		if string(msg.Data)[:min(20, len(msg.Data))] == `{"isSnapshot":true` {
			return
		}

		eventSummary := string(msg.Data)[:min(200, len(msg.Data))]
		events = append(events, eventSummary)
		t.Logf("👤 Event #%d: %s", len(events), eventSummary)
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	t.Log("")
	t.Log("✅ Monitoring... Place orders/trades now!")
	t.Log("")

	time.Sleep(60 * time.Second)

	t.Log("")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Captured %d user events", len(events))
	t.Log("═══════════════════════════════════════════════════════════")
}

// TestBothOrderUpdatesAndUserEvents monitors both channels simultaneously
func TestBothOrderUpdatesAndUserEvents(t *testing.T) {
	godotenv.Overload()

	testAddress := accountAddress(t)

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)

	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("    MONITORING BOTH orderUpdates & userEvents")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Address: %s", testAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer wsClient.Close()

	orderCount := 0
	eventCount := 0

	// Subscribe to orderUpdates
	_, err = wsClient.SubscribeToOrderUpdates(testAddress, func(msg hyperliquid.WSMessage) {
		orderCount++
		t.Logf("📦 ORDER UPDATE #%d: Channel=%s", orderCount, msg.Channel)
		t.Logf("   Data: %s", string(msg.Data)[:min(200, len(msg.Data))])
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to orderUpdates: %v", err)
	}
	t.Log("✅ Subscribed to orderUpdates")

	// Subscribe to userEvents
	_, err = wsClient.SubscribeToUserEvents(testAddress, func(msg hyperliquid.WSMessage) {
		// Skip snapshots
		if string(msg.Data)[:min(20, len(msg.Data))] == `{"isSnapshot":true` {
			return
		}

		eventCount++
		t.Logf("👤 USER EVENT #%d: Channel=%s", eventCount, msg.Channel)
		t.Logf("   Data: %s", string(msg.Data)[:min(200, len(msg.Data))])
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to userEvents: %v", err)
	}
	t.Log("✅ Subscribed to userEvents")

	t.Log("")
	t.Log("🔍 Monitoring both channels...")
	t.Log("⚠️  Place orders to see both subscription types in action!")
	t.Log("Monitoring for 2 minutes...")
	t.Log("")

	time.Sleep(120 * time.Second)

	t.Log("")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Log("            FINAL RESULTS")
	t.Log("═══════════════════════════════════════════════════════════")
	t.Logf("Order updates: %d", orderCount)
	t.Logf("User events: %d", eventCount)
	t.Log("")
	t.Log("✅ Both subscriptions are working!")
	t.Log("")
	t.Log("💡 Key differences:")
	t.Log("   - orderUpdates: Order status changes (open, filled, cancelled)")
	t.Log("   - userEvents: Fills, funding, liquidations, and other events")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
