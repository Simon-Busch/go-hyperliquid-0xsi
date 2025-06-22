package hyperliquid

import (
	"crypto/ecdsa"
	"encoding/json"
	"time"
)

type Exchange struct {
	client       *Client
	privateKey   *ecdsa.PrivateKey
	vault        string
	accountAddr  string
	info         *Info
	expiresAfter *int64
}

func NewExchange(
	privateKey *ecdsa.PrivateKey,
	baseURL string,
	meta *Meta,
	vaultAddr, accountAddr string,
	spotMeta *SpotMeta,
) *Exchange {
	return &Exchange{
		client:      NewClient(baseURL),
		privateKey:  privateKey,
		vault:       vaultAddr,
		accountAddr: accountAddr,
		info:        NewInfo(baseURL, true, meta, spotMeta),
	}
}

// executeAction executes an action and unmarshals the response into the given result
func (e *Exchange) executeAction(action map[string]any, result any) error {
	timestamp := time.Now().UnixMilli()

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(resp, result); err != nil {
		return err
	}

	return nil
}

func (e *Exchange) postAction(
	action map[string]any,
	signature SignatureResult,
	nonce int64,
) ([]byte, error) {
	payload := map[string]any{
		"action":    action,
		"nonce":     nonce,
		"signature": signature,
	}

	if action["type"] != "usdClassTransfer" {
		payload["vaultAddress"] = e.vault
	} else {
		payload["vaultAddress"] = nil
	}

	// Add expiration time if set
	if e.expiresAfter != nil {
		payload["expiresAfter"] = *e.expiresAfter
	}

	return e.client.post("/exchange", payload)
}
