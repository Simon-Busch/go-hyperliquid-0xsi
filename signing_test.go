package hyperliquid

import (
	"encoding/hex"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignL1Action(t *testing.T) {
	// Test private key
	privateKeyHex := "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234"
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	require.NoError(t, err)

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	require.NoError(t, err)

	tests := []struct {
		name         string
		action       map[string]any
		vaultAddress string
		timestamp    int64
		expiresAfter *int64
		isMainnet    bool
		wantErr      bool
		description  string
	}{
		{
			name: "basic_order_action_testnet",
			action: map[string]any{
				"type": "order",
				"orders": []map[string]any{
					{
						"asset":      0,
						"isBuy":      true,
						"limitPx":    "100.0",
						"orderType":  "Limit",
						"reduceOnly": false,
						"size":       "1.0",
						"tif":        "Gtc",
					},
				},
				"grouping": "na",
			},
			vaultAddress: "",
			timestamp:    1703001234567,
			expiresAfter: nil,
			isMainnet:    false,
			wantErr:      false,
			description:  "Basic order action on testnet without expiration",
		},
		{
			name: "basic_order_action_mainnet",
			action: map[string]any{
				"type": "order",
				"orders": []map[string]any{
					{
						"asset":      0,
						"isBuy":      true,
						"limitPx":    "100.0",
						"orderType":  "Limit",
						"reduceOnly": false,
						"size":       "1.0",
						"tif":        "Gtc",
					},
				},
				"grouping": "na",
			},
			vaultAddress: "",
			timestamp:    1703001234567,
			expiresAfter: nil,
			isMainnet:    true,
			wantErr:      false,
			description:  "Basic order action on mainnet without expiration",
		},
		{
			name: "order_with_expiration",
			action: map[string]any{
				"type": "order",
				"orders": []map[string]any{
					{
						"asset":      0,
						"isBuy":      true,
						"limitPx":    "100.0",
						"orderType":  "Limit",
						"reduceOnly": false,
						"size":       "1.0",
						"tif":        "Gtc",
					},
				},
				"grouping": "na",
			},
			vaultAddress: "",
			timestamp:    1703001234567,
			expiresAfter: func() *int64 { e := int64(1703001234567 + 3600000); return &e }(), // 1 hour later
			isMainnet:    false,
			wantErr:      false,
			description:  "Order action with expiration",
		},
		{
			name: "order_with_vault",
			action: map[string]any{
				"type": "order",
				"orders": []map[string]any{
					{
						"asset":      0,
						"isBuy":      true,
						"limitPx":    "100.0",
						"orderType":  "Limit",
						"reduceOnly": false,
						"size":       "1.0",
						"tif":        "Gtc",
					},
				},
				"grouping": "na",
			},
			vaultAddress: "0x1234567890abcdef1234567890abcdef12345678",
			timestamp:    1703001234567,
			expiresAfter: nil,
			isMainnet:    false,
			wantErr:      false,
			description:  "Order action with vault address",
		},
		{
			name: "leverage_update_action",
			action: map[string]any{
				"type":     "updateLeverage",
				"asset":    0,
				"isCross":  true,
				"leverage": 10,
			},
			vaultAddress: "",
			timestamp:    1703001234567,
			expiresAfter: nil,
			isMainnet:    false,
			wantErr:      false,
			description:  "Leverage update action",
		},
		{
			name: "usd_class_transfer_action",
			action: map[string]any{
				"type":   "usdClassTransfer",
				"amount": "100.0",
				"toPerp": true,
			},
			vaultAddress: "",
			timestamp:    1703001234567,
			expiresAfter: nil,
			isMainnet:    false,
			wantErr:      false,
			description:  "USD class transfer action",
		},
		{
			name: "cancel_action",
			action: map[string]any{
				"type": "cancel",
				"cancels": []map[string]any{
					{
						"asset": 0,
						"oid":   12345,
					},
				},
			},
			vaultAddress: "",
			timestamp:    1703001234567,
			expiresAfter: nil,
			isMainnet:    false,
			wantErr:      false,
			description:  "Cancel action without vault",
		},
		{
			name: "vault_action",
			action: map[string]any{
				"type":     "updateLeverage",
				"asset":    0,
				"leverage": 10,
				"isCross":  true,
			},
			vaultAddress: "0x1234567890123456789012345678901234567890",
			timestamp:    1703001234567,
			expiresAfter: nil,
			isMainnet:    true,
			wantErr:      false,
			description:  "Action with vault address",
		},
		{
			name: "empty_vault_with_expiration",
			action: map[string]any{
				"type": "setReferrer",
				"code": "TEST123",
			},
			vaultAddress: "",
			timestamp:    1703001234567,
			expiresAfter: func() *int64 { e := int64(1703001234567 + 86400000); return &e }(), // 24 hours
			isMainnet:    false,
			wantErr:      false,
			description:  "Empty vault with expiration time",
		},
		{
			name: "nil_vault_with_expiration",
			action: map[string]any{
				"type": "createSubAccount",
				"name": "TestAccount",
			},
			vaultAddress: "",
			timestamp:    1703001234567,
			expiresAfter: func() *int64 { e := int64(1703001234567 + 1800000); return &e }(), // 30 minutes
			isMainnet:    true,
			wantErr:      false,
			description:  "Nil vault with expiration time on mainnet",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := SignL1Action(
				privateKey,
				tt.action,
				tt.vaultAddress,
				tt.timestamp,
				tt.expiresAfter,
				tt.isMainnet,
			)

			if tt.wantErr {
				assert.Error(t, err, tt.description)
				return
			}

			require.NoError(t, err, tt.description)
			assert.NotEmpty(t, signature, "Signature should not be empty")
			assert.True(t, signature != "", "Signature should not be empty")
			assert.True(
				t,
				len(signature) >= 132,
				"Signature should be at least 132 characters (0x + 130 hex chars)",
			)
			assert.True(t, signature[:2] == "0x", "Signature should start with 0x")

			// Verify signature is deterministic
			signature2, err2 := SignL1Action(
				privateKey,
				tt.action,
				tt.vaultAddress,
				tt.timestamp,
				tt.expiresAfter,
				tt.isMainnet,
			)
			require.NoError(t, err2)
			assert.Equal(t, signature, signature2, "Signatures should be deterministic")
		})
	}
}

