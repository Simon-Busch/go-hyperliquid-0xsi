package hyperliquid

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// pingInterval is the interval for sending ping messages to keep WebSocket alive
	pingInterval = 10 * time.Second
	// gracefulCloseTimeout is the timeout for graceful WebSocket close
	gracefulCloseTimeout = 10 * time.Second
)

// Connection states
const (
	stateDisconnected int32 = iota
	stateConnecting
	stateConnected
	stateReconnecting
	stateClosed
)

type WebsocketClient struct {
	url           string
	conn          *websocket.Conn
	mu            sync.RWMutex
	writeMu       sync.Mutex
	subscriptions map[subKey]map[int]*subscriptionCallback
	nextSubID     atomic.Int32
	done          chan struct{}
	closed        atomic.Bool
	reconnectWait time.Duration

	// Connection state management
	state              atomic.Int32       // Current connection state
	connReady          chan struct{}      // Signaled when connection is established and ready
	connReadyOnce      sync.Once          // Ensures connReady is closed only once per connection
	stopReconnect      chan struct{}      // Signal to stop reconnection attempts
	connCtx            context.Context    // Connection-scoped context
	connCancel         context.CancelFunc // Cancel function for connection context
	MaxReconnectAttempts int              // Maximum reconnection attempts (0 = unlimited, default)
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
		done:          make(chan struct{}),
		reconnectWait: time.Second,
		connReady:     make(chan struct{}),
		stopReconnect: make(chan struct{}),
	}
	wc.state.Store(stateDisconnected)
	return wc
}

func (w *WebsocketClient) Connect(ctx context.Context) error {
	// Check current state
	currentState := w.state.Load()
	if currentState == stateClosed {
		return fmt.Errorf("client is closed")
	}
	if currentState == stateConnected || currentState == stateConnecting {
		// Already connected or connecting
		return nil
	}

	w.mu.Lock()

	// Double check after acquiring lock
	if w.conn != nil && w.state.Load() == stateConnected {
		w.mu.Unlock()
		return nil
	}

	// Set state to connecting
	w.state.Store(stateConnecting)

	dialer := websocket.Dialer{}

	//nolint:bodyclose // WebSocket connections don't have response bodies to close
	conn, _, err := dialer.DialContext(ctx, w.url, nil)
	if err != nil {
		w.state.Store(stateDisconnected)
		w.mu.Unlock()
		return fmt.Errorf("websocket dial: %w", err)
	}

	w.conn = conn

	// Cancel previous connection context if it exists
	if w.connCancel != nil {
		w.connCancel()
	}

	// Create connection-scoped context (not tied to caller's context)
	w.connCtx, w.connCancel = context.WithCancel(context.Background())

	// Create new connReady channel and reset sync.Once for this connection
	w.connReady = make(chan struct{})
	w.connReadyOnce = sync.Once{}

	// Release lock before launching goroutines to avoid deadlock
	w.mu.Unlock()

	// Launch goroutines with connection-scoped context
	go w.readPump(w.connCtx)
	go w.pingPump(w.connCtx)

	// Mark as connected - readPump will handle validation and set state properly
	// when first message arrives. If connection fails, readPump triggers reconnection.
	w.state.Store(stateConnected)

	// Resubscribe to all previous subscriptions
	if err := w.resubscribeAll(); err != nil {
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
	// Only close the done channel once using atomic flag
	if !w.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	// Set state to closed
	w.state.Store(stateClosed)

	// Cancel connection context to stop goroutines
	if w.connCancel != nil {
		w.connCancel()
	}

	// Signal channels
	close(w.done)
	select {
	case <-w.stopReconnect:
		// Already closed
	default:
		close(w.stopReconnect)
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}

// Private methods

func (w *WebsocketClient) readPump(ctx context.Context) {
	connectionReady := false
	defer func() {
		// Mark state as disconnected
		oldState := w.state.Swap(stateDisconnected)

		w.mu.Lock()
		if w.conn != nil {
			_ = w.conn.Close() // Ignore close error in defer
			w.conn = nil
		}
		w.mu.Unlock()

		// If we were connected and this wasn't a normal close, try to reconnect
		if oldState == stateConnected && !w.closed.Load() {
			go w.reconnect()
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		default:
			w.mu.RLock()
			conn := w.conn
			w.mu.RUnlock()

			if conn == nil {
				return
			}

			_, msg, err := conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					log.Printf("websocket read error: %v", err)
				}
				return
			}

			// Handle the initial connection message
			if string(msg) == "Websocket connection established." {
				if !connectionReady {
					connectionReady = true
					w.state.Store(stateConnected)
					// Signal that connection is ready using sync.Once
					w.connReadyOnce.Do(func() {
						close(w.connReady)
					})
				}
				continue
			}

			// Ensure we're in connected state (in case we didn't get the initial message)
			if !connectionReady {
				connectionReady = true
				w.state.Store(stateConnected)
				// Signal that connection is ready using sync.Once
				w.connReadyOnce.Do(func() {
					close(w.connReady)
				})
			}

			var wsMsg WSMessage
			if err := json.Unmarshal(msg, &wsMsg); err != nil {
				log.Printf("websocket message parse error: %v", err)
				continue
			}

			w.dispatch(wsMsg)
		}
	}
}

