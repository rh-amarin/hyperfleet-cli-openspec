package repos

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// ghHandler returns an http.Handler that mocks the GitHub REST API.
// commitSHAs maps "org/repo" -> SHA string (empty = no commits).
// prURLs maps "org/repo" -> [htmlURL, headRef] (nil = no PRs).
func ghHandler(commitSHAs map[string]string, prURLs map[string][2]string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// path: /repos/{org}/{repo}/commits  or  /repos/{org}/{repo}/pulls
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/repos/"), "/")
		if len(parts) < 3 {
			http.NotFound(w, r)
			return
		}
		key := parts[0] + "/" + parts[1]
		action := parts[2]
		w.Header().Set("Content-Type", "application/json")
		switch action {
		case "commits":
			sha := commitSHAs[key]
			if sha == "" {
				json.NewEncoder(w).Encode([]map[string]any{})
				return
			}
			json.NewEncoder(w).Encode([]map[string]any{{"sha": sha}})
		case "pulls":
			info := prURLs[key]
			if info[0] == "" {
				json.NewEncoder(w).Encode([]map[string]any{})
				return
			}
			json.NewEncoder(w).Encode([]map[string]any{
				{"html_url": info[0], "head": map[string]any{"ref": info[1]}},
			})
		default:
			http.NotFound(w, r)
		}
	})
}

// quayHandler returns an http.Handler that mocks the Quay tag API.
// tags maps "namespace/repo" -> list of quayTag.
func quayHandler(tags map[string][]quayTag) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// path: /api/v1/repository/{ns}/{repo}/tag/
		trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/repository/")
		parts := strings.SplitN(trimmed, "/", 3)
		if len(parts) < 2 {
			http.NotFound(w, r)
			return
		}
		key := parts[0] + "/" + parts[1]
		ts := tags[key]
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(quayTagsResponse{Tags: ts})
	})
}

func TestLatestCommit_found(t *testing.T) {
	commitSHAs := map[string]string{
		defaultGHOrg + "/hyperfleet-api": "abc1234567890",
	}
	ghSrv := httptest.NewServer(ghHandler(commitSHAs, nil))
	defer ghSrv.Close()

	c := newForTesting(ghSrv.URL, "http://unused", defaultQuayNS)
	got := c.latestCommit(context.Background(), "hyperfleet-api")
	if got != "abc1234" {
		t.Errorf("want abc1234, got %q", got)
	}
}

func TestLatestCommit_none(t *testing.T) {
	ghSrv := httptest.NewServer(ghHandler(nil, nil))
	defer ghSrv.Close()

	c := newForTesting(ghSrv.URL, "http://unused", defaultQuayNS)
	got := c.latestCommit(context.Background(), "hyperfleet-api")
	if got != "-" {
		t.Errorf("want -, got %q", got)
	}
}

func TestLatestPR_found(t *testing.T) {
	prURLs := map[string][2]string{
		defaultGHOrg + "/hyperfleet-api": {"https://github.com/org/repo/pull/42", "feature-x"},
	}
	ghSrv := httptest.NewServer(ghHandler(nil, prURLs))
	defer ghSrv.Close()

	c := newForTesting(ghSrv.URL, "http://unused", defaultQuayNS)
	gotURL, gotBranch := c.latestPR(context.Background(), "hyperfleet-api")
	if gotURL != "https://github.com/org/repo/pull/42" {
		t.Errorf("pr url: want https://github.com/org/repo/pull/42, got %q", gotURL)
	}
	if gotBranch != "feature-x" {
		t.Errorf("pr branch: want feature-x, got %q", gotBranch)
	}
}

func TestLatestPR_none(t *testing.T) {
	ghSrv := httptest.NewServer(ghHandler(nil, nil))
	defer ghSrv.Close()

	c := newForTesting(ghSrv.URL, "http://unused", defaultQuayNS)
	gotURL, gotBranch := c.latestPR(context.Background(), "hyperfleet-api")
	if gotURL != "-" || gotBranch != "-" {
		t.Errorf("want -, -, got %q, %q", gotURL, gotBranch)
	}
}

func TestLatestQuayTag_found(t *testing.T) {
	tags := map[string][]quayTag{
		defaultQuayNS + "/hyperfleet-api": {
			{Name: "latest", StartTS: 200, ManifestDigest: "sha256:abc"},
			{Name: "v0.2.0-20240115", StartTS: 150, ManifestDigest: "sha256:abc"},
			{Name: "v0.1.0-20240101", StartTS: 100, ManifestDigest: "sha256:def"},
		},
	}
	quaySrv := httptest.NewServer(quayHandler(tags))
	defer quaySrv.Close()

	c := newForTesting("http://unused", quaySrv.URL, defaultQuayNS)
	tag, aliases := c.latestQuayTag(context.Background(), "hyperfleet-api")
	if tag != "v0.2.0-20240115" {
		t.Errorf("tag: want v0.2.0-20240115, got %q", tag)
	}
	// "latest" shares the same digest as the primary tag → should appear as alias
	if aliases != "latest" {
		t.Errorf("aliases: want \"latest\", got %q", aliases)
	}
}

func TestLatestQuayTag_notFound(t *testing.T) {
	quaySrv := httptest.NewServer(quayHandler(nil))
	defer quaySrv.Close()

	// quayHandler returns empty tags for unknown repos
	c := newForTesting("http://unused", quaySrv.URL, defaultQuayNS)
	tag, aliases := c.latestQuayTag(context.Background(), "architecture")
	if tag != "-" || aliases != "-" {
		t.Errorf("want -, -, got %q, %q", tag, aliases)
	}
}

func TestFetchAll_parallel(t *testing.T) {
	commitSHAs := make(map[string]string)
	for _, repo := range TrackedRepos {
		commitSHAs[defaultGHOrg+"/"+repo] = "deadbeef1234"
	}
	ghSrv := httptest.NewServer(ghHandler(commitSHAs, nil))
	defer ghSrv.Close()
	quaySrv := httptest.NewServer(quayHandler(nil))
	defer quaySrv.Close()

	c := newForTesting(ghSrv.URL, quaySrv.URL, defaultQuayNS)
	results := c.FetchAll(context.Background())

	if len(results) != len(TrackedRepos) {
		t.Fatalf("want %d results, got %d", len(TrackedRepos), len(results))
	}
	for i, r := range results {
		wantRepo := defaultGHOrg + "/" + TrackedRepos[i]
		if r.Repository != wantRepo {
			t.Errorf("result[%d].Repository: want %q, got %q", i, wantRepo, r.Repository)
		}
		if r.Commit != "deadbee" {
			t.Errorf("result[%d].Commit: want deadbee, got %q", i, r.Commit)
		}
	}
}

func TestLatestQuayTag_noAliases(t *testing.T) {
	tags := map[string][]quayTag{
		defaultQuayNS + "/hyperfleet-api": {
			{Name: "v0.1.0-20240101", StartTS: 100, ManifestDigest: "sha256:def"},
		},
	}
	quaySrv := httptest.NewServer(quayHandler(tags))
	defer quaySrv.Close()

	c := newForTesting("http://unused", quaySrv.URL, defaultQuayNS)
	tag, aliases := c.latestQuayTag(context.Background(), "hyperfleet-api")
	if tag != "v0.1.0-20240101" {
		t.Errorf("tag: want v0.1.0-20240101, got %q", tag)
	}
	if aliases != "-" {
		t.Errorf("aliases: want -, got %q", aliases)
	}
}
