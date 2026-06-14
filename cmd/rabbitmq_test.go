package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rh-amarin/hyperfleet-cli/internal/pubsub"
)

// ---- mock RabbitPublisher ----

type mockRabbitPublisher struct {
	publishFn func(ctx context.Context, baseURL, user, password, vhost, exchange, routingKey string, payload []byte) error
}

func (m *mockRabbitPublisher) Publish(ctx context.Context, baseURL, user, password, vhost, exchange, routingKey string, payload []byte) error {
	if m.publishFn != nil {
		return m.publishFn(ctx, baseURL, user, password, vhost, exchange, routingKey, payload)
	}
	return nil
}

// injectMockRabbit sets rabbitFactory for the duration of a test.
func injectMockRabbit(t *testing.T, mock pubsub.RabbitPublisher) {
	t.Helper()
	orig := rabbitFactory
	rabbitFactory = func() pubsub.RabbitPublisher { return mock }
	t.Cleanup(func() { rabbitFactory = orig })
}

// setupRabbitEnv creates a temp dir with clusters+nodepools resource-types.
func setupRabbitEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	makeEnvRaw(t, dir, "test", `hyperfleet:
  api-url: http://api.example.com
  api-version: v1
resource-types:
  clusters:
    path: clusters
  nodepools:
    parent: clusters
    path: "clusters/{cluster_id}/nodepools"
`)
	setActiveEnv(t, dir, "test")
	return dir
}