func (w *WebsocketClient) pingPump(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.done:
			return
		case <-ctx.Done():
			return
		case <-ticker.C:
			state := w.state.Load()
			// Only send ping if we're in connected state
			if state == stateConnected {
				if err := w.sendPing(); err != nil {
					log.Printf("ping error: %v", err)
					// Don't trigger reconnect here - let readPump handle it
					return
				}
			} else if state == stateConnecting || state == stateReconnecting {
				// Log when pings are skipped during connection states
				log.Printf("ping skipped: connection in state %d (not connected)", state)
			}
		}
	}
}

func (w *WebsocketClient) dispatch(msg WSMessage) {
	w.mu.RLock()
	defer w.mu.RUnlock()

	for key, subs := range w.subscriptions {
		if matchSubscription(key, msg) {
			for _, sub := range subs {
				sub.callback(msg)
			}
		}
	}
}

func (w *WebsocketClient) reconnect() {
	// Check if we should even try to reconnect
	if w.closed.Load() {
		return
	}

	// Try to transition to reconnecting state
	if !w.state.CompareAndSwap(stateDisconnected, stateReconnecting) {
		// Already reconnecting or in another state
		return
	}

	log.Printf("Starting reconnection attempts...")

	backoff := w.reconnectWait
	maxRetries := w.MaxReconnectAttempts
	if maxRetries == 0 {
		// 0 means unlimited attempts
		log.Printf("Reconnection attempts: unlimited")
	} else {
		log.Printf("Max reconnection attempts: %d", maxRetries)
	}
	retryCount := 0

	for {
		select {
		case <-w.done:
			w.state.Store(stateDisconnected)
			return
		case <-w.stopReconnect:
			w.state.Store(stateDisconnected)
			return
		default:
			if maxRetries > 0 && retryCount >= maxRetries {
				log.Printf("Max reconnection attempts (%d) reached, giving up", maxRetries)
				w.state.Store(stateDisconnected)
				return
			}

			retryCount++

			// Wait BEFORE dialing to prevent reconnection storms
			if maxRetries > 0 {
				log.Printf("Reconnection attempt %d/%d (waiting %v)...", retryCount, maxRetries, backoff)
			} else {
				log.Printf("Reconnection attempt %d (waiting %v)...", retryCount, backoff)
			}
			time.Sleep(backoff)

			// Reset state to disconnected before attempting connect
			w.state.Store(stateDisconnected)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			err := w.Connect(ctx)
			cancel()

			if err == nil {
				log.Printf("Reconnection successful after %d attempts", retryCount)
				// Reset backoff on success
				w.reconnectWait = time.Second
				return
			}

			log.Printf("Reconnection attempt %d failed: %v", retryCount, err)

			// Exponential backoff for NEXT attempt
			backoff *= 2
			if backoff > time.Minute {
				backoff = time.Minute
			}

			// Set state back to reconnecting for next iteration
			w.state.Store(stateReconnecting)
		}
	}
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

	// Check connection state
	state := w.state.Load()
	if state != stateConnected {
		return fmt.Errorf("connection not ready (state: %d)", state)
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

// SubscribeToOrderUpdates subscribes to order updates
func (w *WebsocketClient) SubscribeToOrderUpdates(callback func(WSMessage)) (int, error) {
	sub := Subscription{Type: "orderUpdates"}
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
	switch key.typ {
	case "allMids":
		return msg.Channel == "allMids"
	case "notification":
		return msg.Channel == "notification"
	case "webData2":
		return msg.Channel == "webData2"
	case "candle":
		return msg.Channel == "candle"
	case "l2Book":
		return msg.Channel == "l2Book"
	case "trades":
		return msg.Channel == "trades"
	case "orderUpdates":
		return msg.Channel == "orderUpdates"
	case "userEvents":
		return msg.Channel == "userEvents"
	case "userFills":
		return msg.Channel == "userFills"
	case "userFundings":
		return msg.Channel == "userFundings"
	case "userNonFundingLedgerUpdates":
		return msg.Channel == "userNonFundingLedgerUpdates"
	case "activeAssetCtx":
		return msg.Channel == "activeAssetCtx"
	case "activeAssetData":
		return msg.Channel == "activeAssetData"
	case "userTwapSliceFills":
		return msg.Channel == "userTwapSliceFills"
	case "userTwapHistory":
		return msg.Channel == "userTwapHistory"
	case "bbo":
		return msg.Channel == "bbo"
	default:
		return false
	}
}
