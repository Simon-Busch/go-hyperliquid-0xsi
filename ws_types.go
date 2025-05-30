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
}

type subKey struct {
	typ      string
	coin     string
	user     string
	interval string
}

func (s Subscription) key() subKey {
	return subKey{
		typ:      s.Type,
		coin:     s.Coin,
		user:     s.User,
		interval: s.Interval,
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
