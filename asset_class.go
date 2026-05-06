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
// Values: 6 for perps (incl. HIP-3 builder perps), 8 for spot, 3 for HIP-4 outcome
// markets (binary, prices in [0.001, 0.999]).
func (c AssetClass) MaxPriceDecimals() int {
	switch c {
	case AssetClassSpot:
		return 8
	case AssetClassOutcome:
		return 3
	default:
		return 6
	}
}

// IsSpotLike reports whether this asset class uses spot pricing rules.
// Retained for callers that previously branched on the old `isSpot bool`;
// new code should use MaxPriceDecimals() directly.
func (c AssetClass) IsSpotLike() bool {
	return c == AssetClassSpot
}
