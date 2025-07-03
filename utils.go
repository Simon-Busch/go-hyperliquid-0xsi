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
	// Format to 8 decimal places
	rounded := fmt.Sprintf("%.8f", x)

	// Check if rounding causes significant error
	parsed, err := strconv.ParseFloat(rounded, 64)
	if err != nil {
		return "", err
	}

	if math.Abs(parsed-x) >= 1e-12 {
		return "", fmt.Errorf("float_to_wire causes rounding: %f", x)
	}

	// Handle -0 case
	if rounded == "-0.00000000" {
		rounded = "0.00000000"
	}

	// Remove trailing zeros and decimal point if not needed
	result := strings.TrimRight(rounded, "0")
	result = strings.TrimRight(result, ".")

	return result, nil
}
