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
	// pingInterval is the interval for sending ping messages to keep WebSocket alive
	pingInterval = 10 * time.Second
)

type WebsocketClient struct {
	url           string
	conn          *websocket.Conn
	mu            sync.RWMutex
	writeMu       sync.Mutex
	subscriptions map[subKey]map[int]*subscriptionCallback
	nextSubID     atomic.Int32

	// Simple state management (like Python)
	running   atomic.Bool // Should the client be running?
	connected atomic.Bool // Is the connection established?

	// Goroutine lifecycle management
	wg sync.WaitGroup // Track goroutines for clean shutdown

	// Reconnection settings
	reconnectWait        time.Duration
	reconnectAttempts    atomic.Int32
	MaxReconnectAttempts int // Maximum reconnection attempts (0 = unlimited, default)
	reconnectTimer       *time.Timer
}

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

	wc := &WebsocketClient{
		url:           wsURL,
		subscriptions: make(map[subKey]map[int]*subscriptionCallback),
		reconnectWait: time.Second,
	}

	// Mark as running so goroutines will start properly
	wc.running.Store(true)

	// Start goroutines once - they'll live for the lifetime of the client (like Python daemon threads)
	wc.wg.Add(2)
	go wc.readLoop()
	go wc.pingLoop()

	return wc
}

func (w *WebsocketClient) Connect(ctx context.Context) error {
	// Check if already connected
	if w.connected.Load() {
		return nil
	}

	w.mu.Lock()

	// Double check after acquiring lock
	if w.conn != nil && w.connected.Load() {
		w.mu.Unlock()
		return nil
	}

	dialer := websocket.Dialer{}

	//nolint:bodyclose // WebSocket connections don't have response bodies to close
	conn, _, err := dialer.DialContext(ctx, w.url, nil)
	if err != nil {
		w.mu.Unlock()
		return fmt.Errorf("websocket dial: %w", err)
	}

	w.conn = conn
	w.mu.Unlock()

	// Mark as connected
	w.connected.Store(true)

	// Reset reconnection attempts on successful connection
	w.reconnectAttempts.Store(0)

	// Resubscribe to all previous subscriptions
	if err := w.resubscribeAll(); err != nil {
		// If resubscribe fails, clean up the connection
		w.mu.Lock()
		if w.conn != nil {
			_ = w.conn.Close()
			w.conn = nil
		}
		w.mu.Unlock()
		w.connected.Store(false)
		return fmt.Errorf("resubscribe failed: %w", err)
	}

	return nil
}

func (w *WebsocketClient) Subscribe(sub Subscription, callback func(WSMessage)) (int, error) {
	if callback == nil {
		return 0, fmt.Errorf("callback cannot be nil")
	}

	w.mu.Lock()
	key := sub.key()
	id := int(w.nextSubID.Add(1))

	if w.subscriptions[key] == nil {
		w.subscriptions[key] = make(map[int]*subscriptionCallback)
	}

	w.subscriptions[key][id] = &subscriptionCallback{
		id:       id,
		callback: callback,
	}
	w.mu.Unlock()

	// Send subscribe outside of lock to avoid deadlock with writeJSON
	if err := w.sendSubscribe(sub); err != nil {
		w.mu.Lock()
		delete(w.subscriptions[key], id)
		w.mu.Unlock()
		return 0, fmt.Errorf("subscribe: %w", err)
	}

	return id, nil
}

func (w *WebsocketClient) Unsubscribe(sub Subscription, id int) error {
	w.mu.Lock()
	key := sub.key()
	subs, ok := w.subscriptions[key]
	if !ok {
		w.mu.Unlock()
		return fmt.Errorf("subscription not found")
	}

	if _, ok := subs[id]; !ok {
		w.mu.Unlock()
		return fmt.Errorf("subscription ID not found")
	}

	delete(subs, id)

	shouldUnsubscribe := len(subs) == 0
	if shouldUnsubscribe {
		delete(w.subscriptions, key)
	}
	w.mu.Unlock()

	// Send unsubscribe outside of lock to avoid deadlock with writeJSON
	if shouldUnsubscribe {
		if err := w.sendUnsubscribe(sub); err != nil {
			return fmt.Errorf("unsubscribe: %w", err)
		}
	}

	return nil
}

func (w *WebsocketClient) Close() error {
	// Mark as not running
	w.running.Store(false)
	w.connected.Store(false)

	// Cancel any pending reconnection timer
	if w.reconnectTimer != nil {
		w.reconnectTimer.Stop()
	}

	// Close connection
	w.mu.Lock()
	if w.conn != nil {
		_ = w.conn.Close()
		w.conn = nil
	}
	w.mu.Unlock()

	// Wait for goroutines to exit (like Python's thread.join())
	w.wg.Wait()

	return nil
}

// Private methods

