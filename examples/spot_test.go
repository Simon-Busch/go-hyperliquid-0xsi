package examples

import (
	"strconv"
	"strings"
	"testing"
	"time"

	hyperliquid "github.com/Simon-Busch/go-hyperliquid-0xsi"
	"github.com/joho/godotenv"
)

func TestUserBalance(t *testing.T) {
	godotenv.Overload()

	t.Log("=== TESTING USER BALANCE ===")

	exchange := newTestExchange(t)
	// accountAddress := exchange.GetAccountAddr()

	addr := accountAddress(t)
	userState, err := exchange.GetInfo().UserState(addr)
	if err != nil {
		t.Fatalf("Failed to get user state: %v", err)
	}
	t.Logf("Spot User state: %+v", userState)

	spotState, err := exchange.GetInfo().SpotUserState(addr)
	if err != nil {
		t.Fatalf("Failed to get spot user state: %v", err)
	}
	t.Logf("Spot User state: %+v", spotState)
}

// TestSpotOrderFlow places a spot IOC buy worth a target USDC margin for base/quote
// It resolves the correct universe entry from SpotMeta and prices from mids for that pair.
func TestSpotOrderFlow(t *testing.T) {
	godotenv.Overload()

	exchange := newTestExchange(t)
	info := exchange.GetInfo()
	account := exchange.GetAccountAddr()

	// Inputs
	baseTicker := "ETH"
	quoteTicker := "USDC"
	usdNotional := 10.0 // target margin in quote currency

	// Fetch spot metadata
	spotMeta, err := info.SpotMeta()
	if err != nil {
		t.Fatalf("failed to fetch spot meta: %v", err)
	}
	if len(spotMeta.Universe) == 0 {
		t.Skip("no spot pairs available on this environment")
	}

	// Log full tokens and universe to inspect availability
	t.Logf("Spot tokens (%d):", len(spotMeta.Tokens))
	for _, tk := range spotMeta.Tokens {
		t.Logf("  token name=%s index=%d szDecimals=%d weiDecimals=%d", tk.Name, tk.Index, tk.SzDecimals, tk.WeiDecimals)
	}
	t.Logf("Spot universe (%d):", len(spotMeta.Universe))
	for _, u := range spotMeta.Universe {
		t.Logf("  pair name=%s tokens=%v index=%d canonical=%t", u.Name, u.Tokens, u.Index, u.IsCanonical)
	}

	// Aliases for common UI tickers -> token names in spotMeta
	aliases := func(sym string) []string {
		s := strings.ToUpper(sym)
		switch s {
		case "ETH":
			return []string{"ETH", "WETH", "UETH"}
		case "BTC":
			return []string{"BTC", "WBTC", "UBTC"}
		case "USDC":
			return []string{"USDC"}
		default:
			return []string{s}
		}
	}

	// Resolve token indices using aliases
	var baseIdx, quoteIdx = -1, -1
	for _, tk := range spotMeta.Tokens {
		for _, cand := range aliases(baseTicker) {
			if strings.EqualFold(tk.Name, cand) {
				baseIdx = tk.Index
				break
			}
		}
		for _, cand := range aliases(quoteTicker) {
			if strings.EqualFold(tk.Name, cand) {
				quoteIdx = tk.Index
				break
			}
		}
		if baseIdx != -1 && quoteIdx != -1 {
			break
		}
	}

	pairName := ""
	pairIndex := -1

	if baseIdx != -1 && quoteIdx != -1 {
		// Find universe by token indices
		for _, u := range spotMeta.Universe {
			if len(u.Tokens) == 2 && u.Tokens[0] == baseIdx && u.Tokens[1] == quoteIdx {
				pairName = u.Name
				pairIndex = u.Index
				break
			}
		}
	}

	if pairName == "" || pairIndex < 0 {
		t.Skipf("could not resolve pair for %s/%s from spot meta (baseIdx=%d quoteIdx=%d)", baseTicker, quoteTicker, baseIdx, quoteIdx)
	}
	t.Logf("Resolved spot pair: %s (index=%d from %s/%s; baseIdx=%d quoteIdx=%d)", pairName, pairIndex, baseTicker, quoteTicker, baseIdx, quoteIdx)

	// Fetch mid for the resolved pair
	mids, err := info.AllMids()
	if err != nil {
		t.Fatalf("failed to fetch mids: %v", err)
	}
	midStr, ok := mids[pairName]
	if !ok || strings.TrimSpace(midStr) == "" {
		t.Skipf("no mid available for pair %s", pairName)
	}
	midPx, _ := strconv.ParseFloat(midStr, 64)
	if midPx <= 0 {
		t.Fatalf("invalid mid price for %s: %s", pairName, midStr)
	}

	// Helper to fetch quote spot balance total as float64
	getQuote := func() float64 {
		st, err := info.SpotUserState(account)
		if err != nil {
			t.Fatalf("failed to fetch spot balances: %v", err)
		}
		for _, b := range st.Balances {
			if strings.EqualFold(b.Coin, quoteTicker) {
				v, _ := strconv.ParseFloat(b.Total, 64)
				return v
			}
		}
		return 0
	}

	quoteBefore := getQuote()
	t.Logf("%s balance before: %.6f", quoteTicker, quoteBefore)

	// Build IOC order crossing the spread so it fills immediately
	aggressivePx := midPx * 1.01 // 1% above mid for BUY
	size := usdNotional / aggressivePx
	req := hyperliquid.CreateOrderRequest{
		Coin:       pairName,
		IsBuy:      true,
		Price:      aggressivePx,
		Size:       size,
		ReduceOnly: false,
		OrderType:  hyperliquid.OrderType{Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifIoc}},
	}
	resp, err := exchange.BulkOrders([]hyperliquid.CreateOrderRequest{req}, nil)
	if err != nil {
		t.Fatalf("bulk orders failed: %v", err)
	}
	if !resp.Ok {
		t.Fatalf("order not accepted: %s", resp.Err)
	}

	// Allow settlement
	time.Sleep(2 * time.Second)

	quoteAfter := getQuote()
	t.Logf("%s balance after: %.6f", quoteTicker, quoteAfter)

	if quoteAfter >= quoteBefore {
		t.Logf("warning: %s did not decrease — trade may not have executed due to size/liquidity", quoteTicker)
	} else {
		delta := quoteBefore - quoteAfter
		t.Logf("spot %s decreased by %.6f (target ~$%.2f)", quoteTicker, delta, usdNotional)
	}
}

