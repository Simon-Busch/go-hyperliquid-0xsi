#!/usr/bin/env python3
"""
Python bridge for Hyperliquid bulk_orders function.
Uses the official Hyperliquid Python SDK for 100% signature compatibility.
"""

import sys
import json
from typing import Optional, Dict, Any, List
from hyperliquid.exchange import Exchange
from eth_account import Account

def bulk_orders_bridge(
    private_key: str,
    is_mainnet: bool,
    order_requests: List[Dict[str, Any]],
    builder: Optional[Dict[str, Any]]
) -> Dict[str, Any]:
    """
    Execute bulk_orders using the official Hyperliquid Python SDK.
    """
    # Create account from private key
    account = Account.from_key(private_key)
    
    # Create exchange instance with explicit base URL
    if is_mainnet:
        base_url = "https://api.hyperliquid.xyz"
    else:
        base_url = "https://api.hyperliquid-testnet.xyz"
    
    exchange = Exchange(account, base_url=base_url)
    
    # Call the official Python SDK bulk_orders function
    result = exchange.bulk_orders(
        order_requests=order_requests,
        builder=builder
    )
    
    return result

def main():
    """
    Command-line interface for the Python bridge.
    Expected arguments: private_key is_mainnet order_requests_json builder_json
    """
    if len(sys.argv) != 5:
        print("Usage: python bulk_orders.py <private_key> <is_mainnet> <order_requests_json> <builder_json>", file=sys.stderr)
        sys.exit(1)
    
    try:
        # Parse arguments
        private_key = sys.argv[1]
        is_mainnet = sys.argv[2].lower() == "true"
        order_requests_json = sys.argv[3]
        builder_json = sys.argv[4]
        
        # Parse order requests
        order_requests = json.loads(order_requests_json)
        
        # Parse builder info
        builder = None
        if builder_json != "null":
            builder = json.loads(builder_json)
        
        # Execute bulk orders
        result = bulk_orders_bridge(
            private_key=private_key,
            is_mainnet=is_mainnet,
            order_requests=order_requests,
            builder=builder
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
