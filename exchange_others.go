package hyperliquid

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/ethereum/go-ethereum/crypto"
)

func (e *Exchange) UpdateLeverage(leverage int, name string, isCross bool) (*UserState, error) {
	leverageType := "isolated"
	if isCross {
		leverageType = "cross"
	}

	action := map[string]any{
		"type":  "updateLeverage",
		"asset": e.info.NameToAsset(name),
		"leverage": map[string]any{
			"type":  leverageType,
			"value": leverage,
		},
	}

	var result UserState
	if err := e.executeAction(action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (e *Exchange) UpdateIsolatedMargin(amount float64, name string) (*UserState, error) {
	action := map[string]any{
		"type":  "updateIsolatedMargin",
		"asset": e.info.NameToAsset(name),
		"isBuy": amount > 0,
		"ntli":  abs(amount),
	}

	var result UserState
	if err := e.executeAction(action, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetExpiresAfter sets the expiration time for actions
// If expiresAfter is nil, actions will not have an expiration time
// If expiresAfter is set, actions will include this expiration timestamp
func (e *Exchange) SetExpiresAfter(expiresAfter *int64) {
	e.expiresAfter = expiresAfter
}

// SlippagePrice calculates the slippage price for market orders
func (e *Exchange) SlippagePrice(
	name string,
	isBuy bool,
	slippage float64,
	px *float64,
) (float64, error) {
	coin := e.info.nameToCoin[name]
	var price float64

	if px != nil {
		price = *px
	} else {
		// Get midprice
		mids, err := e.info.AllMids()
		if err != nil {
			return 0, err
		}
		if midPriceStr, exists := mids[coin]; exists {
			price = parseFloat(midPriceStr)
		} else {
			return 0, fmt.Errorf("could not get mid price for coin: %s", coin)
		}
	}

	asset := e.info.coinToAsset[coin]
	isSpot := asset >= 10000

	// Calculate slippage
	if isBuy {
		price *= (1 + slippage)
	} else {
		price *= (1 - slippage)
	}

	// Round to appropriate decimals
	decimals := 6
	if isSpot {
		decimals = 8
	}
	szDecimals := e.info.assetToDecimal[asset]

	return roundToDecimals(price, decimals-szDecimals), nil
}

// ScheduleCancel schedules cancellation of all open orders
func (e *Exchange) ScheduleCancel(scheduleTime *int64) (*ScheduleCancelResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "scheduleCancel",
	}
	if scheduleTime != nil {
		action["time"] = *scheduleTime
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result ScheduleCancelResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SetReferrer sets a referral code
func (e *Exchange) SetReferrer(code string) (*SetReferrerResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "setReferrer",
		"code": code,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address for referrer
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SetReferrerResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CreateSubAccount creates a new sub-account
func (e *Exchange) CreateSubAccount(name string) (*CreateSubAccountResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "createSubAccount",
		"name": name,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address for sub-account creation
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result CreateSubAccountResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UsdClassTransfer transfers between USD classes
func (e *Exchange) UsdClassTransfer(amount float64, toPerp bool) (*TransferResponse, error) {
	timestamp := time.Now().UnixMilli()

	strAmount := formatFloat(amount)
	if e.vault != "" {
		strAmount += " subaccount:" + e.vault
	}

	action := map[string]any{
		"type":   "usdClassTransfer",
		"amount": strAmount,
		"toPerp": toPerp,
		"nonce":  timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result TransferResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SubAccountTransfer transfers funds to/from sub-account
func (e *Exchange) SubAccountTransfer(
	subAccountUser string,
	isDeposit bool,
	usd int,
) (*TransferResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":           "subAccountTransfer",
		"subAccountUser": subAccountUser,
		"isDeposit":      isDeposit,
		"usd":            usd,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result TransferResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// VaultUsdTransfer transfers to/from vault
func (e *Exchange) VaultUsdTransfer(
	vaultAddress string,
	isDeposit bool,
	usd int,
) (*TransferResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":         "vaultTransfer",
		"vaultAddress": vaultAddress,
		"isDeposit":    isDeposit,
		"usd":          usd,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result TransferResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UsdTransfer transfers USD to another address
func (e *Exchange) UsdTransfer(amount float64, destination string) (*TransferResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"destination": destination,
		"amount":      formatFloat(amount),
		"time":        timestamp,
		"type":        "usdSend",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result TransferResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotTransfer transfers spot tokens to another address
func (e *Exchange) SpotTransfer(
	amount float64,
	destination, token string,
) (*TransferResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"destination": destination,
		"amount":      formatFloat(amount),
		"token":       token,
		"time":        timestamp,
		"type":        "spotSend",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result TransferResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// UseBigBlocks enables or disables big blocks
func (e *Exchange) UseBigBlocks(enable bool) (*ApprovalResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":           "evmUserModify",
		"usingBigBlocks": enable,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result ApprovalResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PerpDexClassTransfer transfers tokens between perp dex classes
func (e *Exchange) PerpDexClassTransfer(
	dex, token string,
	amount float64,
	toPerp bool,
) (*TransferResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":   "perpDexClassTransfer",
		"dex":    dex,
		"token":  token,
		"amount": amount,
		"toPerp": toPerp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result TransferResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SubAccountSpotTransfer transfers spot tokens to/from sub-account
func (e *Exchange) SubAccountSpotTransfer(
	subAccountUser string,
	isDeposit bool,
	token string,
	amount float64,
) (*TransferResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":           "subAccountSpotTransfer",
		"subAccountUser": subAccountUser,
		"isDeposit":      isDeposit,
		"token":          token,
		"amount":         amount,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result TransferResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// TokenDelegate delegates tokens for staking
func (e *Exchange) TokenDelegate(
	validator string,
	wei int,
	isUndelegate bool,
) (*TransferResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":         "tokenDelegate",
		"validator":    validator,
		"wei":          wei,
		"isUndelegate": isUndelegate,
		"nonce":        timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result TransferResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// WithdrawFromBridge withdraws tokens from bridge
func (e *Exchange) WithdrawFromBridge(
	amount float64,
	destination string,
) (*TransferResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":        "withdraw3",
		"destination": destination,
		"amount":      fmt.Sprintf("%.6f", amount),
		"time":        timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result TransferResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ApproveAgent approves an agent to trade on behalf of the user
// Returns the result and the generated agent private key
func (e *Exchange) ApproveAgent(name *string) (*AgentApprovalResponse, string, error) {
	// Generate agent key
	agentBytes := make([]byte, 32)
	if _, err := rand.Read(agentBytes); err != nil {
		return nil, "", fmt.Errorf("failed to generate agent key: %w", err)
	}
	agentKey := "0x" + hex.EncodeToString(agentBytes)

	privateKey, err := crypto.HexToECDSA(agentKey[2:])
	if err != nil {
		return nil, "", fmt.Errorf("failed to create private key: %w", err)
	}

	agentAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":         "approveAgent",
		"agentAddress": agentAddress,
		"nonce":        timestamp,
	}

	if name != nil {
		action["agentName"] = *name
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, "", err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, "", err
	}

	var result AgentApprovalResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, "", err
	}
	return &result, agentKey, nil
}

// ApproveBuilderFee approves builder fee payment
func (e *Exchange) ApproveBuilderFee(builder string, maxFeeRate string) (*ApprovalResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":       "approveBuilderFee",
		"builder":    builder,
		"maxFeeRate": maxFeeRate,
		"nonce":      timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result ApprovalResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ConvertToMultiSigUser converts account to multi-signature user
func (e *Exchange) ConvertToMultiSigUser(
	authorizedUsers []string,
	threshold int,
) (*MultiSigConversionResponse, error) {
	timestamp := time.Now().UnixMilli()

	// Sort users as done in Python
	sort.Strings(authorizedUsers)

	signers := map[string]any{
		"authorizedUsers": authorizedUsers,
		"threshold":       threshold,
	}

	signersJSON, err := json.Marshal(signers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal signers: %w", err)
	}

	action := map[string]any{
		"type":    "convertToMultiSigUser",
		"signers": string(signersJSON),
		"nonce":   timestamp,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result MultiSigConversionResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Spot Deploy Methods

// SpotDeployRegisterToken registers a new spot token
func (e *Exchange) SpotDeployRegisterToken(
	tokenName string,
	szDecimals int,
	weiDecimals int,
	maxGas int,
	fullName string,
) (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "spotDeploy",
		"registerToken2": map[string]any{
			"spec": map[string]any{
				"name":        tokenName,
				"szDecimals":  szDecimals,
				"weiDecimals": weiDecimals,
			},
			"maxGas":   maxGas,
			"fullName": fullName,
		},
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		"", // No vault address for spot deploy
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotDeployUserGenesis initializes user genesis for spot trading
func (e *Exchange) SpotDeployUserGenesis(balances map[string]float64) (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":     "spotDeployUserGenesis",
		"balances": balances,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotDeployEnableFreezePrivilege enables freeze privilege for spot deployer
func (e *Exchange) SpotDeployEnableFreezePrivilege() (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "spotDeployEnableFreezePrivilege",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotDeployFreezeUser freezes a user in spot trading
func (e *Exchange) SpotDeployFreezeUser(userAddress string) (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":        "spotDeployFreezeUser",
		"userAddress": userAddress,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotDeployRevokeFreezePrivilege revokes freeze privilege for spot deployer
func (e *Exchange) SpotDeployRevokeFreezePrivilege() (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "spotDeployRevokeFreezePrivilege",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotDeployGenesis initializes spot genesis
func (e *Exchange) SpotDeployGenesis(deployer string, dexName string) (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":     "spotDeployGenesis",
		"deployer": deployer,
		"dexName":  dexName,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotDeployRegisterSpot registers spot market
func (e *Exchange) SpotDeployRegisterSpot(
	baseToken string,
	quoteToken string,
) (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":       "spotDeployRegisterSpot",
		"baseToken":  baseToken,
		"quoteToken": quoteToken,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotDeployRegisterHyperliquidity registers hyperliquidity spot
func (e *Exchange) SpotDeployRegisterHyperliquidity(
	name string,
	tokens []string,
) (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":   "spotDeployRegisterHyperliquidity",
		"name":   name,
		"tokens": tokens,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// SpotDeploySetDeployerTradingFeeShare sets deployer trading fee share
func (e *Exchange) SpotDeploySetDeployerTradingFeeShare(
	feeShare float64,
) (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":     "spotDeploySetDeployerTradingFeeShare",
		"feeShare": feeShare,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Perp Deploy Methods

// PerpDeployRegisterAsset registers a new perpetual asset
func (e *Exchange) PerpDeployRegisterAsset(
	asset string,
	perpDexInput PerpDexSchemaInput,
) (*PerpDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":         "perpDeployRegisterAsset",
		"asset":        asset,
		"perpDexInput": perpDexInput,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result PerpDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// PerpDeploySetOracle sets oracle for perpetual asset
func (e *Exchange) PerpDeploySetOracle(
	asset string,
	oracleAddress string,
) (*SpotDeployResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":          "perpDeploySetOracle",
		"asset":         asset,
		"oracleAddress": oracleAddress,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result SpotDeployResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CSigner Methods

// CSignerUnjailSelf unjails self as consensus signer
func (e *Exchange) CSignerUnjailSelf() (*ValidatorResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "cSignerUnjailSelf",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result ValidatorResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CSignerJailSelf jails self as consensus signer
func (e *Exchange) CSignerJailSelf() (*ValidatorResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "cSignerJailSelf",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result ValidatorResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CSignerInner executes inner consensus signer action
func (e *Exchange) CSignerInner(innerAction map[string]any) (*ValidatorResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":        "cSignerInner",
		"innerAction": innerAction,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result ValidatorResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CValidator Methods

// CValidatorRegister registers as consensus validator
func (e *Exchange) CValidatorRegister(validatorProfile map[string]any) (*ValidatorResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":             "cValidatorRegister",
		"validatorProfile": validatorProfile,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result ValidatorResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CValidatorChangeProfile changes validator profile
func (e *Exchange) CValidatorChangeProfile(newProfile map[string]any) (*ValidatorResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type":       "cValidatorChangeProfile",
		"newProfile": newProfile,
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result ValidatorResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// CValidatorUnregister unregisters as consensus validator
func (e *Exchange) CValidatorUnregister() (*ValidatorResponse, error) {
	timestamp := time.Now().UnixMilli()

	action := map[string]any{
		"type": "cValidatorUnregister",
	}

	sig, err := SignL1Action(
		e.privateKey,
		action,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(action, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result ValidatorResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (e *Exchange) MultiSig(
	action map[string]any,
	signers []string,
	signatures []string,
) (*MultiSigResponse, error) {
	timestamp := time.Now().UnixMilli()

	multiSigAction := map[string]any{
		"type":       "multiSig",
		"action":     action,
		"signers":    signers,
		"signatures": signatures,
	}

	sig, err := SignL1Action(
		e.privateKey,
		multiSigAction,
		e.vault,
		timestamp,
		e.expiresAfter,
		e.client.baseURL == MainnetAPIURL,
	)
	if err != nil {
		return nil, err
	}

	resp, err := e.postAction(multiSigAction, sig, timestamp)
	if err != nil {
		return nil, err
	}

	var result MultiSigResponse
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