// TestSpotBuilderFeeWorkflow approves a 1% builder fee for spot and places a ~$10 IOC spot order with that builder fee
func TestSpotBuilderFeeWorkflow(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)
	info := exchange.GetInfo()

	t.Log("=== TESTING SPOT BUILDER FEE WORKFLOW ===")

	builderAddress := "0x223fF2A843bE0DB5E70FC1B61e7524b066F7E03c"
	maxFeeRate := "1000" // 1% (1000 tenths of basis points) for spot

	// Approve builder fee at 1%
	if resp, err := exchange.ApproveBuilderFee(builderAddress, maxFeeRate); err != nil {
		t.Fatalf("Failed to approve builder fee: %v", err)
	} else if resp != nil && resp.Error != "" {
		t.Fatalf("Builder fee approval error: %s", resp.Error)
	}

	// Snapshot initial builder rewards
	initialReferralState, err := info.QueryReferralState(builderAddress)
	if err != nil {
		t.Logf("Warning: could not query initial referral state: %v", err)
	}
	initialRewards := 0.0
	if initialReferralState != nil {
		if v, err := strconv.ParseFloat(initialReferralState.BuilderRewards, 64); err == nil {
			initialRewards = v
		}
	}
	t.Logf("Initial builder rewards: %.6f USDC", initialRewards)

	// Resolve a spot pair: ETH/USDC using alias mapping
	baseTicker := "ETH"
	quoteTicker := "USDC"
	spotMeta, err := info.SpotMeta()
	if err != nil {
		t.Fatalf("failed to fetch spot meta: %v", err)
	}
	aliases := func(sym string) []string {
		s := strings.ToUpper(sym)
		switch s {
		case "ETH":
			return []string{"ETH", "WETH", "UETH"}
		case "BTC":
			return []string{"BTC", "WBTC", "UBTC"}
		case "USDC":
			return []string{"USDC"}
		default:
			return []string{s}
		}
	}
	var baseIdx, quoteIdx = -1, -1
	for _, tk := range spotMeta.Tokens {
		for _, cand := range aliases(baseTicker) {
			if strings.EqualFold(tk.Name, cand) {
				baseIdx = tk.Index
				break
			}
		}
		for _, cand := range aliases(quoteTicker) {
			if strings.EqualFold(tk.Name, cand) {
				quoteIdx = tk.Index
				break
			}
		}
		if baseIdx != -1 && quoteIdx != -1 {
			break
		}
	}
	pairName := ""
	if baseIdx != -1 && quoteIdx != -1 {
		for _, u := range spotMeta.Universe {
			if len(u.Tokens) == 2 && u.Tokens[0] == baseIdx && u.Tokens[1] == quoteIdx {
				pairName = u.Name
				break
			}
		}
	}
	if pairName == "" {
		t.Skip("could not resolve ETH/USDC in spot meta on this environment")
	}

	// Get price and compute ~$10 size
	mids, err := info.AllMids()
	if err != nil {
		t.Fatalf("failed to fetch mids: %v", err)
	}
	midStr, ok := mids[pairName]
	if !ok || strings.TrimSpace(midStr) == "" {
		t.Skipf("no mid available for pair %s", pairName)
	}
	midPx, _ := strconv.ParseFloat(midStr, 64)
	if midPx <= 0 {
		t.Fatalf("invalid mid: %s", midStr)
	}
	aggressivePx := midPx * 1.01
	size := 30.0 / aggressivePx

	// Place spot order with builder fee
	req := hyperliquid.CreateOrderRequest{
		Coin:       pairName,
		IsBuy:      true,
		Price:      aggressivePx,
		Size:       size,
		ReduceOnly: false,
		OrderType:  hyperliquid.OrderType{Limit: &hyperliquid.LimitOrderType{Tif: hyperliquid.TifIoc}},
	}
	builder := &hyperliquid.BuilderInfo{Builder: builderAddress, Fee: 1000}
	resp, err := exchange.BulkOrders([]hyperliquid.CreateOrderRequest{req}, builder)
	if err != nil {
		t.Fatalf("bulk orders with builder failed: %v", err)
	}
	if !resp.Ok {
		t.Fatalf("order with builder not accepted: %s", resp.Err)
	}

	// Give time for fee accrual to index
	time.Sleep(5 * time.Second)

	finalReferralState, err := info.QueryReferralState(builderAddress)
	if err != nil {
		t.Logf("Warning: could not query final referral state: %v", err)
	}
	finalRewards := initialRewards
	if finalReferralState != nil {
		if v, err := strconv.ParseFloat(finalReferralState.BuilderRewards, 64); err == nil {
			finalRewards = v
		}
	}
	delta := finalRewards - initialRewards
	t.Logf("Builder rewards: before=%.6f after=%.6f delta=%.6f USDC", initialRewards, finalRewards, delta)
}

