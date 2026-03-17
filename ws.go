package hyperliquid

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// pingInterval matches Hyperliquid's expected interval (upstream uses 50s)
	pingInterval = 50 * time.Second

	// readDeadline is the maximum time to wait for a read before timing out.
	// Set slightly longer than pingInterval to allow for network latency.
	readDeadline = pingInterval + 10*time.Second

	// reconnectBaseWait is the initial wait time before reconnecting
	reconnectBaseWait = time.Second

	// maxReconnectWait caps the exponential backoff
	maxReconnectWait = time.Minute

	// connectTimeout for dial operations
	connectTimeout = 10 * time.Second
)

// WebsocketClient manages a WebSocket connection to the Hyperliquid API.
// It handles automatic reconnection, subscription management, and POST requests.
type WebsocketClient struct {
	url string

	// Connection state (atomic for lock-free reads)
	connected atomic.Bool

	// Connection and synchronization
	conn    *websocket.Conn
	connMu  sync.RWMutex  // Protects conn
	writeMu sync.Mutex    // Serializes writes
	cancel  func()        // Cancels current connection's goroutines
	wg      sync.WaitGroup

	// Subscription management
	subscriptions map[subKey]map[int]*subscriptionCallback
	subMu         sync.RWMutex
	nextSubID     atomic.Int32

	// Reconnection settings
	reconnectMu          sync.Mutex
	reconnectTimer       *time.Timer
	reconnectAttempts    int
	MaxReconnectAttempts int           // 0 = unlimited (default)
	ReconnectWait        time.Duration // Can be customized

	// POST request tracking
	nextPostID      atomic.Int32
	pendingRequests map[int]*pendingRequest
	pendingMu       sync.RWMutex

	// Lifecycle
	closed    atomic.Bool
	closedMu  sync.Mutex
	closeOnce sync.Once
}

type pendingRequest struct {
	responseChan chan WsPostResponseData
}

// NewWebsocketClient creates a new WebSocket client for the given base URL.
// The client is not connected until Connect() is called.
func NewWebsocketClient(baseURL string) *WebsocketClient {
	if baseURL == "" {
		baseURL = MainnetAPIURL
	}

	parsedURL, err := url.Parse(baseURL)
	if err != nil {
		log.Fatalf("invalid URL: %v", err)
	}
	parsedURL.Scheme = "wss"
	parsedURL.Path = "/ws"
	wsURL := parsedURL.String()

	return &WebsocketClient{
		url:             wsURL,
		subscriptions:   make(map[subKey]map[int]*subscriptionCallback),
		pendingRequests: make(map[int]*pendingRequest),
		ReconnectWait:   reconnectBaseWait,
	}
}

