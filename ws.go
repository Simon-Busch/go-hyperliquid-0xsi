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
	pingInterval = 50 * time.Second
	// gracefulCloseTimeout is the timeout for graceful WebSocket close
	gracefulCloseTimeout = 10 * time.Second
)

type WebsocketClient struct {
	url           string
	conn          *websocket.Conn
	mu            sync.RWMutex
	writeMu       sync.Mutex
	subscriptions map[subKey]map[int]*subscriptionCallback
	nextSubID     atomic.Int32
	done          chan struct{}
	reconnectWait time.Duration
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

	return &WebsocketClient{
		url:           wsURL,
		subscriptions: make(map[subKey]map[int]*subscriptionCallback),
		done:          make(chan struct{}),
		reconnectWait: time.Second,
	}
}

func (w *WebsocketClient) Connect(ctx context.Context) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		return nil
	}

	dialer := websocket.Dialer{}

	//nolint:bodyclose // WebSocket connections don't have response bodies to close
	conn, _, err := dialer.DialContext(ctx, w.url, nil)
	if err != nil {
		return fmt.Errorf("websocket dial: %w", err)
	}

	w.conn = conn

	go w.readPump(ctx)
	go w.pingPump(ctx)

	return w.resubscribeAll()
}

func (w *WebsocketClient) Subscribe(sub Subscription, callback func(WSMessage)) (int, error) {
	if callback == nil {
		return 0, fmt.Errorf("callback cannot be nil")
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	key := sub.key()
	id := int(w.nextSubID.Add(1))

	if w.subscriptions[key] == nil {
		w.subscriptions[key] = make(map[int]*subscriptionCallback)
	}

	w.subscriptions[key][id] = &subscriptionCallback{
		id:       id,
		callback: callback,
	}

	if err := w.sendSubscribe(sub); err != nil {
		delete(w.subscriptions[key], id)
		return 0, fmt.Errorf("subscribe: %w", err)
	}

	return id, nil
}

func (w *WebsocketClient) Unsubscribe(sub Subscription, id int) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	key := sub.key()
	subs, ok := w.subscriptions[key]
	if !ok {
		return fmt.Errorf("subscription not found")
	}

	if _, ok := subs[id]; !ok {
		return fmt.Errorf("subscription ID not found")
	}

	delete(subs, id)

	if len(subs) == 0 {
		delete(w.subscriptions, key)
		if err := w.sendUnsubscribe(sub); err != nil {
			return fmt.Errorf("unsubscribe: %w", err)
		}
	}

	return nil
}

func (w *WebsocketClient) Close() error {
	close(w.done)

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.conn != nil {
		return w.conn.Close()
	}
	return nil
}

// Private methods

func (w *WebsocketClient) readPump(ctx context.Context) {
	defer func() {
		w.mu.Lock()
		if w.conn != nil {
			_ = w.conn.Close() // Ignore close error in defer
			w.conn = nil
		}
		w.mu.Unlock()
	}()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.done:
			return
		default:
			_, msg, err := w.conn.ReadMessage()
			if err != nil {
				if !websocket.IsCloseError(err, websocket.CloseNormalClosure) {
					log.Printf("websocket read error: %v", err)
					w.reconnect()
				}
				return
			}

			if string(msg) == "Websocket connection established." {
				continue
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
			if err := w.sendPing(); err != nil {
				log.Printf("ping error: %v", err)
				w.reconnect()
				return
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
	for {
		select {
		case <-w.done:
			return
		default:
			ctx, cancel := context.WithTimeout(context.Background(), gracefulCloseTimeout)
			err := w.Connect(ctx)
			cancel()
			if err == nil {
				return
			}
			time.Sleep(w.reconnectWait)
			w.reconnectWait *= 2
			if w.reconnectWait > time.Minute {
				w.reconnectWait = time.Minute
			}
		}
	}
}

func (w *WebsocketClient) resubscribeAll() error {
	for key, subs := range w.subscriptions {
		if len(subs) > 0 {
			sub := Subscription{
				Type:     key.typ,
				Coin:     key.coin,
				User:     key.user,
				Interval: key.interval,
			}
			if err := w.sendSubscribe(sub); err != nil {
				return fmt.Errorf("resubscribe: %w", err)
			}
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

func matchSubscription(key subKey, msg WSMessage) bool {
	switch key.typ {
	case "l2Book":
		return msg.Channel == "l2Book"
	case "trades":
		return msg.Channel == "trades"
	case "candle":
		return msg.Channel == "candle"
	default:
		return false
	}
}
