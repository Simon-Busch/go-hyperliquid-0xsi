package examples

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/joho/godotenv"
)

func TestWebSocketConnection(t *testing.T) {
	godotenv.Overload()

	// Use testnet URL
	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.TestnetAPIURL)

	// Get test wallet address from environment or use a default test address
	testAddress := os.Getenv("WALLET_ADDRESS")
	if testAddress == "" {
		testAddress = "0x568620375999697d0Dc6B515C3B52CD22B4fDC22" // Default test address
	}

	t.Logf("=== TESTING WEBSOCKET CONNECTION TO TESTNET ===")
	t.Logf("Testnet URL: %s", hyperliquid.TestnetAPIURL)
	t.Logf("Test Address: %s", testAddress)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to WebSocket
	t.Log("Connecting to WebSocket...")
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect to WebSocket: %v", err)
	}
	t.Log("✅ WebSocket connected successfully")

	// Track received messages
	var messageCount int
	var lastMessageTime time.Time

	// Message handler
	messageHandler := func(msg hyperliquid.WSMessage) {
		messageCount++
		lastMessageTime = time.Now()

		// Log the first few messages to see what we're receiving
		if messageCount <= 5 {
			msgData, _ := json.MarshalIndent(msg, "", "  ")
			t.Logf("📨 Message %d: %s", messageCount, string(msgData))
		}
	}

	// Subscribe to various data streams
	subscriptions := []struct {
		name string
		sub  hyperliquid.Subscription
	}{
		{
			name: "All Mid Prices",
			sub:  hyperliquid.Subscription{Type: "allMids"},
		},
		{
			name: "User Events",
			sub:  hyperliquid.Subscription{Type: "userEvents", User: testAddress},
		},
		{
			name: "User Fills",
			sub:  hyperliquid.Subscription{Type: "userFills", User: testAddress},
		},
		{
			name: "Web Data v2",
			sub:  hyperliquid.Subscription{Type: "webData2", User: testAddress},
		},
		{
			name: "SOL Orderbook",
			sub:  hyperliquid.Subscription{Type: "l2Book", Coin: "SOL"},
		},
		{
			name: "SOL Trades",
			sub:  hyperliquid.Subscription{Type: "trades", Coin: "SOL"},
		},
		{
			name: "ETH Orderbook",
			sub:  hyperliquid.Subscription{Type: "l2Book", Coin: "ETH"},
		},
		{
			name: "ETH Trades",
			sub:  hyperliquid.Subscription{Type: "trades", Coin: "ETH"},
		},
	}

	// Subscribe to all streams
	t.Log("\n=== SUBSCRIBING TO DATA STREAMS ===")
	for _, sub := range subscriptions {
		subID, err := wsClient.Subscribe(sub.sub, messageHandler)
		if err != nil {
			t.Logf("❌ Failed to subscribe to %s: %v", sub.name, err)
		} else {
			t.Logf("✅ Subscribed to %s (ID: %d)", sub.name, subID)
		}
	}

	// Wait for messages to arrive
	t.Log("\n=== WAITING FOR MESSAGES ===")
	t.Log("Waiting up to 20 seconds for messages...")

	startTime := time.Now()
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			t.Log("Context cancelled, stopping test")
			break
		case <-ticker.C:
			elapsed := time.Since(startTime)
			t.Logf("⏱️  Elapsed: %v, Messages received: %d", elapsed, messageCount)

			// If we've received messages and waited long enough, we can stop
			if messageCount > 0 && elapsed > 10*time.Second {
				t.Log("✅ Received messages, test successful")
				goto cleanup
			}

			// If we've waited too long without messages, fail the test
			if elapsed > 20*time.Second {
				t.Log("❌ No messages received within timeout")
				goto cleanup
			}
		}
	}

cleanup:
	// Close the connection
	t.Log("\n=== CLEANING UP ===")
	err = wsClient.Close()
	if err != nil {
		t.Logf("Warning: Error closing WebSocket: %v", err)
	} else {
		t.Log("✅ WebSocket closed successfully")
	}

	// Test results
	t.Logf("\n=== TEST RESULTS ===")
	t.Logf("Total messages received: %d", messageCount)
	if messageCount > 0 {
		t.Logf("Last message received: %v", lastMessageTime)
		t.Logf("Test duration: %v", time.Since(startTime))
		t.Log("🎉 WebSocket test PASSED")
	} else {
		t.Fatal("❌ WebSocket test FAILED - no messages received")
	}
}

