package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const (
	overviewChannelID = "chan-001"
	overviewVersionID = "ver-001"
)

var overviewChannelJSON = fmt.Sprintf(`{
  "id": "%s",
  "kind": "Channel",
  "name": "alpha",
  "generation": 3
}`, overviewChannelID)

var overviewVersionJSON = fmt.Sprintf(`{
  "id": "%s",
  "kind": "Version",
  "name": "v1",
  "generation": 1
}`, overviewVersionID)

func TestResourceOverview_DeletionMarker(t *testing.T) {
	deletedChannelJSON := fmt.Sprintf(`{
  "id": "%s",
  "kind": "Channel",
  "name": "gone",
  "generation": 4,
  "deleted_time": "2024-06-01T12:00:00Z"
}`, overviewChannelID)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/api/hyperfleet/v1/channels" {
			fmt.Fprintf(w, `{"items":[%s],"kind":"ChannelList","page":1,"size":1,"total":1}`, deletedChannelJSON)
			return
		}
		if r.URL.Path == "/api/hyperfleet/v1/channels/"+overviewChannelID+"/versions" {
			fmt.Fprint(w, `{"items":[],"kind":"VersionList","page":1,"size":0,"total":0}`)
			return
		}
		t.Errorf("unexpected path: %s", r.URL.Path)
		http.NotFound(w, r)
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)

	resetGenericFlags()
	resourceOverviewWatch = false
	resetResourceRegistrationForTest()
	outputFmt = "table"
	curlMode = false

	out, err := runCmd(t, dir, "rs")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(out, "4 ❌") {
		t.Fatalf("expected GEN deletion marker, got:\n%s", out)
	}
}

func TestResourceOverview_HierarchicalTable(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/hyperfleet/v1/channels":
			fmt.Fprintf(w, `{"items":[%s],"kind":"ChannelList","page":1,"size":1,"total":1}`, overviewChannelJSON)
		case "/api/hyperfleet/v1/channels/" + overviewChannelID + "/versions":
			fmt.Fprintf(w, `{"items":[%s],"kind":"VersionList","page":1,"size":1,"total":1}`, overviewVersionJSON)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)

	resetGenericFlags()
	resourceOverviewWatch = false
	resetResourceRegistrationForTest()
	outputFmt = "table"
	curlMode = false

	out, err := runCmd(t, dir, "rs")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"TYPE", "ID", "NAME", "KIND", "GEN", "channels", overviewChannelID, "alpha", "└─", overviewVersionID, "v1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestResourceOverview_PartialFetchOnChildError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/hyperfleet/v1/channels":
			fmt.Fprintf(w, `{"items":[%s],"kind":"ChannelList","page":1,"size":1,"total":1}`, overviewChannelJSON)
		case "/api/hyperfleet/v1/channels/" + overviewChannelID + "/versions":
			http.Error(w, `{"title":"Not Found","status":404}`, http.StatusNotFound)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	dir := t.TempDir()
	makeResourceEnv(t, dir, ts.URL)

	resetGenericFlags()
	resourceOverviewWatch = false
	resetResourceRegistrationForTest()
	outputFmt = "table"
	curlMode = false

	out, err := runCmd(t, dir, "rs")
	if err != nil {
		t.Fatalf("expected success with partial data, got: %v", err)
	}
	if !strings.Contains(out, overviewChannelID) {
		t.Fatalf("expected parent channel in output:\n%s", out)
	}
	if !strings.Contains(out, "[WARN]") || !strings.Contains(out, "versions") {
		t.Fatalf("expected warning about versions fetch:\n%s", out)
	}
	if strings.Contains(out, overviewVersionID) {
		t.Fatalf("did not expect version row when child list failed:\n%s", out)
	}
}

func TestTreeLinePrefix(t *testing.T) {
	tests := []struct {
		depth  int
		isLast bool
		want   string
	}{
		{0, true, ""},
		{1, false, "├─ "},
		{1, true, "└─ "},
		{2, true, "   └─ "},
		{3, false, "      ├─ "},
	}
	for _, tc := range tests {
		got := treeLinePrefix(tc.depth, tc.isLast)
		if got != tc.want {
			t.Errorf("treeLinePrefix(%d,%v) = %q, want %q", tc.depth, tc.isLast, got, tc.want)
		}
	}
}
