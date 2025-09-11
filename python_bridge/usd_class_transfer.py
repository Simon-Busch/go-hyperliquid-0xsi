#!/usr/bin/env python3
"""
Python bridge for Hyperliquid usdClassTransfer function.
Uses the official Hyperliquid Python SDK for 100% signature compatibility.
"""

import sys
import json
from typing import Dict, Any
from hyperliquid.exchange import Exchange
from eth_account import Account


def usd_class_transfer_bridge(
    private_key: str,
    is_mainnet: bool,
    amount: float,
    to_perp: bool,
) -> Dict[str, Any]:
    """
    Execute usd_class_transfer using the official Hyperliquid Python SDK.
    """
    # Create account from private key
    account = Account.from_key(private_key)

    # Create exchange instance with explicit base URL
    base_url = "https://api.hyperliquid.xyz" if is_mainnet else "https://api.hyperliquid-testnet.xyz"
    exchange = Exchange(account, base_url=base_url)

    # Call the official Python SDK
    result = exchange.usd_class_transfer(amount=amount, to_perp=to_perp)
    return result


def main():
    """
    Command-line interface for the Python bridge.
    Expected arguments: private_key is_mainnet amount to_perp
    """
    if len(sys.argv) != 5:
        print(
            "Usage: python usd_class_transfer.py <private_key> <is_mainnet> <amount> <to_perp>",
            file=sys.stderr,
        )
        sys.exit(1)

    try:
        private_key = sys.argv[1]
        is_mainnet = sys.argv[2].lower() == "true"
        amount = float(sys.argv[3])
        to_perp = sys.argv[4].lower() == "true"

        result = usd_class_transfer_bridge(
            private_key=private_key,
            is_mainnet=is_mainnet,
            amount=amount,
            to_perp=to_perp,
        )

        print(json.dumps(result))
    except Exception as e:
        error_result = {"error": str(e), "type": type(e).__name__}
        print(json.dumps(error_result), file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