func TestWebSocketReconnection(t *testing.T) {
	godotenv.Overload()

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.TestnetAPIURL)

	t.Log("=== TESTING WEBSOCKET RECONNECTION ===")

	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Connect initially
	t.Log("Initial connection...")
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect initially: %v", err)
	}
	t.Log("✅ Initial connection successful")

	// Subscribe to a simple stream
	messageCount := 0
	subID, err := wsClient.Subscribe(hyperliquid.Subscription{Type: "allMids"}, func(msg hyperliquid.WSMessage) {
		messageCount++
	})
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	t.Logf("✅ Subscribed to allMids (ID: %d)", subID)

	// Wait for a message
	time.Sleep(3 * time.Second)

	// Close connection manually
	t.Log("Manually closing connection...")
	err = wsClient.Close()
	if err != nil {
		t.Logf("Warning: Error closing: %v", err)
	}

	// Try to reconnect
	t.Log("Attempting to reconnect...")
	err = wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to reconnect: %v", err)
	}
	t.Log("✅ Reconnection successful")

	// Wait for messages after reconnection
	time.Sleep(3 * time.Second)

	// Cleanup
	wsClient.Close()

	if messageCount > 0 {
		t.Logf("✅ Reconnection test PASSED - received %d messages", messageCount)
	} else {
		t.Log("⚠️  Reconnection test completed but no messages received")
	}
}