// Connect establishes the WebSocket connection.
// It's safe to call multiple times - subsequent calls return immediately if already connected.
func (w *WebsocketClient) Connect(ctx context.Context) error {
	if w.closed.Load() {
		return fmt.Errorf("client is closed")
	}

	if w.connected.Load() {
		return nil
	}

	w.connMu.Lock()
	defer w.connMu.Unlock()

	// Double-check after acquiring lock
	if w.conn != nil && w.connected.Load() {
		return nil
	}

	// Clean up any existing connection
	if w.conn != nil {
		_ = w.conn.Close()
		w.conn = nil
	}

	// Cancel previous goroutines if any
	if w.cancel != nil {
		w.cancel()
		w.wg.Wait()
	}

	// Dial new connection
	dialer := websocket.Dialer{
		ReadBufferSize:  16384,
		WriteBufferSize: 4096,
	}
	dialCtx, dialCancel := context.WithTimeout(ctx, connectTimeout)
	defer dialCancel()

	//nolint:bodyclose // WebSocket connections don't have response bodies to close
	conn, _, err := dialer.DialContext(dialCtx, w.url, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	w.conn = conn
	w.connected.Store(true)

	// Reset reconnection state on successful connection
	w.reconnectMu.Lock()
	w.reconnectAttempts = 0
	w.ReconnectWait = reconnectBaseWait
	w.reconnectMu.Unlock()

	// Create context for this connection's goroutines
	connCtx, cancel := context.WithCancel(context.Background())
	w.cancel = cancel

	// Start read and ping pumps (per-connection, like upstream)
	w.wg.Add(2)
	go w.readPump(connCtx)
	go w.pingPump(connCtx)

	// Resubscribe to all previous subscriptions
	if err := w.resubscribeAll(); err != nil {
		w.connected.Store(false)
		cancel()
		_ = conn.Close()
		w.conn = nil
		return fmt.Errorf("resubscribe failed: %w", err)
	}

	return nil
}

// Subscribe registers a callback for the given subscription.
// Returns a subscription ID that can be used to unsubscribe.
func (w *WebsocketClient) Subscribe(sub Subscription, callback func(WSMessage)) (int, error) {
	if callback == nil {
		return 0, fmt.Errorf("callback cannot be nil")
	}

	w.subMu.Lock()
	key := sub.key()
	id := int(w.nextSubID.Add(1))

	if w.subscriptions[key] == nil {
		w.subscriptions[key] = make(map[int]*subscriptionCallback)
	}

	w.subscriptions[key][id] = &subscriptionCallback{
		id:       id,
		callback: callback,
	}
	w.subMu.Unlock()

	// Send subscribe message (outside lock to avoid deadlock)
	if err := w.sendSubscribe(sub); err != nil {
		w.subMu.Lock()
		delete(w.subscriptions[key], id)
		w.subMu.Unlock()
		return 0, fmt.Errorf("subscribe: %w", err)
	}

	return id, nil
}

// Unsubscribe removes a subscription by ID.
func (w *WebsocketClient) Unsubscribe(sub Subscription, id int) error {
	w.subMu.Lock()
	key := sub.key()
	subs, ok := w.subscriptions[key]
	if !ok {
		w.subMu.Unlock()
		return fmt.Errorf("subscription not found")
	}

	if _, ok := subs[id]; !ok {
		w.subMu.Unlock()
		return fmt.Errorf("subscription ID not found")
	}

	delete(subs, id)

	shouldUnsubscribe := len(subs) == 0
	if shouldUnsubscribe {
		delete(w.subscriptions, key)
	}
	w.subMu.Unlock()

	if shouldUnsubscribe {
		if err := w.sendUnsubscribe(sub); err != nil {
			return fmt.Errorf("unsubscribe: %w", err)
		}
	}

	return nil
}

// Close shuts down the WebSocket client and releases all resources.
func (w *WebsocketClient) Close() error {
	var err error
	w.closeOnce.Do(func() {
		w.closed.Store(true)
		w.connected.Store(false)

		// Cancel reconnection timer
		w.reconnectMu.Lock()
		if w.reconnectTimer != nil {
			w.reconnectTimer.Stop()
			w.reconnectTimer = nil
		}
		w.reconnectMu.Unlock()

		// Clean up pending POST requests
		w.pendingMu.Lock()
		for id, pending := range w.pendingRequests {
			close(pending.responseChan)
			delete(w.pendingRequests, id)
		}
		w.pendingMu.Unlock()

		// Cancel goroutines and close connection
		w.connMu.Lock()
		if w.cancel != nil {
			w.cancel()
		}
		if w.conn != nil {
			err = w.conn.Close()
			w.conn = nil
		}
		w.connMu.Unlock()

		// Wait for goroutines to finish
		w.wg.Wait()
	})
	return err
}

// PostRequest sends a POST-style request over WebSocket and waits for response.
func (w *WebsocketClient) PostRequest(
	requestType string,
	payload any,
	timeout time.Duration,
) (*WsPostResponseData, error) {
	if !w.connected.Load() {
		return nil, fmt.Errorf("not connected")
	}

	id := int(w.nextPostID.Add(1))
	responseChan := make(chan WsPostResponseData, 1)

	pending := &pendingRequest{
		responseChan: responseChan,
	}

	w.pendingMu.Lock()
	w.pendingRequests[id] = pending
	w.pendingMu.Unlock()

	// Cleanup on exit
	defer func() {
		w.pendingMu.Lock()
		delete(w.pendingRequests, id)
		w.pendingMu.Unlock()
	}()

	// Send request
	request := WsPostRequest{
		Method: "post",
		ID:     id,
		Request: WsRequest{
			Type:    requestType,
			Payload: payload,
		},
	}

	if err := w.writeJSON(request); err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Wait for response with timer (no goroutine spawned)
	timer := time.NewTimer(timeout)
	defer timer.Stop()

	select {
	case response, ok := <-responseChan:
		if !ok {
			return nil, fmt.Errorf("request cancelled")
		}
		return &response, nil
	case <-timer.C:
		return nil, fmt.Errorf("request timeout")
	}
}

// PostInfoRequest sends an info request over WebSocket.
func (w *WebsocketClient) PostInfoRequest(
	payload map[string]any,
	timeout time.Duration,
) (json.RawMessage, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	resp, err := w.PostRequest("info", payload, timeout)
	if err != nil {
		return nil, err
	}

	if resp.Response.Type == "error" {
		return nil, fmt.Errorf("info request error: %s", string(resp.Response.Payload))
	}

	return resp.Response.Payload, nil
}

// PostActionRequest sends a signed action request over WebSocket.
func (w *WebsocketClient) PostActionRequest(
	action any,
	signature SignatureResult,
	nonce int64,
	vaultAddress string,
	timeout time.Duration,
) (json.RawMessage, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	payload := map[string]any{
		"action":    action,
		"nonce":     nonce,
		"signature": signature,
	}

	if vaultAddress != "" {
		payload["vaultAddress"] = vaultAddress
	} else {
		payload["vaultAddress"] = nil
	}

	resp, err := w.PostRequest("action", payload, timeout)
	if err != nil {
		return nil, err
	}

	if resp.Response.Type == "error" {
		return nil, fmt.Errorf("action request error: %s", string(resp.Response.Payload))
	}

	return resp.Response.Payload, nil
}

// readPump reads messages from the WebSocket connection.
// Runs for the lifetime of a single connection (context-aware, like upstream).
func (w *WebsocketClient) readPump(ctx context.Context) {
	defer w.wg.Done()
	defer w.handleDisconnect()

	// Grab conn once — if it changes, context will be cancelled and we exit.
	w.connMu.RLock()
	conn := w.conn
	w.connMu.RUnlock()
	if conn == nil {
		return
	}

	for {
		// Set read deadline to detect dead connections.
		_ = conn.SetReadDeadline(time.Now().Add(readDeadline))

		_, msg, err := conn.ReadMessage()
		if err != nil {
			if ctx.Err() == nil && !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			return // Exit pump, handleDisconnect will trigger reconnection
		}

		// Skip server hello
		if string(msg) == "Websocket connection established." {
			continue
		}

		// Parse and dispatch
		var wsMsg WSMessage
		if err := json.Unmarshal(msg, &wsMsg); err != nil {
			log.Printf("websocket message parse error: %v", err)
			continue
		}

		w.dispatch(wsMsg)
	}
}

// pingPump sends periodic ping messages to keep the connection alive.
// Runs for the lifetime of a single connection (context-aware, like upstream).
func (w *WebsocketClient) pingPump(ctx context.Context) {
	defer w.wg.Done()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := w.sendPing(); err != nil {
				log.Printf("ping error: %v", err)
				// Don't return here - readPump will detect the error
				// and handleDisconnect will trigger reconnection
			}
		}
	}
}

