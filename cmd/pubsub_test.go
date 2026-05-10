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

// mockGCPPublisher is a test double for pubsub.GCPPublisher.
type mockGCPPublisher struct {
	publishFn    func(ctx context.Context, projectID, topicID string, data []byte) (string, error)
	listTopicsFn func(ctx context.Context, projectID string) ([]pubsub.TopicGroup, error)
}

func (m *mockGCPPublisher) Publish(ctx context.Context, projectID, topicID string, data []byte) (string, error) {
	if m.publishFn != nil {
		return m.publishFn(ctx, projectID, topicID, data)
	}
	return "msg-001", nil
}

func (m *mockGCPPublisher) ListTopics(ctx context.Context, projectID string) ([]pubsub.TopicGroup, error) {
	if m.listTopicsFn != nil {
		return m.listTopicsFn(ctx, projectID)
	}
	return nil, nil
}

// injectMockGCP sets gcpFactory to return the given mock for the duration of a test.
func injectMockGCP(t *testing.T, mock pubsub.GCPPublisher) {
	t.Helper()
	orig := gcpFactory
	gcpFactory = func(_ context.Context) (pubsub.GCPPublisher, error) { return mock, nil }
	t.Cleanup(func() { gcpFactory = orig })
}

// ---- state helpers ----