// setRabbitState writes state.yaml with the given IDs.
func setRabbitState(t *testing.T, dir, cid, npid string) {
	t.Helper()
	statePath := filepath.Join(dir, "state.yaml")
	content := fmt.Sprintf("active-environment: test\nclusters: %s\n", cid)
	if npid != "" {
		content += "nodepools: " + npid + "\n"
	}
	if err := os.WriteFile(statePath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// ---- hf rabbitmq publish clusters ----

func TestRabbitPublish_UnknownResourceType(t *testing.T) {
	injectMockRabbit(t, &mockRabbitPublisher{})
	dir := setupRabbitEnv(t)
	_, err := runCmd(t, dir, "rabbitmq", "publish", "widgets", "my-exchange")
	if err == nil {
		t.Fatal("expected error for unknown resource type")
	}
	if !strings.Contains(err.Error(), "unknown resource type") {
		t.Errorf("error = %q, want 'unknown resource type'", err.Error())
	}
}

func TestRabbitPublish_MissingClusterState(t *testing.T) {
	injectMockRabbit(t, &mockRabbitPublisher{})
	dir := setupRabbitEnv(t)
	_, err := runCmd(t, dir, "rabbitmq", "publish", "clusters", "my-exchange")
	if err == nil {
		t.Fatal("expected error for missing clusters state")
	}
	if !strings.Contains(err.Error(), "No clusters") {
		t.Errorf("error = %q, want 'No clusters'", err.Error())
	}
}

func TestRabbitPublish_RootResource_Success(t *testing.T) {
	var gotExchange, gotRoutingKey string
	var gotPayload []byte

	injectMockRabbit(t, &mockRabbitPublisher{
		publishFn: func(_ context.Context, _, _, _, _, exchange, routingKey string, payload []byte) error {
			gotExchange = exchange
			gotRoutingKey = routingKey
			gotPayload = payload
			return nil
		},
	})
	dir := setupRabbitEnv(t)
	setRabbitState(t, dir, clusterID, "")

	out, err := runCmd(t, dir, "rabbitmq", "publish", "clusters", "hf-exchange", "hf.clusters")
	if err != nil {
		t.Fatalf("rabbitmq publish clusters: %v", err)
	}
	if gotExchange != "hf-exchange" {
		t.Errorf("exchange = %q, want %q", gotExchange, "hf-exchange")
	}
	if gotRoutingKey != "hf.clusters" {
		t.Errorf("routing_key = %q, want %q", gotRoutingKey, "hf.clusters")
	}
	if !strings.Contains(string(gotPayload), "clusters.reconcile") {
		t.Errorf("expected clusters event type in payload, got: %s", gotPayload)
	}
	if !strings.Contains(out, "[INFO] Published clusters") {
		t.Errorf("expected success log, got: %q", out)
	}
}

func TestRabbitPublish_DefaultRoutingKey(t *testing.T) {
	var gotRoutingKey string
	injectMockRabbit(t, &mockRabbitPublisher{
		publishFn: func(_ context.Context, _, _, _, _, _, routingKey string, _ []byte) error {
			gotRoutingKey = routingKey
			return nil
		},
	})
	dir := setupRabbitEnv(t)
	setRabbitState(t, dir, clusterID, "")

	_, err := runCmd(t, dir, "rabbitmq", "publish", "clusters", "hf-exchange")
	if err != nil {
		t.Fatalf("rabbitmq publish clusters: %v", err)
	}
	if gotRoutingKey != "" {
		t.Errorf("routing_key = %q, want empty string", gotRoutingKey)
	}
}

// ---- hf rabbitmq publish nodepools ----

func TestRabbitPublish_ChildResource_MissingParentState(t *testing.T) {
	injectMockRabbit(t, &mockRabbitPublisher{})
	dir := setupRabbitEnv(t)
	_, err := runCmd(t, dir, "rabbitmq", "publish", "nodepools", "my-exchange")
	if err == nil {
		t.Fatal("expected error for missing clusters state when publishing nodepools")
	}
	if !strings.Contains(err.Error(), "No clusters") {
		t.Errorf("error = %q, want 'No clusters'", err.Error())
	}
}

func TestRabbitPublish_ChildResource_MissingOwnState(t *testing.T) {
	injectMockRabbit(t, &mockRabbitPublisher{})
	dir := setupRabbitEnv(t)
	setRabbitState(t, dir, clusterID, "")
	_, err := runCmd(t, dir, "rabbitmq", "publish", "nodepools", "my-exchange")
	if err == nil {
		t.Fatal("expected error for missing nodepools state")
	}
	if !strings.Contains(err.Error(), "No nodepools") {
		t.Errorf("error = %q, want 'No nodepools'", err.Error())
	}
}

func TestRabbitPublish_ChildResource_Success(t *testing.T) {
	var gotPayload []byte
	injectMockRabbit(t, &mockRabbitPublisher{
		publishFn: func(_ context.Context, _, _, _, _, _, _ string, payload []byte) error {
			gotPayload = payload
			return nil
		},
	})
	dir := setupRabbitEnv(t)
	setRabbitState(t, dir, clusterID, nodepoolID)

	out, err := runCmd(t, dir, "rabbitmq", "publish", "nodepools", "hf-exchange")
	if err != nil {
		t.Fatalf("rabbitmq publish nodepools: %v", err)
	}
	if !strings.Contains(string(gotPayload), "nodepools.reconcile") {
		t.Errorf("expected nodepools event type in payload, got: %s", gotPayload)
	}
	if !strings.Contains(string(gotPayload), "owner_references") {
		t.Errorf("expected owner_references in payload, got: %s", gotPayload)
	}
	if !strings.Contains(out, "[INFO] Published nodepools") {
		t.Errorf("expected success log, got: %q", out)
	}
}

func TestRabbitPublish_PublishError(t *testing.T) {
	injectMockRabbit(t, &mockRabbitPublisher{
		publishFn: func(_ context.Context, _, _, _, _, _, _ string, _ []byte) error {
			return fmt.Errorf("connection refused")
		},
	})
	dir := setupRabbitEnv(t)
	setRabbitState(t, dir, clusterID, "")

	_, err := runCmd(t, dir, "rabbitmq", "publish", "clusters", "hf-exchange")
	if err == nil {
		t.Fatal("expected error when publish fails")
	}
	if !strings.Contains(err.Error(), "Failed to publish") {
		t.Errorf("error = %q, want 'Failed to publish'", err.Error())
	}
}