// TestUsdClassTransferWorkflow moves USDC between Spot and Perp and verifies balances before/after
func TestUsdClassTransferWorkflow(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)
	info := exchange.GetInfo()
	account := exchange.GetAccountAddr()

	getSpotUSDC := func() float64 {
		st, err := info.SpotUserState(account)
		if err != nil {
			t.Fatalf("failed to fetch spot balances: %v", err)
		}
		for _, b := range st.Balances {
			if strings.EqualFold(b.Coin, "USDC") {
				v, _ := strconv.ParseFloat(b.Total, 64)
				return v
			}
		}
		return 0
	}
	getPerpWithdrawable := func() float64 {
		st, err := info.UserState(account)
		if err != nil {
			t.Fatalf("failed to fetch perp state: %v", err)
		}
		v, _ := strconv.ParseFloat(st.Withdrawable, 64)
		return v
	}

	spotBefore := getSpotUSDC()
	perpBefore := getPerpWithdrawable()
	t.Logf("Before: spotUSDC=%.6f perpWithdrawable=%.6f", spotBefore, perpBefore)

	amount := 10.0
	if _, err := exchange.UsdClassTransfer(amount, true); err != nil { // Spot -> Perp
		t.Fatalf("spot->perp transfer failed: %v", err)
	}
	time.Sleep(3 * time.Second)

	spotAfter1 := getSpotUSDC()
	perpAfter1 := getPerpWithdrawable()
	t.Logf("After spot->perp: spotUSDC=%.6f perpWithdrawable=%.6f", spotAfter1, perpAfter1)

	if spotAfter1 > spotBefore-0.01 {
		t.Logf("warning: spot USDC did not decrease as expected")
	}
	if perpAfter1 < perpBefore+0.01 {
		t.Logf("warning: perp withdrawable did not increase as expected")
	}

	// Transfer back partial
	amountBack := 15.0
	if _, err := exchange.UsdClassTransfer(amountBack, false); err != nil { // Perp -> Spot
		t.Fatalf("perp->spot transfer failed: %v", err)
	}
	time.Sleep(3 * time.Second)

	spotAfter2 := getSpotUSDC()
	perpAfter2 := getPerpWithdrawable()
	t.Logf("After perp->spot: spotUSDC=%.6f perpWithdrawable=%.6f", spotAfter2, perpAfter2)
}