// setFullState writes state.yaml with active-env, cluster-id, and nodepool-id.
func setFullState(t *testing.T, dir, envName, cid, npid string) {
	t.Helper()
	statePath := filepath.Join(dir, "state.yaml")
	content := fmt.Sprintf("active-environment: %s\ncluster-id: %s\nnodepool-id: %s\n", envName, cid, npid)
	if err := os.WriteFile(statePath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
}

// setupPubSubEnv creates a temp dir with an active env (no HTTP server needed).
func setupPubSubEnv(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	makeEnv(t, dir, "test", "http://api.example.com")
	setActiveEnv(t, dir, "test")
	return dir
}

// ---- hf pubsub list ----

func TestPubSubList_NoTopics(t *testing.T) {
	mock := &mockGCPPublisher{
		listTopicsFn: func(_ context.Context, _ string) ([]pubsub.TopicGroup, error) {
			return nil, nil
		},
	}
	injectMockGCP(t, mock)
	dir := setupPubSubEnv(t)
	out, err := runCmd(t, dir, "pubsub", "list")
	if err != nil {
		t.Fatalf("pubsub list: %v", err)
	}
	if !strings.Contains(out, "No topics found.") {
		t.Errorf("expected 'No topics found.', got: %q", out)
	}
}

func TestPubSubList_WithTopics(t *testing.T) {
	mock := &mockGCPPublisher{
		listTopicsFn: func(_ context.Context, _ string) ([]pubsub.TopicGroup, error) {
			return []pubsub.TopicGroup{
				{Topic: "topic-alpha", Subscriptions: []string{"sub-1", "sub-2"}},
				{Topic: "topic-beta", Subscriptions: nil},
			}, nil
		},
	}
	injectMockGCP(t, mock)
	dir := setupPubSubEnv(t)
	out, err := runCmd(t, dir, "pubsub", "list")
	if err != nil {
		t.Fatalf("pubsub list: %v", err)
	}
	if !strings.Contains(out, "topic-alpha") {
		t.Errorf("expected 'topic-alpha' in output, got: %q", out)
	}
	if !strings.Contains(out, "    sub-1") {
		t.Errorf("expected '    sub-1' (4-space indent) in output, got: %q", out)
	}
	if !strings.Contains(out, "topic-beta") {
		t.Errorf("expected 'topic-beta' in output, got: %q", out)
	}
}

func TestPubSubList_Filter(t *testing.T) {
	mock := &mockGCPPublisher{
		listTopicsFn: func(_ context.Context, _ string) ([]pubsub.TopicGroup, error) {
			return []pubsub.TopicGroup{
				{Topic: "fleet-events", Subscriptions: []string{"fleet-sub"}},
				{Topic: "other-topic", Subscriptions: []string{"other-sub"}},
			}, nil
		},
	}
	injectMockGCP(t, mock)
	dir := setupPubSubEnv(t)
	out, err := runCmd(t, dir, "pubsub", "list", "fleet")
	if err != nil {
		t.Fatalf("pubsub list: %v", err)
	}
	if !strings.Contains(out, "fleet-events") {
		t.Errorf("expected 'fleet-events' in output")
	}
	if strings.Contains(out, "other-topic") {
		t.Errorf("expected 'other-topic' to be filtered out")
	}
}

// ---- hf pubsub publish cluster ----

func TestPubSubPublishCluster_MissingClusterID(t *testing.T) {
	injectMockGCP(t, &mockGCPPublisher{})
	dir := setupPubSubEnv(t)
	// No cluster-id in state
	_, err := runCmd(t, dir, "pubsub", "publish", "cluster", "my-topic")
	if err == nil {
		t.Fatal("expected error for missing cluster-id")
	}
	if !strings.Contains(err.Error(), "No cluster-id") {
		t.Errorf("error = %q, want to contain 'No cluster-id'", err.Error())
	}
}

func TestPubSubPublishCluster_Success(t *testing.T) {
	var gotTopic string
	mock := &mockGCPPublisher{
		publishFn: func(_ context.Context, _, topicID string, _ []byte) (string, error) {
			gotTopic = topicID
			return "msg-xyz", nil
		},
	}
	injectMockGCP(t, mock)
	dir := setupPubSubEnv(t)
	setClusterIDInState(t, dir, "test", clusterID)

	out, err := runCmd(t, dir, "pubsub", "publish", "cluster", "my-topic")
	if err != nil {
		t.Fatalf("pubsub publish cluster: %v", err)
	}
	if gotTopic != "my-topic" {
		t.Errorf("published to topic %q, want %q", gotTopic, "my-topic")
	}
	if !strings.Contains(out, "[INFO] Published cluster") {
		t.Errorf("expected success log, got: %q", out)
	}
	if !strings.Contains(out, "msg-xyz") {
		t.Errorf("expected msg-id in output, got: %q", out)
	}
}

// ---- hf pubsub publish nodepool ----

func TestPubSubPublishNodePool_MissingClusterID(t *testing.T) {
	injectMockGCP(t, &mockGCPPublisher{})
	dir := setupPubSubEnv(t)
	_, err := runCmd(t, dir, "pubsub", "publish", "nodepool", "my-topic")
	if err == nil {
		t.Fatal("expected error for missing cluster-id")
	}
	if !strings.Contains(err.Error(), "No cluster-id") {
		t.Errorf("error = %q, want 'No cluster-id'", err.Error())
	}
}

func TestPubSubPublishNodePool_MissingNodePoolID(t *testing.T) {
	injectMockGCP(t, &mockGCPPublisher{})
	dir := setupPubSubEnv(t)
	setClusterIDInState(t, dir, "test", clusterID)
	// nodepool-id not set
	_, err := runCmd(t, dir, "pubsub", "publish", "nodepool", "my-topic")
	if err == nil {
		t.Fatal("expected error for missing nodepool-id")
	}
	if !strings.Contains(err.Error(), "No nodepool-id") {
		t.Errorf("error = %q, want 'No nodepool-id'", err.Error())
	}
}

func TestPubSubPublishNodePool_Success(t *testing.T) {
	var gotData []byte
	mock := &mockGCPPublisher{
		publishFn: func(_ context.Context, _, _ string, data []byte) (string, error) {
			gotData = data
			return "msg-np-001", nil
		},
	}
	injectMockGCP(t, mock)
	dir := setupPubSubEnv(t)
	setFullState(t, dir, "test", clusterID, nodepoolID)

	out, err := runCmd(t, dir, "pubsub", "publish", "nodepool", "np-topic")
	if err != nil {
		t.Fatalf("pubsub publish nodepool: %v", err)
	}
	if !strings.Contains(string(gotData), "nodepool.reconcile") {
		t.Errorf("expected nodepool event type in payload, got: %s", gotData)
	}
	if !strings.Contains(out, "[INFO] Published nodepool") {
		t.Errorf("expected success log, got: %q", out)
	}
}

// ---- GCP credentials error ----

func TestPubSubList_NoCredentials(t *testing.T) {
	orig := gcpFactory
	gcpFactory = func(_ context.Context) (pubsub.GCPPublisher, error) {
		return nil, pubsub.ErrNoCredentials
	}
	t.Cleanup(func() { gcpFactory = orig })

	dir := setupPubSubEnv(t)
	_, err := runCmd(t, dir, "pubsub", "list")
	if err == nil {
		t.Fatal("expected error for missing GCP credentials")
	}
	if !strings.Contains(err.Error(), "GCP credentials not found") {
		t.Errorf("error = %q, want 'GCP credentials not found'", err.Error())
	}
}

