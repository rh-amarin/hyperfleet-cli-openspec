// Package api provides a generic typed HTTP client for the HyperFleet API.
// All requests use a shared Client that handles Bearer token auth, RFC 7807
// error parsing, verbose debug logging, and a 30-second timeout.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"
)

// Client is an HTTP client configured for the HyperFleet API.
type Client struct {
	baseURL string
	token   string
	verbose bool
	http    *http.Client
}

// NewClient creates a Client.
// baseURL should be the full base URL including scheme (e.g., "http://localhost:8000").
// The client appends "/api/hyperfleet/{apiVersion}/" automatically via ResourceHref.
// Pass the full base URL as: "{api-url}/api/hyperfleet/{api-version}/".
func NewClient(baseURL, token string, verbose bool) *Client {
	if !strings.HasSuffix(baseURL, "/") {
		baseURL += "/"
	}
	return &Client{
		baseURL: baseURL,
		token:   token,
		verbose: verbose,
		http:    &http.Client{Timeout: 30 * time.Second},
	}
}

// ResourceHref builds a canonical resource URL for the given path.
// It prepends the base URL. Callers pass just the resource path (e.g., "clusters/id").
func (c *Client) ResourceHref(resourcePath string) string {
	return c.baseURL + strings.TrimPrefix(resourcePath, "/")
}

// do executes an HTTP request, returning the response or an error.
func (c *Client) do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	url := c.ResourceHref(path)

	var bodyReader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	var req *http.Request
	var err error
	if bodyReader != nil {
		req, err = http.NewRequestWithContext(ctx, method, url, bodyReader)
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	start := time.Now()
	resp, err := c.http.Do(req)
	elapsed := time.Since(start).Milliseconds()

	if err != nil {
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			return nil, fmt.Errorf("[ERROR] Request to %s timed out after 30s. Check your network connection and API server.", req.URL)
		}
		return nil, fmt.Errorf("[ERROR] %w", err)
	}

	if c.verbose {
		fmt.Fprintf(os.Stderr, "[DEBUG] %s %s → %d (%dms)\n", method, url, resp.StatusCode, elapsed)
	}

	return resp, nil
}

// decode reads a JSON response body into v and closes the body.
func decode[T any](resp *http.Response) (T, error) {
	defer resp.Body.Close()
	var zero T
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return zero, parseError(resp)
	}
	var result T
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return zero, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

// Get sends a GET request and decodes the response into T.
func Get[T any](ctx context.Context, c *Client, path string) (T, error) {
	resp, err := c.do(ctx, http.MethodGet, path, nil)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](resp)
}

// Post sends a POST request with body and decodes the response into T.
func Post[T any](ctx context.Context, c *Client, path string, body any) (T, error) {
	resp, err := c.do(ctx, http.MethodPost, path, body)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](resp)
}

// Patch sends a PATCH request with body and decodes the response into T.
func Patch[T any](ctx context.Context, c *Client, path string, body any) (T, error) {
	resp, err := c.do(ctx, http.MethodPatch, path, body)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](resp)
}

// Delete sends a DELETE request and decodes the response into T.
func Delete[T any](ctx context.Context, c *Client, path string) (T, error) {
	resp, err := c.do(ctx, http.MethodDelete, path, nil)
	if err != nil {
		var zero T
		return zero, err
	}
	return decode[T](resp)
}
