package pubsub

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRabbitClient_Publish(t *testing.T) {
	var gotReq *http.Request
	var gotBody []byte

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotReq = r
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"routed":true}`))
	}))
	defer srv.Close()

	client := &RabbitClient{httpClient: srv.Client()}
	payload := []byte(`{"specversion":"1.0"}`)
	err := client.Publish(context.Background(), srv.URL, "guest", "guest", "/", "my-exchange", "my-key", payload)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify request path — vhost "/" must be encoded as "%2F" (use EscapedPath to see the raw form).
	wantPath := "/api/exchanges/%2F/my-exchange/publish"
	gotPath := gotReq.URL.EscapedPath()
	if gotPath != wantPath {
		t.Errorf("path = %q, want %q", gotPath, wantPath)
	}

	// Verify method
	if gotReq.Method != http.MethodPost {
		t.Errorf("method = %q, want POST", gotReq.Method)
	}

	// Verify BasicAuth
	u, p, ok := gotReq.BasicAuth()
	if !ok || u != "guest" || p != "guest" {
		t.Errorf("basic auth = %q/%q/%v, want guest/guest/true", u, p, ok)
	}

	// Verify body fields
	var body map[string]any
	if err := json.Unmarshal(gotBody, &body); err != nil {
		t.Fatalf("invalid body JSON: %v", err)
	}
	if body["routing_key"] != "my-key" {
		t.Errorf("routing_key = %v, want %q", body["routing_key"], "my-key")
	}
	if body["payload"] != string(payload) {
		t.Errorf("payload = %v, want %q", body["payload"], string(payload))
	}
	if body["payload_encoding"] != "string" {
		t.Errorf("payload_encoding = %v, want %q", body["payload_encoding"], "string")
	}
}

func TestRabbitClient_Publish_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	client := &RabbitClient{httpClient: srv.Client()}
	err := client.Publish(context.Background(), srv.URL, "bad", "creds", "/", "ex", "", []byte(`{}`))
	if err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("error = %q, want to contain 401", err.Error())
	}
}

func TestRabbitClient_Publish_EmptyRoutingKey(t *testing.T) {
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := &RabbitClient{httpClient: srv.Client()}
	_ = client.Publish(context.Background(), srv.URL, "u", "p", "/", "ex", "", []byte(`{}`))

	var body map[string]any
	_ = json.Unmarshal(gotBody, &body)
	if body["routing_key"] != "" {
		t.Errorf("routing_key = %v, want empty string", body["routing_key"])
	}
}
