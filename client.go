// Package hyperliquid provides a Go client library for the Hyperliquid exchange API.
// It includes support for both REST API and WebSocket connections, allowing users to
// access market data, manage orders, and handle user account operations.
package hyperliquid

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	MainnetAPIURL = "https://api.hyperliquid.xyz"
	TestnetAPIURL = "https://api.hyperliquid-testnet.xyz"
	LocalAPIURL   = "http://localhost:3001"

	// httpErrorStatusCode is the minimum status code considered an error
	httpErrorStatusCode = 400
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	if baseURL == "" {
		baseURL = MainnetAPIURL
	}

	return &Client{
		baseURL:    baseURL,
		httpClient: new(http.Client),
	}
}

func (c *Client) post(path string, payload any) ([]byte, error) {
	jsonData, err := json.Marshal(payload)
	fmt.Println(string(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to marshal payload: %w", err)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(
		context.Background(),
		"POST",
		url,
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body := make([]byte, 0)
	if resp.Body != nil {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
	}

	if resp.StatusCode >= httpErrorStatusCode {
		var apiErr APIError
		if err := json.Unmarshal(body, &apiErr); err != nil {
			return nil, fmt.Errorf("status %d: %s", resp.StatusCode, string(body))
		}
		return nil, apiErr
	}

	return body, nil
}
