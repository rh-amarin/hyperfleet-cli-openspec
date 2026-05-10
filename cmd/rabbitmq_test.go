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

// ---- helpers ----

// setupRabbitEnv creates a temp dir with an active env.
func setupRabbitEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	makeEnv(t, dir, "test", "http://api.example.com")
	setActiveEnv(t, dir, "test")
	return dir
}

// setRabbitState writes state.yaml with cluster-id and optional nodepool-id.
func setRabbitState(t *testing.T, dir, cid, npid string) {
	t.Helper()
	statePath := filepath.Join(dir, "state.yaml")
	content := fmt.Sprintf("active-environment: test\ncluster-id: %s\n", cid)
	if npid != "" {
		content += "nodepool-id: " + npid + "\n"
	}
	if err := os.WriteFile(statePath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// ---- hf rabbitmq publish cluster ----

func TestRabbitPublishCluster_MissingClusterID(t *testing.T) {
	injectMockRabbit(t, &mockRabbitPublisher{})
	dir := setupRabbitEnv(t)
	_, err := runCmd(t, dir, "rabbitmq", "publish", "cluster", "my-exchange")
	if err == nil {
		t.Fatal("expected error for missing cluster-id")
	}
	if !strings.Contains(err.Error(), "No cluster-id") {
		t.Errorf("error = %q, want 'No cluster-id'", err.Error())
	}
}

func TestRabbitPublishCluster_Success(t *testing.T) {
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

	out, err := runCmd(t, dir, "rabbitmq", "publish", "cluster", "hf-exchange", "hf.cluster")
	if err != nil {
		t.Fatalf("rabbitmq publish cluster: %v", err)
	}
	if gotExchange != "hf-exchange" {
		t.Errorf("exchange = %q, want %q", gotExchange, "hf-exchange")
	}
	if gotRoutingKey != "hf.cluster" {
		t.Errorf("routing_key = %q, want %q", gotRoutingKey, "hf.cluster")
	}
	if !strings.Contains(string(gotPayload), "cluster.reconcile") {
		t.Errorf("expected cluster event type in payload, got: %s", gotPayload)
	}
	if !strings.Contains(out, "[INFO] Published cluster") {
		t.Errorf("expected success log, got: %q", out)
	}
}

func TestRabbitPublishCluster_DefaultRoutingKey(t *testing.T) {
	var gotRoutingKey string
	injectMockRabbit(t, &mockRabbitPublisher{
		publishFn: func(_ context.Context, _, _, _, _, _, routingKey string, _ []byte) error {
			gotRoutingKey = routingKey
			return nil
		},
	})
	dir := setupRabbitEnv(t)
	setRabbitState(t, dir, clusterID, "")

	_, err := runCmd(t, dir, "rabbitmq", "publish", "cluster", "hf-exchange")
	if err != nil {
		t.Fatalf("rabbitmq publish cluster: %v", err)
	}
	if gotRoutingKey != "" {
		t.Errorf("routing_key = %q, want empty string", gotRoutingKey)
	}
}

// ---- hf rabbitmq publish nodepool ----

func TestRabbitPublishNodePool_MissingClusterID(t *testing.T) {
	injectMockRabbit(t, &mockRabbitPublisher{})
	dir := setupRabbitEnv(t)
	_, err := runCmd(t, dir, "rabbitmq", "publish", "nodepool", "my-exchange")
	if err == nil {
		t.Fatal("expected error for missing cluster-id")
	}
	if !strings.Contains(err.Error(), "No cluster-id") {
		t.Errorf("error = %q, want 'No cluster-id'", err.Error())
	}
}

func TestRabbitPublishNodePool_MissingNodePoolID(t *testing.T) {
	injectMockRabbit(t, &mockRabbitPublisher{})
	dir := setupRabbitEnv(t)
	setRabbitState(t, dir, clusterID, "")
	_, err := runCmd(t, dir, "rabbitmq", "publish", "nodepool", "my-exchange")
	if err == nil {
		t.Fatal("expected error for missing nodepool-id")
	}
	if !strings.Contains(err.Error(), "No nodepool-id") {
		t.Errorf("error = %q, want 'No nodepool-id'", err.Error())
	}
}

func TestRabbitPublishNodePool_Success(t *testing.T) {
	var gotPayload []byte
	injectMockRabbit(t, &mockRabbitPublisher{
		publishFn: func(_ context.Context, _, _, _, _, _, _ string, payload []byte) error {
			gotPayload = payload
			return nil
		},
	})
	dir := setupRabbitEnv(t)
	setRabbitState(t, dir, clusterID, nodepoolID)

	out, err := runCmd(t, dir, "rabbitmq", "publish", "nodepool", "hf-exchange")
	if err != nil {
		t.Fatalf("rabbitmq publish nodepool: %v", err)
	}
	if !strings.Contains(string(gotPayload), "nodepool.reconcile") {
		t.Errorf("expected nodepool event type in payload, got: %s", gotPayload)
	}
	if !strings.Contains(out, "[INFO] Published nodepool") {
		t.Errorf("expected success log, got: %q", out)
	}
}

// TestRabbitPublishCluster_PublishError verifies that publish errors are surfaced.
func TestRabbitPublishCluster_PublishError(t *testing.T) {
	injectMockRabbit(t, &mockRabbitPublisher{
		publishFn: func(_ context.Context, _, _, _, _, _, _ string, _ []byte) error {
			return fmt.Errorf("connection refused")
		},
	})
	dir := setupRabbitEnv(t)
	setRabbitState(t, dir, clusterID, "")

	_, err := runCmd(t, dir, "rabbitmq", "publish", "cluster", "hf-exchange")
	if err == nil {
		t.Fatal("expected error when publish fails")
	}
	if !strings.Contains(err.Error(), "Failed to publish") {
		t.Errorf("error = %q, want 'Failed to publish'", err.Error())
	}
}