// handleDisconnect is called when the connection is lost.
// It marks the client as disconnected and schedules reconnection.
func (w *WebsocketClient) handleDisconnect() {
	if w.closed.Load() {
		return
	}

	w.connected.Store(false)

	// Close connection cleanly
	w.connMu.Lock()
	if w.conn != nil {
		_ = w.conn.Close()
		w.conn = nil
	}
	w.connMu.Unlock()

	// Schedule reconnection
	w.scheduleReconnect()
}

// scheduleReconnect schedules an asynchronous reconnection attempt with exponential backoff and jitter.
func (w *WebsocketClient) scheduleReconnect() {
	if w.closed.Load() {
		return
	}

	w.reconnectMu.Lock()
	defer w.reconnectMu.Unlock()

	// Stop existing timer
	if w.reconnectTimer != nil {
		w.reconnectTimer.Stop()
	}

	w.reconnectAttempts++
	attempts := w.reconnectAttempts

	// Check max attempts
	if w.MaxReconnectAttempts > 0 && attempts > w.MaxReconnectAttempts {
		log.Printf("Max reconnection attempts (%d) reached, giving up", w.MaxReconnectAttempts)
		return
	}

	// Calculate backoff with jitter (±20%)
	backoff := w.ReconnectWait * time.Duration(1<<(attempts-1))
	if backoff > maxReconnectWait {
		backoff = maxReconnectWait
	}
	jitter := time.Duration(float64(backoff) * 0.2 * (2*rand.Float64() - 1))
	delay := backoff + jitter
	if delay < time.Second {
		delay = time.Second
	}

	log.Printf("Reconnection attempt %d in %v...", attempts, delay)

	// Schedule reconnection (non-blocking, like timer-based approach)
	w.reconnectTimer = time.AfterFunc(delay, func() {
		if w.closed.Load() || w.connected.Load() {
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
		err := w.Connect(ctx)
		cancel()

		if err != nil {
			log.Printf("Reconnection attempt %d failed: %v", attempts, err)
			w.scheduleReconnect()
		} else {
			log.Printf("Reconnection successful after %d attempts", attempts)
		}
	})
}

// dispatch routes messages to appropriate callbacks.
func (w *WebsocketClient) dispatch(msg WSMessage) {
	// Handle pong responses (sent by Hyperliquid as JSON, not WebSocket protocol pongs)
	if msg.Channel == "pong" {
		// Pong received - connection is alive, nothing to do
		return
	}

	// Handle subscription confirmations
	if msg.Channel == "subscriptionResponse" {
		// Subscription confirmed by server, nothing to do
		return
	}

	// Handle POST responses
	if msg.Channel == "post" {
		var postResp WsPostResponseData
		if err := json.Unmarshal(msg.Data, &postResp); err != nil {
			log.Printf("failed to unmarshal post response: %v", err)
			return
		}

		w.pendingMu.RLock()
		pending, ok := w.pendingRequests[postResp.ID]
		w.pendingMu.RUnlock()

		if ok {
			select {
			case pending.responseChan <- postResp:
			default:
				// Channel closed (timeout) or full
			}
		}
		return
	}

	// Copy matching callbacks under lock
	w.subMu.RLock()
	var callbacks []func(WSMessage)
	for key, subs := range w.subscriptions {
		if matchSubscription(key, msg) {
			for _, sub := range subs {
				callbacks = append(callbacks, sub.callback)
			}
		}
	}
	w.subMu.RUnlock()

	// Execute callbacks without holding lock
	for _, cb := range callbacks {
		cb(msg)
	}
}

// resubscribeAll resends subscribe messages for all active subscriptions.
func (w *WebsocketClient) resubscribeAll() error {
	w.subMu.RLock()
	keys := make([]subKey, 0, len(w.subscriptions))
	for key, subs := range w.subscriptions {
		if len(subs) > 0 {
			keys = append(keys, key)
		}
	}
	w.subMu.RUnlock()

	for _, key := range keys {
		sub := Subscription{
			Type:     key.typ,
			Coin:     key.coin,
			User:     key.user,
			Interval: key.interval,
			Dex:      key.dex,
		}
		if err := w.sendSubscribe(sub); err != nil {
			return fmt.Errorf("resubscribe %s: %w", key.typ, err)
		}
	}
	return nil
}

func (w *WebsocketClient) sendSubscribe(sub Subscription) error {
	return w.writeJSON(WsCommand{
		Method:       "subscribe",
		Subscription: &sub,
	})
}

func (w *WebsocketClient) sendUnsubscribe(sub Subscription) error {
	return w.writeJSON(WsCommand{
		Method:       "unsubscribe",
		Subscription: &sub,
	})
}

func (w *WebsocketClient) sendPing() error {
	return w.writeJSON(WsCommand{Method: "ping"})
}

func (w *WebsocketClient) writeJSON(v any) error {
	// Marshal outside the lock so serialization doesn't block other writers
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	if !w.connected.Load() {
		return fmt.Errorf("not connected")
	}

	w.connMu.RLock()
	conn := w.conn
	w.connMu.RUnlock()

	if conn == nil {
		return fmt.Errorf("connection closed")
	}

	// WriteMessage is a single frame write — no NextWriter/Close dance
	return conn.WriteMessage(websocket.TextMessage, data)
}

// Convenience subscription methods

func (w *WebsocketClient) SubscribeToTrades(coin string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "trades", Coin: coin}, callback)
}

