package hyperliquid

import (
	_ "embed"
)

// Embedded Python scripts for the Python bridge functionality
// These scripts are embedded into the Go binary to ensure they're always available

//go:embed python_bridge/bulk_orders.py
var bulkOrdersPythonScript string

//go:embed python_bridge/approve_builder_fee.py
var approveBuilderFeePythonScript string
