// Package maestro provides an HTTP client for the Maestro resource-bundle API.
package maestro

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
)

// ErrDryRun is returned when curlMode is enabled and the request was not sent.
var ErrDryRun = errors.New("dry-run")

// ManifestSummary is a lightweight summary of a single Kubernetes manifest.
type ManifestSummary struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

// BundleCondition is a status condition on a ResourceBundle.
type BundleCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
	Reason string `json:"reason,omitempty"`
}

// ResourceBundle is a Maestro resource-bundle record.
type ResourceBundle struct {
	ID            string            `json:"id"`
	Name          string            `json:"name"`
	ConsumerName  string            `json:"consumer_name"`
	Version       int               `json:"version"`
	ManifestCount int               `json:"manifest_count"`
	Manifests     []ManifestSummary `json:"manifests"`
	Conditions    []BundleCondition `json:"conditions"`
}

// ResourceBundleList is the response shape for listing resource-bundles.
type ResourceBundleList struct {
	Items []ResourceBundle `json:"items"`
	Kind  string           `json:"kind"`
	Total int              `json:"total"`
}

// Consumer is a Maestro consumer record.
type Consumer struct {
	ID   string `json:"id"`
	Kind string `json:"kind"`
	Name string `json:"name"`
}

// ConsumerList is the response shape for listing consumers.
type ConsumerList struct {
	Items []Consumer `json:"items"`
	Kind  string     `json:"kind"`
	Total int        `json:"total"`
}

// Client is an HTTP client for the Maestro REST API.
type Client struct {
	httpEndpoint string
	curlMode     bool
	http         *http.Client
}

// NewFromConfig creates a Client from the active config store.
func NewFromConfig(s *config.Store, curlMode bool) *Client {
	endpoint := s.Get("maestro", "http-endpoint")
	return &Client{
		httpEndpoint: strings.TrimRight(endpoint, "/"),
		curlMode:     curlMode,
		http:         &http.Client{Timeout: 30 * time.Second},
	}
}

// get sends a GET to the given path (relative to the Maestro API root) and
// decodes the JSON response into v.
func (c *Client) get(ctx context.Context, path string, v any) error {
	url := c.httpEndpoint + "/api/maestro/v1/" + strings.TrimPrefix(path, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")

	if c.curlMode {
		fmt.Fprintf(os.Stderr, "[CURL] curl -s \"%s\" \\\n  -H 'Accept: application/json'\n", url)
		return ErrDryRun
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("[ERROR] maestro: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("[ERROR] maestro: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

// delete sends a DELETE to the given path.
func (c *Client) delete(ctx context.Context, path string) error {
	url := c.httpEndpoint + "/api/maestro/v1/" + strings.TrimPrefix(path, "/")
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}

	if c.curlMode {
		fmt.Fprintf(os.Stderr, "[CURL] curl -s -X DELETE \"%s\"\n", url)
		return ErrDryRun
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("[ERROR] maestro: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("[ERROR] maestro: HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

// ListResourceBundles lists all resource-bundles (no consumer filter).
func (c *Client) ListResourceBundles(ctx context.Context) (*ResourceBundleList, error) {
	var result ResourceBundleList
	if err := c.get(ctx, "resource-bundles", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// ListResourceBundlesByConsumer lists resource-bundles for a specific consumer.
func (c *Client) ListResourceBundlesByConsumer(ctx context.Context, consumer string) (*ResourceBundleList, error) {
	var result ResourceBundleList
	path := "resource-bundles?search=consumer_name='" + consumer + "'"
	if err := c.get(ctx, path, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetResourceBundle fetches all resource-bundles and returns the one matching name.
// Returns (nil, nil) if not found.
func (c *Client) GetResourceBundle(ctx context.Context, name string) (*ResourceBundle, error) {
	list, err := c.ListResourceBundles(ctx)
	if err != nil {
		return nil, err
	}
	for i := range list.Items {
		if list.Items[i].Name == name {
			return &list.Items[i], nil
		}
	}
	return nil, nil
}

// DeleteResourceBundle deletes the resource-bundle with the given ID.
func (c *Client) DeleteResourceBundle(ctx context.Context, id string) error {
	return c.delete(ctx, "resource-bundles/"+id)
}

// ListConsumers lists all Maestro consumers.
func (c *Client) ListConsumers(ctx context.Context) (*ConsumerList, error) {
	var result ConsumerList
	if err := c.get(ctx, "consumers", &result); err != nil {
		return nil, err
	}
	return &result, nil
}
