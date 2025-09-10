#!/usr/bin/env python3
"""
Python bridge for Hyperliquid updateIsolatedMargin function.
Uses the official Hyperliquid Python SDK for 100% signature compatibility.
"""

import sys
import json
from typing import Dict, Any
from hyperliquid.exchange import Exchange
from eth_account import Account

def update_isolated_margin_bridge(
    private_key: str,
    is_mainnet: bool,
    amount: float,
    name: str
) -> Dict[str, Any]:
    """
    Execute updateIsolatedMargin using the official Hyperliquid Python SDK.
    """
    # Create account from private key
    account = Account.from_key(private_key)
    
    # Create exchange instance with explicit base URL
    if is_mainnet:
        base_url = "https://api.hyperliquid.xyz"
    else:
        base_url = "https://api.hyperliquid-testnet.xyz"
    
    exchange = Exchange(account, base_url=base_url)
    
    # Call the official Python SDK update_isolated_margin function
    result = exchange.update_isolated_margin(
        amount=amount,
        name=name
    )
    
    return result

def main():
    """
    Command-line interface for the Python bridge.
    Expected arguments: private_key is_mainnet amount name
    """
    if len(sys.argv) != 5:
        print("Usage: python update_isolated_margin.py <private_key> <is_mainnet> <amount> <name>", file=sys.stderr)
        sys.exit(1)
    
    try:
        # Parse arguments
        private_key = sys.argv[1]
        is_mainnet = sys.argv[2].lower() == "true"
        amount = float(sys.argv[3])
        name = sys.argv[4]
        
        # Execute update isolated margin
        result = update_isolated_margin_bridge(
            private_key=private_key,
            is_mainnet=is_mainnet,
            amount=amount,
            name=name
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
