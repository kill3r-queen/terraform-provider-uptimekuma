// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package client

import (
	"bytes" // Import bytes for handling request body.
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Config holds the configuration for the Uptime Kuma client.
type Config struct {
	BaseURL          string
	Username         string
	Password         string
	Timeout          time.Duration
	InsecureHTTPS    bool // You might need to handle this in http.Client transport.
	CustomHTTPClient *http.Client
}

// Client is the API client for Uptime Kuma.
type Client struct {
	config     *Config
	authClient *AuthClient
	httpClient *http.Client // This client should handle auth automatically.
}

// New creates a new Uptime Kuma API client.
func New(config *Config) (*Client, error) {
	// Validate config
	if config.BaseURL == "" {
		return nil, fmt.Errorf("base URL is required")
	}
	if config.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if config.Password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Set defaults.
	if config.Timeout == 0 {
		config.Timeout = 30 * time.Second
	}

	// Create base HTTP client (used by AuthClient for token fetching).
	baseHttpClient := config.CustomHTTPClient
	if baseHttpClient == nil {
		baseHttpClient = &http.Client{
			Timeout: config.Timeout,
			// TODO: Handle config.InsecureHTTPS if needed, e.g., using tls.Config.
		}
	}

	// Create auth client.
	authClient := NewAuthClient(
		config.BaseURL,
		config.Username,
		config.Password,
		baseHttpClient, // Pass the base client here.
	)

	// Create the main API client with an *authenticated* http client.
	// The authenticated client uses the authClient internally via its transport.
	authenticatedHttpClient := authClient.AuthenticatedClient()

	return &Client{
		config:     config,
		authClient: authClient, // Store authClient if needed elsewhere, otherwise optional.
		httpClient: authenticatedHttpClient,
	}, nil
}

// doRequest performs an HTTP request and decodes the response.
func (c *Client) doRequest(ctx context.Context, method, path string, requestBody interface{}, result interface{}) error {
	// Create request URL.
	url := fmt.Sprintf("%s%s", c.config.BaseURL, path)

	// Marshal request body if provided.
	var bodyReader io.Reader
	if requestBody != nil {
		bodyBytes, err := json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(bodyBytes)
	}

	// Create request.
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set common headers.
	req.Header.Set("Accept", "application/json")
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Execute request using the authenticated client.
	// The authTransport within c.httpClient will handle adding the Bearer token.
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read body first for better error messages.
	respBodyBytes, readErr := io.ReadAll(resp.Body)
	if readErr != nil {
		// Log reading error but still check status code.
		fmt.Printf("Warning: failed to read response body: %v\n", readErr)
	}

	// Check status code.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status %d: %s", resp.StatusCode, string(respBodyBytes))
	}

	// Decode response if result is provided and body was read successfully.
	if result != nil {
		if readErr != nil {
			return fmt.Errorf("failed to decode response body due to read error: %w", readErr)
		}
		if err := json.Unmarshal(respBodyBytes, result); err != nil {
			return fmt.Errorf("failed to decode response body: %w (body: %s)", err, string(respBodyBytes))
		}
	}

	return nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodGet, path, nil, result)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, requestBody interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPost, path, requestBody, result)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, requestBody interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPut, path, requestBody, result)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, requestBody interface{}, result interface{}) error {
	return c.doRequest(ctx, http.MethodPatch, path, requestBody, result)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string, result interface{}) error {
	return c.doRequest(ctx, http.MethodDelete, path, nil, result)
}
