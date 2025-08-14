package hyperliquid

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// roundToDecimals rounds a float64 to the specified number of decimals.
func roundToDecimals(value float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(value*pow) / pow
}

// parseFloat parses a string to float64, returns 0.0 if parsing fails.
func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0.0
	}
	return f
}

// abs returns the absolute value of a float64.
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// formatFloat formats a float64 to string with 6 decimal places.
func formatFloat(f float64) string {
	return fmt.Sprintf("%.6f", f)
}

// floatToWire converts a float64 to a wire-compatible string format
func floatToWire(x float64) (string, error) {
	// Format to 8 decimal places (for prices)
	rounded := fmt.Sprintf("%.8f", x)

	// Handle -0 case
	if rounded == "-0.00000000" {
		rounded = "0.00000000"
	}

	// Remove trailing zeros and decimal point if not needed
	result := strings.TrimRight(rounded, "0")
	result = strings.TrimRight(result, ".")

	return result, nil
}

// sizeToWire converts a float64 size to a wire-compatible string format
// conforming to Hyperliquid's lot size constraints (typically 2-3 decimal places)
func sizeToWire(x float64) (string, error) {
	// Round to 3 decimal places for lot size compliance
	rounded := fmt.Sprintf("%.3f", x)

	// Handle -0 case
	if rounded == "-0.000" {
		rounded = "0.000"
	}

	// Remove trailing zeros and decimal point if not needed
	result := strings.TrimRight(rounded, "0")
	result = strings.TrimRight(result, ".")

	return result, nil
}

// sizeToWireWithAsset converts a float64 size to a wire-compatible string format
// using the asset-specific decimal constraints from Hyperliquid
// According to docs: "Sizes are rounded to the szDecimals of that asset"
func sizeToWireWithAsset(x float64, asset int, info *Info) (string, error) {
	// Get the asset-specific decimal constraints
	szDecimals, exists := info.assetToDecimal[asset]
	if !exists {
		return sizeToWire(x)
	}

	// Round to the asset's szDecimals (this is what the docs specify)
	rounded := fmt.Sprintf("%.*f", szDecimals, x)

	// Handle -0 case
	if strings.HasPrefix(rounded, "-0.") && strings.TrimRight(strings.TrimPrefix(rounded, "-0."), "0") == "" {
		rounded = "0" + strings.TrimPrefix(rounded, "-0")
	}

	// Remove trailing zeros and decimal point if not needed (docs requirement for signing)
	result := strings.TrimRight(rounded, "0")
	result = strings.TrimRight(result, ".")

	return result, nil
}

// PriceToWire converts a float64 price to a wire-compatible string format
// following Hyperliquid's price constraints:
// - Up to 5 significant figures
// - No more than MAX_DECIMALS - szDecimals decimal places
// - MAX_DECIMALS is 6 for perps, 8 for spot
func PriceToWire(x float64, asset int, info *Info, isSpot bool) (string, error) {
	// Get the asset-specific decimal constraints
	szDecimals, exists := info.assetToDecimal[asset]
	if !exists {
		// Fallback to default behavior
		return floatToWire(x)
	}

	// Determine MAX_DECIMALS based on asset type
	maxDecimals := 6 // perps
	if isSpot {
		maxDecimals = 8
	}

	// Calculate allowed decimal places: MAX_DECIMALS - szDecimals
	allowedDecimals := maxDecimals - szDecimals
	if allowedDecimals < 0 {
		allowedDecimals = 0
	}

	// Format to allowed decimal places
	rounded := fmt.Sprintf("%.*f", allowedDecimals, x)

	// Handle -0 case
	if strings.HasPrefix(rounded, "-0.") && strings.TrimRight(strings.TrimPrefix(rounded, "-0."), "0") == "" {
		rounded = "0" + strings.TrimPrefix(rounded, "-0")
	}

	// Remove trailing zeros and decimal point if not needed (docs requirement for signing)
	result := strings.TrimRight(rounded, "0")
	result = strings.TrimRight(result, ".")

	return result, nil
}
