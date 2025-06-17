package hyperliquid

import (
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"time"

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

// SignUsdClassTransferAction signs USD class transfer action
func SignUsdClassTransferAction(
	privateKey *ecdsa.PrivateKey,
	amount float64,
	toPerp bool,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":   "usdClassTransfer",
		"amount": amount,
		"toPerp": toPerp,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// SignSpotTransferAction signs spot transfer action
func SignSpotTransferAction(
	privateKey *ecdsa.PrivateKey,
	amount float64,
	destination, token string,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":        "spotTransfer",
		"amount":      amount,
		"destination": destination,
		"token":       token,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// SignUsdTransferAction signs USD transfer action
func SignUsdTransferAction(
	privateKey *ecdsa.PrivateKey,
	amount float64,
	destination string,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":        "usdTransfer",
		"amount":      amount,
		"destination": destination,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// SignPerpDexClassTransferAction signs perp dex class transfer action
func SignPerpDexClassTransferAction(
	privateKey *ecdsa.PrivateKey,
	dex, token string,
	amount float64,
	toPerp bool,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":   "perpDexClassTransfer",
		"dex":    dex,
		"token":  token,
		"amount": amount,
		"toPerp": toPerp,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// SignTokenDelegateAction signs token delegate action
func SignTokenDelegateAction(
	privateKey *ecdsa.PrivateKey,
	token string,
	amount float64,
	validatorAddress string,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":             "tokenDelegate",
		"token":            token,
		"amount":           amount,
		"validatorAddress": validatorAddress,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// SignWithdrawFromBridgeAction signs withdraw from bridge action
func SignWithdrawFromBridgeAction(
	privateKey *ecdsa.PrivateKey,
	destination string,
	amount, fee float64,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":        "withdrawFromBridge",
		"destination": destination,
		"amount":      amount,
		"fee":         fee,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// SignAgent signs agent approval action
func SignAgent(
	privateKey *ecdsa.PrivateKey,
	agentAddress, agentName string,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":         "approveAgent",
		"agentAddress": agentAddress,
		"agentName":    agentName,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// SignApproveBuilderFee signs approve builder fee action
func SignApproveBuilderFee(
	privateKey *ecdsa.PrivateKey,
	builderAddress string,
	maxFeeRate float64,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":           "approveBuilderFee",
		"builderAddress": builderAddress,
		"maxFeeRate":     maxFeeRate,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// SignConvertToMultiSigUserAction signs convert to multi-sig user action
func SignConvertToMultiSigUserAction(
	privateKey *ecdsa.PrivateKey,
	signers []string,
	threshold int,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":      "convertToMultiSigUser",
		"signers":   signers,
		"threshold": threshold,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// SignMultiSigAction signs multi-signature action
func SignMultiSigAction(
	privateKey *ecdsa.PrivateKey,
	innerAction map[string]any,
	signers []string,
	signatures []string,
	timestamp int64,
	isMainnet bool,
) (string, error) {
	action := map[string]any{
		"type":       "multiSig",
		"action":     innerAction,
		"signers":    signers,
		"signatures": signatures,
	}

	return SignL1Action(privateKey, action, "", timestamp, isMainnet)
}

// Utility function to convert float to USD integer representation
func FloatToUsdInt(value float64) int {
	// Convert float USD to integer representation (assuming 6 decimals for USDC)
	return int(value * 1e6)
}

// GetTimestampMs returns current timestamp in milliseconds
func GetTimestampMs() int64 {
	return time.Now().UnixMilli()
}
