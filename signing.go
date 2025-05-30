package hyperliquid

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

func SignL1Action(
	privateKey *ecdsa.PrivateKey,
	action any,
	vaultAddress string,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	chainID := "0x1"
	if !isMainnet {
		chainID = "0x66eee"
	}

	actionJSON, err := json.Marshal(action)
	if err != nil {
		return "", fmt.Errorf("failed to marshal action: %w", err)
	}

	msg := map[string]any{
		"action":       string(actionJSON),
		"chainId":      chainID,
		"nonce":        timestamp,
		"vaultAddress": vaultAddress,
	}

	msgJSON, err := json.Marshal(msg)
	if err != nil {
		return "", fmt.Errorf("failed to marshal message: %w", err)
	}

	hash := crypto.Keccak256Hash(msgJSON)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to sign message: %w", err)
	}

	// Convert to Ethereum signature format
	signature[64] += 27

	return hexutil.Encode(signature), nil
}

func OrderRequestToWire(req OrderRequest, asset int) OrderWire {
	wire := OrderWire{
		Asset:      asset,
		IsBuy:      req.IsBuy,
		LimitPx:    req.LimitPx,
		ReduceOnly: req.ReduceOnly,
		Size:       req.Size,
	}

	if req.OrderType.Limit != nil {
		wire.OrderType = "Limit"
		wire.Tif = req.OrderType.Limit.Tif
	} else if req.OrderType.Trigger != nil {
		wire.OrderType = "Trigger"
		wire.TriggerPx = req.OrderType.Trigger.TriggerPx
		wire.IsMarket = req.OrderType.Trigger.IsMarket
		wire.Tpsl = req.OrderType.Trigger.Tpsl
	}

	if req.Cloid != nil {
		wire.Cloid = *req.Cloid
	}

	return wire
}

// ... Add other signing helper functions