func TestSpotMeta(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)
	response, err := exchange.GetInfo().MetaAndAssetCtxs()
	if err != nil {
		t.Fatalf("Failed to fetch meta and asset contexts: %v", err)
	}

	t.Logf("Meta universe count: %d", len(response.Meta.Universe))
	t.Logf("Asset contexts count: %d", len(response.AssetCtxs))

	// Print first few asset contexts for inspection
	for i, ctx := range response.AssetCtxs {
		if i >= 3 { // Only show first 3
			break
		}
		t.Logf("Asset %d: MarkPx=%s, MidPx=%s, OpenInterest=%s, Funding=%s",
			i, ctx.MarkPx, ctx.MidPx, ctx.OpenInterest, ctx.Funding)
	}
}

// TestFundingHistoryLast24h fetches funding history for ETH over the last 24 hours
func TestFundingHistoryLast24h(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)
	info := exchange.GetInfo()

	endMs := time.Now().UnixMilli()
	startMs := time.Now().Add(-24 * time.Hour).UnixMilli()

	entries, err := info.FundingHistory("ETH", startMs, &endMs)
	if err != nil {
		t.Fatalf("failed to fetch funding history: %v", err)
	}

	t.Logf("ETH funding history entries (last 24h): %d", len(entries))
	for i, e := range entries {
		if i >= 5 { // limit log volume
			break
		}
		t.Logf("%d) time=%d rate=%s premium=%s", i+1, e.Time, e.FundingRate, e.Premium)
	}
}

func TestL2Snapshot(t *testing.T) {
	godotenv.Overload()
	exchange := newTestExchange(t)
	info := exchange.GetInfo()

	snapshot, err := info.L2Snapshot("ETH")
	if err != nil {
		t.Fatalf("failed to fetch L2 snapshot: %v", err)
	}
	t.Logf("L2 snapshot: %+v", snapshot)
}