package examples

import (
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
	"github.com/sonirico/go-hyperliquid"
)

func init() {
	// Load environment variables from .test.env file
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}
}

func newTestExchange(t *testing.T) *hyperliquid.Exchange {
	t.Helper()

	privKeyHex := os.Getenv("HL_PRIVATE_KEY")
	fmt.Println(privKeyHex)
	accountAddr := os.Getenv("HL_ACCOUNT_ADDRESS") // main user wallet address
	// vaultAddr := os.Getenv("HL_VAULT_ADDRESS")
	testPrivateKey, err := crypto.HexToECDSA(privKeyHex)

	if err != nil {
		t.Fatalf("Failed to create test private key: %v", err)
	}

	// Log the agent (signing) address and the account address
	agentAddress := crypto.PubkeyToAddress(testPrivateKey.PublicKey).Hex()
	if accountAddr == "" {
		// Fallback: use the agent address as account if none provided
		accountAddr = agentAddress
	}
	t.Logf("Agent (signer) address: %s", agentAddress)
	t.Logf("Account address: %s", accountAddr)

	// Initialize test exchange
	return hyperliquid.NewExchange(
		testPrivateKey,
		hyperliquid.TestnetAPIURL,
		nil,
		"",
		accountAddr,
		nil,
	)
}
