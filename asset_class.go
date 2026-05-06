package hyperliquid

// AssetClass categorises a numeric asset ID by its origin and tick rules.
//
// Ranges follow the Hyperliquid asset-IDs reference:
//   - default perp:    0..9_999
//   - spot:            10_000..99_999
//   - builder perp:    100_000..99_999_999       (HIP-3)
//   - outcome market:  100_000_000+              (HIP-4)
type AssetClass int

const (
	AssetClassPerp AssetClass = iota
	AssetClassSpot
	AssetClassBuilderPerp
	AssetClassOutcome
)

// ClassifyAsset maps a numeric asset ID to its AssetClass.
func ClassifyAsset(asset int) AssetClass {
	switch {
	case asset >= outcomeAssetBase:
		return AssetClassOutcome
	case asset >= builderPerpAssetBase:
		return AssetClassBuilderPerp
	case asset >= spotAssetIndexOffset:
		return AssetClassSpot
	default:
		return AssetClassPerp
	}
}

// MaxPriceDecimals returns MAX_DECIMALS used in the tick-size formula:
//
//	allowedPriceDecimals = MaxPriceDecimals() - szDecimals
//
// Values: 8 for spot, 6 for everything else (perps, HIP-3 builder perps,
// HIP-4 outcome markets). For HIP-4 the value was confirmed empirically:
// mainnet L2 books for outcomes show prices up to 5 decimals (e.g. 0.66782)
// with szDecimals=1, i.e. 6 - 1 = 5.
func (c AssetClass) MaxPriceDecimals() int {
	if c == AssetClassSpot {
		return 8
	}
	return 6
}

// IsSpotLike reports whether this asset class uses spot pricing rules.
// Retained for callers that previously branched on the old `isSpot bool`;
// new code should use MaxPriceDecimals() directly.
func (c AssetClass) IsSpotLike() bool {
	return c == AssetClassSpot
}