// Test helper functions
func TestSigningHelperFunctions(t *testing.T) {
	privateKeyHex := "abcd1234567890abcd1234567890abcd1234567890abcd1234567890abcd1234"
	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	require.NoError(t, err)

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	require.NoError(t, err)

	timestamp := int64(1703001234567)

	tests := []struct {
		name        string
		testFunc    func() (string, error)
		description string
	}{
		{
			name: "SignUsdClassTransferAction",
			testFunc: func() (string, error) {
				return SignUsdClassTransferAction(privateKey, 100.5, true, timestamp, false)
			},
			description: "USD class transfer action",
		},
		{
			name: "SignSpotTransferAction",
			testFunc: func() (string, error) {
				return SignSpotTransferAction(
					privateKey,
					50.25,
					"0x1234567890123456789012345678901234567890",
					"USDC",
					timestamp,
					false,
				)
			},
			description: "Spot transfer action",
		},
		{
			name: "SignUsdTransferAction",
			testFunc: func() (string, error) {
				return SignUsdTransferAction(
					privateKey,
					75.0,
					"0x1234567890123456789012345678901234567890",
					timestamp,
					true,
				)
			},
			description: "USD transfer action",
		},
		{
			name: "SignPerpDexClassTransferAction",
			testFunc: func() (string, error) {
				return SignPerpDexClassTransferAction(
					privateKey,
					"testdex",
					"ETH",
					1.5,
					false,
					timestamp,
					false,
				)
			},
			description: "Perp dex class transfer action",
		},
		{
			name: "SignTokenDelegateAction",
			testFunc: func() (string, error) {
				return SignTokenDelegateAction(
					privateKey,
					"ETH",
					2.0,
					"0x1234567890123456789012345678901234567890",
					timestamp,
					true,
				)
			},
			description: "Token delegate action",
		},
		{
			name: "SignWithdrawFromBridgeAction",
			testFunc: func() (string, error) {
				return SignWithdrawFromBridgeAction(
					privateKey,
					"0x1234567890123456789012345678901234567890",
					100.0,
					0.1,
					timestamp,
					false,
				)
			},
			description: "Withdraw from bridge action",
		},
		{
			name: "SignAgent",
			testFunc: func() (string, error) {
				return SignAgent(
					privateKey,
					"0x1234567890123456789012345678901234567890",
					"TestAgent",
					timestamp,
					true,
				)
			},
			description: "Sign agent action",
		},
		{
			name: "SignApproveBuilderFee",
			testFunc: func() (string, error) {
				return SignApproveBuilderFee(
					privateKey,
					"0x1234567890123456789012345678901234567890",
					0.001,
					timestamp,
					false,
				)
			},
			description: "Approve builder fee action",
		},
		{
			name: "SignConvertToMultiSigUserAction",
			testFunc: func() (string, error) {
				signers := []string{
					"0x1234567890123456789012345678901234567890",
					"0x0987654321098765432109876543210987654321",
				}
				return SignConvertToMultiSigUserAction(privateKey, signers, 2, timestamp, true)
			},
			description: "Convert to multi-sig user action",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signature, err := tt.testFunc()

			require.NoError(t, err, tt.description)
			assert.NotEmpty(t, signature, "Signature should not be empty")
			assert.True(t, len(signature) >= 132, "Signature should be at least 132 characters")
			assert.True(t, signature[:2] == "0x", "Signature should start with 0x")

			// Test deterministic behavior
			signature2, err2 := tt.testFunc()
			require.NoError(t, err2)
			assert.Equal(
				t,
				signature,
				signature2,
				"Helper function signatures should be deterministic",
			)
		})
	}
}

