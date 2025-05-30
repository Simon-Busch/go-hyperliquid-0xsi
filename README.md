# go-hyperliquid

[![Go Reference](https://pkg.go.dev/badge/github.com/sonirico/go-hyperliquid.svg)](https://pkg.go.dev/github.com/sonirico/go-hyperliquid)
[![Go Report Card](https://goreportcard.com/badge/github.com/sonirico/go-hyperliquid)](https://goreportcard.com/report/github.com/sonirico/go-hyperliquid)
[![CI](https://github.com/sonirico/go-hyperliquid/actions/workflows/ci.yml/badge.svg)](https://github.com/sonirico/go-hyperliquid/actions/workflows/ci.yml)
[![Coverage Status](https://coveralls.io/repos/github/sonirico/go-hyperliquid/badge.svg?branch=main)](https://coveralls.io/github/sonirico/go-hyperliquid?branch=main)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sonirico/go-hyperliquid)](https://go.dev/)

Unofficial Go client for the Hyperliquid exchange API. This implementation follows the same philosophy and patterns as the official Python SDK.

## Installation

```bash
go get github.com/sonirico/go-hyperliquid
```

## Features

- Complete WebSocket implementation for real-time market data and user events
- Full REST API support including:
  - Market data (L2 order book, trades, candles)
  - Trading operations (orders, positions, leverage)
  - User account management
  - Wallet operations
- Both mainnet and testnet environments
- Proper error handling and type safety
- Built-in reconnection and recovery mechanisms
- Concurrent-safe operations

## Usage

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "github.com/ethereum/go-ethereum/crypto"
    hyperliquid "github.com/sonirico/go-hyperliquid"
)

func main() {
    // Initialize client
    client := hyperliquid.NewClient(hyperliquid.MainnetAPIURL)
    
    // For trading, create an Exchange with your private key
    privateKey, _ := crypto.HexToECDSA("your-private-key")
    exchange := hyperliquid.NewExchange(
        privateKey,
        hyperliquid.MainnetAPIURL,
        nil,    // Meta will be fetched automatically
        "vault-address",
        "account-address",
        nil,    // SpotMeta will be fetched automatically
    )
    
    // Place a limit order
    order := hyperliquid.OrderRequest{
        Coin:    "BTC",
        IsBuy:   true,
        Size:    0.1,
        LimitPx: 40000.0,
        OrderType: hyperliquid.OrderType{
            Limit: &hyperliquid.LimitOrderType{
                Tif: "Gtc",
            },
        },
    }
    
    resp, err := exchange.Order(order, nil)
    if err != nil {
        log.Fatal(err)
    }
    
    // Subscribe to WebSocket updates
    ws := hyperliquid.NewWebsocketClient(hyperliquid.MainnetAPIURL)
    if err := ws.Connect(context.Background()); err != nil {
        log.Fatal(err)
    }
    defer ws.Close()
    
    // Subscribe to BTC trades
    _, err = ws.Subscribe(hyperliquid.Subscription{
        Type: "trades",
        Coin: "BTC",
    }, func(msg hyperliquid.WSMessage) {
        fmt.Printf("Trade: %+v\n", msg)
    })
}
```

## Documentation

For detailed API documentation, please refer to:
- [Official Hyperliquid API docs](https://hyperliquid.xyz/docs)
- [Go package documentation](https://pkg.go.dev/github.com/sonirico/go-hyperliquid)

### Examples

Check the `examples/` directory for more usage examples:
- WebSocket subscriptions
- Order management
- Position handling
- Market data retrieval

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Quick Start for Contributors

```bash
# Clone the repository
git clone https://github.com/sonirico/go-hyperliquid.git
cd go-hyperliquid

# Install dependencies and tools
make deps install-tools

# Run all checks
make ci-full

# Run tests (excluding examples)
make ci-test
```

## Roadmap

- [x] Complete WebSocket API implementation
- [x] REST API client
- [x] Order management
- [x] User account operations
- [ ] Advanced order types
- [ ] Historical data API
- [ ] Rate limiting improvements
- [ ] Connection pooling

## License

MIT License

Copyright (c) 2025

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
