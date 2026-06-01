package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	testChannelID = "chan-001"
)

var channelJSON = fmt.Sprintf(`{
  "id": "%s",
  "kind": "Channel",
  "name": "test-channel",
  "generation": 1
}`, testChannelID)

var channelListJSON = fmt.Sprintf(`{
  "items": [%s],
  "kind": "ChannelList",
  "page": 1,
  "size": 1,
  "total": 1
}`, channelJSON)

var versionListJSON = `{
  "items": [],
  "kind": "VersionList",
  "page": 1,
  "size": 0,
  "total": 0
}`

func makeResourceEnv(t *testing.T, dir, apiURL string) {
	t.Helper()
	content := fmt.Sprintf(`hyperfleet:
  api-url: %s
  api-version: v1
  token: test-token
resource-types:
  channels:
    path: channels
    create-template: channels.json
  versions:
    parent: channels
    path: "channels/{channel_id}/versions"
`, apiURL)
	makeEnvRaw(t, dir, "test", content)
	setActiveEnv(t, dir, "test")
}

func runResourceCmd(t *testing.T, dir string, args ...string) (string, error) {
	t.Helper()
	resetGenericFlags()
	resetResourceRegistrationForTest()
	outputFmt = "json"
	curlMode = false
	return runCmd(t, dir, args...)
}

func TestResourceTypes_Empty(t *testing.T) {
	dir := t.TempDir()
	makeEnv(t, dir, "test", "http://localhost:8000")
	setActiveEnv(t, dir, "test")

	out, err := runResourceCmd(t, dir, "resource", "types")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "No resource types configured") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestResourceChannelsList(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/hyperfleet/v1/channels" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, channelListJSON)
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)

	out, err := runResourceCmd(t, dir, "resource", "channels", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, testChannelID) {
		t.Fatalf("expected channel id in output: %q", out)
	}
}

func TestResourceVersionsList_RequiresParentState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)

	_, err := runResourceCmd(t, dir, "resource", "versions", "list")
	if err == nil {
		t.Fatal("expected error without channels")
	}
	if !strings.Contains(err.Error(), "channels") {
		t.Fatalf("error: %v", err)
	}
}

func TestResourceVersionsList_WithParentState(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, versionListJSON)
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)
	statePath := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(statePath, []byte("active-environment: test\nchannels: "+testChannelID+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := runResourceCmd(t, dir, "resource", "versions", "list")
	if err != nil {
		t.Fatal(err)
	}
	want := "/api/hyperfleet/v1/channels/" + testChannelID + "/versions"
	if gotPath != want {
		t.Fatalf("path: got %q want %q", gotPath, want)
	}
}

func TestResourceChannelsSearch_SetsState(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, channelListJSON)
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)

	_, err := runResourceCmd(t, dir, "resource", "channels", "search", "test-channel")
	if err != nil {
		t.Fatal(err)
	}
	state, err := os.ReadFile(filepath.Join(dir, "state.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(state), "channels: "+testChannelID) {
		t.Fatalf("state not updated: %q", state)
	}
}

func TestResourceChannelsCreate_CurlDryRun(t *testing.T) {
	var postCount int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			postCount++
		}
		t.Error("unexpected HTTP request in curl dry-run mode")
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)
	tmplDir := filepath.Join(dir, "templates")
	if err := os.MkdirAll(tmplDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(tmplDir, "channels.json"), []byte(`{"name":"from-template","kind":"Channel"}`), 0600); err != nil {
		t.Fatal(err)
	}

	var stdout string
	stderr := captureStderr(t, func() {
		var err error
		stdout, err = runResourceCmd(t, dir, "resource", "channels", "create", "--name", "new-channel", "--curl")
		if err != nil {
			t.Fatal(err)
		}
	})
	if postCount != 0 {
		t.Fatalf("expected no POST in dry-run, got %d", postCount)
	}
	if !strings.Contains(stderr, "[CURL] curl -s -X POST") {
		t.Fatalf("expected POST curl on stderr, got: %q", stderr)
	}
	if strings.Contains(stdout, testChannelID) {
		t.Fatalf("expected no create response on stdout: %q", stdout)
	}
}

func TestResourceChannelsAdapterReport(t *testing.T) {
	var capturedBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		want := "/api/hyperfleet/v1/channels/" + testChannelID + "/statuses"
		if r.Method == http.MethodPut && r.URL.Path == want {
			capturedBody, _ = io.ReadAll(r.Body)
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"adapter": "ch-adapter",
				"observed_generation": 2,
				"observed_time": "2026-05-10T00:00:00Z",
				"conditions": [
					{"type": "Available", "status": "True", "reason": "ManualStatusPost", "message": "Status reported via hf rs adapter-report"}
				],
				"created_time": "2026-05-10T00:00:00Z",
				"last_report_time": "2026-05-10T00:00:00Z"
			}`)
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)
	statePath := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(statePath, []byte("active-environment: test\nchannels: "+testChannelID+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	out, err := runResourceCmd(t, dir, "rs", "channels", "adapter-report", "ch-adapter", "True", "2")
	if err != nil {
		t.Fatalf("adapter-report: %v", err)
	}
	if !strings.Contains(out, "ch-adapter") {
		t.Fatalf("expected adapter in output: %q", out)
	}
	if len(capturedBody) == 0 {
		t.Fatal("expected request body to be captured")
	}
	var body map[string]any
	if err := json.Unmarshal(capturedBody, &body); err != nil {
		t.Fatalf("captured body is not JSON: %v", err)
	}
	if body["adapter"] != "ch-adapter" {
		t.Errorf("body adapter: got %v", body["adapter"])
	}
	conds, ok := body["conditions"].([]any)
	if !ok || len(conds) != 4 {
		t.Errorf("expected 4 conditions, got %v", body["conditions"])
	}
}

func TestResourceChannelsAdapterReport_InvalidStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("HTTP request should not be made for invalid status")
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)
	statePath := filepath.Join(dir, "state.yaml")
	if err := os.WriteFile(statePath, []byte("active-environment: test\nchannels: "+testChannelID+"\n"), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := runResourceCmd(t, dir, "rs", "channels", "adapter-report", "ch-adapter", "INVALID", "2")
	if err == nil {
		t.Fatal("expected error for invalid status value")
	}
	if !strings.Contains(err.Error(), "Invalid status value") {
		t.Errorf("expected 'Invalid status value' error, got: %v", err)
	}
}

func TestResourceRsAlias(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, channelListJSON)
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)

	out, err := runResourceCmd(t, dir, "rs", "channels", "list")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, testChannelID) {
		t.Fatalf("expected channel id via rs alias: %q", out)
	}
}

func TestEnvShow_ListsResourceTypes(t *testing.T) {
	dir := t.TempDir()
	makeResourceEnv(t, dir, "http://localhost:8000")

	out, err := runCmd(t, dir, "env", "show", "--no-color")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"resource-types:", "channels:", "path: channels", "versions:", "parent: channels"} {
		if !strings.Contains(out, want) {
			t.Fatalf("expected %q in env show output:\n%s", want, out)
		}
	}
}