// Test utility functions
func TestUtilityFunctions(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() bool
		expected bool
	}{
		{
			name: "GetTimestampMs",
			testFunc: func() bool {
				ts := GetTimestampMs()
				return ts > 0 && ts > 1700000000000 // After 2023
			},
			expected: true,
		},
		{
			name: "FloatToUsdInt",
			testFunc: func() bool {
				result := FloatToUsdInt(123.456789)
				expected := 123456789 // 123.456789 * 1e6
				return result == expected
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.testFunc()
			assert.Equal(t, tt.expected, result, tt.name+" should work correctly")
		})
	}
}

// Test OrderRequestToWire function
func TestOrderRequestToWire(t *testing.T) {
	tests := []struct {
		name     string
		req      OrderRequest
		asset    int
		expected OrderWire
	}{
		{
			name: "limit_order",
			req: OrderRequest{
				Coin:       "ETH",
				IsBuy:      true,
				Size:       1.5,
				LimitPx:    2000.50,
				ReduceOnly: false,
				OrderType: OrderType{
					Limit: &LimitOrderType{
						Tif: "Gtc",
					},
				},
				Cloid: func() *string { s := "test123"; return &s }(),
			},
			asset: 0,
			expected: OrderWire{
				Asset:      0,
				IsBuy:      true,
				Size:       1.5,
				LimitPx:    2000.50,
				ReduceOnly: false,
				OrderType:  "Limit",
				Tif:        "Gtc",
				Cloid:      "test123",
			},
		},
		{
			name: "trigger_order",
			req: OrderRequest{
				Coin:       "BTC",
				IsBuy:      false,
				Size:       0.1,
				LimitPx:    45000.00,
				ReduceOnly: true,
				OrderType: OrderType{
					Trigger: &TriggerOrderType{
						TriggerPx: 44000.00,
						IsMarket:  true,
						Tpsl:      "tp",
					},
				},
			},
			asset: 1,
			expected: OrderWire{
				Asset:      1,
				IsBuy:      false,
				Size:       0.1,
				LimitPx:    45000.00,
				ReduceOnly: true,
				OrderType:  "Trigger",
				TriggerPx:  44000.00,
				IsMarket:   true,
				Tpsl:       "tp",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := OrderRequestToWire(tt.req, tt.asset)
			assert.Equal(t, tt.expected, result, "OrderRequestToWire should convert correctly")
		})
	}
}

// TestDebugActionHash helps debug the action hash generation
func TestDebugActionHash(t *testing.T) {
	// Use the same test data as Python
	action := map[string]any{
		"type": "order",
		"orders": []OrderWire{{
			Asset:     0,
			IsBuy:     true,
			LimitPx:   100.5,
			Size:      1.0,
			OrderType: "Limit",
			Tif:       "Gtc",
		}},
		"grouping": "na",
	}

	privateKey, _ := crypto.HexToECDSA(
		"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	)
	vaultAddress := ""
	timestamp := int64(1640995200000) // Fixed timestamp
	var expiresAfter *int64 = nil
	isMainnet := false

	// Debug: Print action hash components
	hash := actionHash(action, vaultAddress, timestamp, expiresAfter)
	t.Logf("Action hash: %x", hash)

	// Debug: Print phantom agent
	phantomAgent := constructPhantomAgent(hash, isMainnet)
	t.Logf("Phantom agent: %+v", phantomAgent)

	// Generate signature
	signature, err := SignL1Action(
		privateKey,
		action,
		vaultAddress,
		timestamp,
		expiresAfter,
		isMainnet,
	)
	require.NoError(t, err)
	t.Logf("Generated signature: %s", signature)
}