func TestWebSocketSpecificAddress(t *testing.T) {
	godotenv.Overload()

	// Get specific address to test
	testAddress := os.Getenv("WALLET_ADDRESS")
	if testAddress == "" {
		t.Skip("WALLET_ADDRESS not set, skipping specific address test")
	}

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.TestnetAPIURL)

	t.Logf("=== TESTING WEBSOCKET FOR SPECIFIC ADDRESS ===")
	t.Logf("Address: %s", testAddress)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Connect
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Subscribe to user-specific data
	userDataReceived := false
	userFillsReceived := false

	// Subscribe to user events
	_, err = wsClient.Subscribe(hyperliquid.Subscription{Type: "userEvents", User: testAddress}, func(msg hyperliquid.WSMessage) {
		userDataReceived = true
		t.Logf("📊 User event received: %s", string(msg.Data))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to user events: %v", err)
	}

	// Subscribe to user fills
	_, err = wsClient.Subscribe(hyperliquid.Subscription{Type: "userFills", User: testAddress}, func(msg hyperliquid.WSMessage) {
		userFillsReceived = true
		t.Logf("💰 User fill received: %s", string(msg.Data))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to user fills: %v", err)
	}

	// Wait for data
	t.Log("Waiting for user-specific data...")
	time.Sleep(10 * time.Second)

	// Cleanup
	wsClient.Close()

	// Results
	t.Logf("\n=== ADDRESS-SPECIFIC TEST RESULTS ===")
	t.Logf("User events received: %t", userDataReceived)
	t.Logf("User fills received: %t", userFillsReceived)

	if userDataReceived || userFillsReceived {
		t.Log("✅ Address-specific test PASSED")
	} else {
		t.Log("⚠️  Address-specific test completed but no user data received")
	}
}

func TestRealTimeOrderMonitoring(t *testing.T) {
	godotenv.Overload()

	// Get your wallet address
	testAddress := os.Getenv("WALLET_ADDRESS")
	if testAddress == "" {
		testAddress = "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"
	}

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.TestnetAPIURL)

	t.Logf("=== REAL-TIME ORDER MONITORING ===")
	t.Logf("Monitoring address: %s", testAddress)
	t.Logf("⚠️  OPEN AN ORDER IN ANOTHER TERMINAL NOW!")
	t.Logf("This test will watch for your order events...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Connect
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Track different types of events
	events := map[string]int{
		"userFills": 0,
		"webData2":  0,
		"trades":    0,
		"allMids":   0,
	}

	// Subscribe to user fills (WORKING - shows when your orders are filled)
	_, err = wsClient.Subscribe(hyperliquid.Subscription{Type: "userFills", User: testAddress}, func(msg hyperliquid.WSMessage) {
		events["userFills"]++
		t.Logf("💰 USER FILL #%d: %s", events["userFills"], string(msg.Data))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to user fills: %v", err)
	}

	// Subscribe to web data (WORKING - shows user account updates)
	_, err = wsClient.Subscribe(hyperliquid.Subscription{Type: "webData2", User: testAddress}, func(msg hyperliquid.WSMessage) {
		events["webData2"]++
		t.Logf("📊 WEB DATA #%d: %s", events["webData2"], string(msg.Data))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to web data: %v", err)
	}

	// Subscribe to trades for SOL (WORKING - shows all trades including yours)
	_, err = wsClient.Subscribe(hyperliquid.Subscription{Type: "trades", Coin: "SOL"}, func(msg hyperliquid.WSMessage) {
		events["trades"]++
		t.Logf("📈 TRADE #%d: %s", events["trades"], string(msg.Data))
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to trades: %v", err)
	}

	// Subscribe to all mids (WORKING - shows market price updates)
	_, err = wsClient.Subscribe(hyperliquid.Subscription{Type: "allMids"}, func(msg hyperliquid.WSMessage) {
		events["allMids"]++
		if events["allMids"] <= 3 { // Only log first 3 to avoid spam
			t.Logf("📊 MID PRICE #%d: %s", events["allMids"], string(msg.Data))
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to all mids: %v", err)
	}

	t.Log("✅ All subscriptions active")
	t.Log("Now open an order in another terminal to see it here!")
	t.Log("💡 Working channels: userFills, webData2, trades, allMids")

	// Monitor for 30 seconds
	time.Sleep(30 * time.Second)

	// Cleanup
	wsClient.Close()

	// Results
	t.Logf("\n=== ORDER MONITORING RESULTS ===")
	t.Logf("User fills received: %d", events["userFills"])
	t.Logf("Web data updates: %d", events["webData2"])
	t.Logf("Trades received: %d", events["trades"])
	t.Logf("Mid price updates: %d", events["allMids"])

	if events["userFills"] > 0 || events["webData2"] > 0 {
		t.Log("🎉 Order monitoring successful - you should see your orders!")
	} else {
		t.Log("⚠️  No order events received - try opening an order while this test runs")
		t.Log("💡 Note: userEvents and orderUpdates don't work on Hyperliquid WebSocket")
	}
}

func TestWebSocketChannelDiscovery(t *testing.T) {
	godotenv.Overload()

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.TestnetAPIURL)

	t.Log("=== WEBSOCKET CHANNEL DISCOVERY ===")
	t.Log("Testing different subscription types to see what works...")

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Connect
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Test different subscription types
	testSubscriptions := []struct {
		name string
		sub  hyperliquid.Subscription
	}{
		{"allMids", hyperliquid.Subscription{Type: "allMids"}},
		{"l2Book SOL", hyperliquid.Subscription{Type: "l2Book", Coin: "SOL"}},
		{"trades SOL", hyperliquid.Subscription{Type: "trades", Coin: "SOL"}},
		{"userEvents", hyperliquid.Subscription{Type: "userEvents", User: "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"}},
		{"userFills", hyperliquid.Subscription{Type: "userFills", User: "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"}},
		{"orderUpdates", hyperliquid.Subscription{Type: "orderUpdates"}},
		{"webData2", hyperliquid.Subscription{Type: "webData2", User: "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"}},
		{"userNonFundingLedgerUpdates", hyperliquid.Subscription{Type: "userNonFundingLedgerUpdates", User: "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"}},
		{"userFundings", hyperliquid.Subscription{Type: "userFundings", User: "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"}},
		{"clearinghouseState", hyperliquid.Subscription{Type: "clearinghouseState", User: "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"}},
		{"frontendOpenOrders", hyperliquid.Subscription{Type: "frontendOpenOrders", User: "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"}},
	}

	// Track which subscriptions work
	workingSubs := make(map[string]bool)
	messageCounts := make(map[string]int)

	// Subscribe to all test channels
	for _, testSub := range testSubscriptions {
		_, err := wsClient.Subscribe(testSub.sub, func(msg hyperliquid.WSMessage) {
			messageCounts[testSub.name]++
			if messageCounts[testSub.name] <= 3 { // Log first 3 messages
				t.Logf("📨 %s: %s", testSub.name, string(msg.Data))
			}
		})

		if err != nil {
			t.Logf("❌ Failed to subscribe to %s: %v", testSub.name, err)
			workingSubs[testSub.name] = false
		} else {
			t.Logf("✅ Subscribed to %s", testSub.name)
			workingSubs[testSub.name] = true
		}
	}

	t.Log("\n=== WAITING FOR MESSAGES ===")
	t.Log("Waiting 30 seconds to see which channels are active...")

	// Wait and monitor
	time.Sleep(30 * time.Second)

	// Cleanup
	wsClient.Close()

	// Results
	t.Logf("\n=== CHANNEL DISCOVERY RESULTS ===")
	for _, testSub := range testSubscriptions {
		status := "❌"
		if workingSubs[testSub.name] {
			status = "✅"
		}
		t.Logf("%s %s: %d messages", status, testSub.name, messageCounts[testSub.name])
	}

	// Summary
	workingCount := 0
	for _, working := range workingSubs {
		if working {
			workingCount++
		}
	}

	t.Logf("\nWorking channels: %d/%d", workingCount, len(testSubscriptions))
}

func TestUserNotifications(t *testing.T) {
	godotenv.Overload()

	// Get your wallet address
	testAddress := os.Getenv("WALLET_ADDRESS")
	if testAddress == "" {
		testAddress = "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"
	}

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.TestnetAPIURL)

	t.Logf("=== USER NOTIFICATIONS TEST ===")
	t.Logf("Listening for notifications for address: %s", testAddress)
	t.Logf("This test will show any notifications sent to your wallet...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Connect
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Track notifications
	notificationCount := 0
	lastNotificationTime := time.Now()

	// Subscribe to user notifications
	_, err = wsClient.Subscribe(hyperliquid.Subscription{Type: "notification", User: testAddress}, func(msg hyperliquid.WSMessage) {
		notificationCount++
		lastNotificationTime = time.Now()

		t.Logf("🔔 NOTIFICATION #%d received at %s", notificationCount, lastNotificationTime.Format("15:04:05"))
		t.Logf("   Channel: %s", msg.Channel)
		t.Logf("   Data: %s", string(msg.Data))

		// Try to parse the notification data
		var notification struct {
			Notification string `json:"notification"`
		}
		if err := json.Unmarshal(msg.Data, &notification); err == nil {
			t.Logf("   Message: %s", notification.Notification)
		}
		t.Logf("   ---")
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to notifications: %v", err)
	}

	t.Log("✅ Notification subscription active")
	t.Log("Waiting for notifications...")
	t.Log("💡 Notifications may include:")
	t.Log("   - Order status updates")
	t.Log("   - Position changes")
	t.Log("   - Account alerts")
	t.Log("   - System messages")

	// Monitor for 45 seconds
	time.Sleep(45 * time.Second)

	// Cleanup
	wsClient.Close()

	// Results
	t.Logf("\n=== NOTIFICATION RESULTS ===")
	t.Logf("Total notifications received: %d", notificationCount)

	if notificationCount > 0 {
		t.Logf("Last notification at: %s", lastNotificationTime.Format("15:04:05"))
		t.Log("🎉 Successfully received user notifications!")
	} else {
		t.Log("⚠️  No notifications received during test period")
		t.Log("💡 This is normal - notifications are only sent when there are events")
		t.Log("💡 Try placing an order or making a trade to trigger notifications")
	}
}

func TestCompleteUserActivityMonitoring(t *testing.T) {
	godotenv.Overload()

	// Get your wallet address
	testAddress := os.Getenv("WALLET_ADDRESS")
	if testAddress == "" {
		testAddress = "0x568620375999697d0Dc6B515C3B52CD22B4fDC22"
	}

	wsClient := hyperliquid.NewWebsocketClient(hyperliquid.TestnetAPIURL)

	t.Logf("=== COMPLETE USER ACTIVITY MONITORING ===")
	t.Logf("Monitoring address: %s", testAddress)
	t.Logf("�� This test monitors user fills and notifications...")
	t.Logf("⚠️  Now try: opening/closing positions, placing orders, etc.")

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Connect
	err := wsClient.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Track different types of events
	events := map[string]int{
		"userFills":    0,
		"notification": 0,
	}

	// Subscribe to user fills (when orders are executed)
	_, err = wsClient.Subscribe(hyperliquid.Subscription{Type: "userFills", User: testAddress}, func(msg hyperliquid.WSMessage) {
		events["userFills"]++

		// Parse the userFills message structure
		var fillMessage struct {
			IsSnapshot bool   `json:"isSnapshot,omitempty"`
			User       string `json:"user"`
			Fills      []struct {
				Coin      string `json:"coin"`
				Price     string `json:"px"`
				Size      string `json:"sz"`
				Side      string `json:"side"`
				Direction string `json:"dir"`
				ClosedPnL string `json:"closedPnl"`
				Hash      string `json:"hash"`
				OID       int64  `json:"oid"`
				Fee       string `json:"fee"`
				FeeToken  string `json:"feeToken"`
			} `json:"fills"`
		}

		if err := json.Unmarshal(msg.Data, &fillMessage); err == nil {
			if fillMessage.IsSnapshot {
				t.Logf("💰 SNAPSHOT #%d: %d fills loaded", events["userFills"], len(fillMessage.Fills))
			} else {
				t.Logf("💰 NEW FILLS #%d: %d fills received", events["userFills"], len(fillMessage.Fills))
			}

			// Show details for each fill
			for i, fill := range fillMessage.Fills {
				side := "BUY"
				if fill.Side == "A" {
					side = "SELL"
				}

				t.Logf("   %d. %s %s %s @ $%s | PnL: $%s | Fee: $%s %s",
					i+1, side, fill.Size, fill.Coin, fill.Price, fill.ClosedPnL, fill.Fee, fill.FeeToken)
			}
		} else {
			t.Logf("💰 FILL #%d: Failed to parse: %v", events["userFills"], err)
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to userFills: %v", err)
	}

	// Subscribe to notifications
	_, err = wsClient.Subscribe(hyperliquid.Subscription{Type: "notification", User: testAddress}, func(msg hyperliquid.WSMessage) {
		events["notification"]++

		// Try to parse notification content for key info
		var notification struct {
			Message string `json:"message"`
			Type    string `json:"type"`
		}
		if json.Unmarshal(msg.Data, &notification) == nil {
			t.Logf("🔔 NOTIFICATION #%d: %s (%s)",
				events["notification"],
				notification.Message,
				notification.Type)
		} else {
			t.Logf("🔔 NOTIFICATION #%d: %s", events["notification"], string(msg.Data))
		}
	})
	if err != nil {
		t.Fatalf("Failed to subscribe to notifications: %v", err)
	}

	t.Log("✅ Subscriptions active!")
	t.Log("📋 Monitoring channels:")
	t.Log("   - userFills: Your order executions")
	t.Log("   - notification: System alerts & messages")
	t.Log("")
	t.Log("🚀 Now perform some actions:")
	t.Log("   - Open/close positions")
	t.Log("   - Place/cancel orders")
	t.Log("   - Check for liquidations")
	t.Log("")

	// Monitor for 75 seconds
	time.Sleep(75 * time.Second)

	// Cleanup
	wsClient.Close()

	// Results
	t.Logf("\n=== MONITORING RESULTS ===")
	t.Logf("Total events received:")
	t.Logf("   - User Fills: %d", events["userFills"])
	t.Logf("   - Notifications: %d", events["notification"])

	totalEvents := events["userFills"] + events["notification"]
	t.Logf("Total events: %d", totalEvents)

	if totalEvents == 0 {
		t.Log("⚠️  No events received. This might mean:")
		t.Log("   - No trading activity during test")
		t.Log("   - Need to perform actions during test")
	} else {
		t.Log("✅ Events received successfully!")
	}
}
