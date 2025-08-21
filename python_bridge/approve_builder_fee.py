#!/usr/bin/env python3
"""
Python bridge for Hyperliquid approve_builder_fee function.
Uses the official Hyperliquid Python SDK for 100% signature compatibility.
"""

import sys
import json
from typing import Dict, Any
from hyperliquid.exchange import Exchange
from eth_account import Account

def approve_builder_fee_bridge(
    private_key: str,
    is_mainnet: bool,
    builder: str,
    max_fee_rate: str
) -> Dict[str, Any]:
    """
    Execute approve_builder_fee using the official Hyperliquid Python SDK.
    """
    # Create account from private key
    account = Account.from_key(private_key)
    
    # Create exchange instance with explicit base URL
    if is_mainnet:
        base_url = "https://api.hyperliquid.xyz"
    else:
        base_url = "https://api.hyperliquid-testnet.xyz"
    
    exchange = Exchange(account, base_url=base_url)
    
    # Convert the max_fee_rate from "tenths of basis points" to percentage format
    # Example: "100" (100 tenths of basis points) = 1% = "1%"
    # Formula: tenths_of_basis_points / 1000 = percentage
    try:
        tenths_of_basis_points = int(max_fee_rate)
        percentage = tenths_of_basis_points / 1000.0
        fee_rate_str = f"{percentage}%"
    except ValueError:
        # If it's already in the right format, use as-is
        fee_rate_str = max_fee_rate
    
    # Call the official Python SDK approve_builder_fee function
    result = exchange.approve_builder_fee(
        builder=builder,
        max_fee_rate=fee_rate_str
    )
    
    return result

def main():
    """
    Command-line interface for the Python bridge.
    Expected arguments: private_key is_mainnet builder max_fee_rate
    """
    if len(sys.argv) != 5:
        print("Usage: python approve_builder_fee.py <private_key> <is_mainnet> <builder> <max_fee_rate>", file=sys.stderr)
        sys.exit(1)
    
    try:
        # Parse arguments
        private_key = sys.argv[1]
        is_mainnet = sys.argv[2].lower() == "true"
        builder = sys.argv[3]
        max_fee_rate = sys.argv[4]
        
        # Execute approve builder fee
        result = approve_builder_fee_bridge(
            private_key=private_key,
            is_mainnet=is_mainnet,
            builder=builder,
            max_fee_rate=max_fee_rate
        )
        
        # Output result as JSON
        print(json.dumps(result))
        
    except Exception as e:
        # Output error in JSON format for easy parsing
        error_result = {
            "error": str(e),
            "type": type(e).__name__
        }
        print(json.dumps(error_result), file=sys.stderr)
        sys.exit(1)

if __name__ == "__main__":
    main()
