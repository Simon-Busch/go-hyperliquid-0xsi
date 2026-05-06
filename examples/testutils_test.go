package examples

import (
	"log"
	"os"
	"testing"

	"github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/joho/godotenv"
)

func init() {
	// Load environment variables from .test.env file
	err := godotenv.Load("../.env")
	if err != nil {
		log.Printf("Warning: Could not load .env file: %v", err)
	}
}

// accountAddress returns the user's HL account address from the
// HL_ACCOUNT_ADDRESS env var (loaded from .env via the package init()).
// Fails the test if the variable is empty so tests don't silently target
// some other address.
func accountAddress(t *testing.T) string {
	t.Helper()
	addr := os.Getenv("HL_ACCOUNT_ADDRESS")
	if addr == "" {
		t.Fatal("HL_ACCOUNT_ADDRESS not set in environment (.env)")
	}
	return addr
}

func newTestExchange(t *testing.T) *hyperliquid.Exchange {
	t.Helper()

	privKeyHex := os.Getenv("HL_PRIVATE_KEY")
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

	// Initialize test exchange (default perp dex)
	return hyperliquid.NewExchange(
		testPrivateKey,
		hyperliquid.MainnetAPIURL,
		nil,
		"",
		accountAddr,
		nil,
		nil, // perpDexs
		"",  // perpDexName (empty = default dex)
	)
}
