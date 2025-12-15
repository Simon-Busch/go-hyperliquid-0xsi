package examples

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	hyperliquid "github.com/Simon-Busch/go-hyperliquid-0xsi"
)

// TestWebSocketPostInfo tests sending info requests via WebSocket POST
func TestWebSocketPostInfo(t *testing.T) {
	// Create WebSocket client
	ws := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)
	defer ws.Close()

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := ws.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait a bit for connection to stabilize
	time.Sleep(500 * time.Millisecond)

	t.Run("Meta request", func(t *testing.T) {
		payload := map[string]any{
			"type": "meta",
		}

		resp, err := ws.PostInfoRequest(payload, 10*time.Second)
		if err != nil {
			t.Fatalf("Failed to get meta: %v", err)
		}

		// WebSocket POST wraps the response with {"type":"...", "data":{...}}
		// We need to extract the "data" field
		var wrapper struct {
			Type string               `json:"type"`
			Data hyperliquid.Meta `json:"data"`
		}
		if err := json.Unmarshal(resp, &wrapper); err != nil {
			t.Fatalf("Failed to unmarshal meta wrapper: %v", err)
		}

		if len(wrapper.Data.Universe) == 0 {
			t.Error("Expected non-empty universe")
		}

		t.Logf("Got %d assets from meta via WebSocket POST", len(wrapper.Data.Universe))
	})

	t.Run("AllMids request", func(t *testing.T) {
		payload := map[string]any{
			"type": "allMids",
		}

		resp, err := ws.PostInfoRequest(payload, 10*time.Second)
		if err != nil {
			t.Fatalf("Failed to get allMids: %v", err)
		}

		// WebSocket POST wraps the response
		var wrapper struct {
			Type string            `json:"type"`
			Data map[string]string `json:"data"`
		}
		if err := json.Unmarshal(resp, &wrapper); err != nil {
			t.Fatalf("Failed to unmarshal allMids: %v", err)
		}

		if len(wrapper.Data) == 0 {
			t.Error("Expected non-empty mids")
		}

		t.Logf("Got %d market mids via WebSocket POST", len(wrapper.Data))
	})

	t.Run("L2Book request", func(t *testing.T) {
		payload := map[string]any{
			"type": "l2Book",
			"coin": "BTC",
		}

		resp, err := ws.PostInfoRequest(payload, 10*time.Second)
		if err != nil {
			t.Fatalf("Failed to get l2Book: %v", err)
		}

		// WebSocket POST wraps the response
		var wrapper struct {
			Type string             `json:"type"`
			Data hyperliquid.L2Book `json:"data"`
		}
		if err := json.Unmarshal(resp, &wrapper); err != nil {
			t.Fatalf("Failed to unmarshal l2Book: %v", err)
		}

		if len(wrapper.Data.Levels) == 0 {
			t.Error("Expected non-empty order book levels")
		}

		t.Logf("Got %d levels in BTC order book via WebSocket POST", len(wrapper.Data.Levels))
	})
}

// TestWebSocketPostConcurrent tests concurrent POST requests
func TestWebSocketPostConcurrent(t *testing.T) {
	// Create WebSocket client
	ws := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)
	defer ws.Close()

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := ws.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// Wait a bit for connection to stabilize
	time.Sleep(500 * time.Millisecond)

	// Send multiple concurrent requests
	numRequests := 5
	results := make(chan error, numRequests)

	for i := 0; i < numRequests; i++ {
		go func(index int) {
			payload := map[string]any{
				"type": "allMids",
			}

			_, err := ws.PostInfoRequest(payload, 10*time.Second)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numRequests; i++ {
		err := <-results
		if err != nil {
			t.Errorf("Request %d failed: %v", i, err)
		}
	}

	t.Logf("Successfully completed %d concurrent POST requests", numRequests)
}

// TestWebSocketPostTimeout tests timeout behavior
func TestWebSocketPostTimeout(t *testing.T) {
	// Create WebSocket client but don't connect
	ws := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)
	defer ws.Close()

	// Try to send request without connecting
	payload := map[string]any{
		"type": "meta",
	}

	_, err := ws.PostInfoRequest(payload, 1*time.Second)
	if err == nil {
		t.Error("Expected error when not connected")
	}

	t.Logf("Got expected error: %v", err)
}

// TestWebSocketPostVsHTTP compares WebSocket POST with HTTP for the same request
func TestWebSocketPostVsHTTP(t *testing.T) {
	// Create Info client for HTTP requests
	info := hyperliquid.NewInfo(hyperliquid.MainnetAPIURL, true, nil, nil)

	// Get meta via HTTP
	httpStart := time.Now()
	httpMeta, err := info.Meta()
	httpDuration := time.Since(httpStart)
	if err != nil {
		t.Fatalf("HTTP Meta failed: %v", err)
	}

	// Create WebSocket client
	ws := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)
	defer ws.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = ws.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Get meta via WebSocket POST
	wsStart := time.Now()
	payload := map[string]any{
		"type": "meta",
	}
	resp, err := ws.PostInfoRequest(payload, 10*time.Second)
	if err != nil {
		t.Fatalf("WebSocket Meta failed: %v", err)
	}

	// WebSocket POST wraps the response
	var wrapper struct {
		Type string            `json:"type"`
		Data hyperliquid.Meta `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		t.Fatalf("Failed to unmarshal WebSocket meta: %v", err)
	}
	wsMeta := wrapper.Data
	wsDuration := time.Since(wsStart)

	// Compare results
	if len(httpMeta.Universe) != len(wsMeta.Universe) {
		t.Errorf("Universe size mismatch: HTTP=%d, WS=%d", len(httpMeta.Universe), len(wsMeta.Universe))
	}

	t.Logf("HTTP Meta took: %v", httpDuration)
	t.Logf("WebSocket Meta took: %v", wsDuration)
	t.Logf("Both returned %d assets", len(httpMeta.Universe))
}

// TestWebSocketPostUserState tests querying user state via WebSocket POST
func TestWebSocketPostUserState(t *testing.T) {
	// Create WebSocket client
	ws := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)
	defer ws.Close()

	// Connect
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := ws.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	time.Sleep(500 * time.Millisecond)

	// Use a test address (Hyperliquid's example address)
	testAddress := "0x0000000000000000000000000000000000000000"

	payload := map[string]any{
		"type": "clearinghouseState",
		"user": testAddress,
	}

	resp, err := ws.PostInfoRequest(payload, 10*time.Second)
	if err != nil {
		t.Fatalf("Failed to get user state: %v", err)
	}

	// WebSocket POST wraps the response
	var wrapper struct {
		Type string                 `json:"type"`
		Data hyperliquid.UserState `json:"data"`
	}
	if err := json.Unmarshal(resp, &wrapper); err != nil {
		t.Fatalf("Failed to unmarshal user state: %v", err)
	}

	t.Logf("Got user state for %s via WebSocket POST", testAddress)
	t.Logf("Cross margin summary: %+v", wrapper.Data.CrossMarginSummary)
}