// readLoop runs for the lifetime of the client (like Python's daemon thread)
func (w *WebsocketClient) readLoop() {
	defer w.wg.Done()

	for w.running.Load() {
		// Idle while not connected
		if !w.connected.Load() {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		w.mu.RLock()
		conn := w.conn
		w.mu.RUnlock()

		if conn == nil {
			w.connected.Store(false)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		_, msg, err := conn.ReadMessage()
		if err != nil {
			if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
				log.Printf("websocket read error: %v", err)
			}
			// Mark as disconnected
			w.connected.Store(false)

			// Close connection
			w.mu.Lock()
			if w.conn != nil {
				_ = w.conn.Close()
				w.conn = nil
			}
			w.mu.Unlock()

			// Trigger reconnection
			if w.running.Load() {
				w.scheduleReconnect()
			}
			continue
		}

		// Skip server hello message
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

// pingLoop runs for the lifetime of the client (like Python's daemon thread)
func (w *WebsocketClient) pingLoop() {
	defer w.wg.Done()

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for w.running.Load() {
		<-ticker.C

		// Only send ping if connected
		if w.connected.Load() {
			if err := w.sendPing(); err != nil {
				log.Printf("ping error: %v", err)
				// Don't exit - readLoop will handle disconnection
			}
		}
	}
}

// dispatch executes callbacks (copies callbacks first to avoid holding lock - like Python)
func (w *WebsocketClient) dispatch(msg WSMessage) {
	// Copy callbacks under lock
	w.mu.RLock()
	var callbacks []func(WSMessage)
	for key, subs := range w.subscriptions {
		if matchSubscription(key, msg) {
			for _, sub := range subs {
				callbacks = append(callbacks, sub.callback)
			}
		}
	}
	w.mu.RUnlock()

	// Execute callbacks without holding lock (prevents callback from blocking other operations)
	for _, cb := range callbacks {
		cb(msg)
	}
}

// scheduleReconnect uses timer-based reconnection (like Python's threading.Timer)
func (w *WebsocketClient) scheduleReconnect() {
	attempts := w.reconnectAttempts.Add(1)

	// Check max attempts
	maxRetries := w.MaxReconnectAttempts
	if maxRetries > 0 && int(attempts) > maxRetries {
		log.Printf("Max reconnection attempts (%d) reached, giving up", maxRetries)
		w.running.Store(false)
		return
	}

	// Calculate backoff with jitter (like Python)
	backoff := w.reconnectWait * time.Duration(1<<(attempts-1))
	if backoff > time.Minute {
		backoff = time.Minute
	}
	// Add jitter (±20%)
	jitter := time.Duration(float64(backoff) * 0.2 * (2*rand.Float64() - 1))
	delay := backoff + jitter
	if delay < time.Second {
		delay = time.Second
	}

	log.Printf("Reconnection attempt %d (waiting %v)...", attempts, delay)

	// Stop existing timer if any
	if w.reconnectTimer != nil {
		w.reconnectTimer.Stop()
	}

	// Single-shot timer (like Python's threading.Timer)
	w.reconnectTimer = time.AfterFunc(delay, func() {
		if w.running.Load() && !w.connected.Load() {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := w.Connect(ctx)
			cancel()

			if err == nil {
				log.Printf("Reconnection successful after %d attempts", attempts)
				w.reconnectWait = time.Second // Reset backoff
			} else {
				log.Printf("Reconnection attempt %d failed: %v", attempts, err)
				// Schedule next attempt
				w.scheduleReconnect()
			}
		}
	})
}

func (w *WebsocketClient) resubscribeAll() error {
	// Copy subscriptions under lock to avoid concurrent map iteration and write
	w.mu.RLock()
	subsCopy := make(map[subKey]bool)
	for key, subs := range w.subscriptions {
		if len(subs) > 0 {
			subsCopy[key] = true
		}
	}
	w.mu.RUnlock()

	// Send subscribe messages outside of lock
	for key := range subsCopy {
		sub := Subscription{
			Type:     key.typ,
			Coin:     key.coin,
			User:     key.user,
			Interval: key.interval,
			Dex:      key.dex,
		}
		if err := w.sendSubscribe(sub); err != nil {
			return fmt.Errorf("resubscribe: %w", err)
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
	w.writeMu.Lock()
	defer w.writeMu.Unlock()

	// Check if connected
	if !w.connected.Load() {
		return fmt.Errorf("not connected")
	}

	// Hold read lock while using conn to prevent TOCTOU race
	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.conn == nil {
		return fmt.Errorf("connection closed")
	}

	return w.conn.WriteJSON(v)
}

func (w *WebsocketClient) SubscribeToTrades(coin string, callback func(WSMessage)) (int, error) {
	sub := Subscription{Type: "trades", Coin: coin}
	return w.Subscribe(sub, callback)
}

func (w *WebsocketClient) SubscribeToOrderbook(coin string, callback func(WSMessage)) (int, error) {
	sub := Subscription{Type: "l2Book", Coin: coin}
	return w.Subscribe(sub, callback)
}

// SubscribeToAllMids subscribes to all mid prices
func (w *WebsocketClient) SubscribeToAllMids(callback func(WSMessage)) (int, error) {
	sub := Subscription{Type: "allMids"}
	return w.Subscribe(sub, callback)
}

// SubscribeToAllMidsWithDex subscribes to all mid prices with specific DEX
func (w *WebsocketClient) SubscribeToAllMidsWithDex(dex string, callback func(WSMessage)) (int, error) {
	sub := Subscription{Type: "allMids", Dex: dex}
	return w.Subscribe(sub, callback)
}

// SubscribeToUserEvents subscribes to user events
func (w *WebsocketClient) SubscribeToUserEvents(
	user string,
	callback func(WSMessage),
) (int, error) {
	sub := Subscription{Type: "userEvents", User: user}
	return w.Subscribe(sub, callback)
}

// SubscribeToUserFills subscribes to user fills
func (w *WebsocketClient) SubscribeToUserFills(user string, callback func(WSMessage)) (int, error) {
	sub := Subscription{Type: "userFills", User: user}
	return w.Subscribe(sub, callback)
}

// SubscribeToCandles subscribes to candle data
func (w *WebsocketClient) SubscribeToCandles(
	coin, interval string,
	callback func(WSMessage),
) (int, error) {
	sub := Subscription{Type: "candle", Coin: coin, Interval: interval}
	return w.Subscribe(sub, callback)
}

// SubscribeToOrderUpdates subscribes to order updates for a specific user
func (w *WebsocketClient) SubscribeToOrderUpdates(user string, callback func(WSMessage)) (int, error) {
	sub := Subscription{Type: "orderUpdates", User: user}
	return w.Subscribe(sub, callback)
}

// SubscribeToUserFundings subscribes to user funding updates
func (w *WebsocketClient) SubscribeToUserFundings(
	user string,
	callback func(WSMessage),
) (int, error) {
	sub := Subscription{Type: "userFundings", User: user}
	return w.Subscribe(sub, callback)
}

// SubscribeToUserNonFundingLedgerUpdates subscribes to user non-funding ledger updates
func (w *WebsocketClient) SubscribeToUserNonFundingLedgerUpdates(
	user string,
	callback func(WSMessage),
) (int, error) {
	sub := Subscription{Type: "userNonFundingLedgerUpdates", User: user}
	return w.Subscribe(sub, callback)
}

// SubscribeToWebData2 subscribes to web data v2
func (w *WebsocketClient) SubscribeToWebData2(user string, callback func(WSMessage)) (int, error) {
	sub := Subscription{Type: "webData2", User: user}
	return w.Subscribe(sub, callback)
}

// SubscribeToBBO subscribes to best bid/offer data
func (w *WebsocketClient) SubscribeToBBO(coin string, callback func(WSMessage)) (int, error) {
	sub := Subscription{Type: "bbo", Coin: coin}
	return w.Subscribe(sub, callback)
}

// SubscribeToActiveAssetCtx subscribes to active asset context
func (w *WebsocketClient) SubscribeToActiveAssetCtx(
	coin string,
	callback func(WSMessage),
) (int, error) {
	sub := Subscription{Type: "activeAssetCtx", Coin: coin}
	return w.Subscribe(sub, callback)
}

// SubscribeToNotification subscribes to notifications for a specific user
func (w *WebsocketClient) SubscribeToNotification(
	user string,
	callback func(WSMessage),
) (int, error) {
	sub := Subscription{Type: "notification", User: user}
	return w.Subscribe(sub, callback)
}

// SubscribeToActiveAssetData subscribes to active asset data for a user and coin
func (w *WebsocketClient) SubscribeToActiveAssetData(
	user, coin string,
	callback func(WSMessage),
) (int, error) {
	sub := Subscription{Type: "activeAssetData", User: user, Coin: coin}
	return w.Subscribe(sub, callback)
}

// SubscribeToUserTwapSliceFills subscribes to user TWAP slice fills
func (w *WebsocketClient) SubscribeToUserTwapSliceFills(
	user string,
	callback func(WSMessage),
) (int, error) {
	sub := Subscription{Type: "userTwapSliceFills", User: user}
	return w.Subscribe(sub, callback)
}

// SubscribeToUserTwapHistory subscribes to user TWAP history
func (w *WebsocketClient) SubscribeToUserTwapHistory(
	user string,
	callback func(WSMessage),
) (int, error) {
	sub := Subscription{Type: "userTwapHistory", User: user}
	return w.Subscribe(sub, callback)
}

func matchSubscription(key subKey, msg WSMessage) bool {
	// First check if channel matches
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
		// Note: userEvents subscription sends messages on "user" channel, not "userEvents"
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

	// For subscriptions that include a coin, check if the coin matches
	if key.coin != "" {
		var msgData struct {
			Coin string `json:"coin"`
		}
		if err := json.Unmarshal(msg.Data, &msgData); err != nil {
			return false
		}
		if msgData.Coin != key.coin {
			return false
		}
	}

	// For subscriptions that include a user, check if the user matches
	// NOTE: orderUpdates messages don't include user in the data - user is implicit from subscription
	if key.user != "" && key.typ != "orderUpdates" {
		var msgData struct {
			User string `json:"user"`
		}
		if err := json.Unmarshal(msg.Data, &msgData); err != nil {
			return false
		}
		// Case-insensitive comparison for Ethereum addresses
		if !strings.EqualFold(msgData.User, key.user) {
			return false
		}
	}

	return true
}
