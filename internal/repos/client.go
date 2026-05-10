// Package repos provides a GitHub + Quay registry client for the HyperFleet
// repository status overview.
package repos

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"sync"

	"github.com/google/go-github/v62/github"
	"golang.org/x/oauth2"
)

const (
	defaultGHOrg    = "openshift-hyperfleet"
	defaultQuayBase = "https://quay.io"
	defaultQuayNS   = "openshift-hyperfleet"
)

// TrackedRepos is the ordered list of HyperFleet repositories to report on.
var TrackedRepos = []string{
	"hyperfleet-api-spec",
	"hyperfleet-api",
	"hyperfleet-sentinel",
	"hyperfleet-adapter",
	"hyperfleet-infra",
	"hyperfleet-e2e",
	"architecture",
}

// RepoStatus holds the display data for one repository.
type RepoStatus struct {
	Repository  string `json:"repository"`
	Commit      string `json:"commit"`
	PRURL       string `json:"pr_url"`
	PRBranch    string `json:"pr_branch"`
	QuayTag     string `json:"quay_tag"`
	QuayAliases string `json:"quay_aliases"`
}

// Client fetches repository status from GitHub and Quay.
type Client struct {
	gh         *github.Client
	ghOrg      string
	quayBase   string
	quayNS     string
	httpClient *http.Client
}

// New creates a production Client.
// token is the GitHub personal access token (empty = unauthenticated, 60 req/h).
// quayBase: Quay API base URL (empty = "https://quay.io").
// quayNS: Quay namespace (empty = "openshift-hyperfleet").
func New(token, quayBase, quayNS string) *Client {
	var authClient *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		authClient = oauth2.NewClient(context.Background(), ts)
	}
	gh := github.NewClient(authClient)
	if quayBase == "" {
		quayBase = defaultQuayBase
	}
	if quayNS == "" {
		quayNS = defaultQuayNS
	}
	return &Client{
		gh:         gh,
		ghOrg:      defaultGHOrg,
		quayBase:   quayBase,
		quayNS:     quayNS,
		httpClient: http.DefaultClient,
	}
}

// newForTesting creates a Client that targets mock servers instead of live APIs.
func newForTesting(ghBaseURL, quayBase, quayNS string) *Client {
	gh := github.NewClient(nil)
	u, _ := url.Parse(ghBaseURL)
	if !strings.HasSuffix(u.Path, "/") {
		u.Path += "/"
	}
	gh.BaseURL = u
	gh.UploadURL = u
	return &Client{
		gh:         gh,
		ghOrg:      defaultGHOrg,
		quayBase:   quayBase,
		quayNS:     quayNS,
		httpClient: http.DefaultClient,
	}
}

// FetchAll fetches status for all tracked repos in parallel, returning results
// in the same order as TrackedRepos.
func (c *Client) FetchAll(ctx context.Context) []RepoStatus {
	results := make([]RepoStatus, len(TrackedRepos))
	var wg sync.WaitGroup
	wg.Add(len(TrackedRepos))
	for i, repo := range TrackedRepos {
		i, repo := i, repo
		go func() {
			defer wg.Done()
			results[i] = c.fetchOne(ctx, repo)
		}()
	}
	wg.Wait()
	return results
}

// fetchOne fetches status for a single repository.
func (c *Client) fetchOne(ctx context.Context, repo string) RepoStatus {
	s := RepoStatus{
		Repository:  c.ghOrg + "/" + repo,
		Commit:      "-",
		PRURL:       "-",
		PRBranch:    "-",
		QuayTag:     "-",
		QuayAliases: "-",
	}
	s.Commit = c.latestCommit(ctx, repo)
	s.PRURL, s.PRBranch = c.latestPR(ctx, repo)
	s.QuayTag, s.QuayAliases = c.latestQuayTag(ctx, repo)
	return s
}

// latestCommit returns the 7-char short SHA of the latest commit on the default branch.
func (c *Client) latestCommit(ctx context.Context, repo string) string {
	commits, _, err := c.gh.Repositories.ListCommits(ctx, c.ghOrg, repo,
		&github.CommitsListOptions{ListOptions: github.ListOptions{PerPage: 1}})
	if err != nil || len(commits) == 0 {
		return "-"
	}
	sha := commits[0].GetSHA()
	if len(sha) >= 7 {
		return sha[:7]
	}
	return sha
}

// latestPR returns the HTML URL and head branch of the most recently created open PR.
func (c *Client) latestPR(ctx context.Context, repo string) (string, string) {
	prs, _, err := c.gh.PullRequests.List(ctx, c.ghOrg, repo, &github.PullRequestListOptions{
		State:       "open",
		ListOptions: github.ListOptions{PerPage: 1},
	})
	if err != nil || len(prs) == 0 {
		return "-", "-"
	}
	pr := prs[0]
	prURL := pr.GetHTMLURL()
	branch := ""
	if pr.Head != nil {
		branch = pr.Head.GetRef()
	}
	if prURL == "" {
		prURL = "-"
	}
	if branch == "" {
		branch = "-"
	}
	return prURL, branch
}

// quayTagsResponse is the Quay API tag list response.
type quayTagsResponse struct {
	Tags []quayTag `json:"tags"`
}

type quayTag struct {
	Name           string `json:"name"`
	StartTS        int64  `json:"start_ts"`
	ManifestDigest string `json:"manifest_digest"`
}

// latestQuayTag returns the most recent non-"latest" tag name and comma-separated aliases.
func (c *Client) latestQuayTag(ctx context.Context, repo string) (string, string) {
	reqURL := fmt.Sprintf("%s/api/v1/repository/%s/%s/tag/?limit=10", c.quayBase, c.quayNS, repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "-", "-"
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "-", "-"
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "-", "-"
	}
	var data quayTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return "-", "-"
	}
	if len(data.Tags) == 0 {
		return "-", "-"
	}
	sort.Slice(data.Tags, func(i, j int) bool {
		return data.Tags[i].StartTS > data.Tags[j].StartTS
	})
	var primary *quayTag
	for i := range data.Tags {
		n := data.Tags[i].Name
		if n != "latest" && !strings.HasPrefix(n, "sha256:") {
			primary = &data.Tags[i]
			break
		}
	}
	if primary == nil {
		primary = &data.Tags[0]
	}
	var aliases []string
	for _, t := range data.Tags {
		if t.Name != primary.Name && t.ManifestDigest == primary.ManifestDigest {
			aliases = append(aliases, t.Name)
		}
	}
	aliasStr := "-"
	if len(aliases) > 0 {
		aliasStr = strings.Join(aliases, ", ")
	}
	return primary.Name, aliasStr
}
