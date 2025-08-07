package hyperliquid

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
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
func (e *Exchange) executeAction(action any, result any) error {
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


	err = json.Unmarshal(resp, result)
	if err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return nil
}

func (e *Exchange) postAction(
	action any,
	signature SignatureResult,
	nonce int64,
) ([]byte, error) {
	payload := map[string]any{
		"action":    action,
		"nonce":     nonce,
		"signature": signature,
	}

	// Handle vault address based on action type
	if actionMap, ok := action.(map[string]any); ok {
		if actionMap["type"] == "usdClassTransfer" {
			// For usdClassTransfer, explicitly set vaultAddress to nil
			payload["vaultAddress"] = nil
		}
		// For all other action types, only include vaultAddress if it's not empty
		if e.vault != "" {
			payload["vaultAddress"] = e.vault
		}
	} else {
		// For struct types, we need to use reflection or type assertion
		// For now, assume it's not usdClassTransfer
		if e.vault != "" {
			payload["vaultAddress"] = e.vault
		}
	}

	return e.client.post("/exchange", payload)
}