func (w *WebsocketClient) SubscribeToOrderbook(coin string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "l2Book", Coin: coin}, callback)
}

func (w *WebsocketClient) SubscribeToAllMids(callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "allMids"}, callback)
}

func (w *WebsocketClient) SubscribeToAllMidsWithDex(dex string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "allMids", Dex: dex}, callback)
}

func (w *WebsocketClient) SubscribeToUserEvents(user string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "userEvents", User: user}, callback)
}

func (w *WebsocketClient) SubscribeToUserFills(user string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "userFills", User: user}, callback)
}

func (w *WebsocketClient) SubscribeToCandles(coin, interval string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "candle", Coin: coin, Interval: interval}, callback)
}

func (w *WebsocketClient) SubscribeToOrderUpdates(user string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "orderUpdates", User: user}, callback)
}

func (w *WebsocketClient) SubscribeToUserFundings(user string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "userFundings", User: user}, callback)
}

func (w *WebsocketClient) SubscribeToUserNonFundingLedgerUpdates(user string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "userNonFundingLedgerUpdates", User: user}, callback)
}

func (w *WebsocketClient) SubscribeToWebData2(user string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "webData2", User: user}, callback)
}

func (w *WebsocketClient) SubscribeToBBO(coin string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "bbo", Coin: coin}, callback)
}

