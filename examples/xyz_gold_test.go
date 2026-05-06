package examples

import (
	"testing"

	hyperliquid "github.com/Simon-Busch/go-hyperliquid-0xsi"
)

// TestXYZDexGoldPosition tests querying positions on the xyz builder dex
// which hosts commodities like GOLD, SILVER, etc.
func TestXYZDexGoldPosition(t *testing.T) {
	address := accountAddress(t)

	// Create Info instance for the xyz dex
	info := hyperliquid.NewInfo(
		hyperliquid.MainnetAPIURL,
		true,  // skipWS
		nil,   // meta (will be fetched)
		nil,   // spotMeta (will be fetched)
		nil,   // perpDexs (will be fetched)
		"xyz", // perpDexName - the builder dex that has GOLD
	)

	// Query user state on xyz dex
	userState, err := info.UserState(address, "xyz")
	if err != nil {
		t.Fatalf("Failed to get user state: %v", err)
	}

	t.Logf("Account Value: %s", userState.MarginSummary.AccountValue)
	t.Logf("Total Positions: %d", len(userState.AssetPositions))

	// Check for GOLD position
	foundGold := false
	for _, assetPos := range userState.AssetPositions {
		pos := assetPos.Position
		entryPx := "N/A"
		if pos.EntryPx != nil {
			entryPx = *pos.EntryPx
		}
		t.Logf("Position: %s, Size: %s, Entry: %s, Leverage: %+v",
			pos.Coin, pos.Szi, entryPx, pos.Leverage)

		if pos.Coin == "xyz:GOLD" {
			foundGold = true
			t.Logf("Found GOLD position!")
		}
	}

	if !foundGold {
		t.Log("No GOLD position found (this is OK if the position was closed)")
	}
}

// TestXYZDexMeta tests fetching metadata for the xyz builder dex
func TestXYZDexMeta(t *testing.T) {
	info := hyperliquid.NewInfo(
		hyperliquid.MainnetAPIURL,
		true,
		nil,
		nil,
		nil,
		"xyz",
	)

	// Fetch meta for xyz dex
	meta, err := info.Meta("xyz")
	if err != nil {
		t.Fatalf("Failed to get xyz meta: %v", err)
	}

	t.Logf("xyz dex has %d assets", len(meta.Universe))

	// Look for GOLD in the universe
	for _, asset := range meta.Universe {
		if asset.Name == "xyz:GOLD" {
			t.Logf("Found GOLD: szDecimals=%d, maxLeverage=%d, onlyIsolated=%v",
				asset.SzDecimals, asset.MaxLeverage, asset.OnlyIsolated)
		}
	}
}

// TestXYZDexAllMids tests fetching mid prices for xyz dex assets
func TestXYZDexAllMids(t *testing.T) {
	info := hyperliquid.NewInfo(
		hyperliquid.MainnetAPIURL,
		true,
		nil,
		nil,
		nil,
		"xyz",
	)

	// Fetch all mids for xyz dex
	mids, err := info.AllMids("xyz")
	if err != nil {
		t.Fatalf("Failed to get xyz mids: %v", err)
	}

	t.Logf("xyz dex has %d assets with mid prices", len(mids))

	// Show some commodity prices
	commodities := []string{"xyz:GOLD", "xyz:SILVER", "xyz:COPPER", "xyz:PLATINUM"}
	for _, coin := range commodities {
		if mid, ok := mids[coin]; ok {
			t.Logf("%s mid price: $%s", coin, mid)
		}
	}
}

// TestDefaultVsXYZDex compares positions between default dex and xyz dex
func TestDefaultVsXYZDex(t *testing.T) {
	address := accountAddress(t)

	// Create Info for default dex
	defaultInfo := hyperliquid.NewInfo(
		hyperliquid.MainnetAPIURL,
		true,
		nil,
		nil,
		nil,
		"", // empty = default dex
	)

	// Create Info for xyz dex
	xyzInfo := hyperliquid.NewInfo(
		hyperliquid.MainnetAPIURL,
		true,
		nil,
		nil,
		nil,
		"xyz",
	)

	// Query default dex positions
	defaultState, err := defaultInfo.UserState(address)
	if err != nil {
		t.Fatalf("Failed to get default user state: %v", err)
	}

	t.Log("=== Default Dex Positions ===")
	for _, assetPos := range defaultState.AssetPositions {
		pos := assetPos.Position
		entryPx := "N/A"
		if pos.EntryPx != nil {
			entryPx = *pos.EntryPx
		}
		t.Logf("  %s: size=%s, entry=%s", pos.Coin, pos.Szi, entryPx)
	}

	// Query xyz dex positions
	xyzState, err := xyzInfo.UserState(address, "xyz")
	if err != nil {
		t.Fatalf("Failed to get xyz user state: %v", err)
	}

	t.Log("=== XYZ Dex Positions ===")
	for _, assetPos := range xyzState.AssetPositions {
		pos := assetPos.Position
		entryPx := "N/A"
		if pos.EntryPx != nil {
			entryPx = *pos.EntryPx
		}
		t.Logf("  %s: size=%s, entry=%s", pos.Coin, pos.Szi, entryPx)
	}
}

// TestPerpDexsList tests listing all available perp dexes
func TestPerpDexsList(t *testing.T) {
	info := hyperliquid.NewInfo(
		hyperliquid.MainnetAPIURL,
		true,
		nil,
		nil,
		nil,
		"",
	)

	perpDexs, err := info.PerpDexs()
	if err != nil {
		t.Fatalf("Failed to get perp dexs: %v", err)
	}

	t.Logf("Found %d perp dexes:", len(perpDexs))
	for i, mv := range perpDexs {
		if mv.Type() == "null" {
			t.Logf("  [%d] Default Perp Dex (null)", i)
		} else if mv.Type() == "object" {
			var pd hyperliquid.PerpDex
			if err := mv.Parse(&pd); err == nil {
				t.Logf("  [%d] %s (%s) - deployer: %s", i, pd.Name, pd.FullName, pd.Deployer)
			}
		}
	}
}
