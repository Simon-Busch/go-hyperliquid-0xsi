package hyperliquid

import (
	"encoding/json"
)

//go:generate easyjson -all ws_types.go

type WSMessage struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type Subscription struct {
	Type     string `json:"type"`
	Coin     string `json:"coin,omitempty"`
	User     string `json:"user,omitempty"`
	Interval string `json:"interval,omitempty"`
	Dex      string `json:"dex,omitempty"`
}

type subKey struct {
	typ      string
	coin     string
	user     string
	interval string
	dex      string
}

func (s Subscription) key() subKey {
	return subKey{
		typ:      s.Type,
		coin:     s.Coin,
		user:     s.User,
		interval: s.Interval,
		dex:      s.Dex,
	}
}

type WsCommand struct {
	Method       string        `json:"method"`
	Subscription *Subscription `json:"subscription,omitempty"`
}

type subscriptionCallback struct {
	id       int
	callback func(WSMessage)
}

// WebSocket POST request types

// WsPostRequest is the request structure for WebSocket POST requests
type WsPostRequest struct {
	Method  string    `json:"method"` // Always "post"
	ID      int       `json:"id"`
	Request WsRequest `json:"request"`
}

// WsRequest wraps the actual request payload
type WsRequest struct {
	Type    string `json:"type"` // "info" or "action"
	Payload any    `json:"payload"`
}

// WsPostResponseData is the data field of a POST response message
type WsPostResponseData struct {
	ID       int        `json:"id"`
	Response WsResponse `json:"response"`
}

// WsResponse contains the actual response payload
type WsResponse struct {
	Type    string          `json:"type"` // "info", "action", or "error"
	Payload json.RawMessage `json:"payload"`
}