func (w *WebsocketClient) SubscribeToActiveAssetCtx(coin string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "activeAssetCtx", Coin: coin}, callback)
}

func (w *WebsocketClient) SubscribeToNotification(user string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "notification", User: user}, callback)
}

func (w *WebsocketClient) SubscribeToActiveAssetData(user, coin string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "activeAssetData", User: user, Coin: coin}, callback)
}

func (w *WebsocketClient) SubscribeToUserTwapSliceFills(user string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "userTwapSliceFills", User: user}, callback)
}

func (w *WebsocketClient) SubscribeToUserTwapHistory(user string, callback func(WSMessage)) (int, error) {
	return w.Subscribe(Subscription{Type: "userTwapHistory", User: user}, callback)
}

// matchSubscription checks if a message matches a subscription key.
func matchSubscription(key subKey, msg WSMessage) bool {
	// Channel matching
	// NOTE: Most subscription types have matching channel names, but there are exceptions:
	// - "userEvents" subscription sends messages on "user" channel (Hyperliquid API quirk)
	channelMatch := false
	switch key.typ {
	case "allMids":
		channelMatch = msg.Channel == "allMids"
	case "notification":
		channelMatch = msg.Channel == "notification"
	case "webData2":
		channelMatch = msg.Channel == "webData2"
	case "candle":
		channelMatch = msg.Channel == "candle"
	case "l2Book":
		channelMatch = msg.Channel == "l2Book"
	case "trades":
		channelMatch = msg.Channel == "trades"
	case "orderUpdates":
		channelMatch = msg.Channel == "orderUpdates"
	case "userEvents":
		// API quirk: "userEvents" subscription type receives messages on "user" channel
		channelMatch = msg.Channel == "user"
	case "userFills":
		channelMatch = msg.Channel == "userFills"
	case "userFundings":
		channelMatch = msg.Channel == "userFundings"
	case "userNonFundingLedgerUpdates":
		channelMatch = msg.Channel == "userNonFundingLedgerUpdates"
	case "activeAssetCtx":
		channelMatch = msg.Channel == "activeAssetCtx"
	case "activeAssetData":
		channelMatch = msg.Channel == "activeAssetData"
	case "userTwapSliceFills":
		channelMatch = msg.Channel == "userTwapSliceFills"
	case "userTwapHistory":
		channelMatch = msg.Channel == "userTwapHistory"
	case "bbo":
		channelMatch = msg.Channel == "bbo"
	default:
		return false
	}

	if !channelMatch {
		return false
	}

	// Early return if no additional filtering needed
	if key.coin == "" && key.user == "" {
		return true
	}

	// orderUpdates data is a JSON array, not an object — matching is purely by channel.
	// User is implicit from the subscription, not present in message data.
	if key.typ == "orderUpdates" {
		return true
	}

	// Single unmarshal for both coin and user matching
	var msgData struct {
		Coin string `json:"coin"`
		User string `json:"user"`
	}
	if err := json.Unmarshal(msg.Data, &msgData); err != nil {
		return false
	}

	// Coin matching
	if key.coin != "" && msgData.Coin != key.coin {
		return false
	}

	// User matching
	if key.user != "" {
		if !strings.EqualFold(msgData.User, key.user) {
			return false
		}
	}

	return true
}
