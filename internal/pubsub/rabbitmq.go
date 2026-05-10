package pubsub

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// RabbitClient implements RabbitPublisher using the RabbitMQ HTTP Management API.
type RabbitClient struct {
	httpClient *http.Client
}

// NewRabbitClient returns a RabbitClient using the default HTTP client.
func NewRabbitClient() *RabbitClient {
	return &RabbitClient{httpClient: http.DefaultClient}
}

// Publish sends payload to exchange via the RabbitMQ HTTP Management API.
// baseURL should be http://{host}:{mgmt-port}. vhost "/" is URL-encoded to "%2F".
func (r *RabbitClient) Publish(ctx context.Context, baseURL, user, password, vhost, exchange, routingKey string, payload []byte) error {
	base, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil {
		return fmt.Errorf("invalid baseURL: %w", err)
	}
	// Build decoded and raw (encoded) paths so Go's HTTP client sends %2F literally.
	decodedPath := fmt.Sprintf("/api/exchanges/%s/%s/publish", vhost, exchange)
	rawPath := fmt.Sprintf("/api/exchanges/%s/%s/publish", url.PathEscape(vhost), url.PathEscape(exchange))
	base.Path = decodedPath
	base.RawPath = rawPath
	endpoint := base.String()

	body, err := json.Marshal(map[string]any{
		"properties":       map[string]any{},
		"routing_key":      routingKey,
		"payload":          string(payload),
		"payload_encoding": "string",
	})
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(user, password)

	resp, err := r.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("http request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rabbitmq returned HTTP %d", resp.StatusCode)
	}
	return nil
}
