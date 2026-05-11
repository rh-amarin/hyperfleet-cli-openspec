// Package insights parses HyperFleet component log lines into structured summaries.
package insights

import (
	"encoding/json"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/kube"
)

// ---- API ----

// APIEndpointStat holds request counts for one method+path combination.
type APIEndpointStat struct {
	Method string
	Path   string
	OK     int
	Err    int
}

// APIInsights is the parsed summary from API pod logs.
type APIInsights struct {
	Endpoints []APIEndpointStat
}

var uuidRE = regexp.MustCompile(`[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)

// ParseAPILogs parses JSON API log lines and returns per-endpoint request counts.
func ParseAPILogs(lines []string) APIInsights {
	type key struct{ method, path string }
	counts := map[key]*APIEndpointStat{}

	for _, line := range lines {
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		if m["message"] != "HTTP request completed" {
			continue
		}
		method, _ := m["method"].(string)
		path, _ := m["path"].(string)
		if method == "" || path == "" {
			continue
		}
		path = uuidRE.ReplaceAllString(path, ":id")
		statusCode := 0
		switch v := m["status_code"].(type) {
		case float64:
			statusCode = int(v)
		}
		k := key{method, path}
		if counts[k] == nil {
			counts[k] = &APIEndpointStat{Method: method, Path: path}
		}
		if statusCode >= 400 {
			counts[k].Err++
		} else {
			counts[k].OK++
		}
	}

	stats := make([]APIEndpointStat, 0, len(counts))
	for _, v := range counts {
		stats = append(stats, *v)
	}
	sort.Slice(stats, func(i, j int) bool {
		if stats[i].Path != stats[j].Path {
			return stats[i].Path < stats[j].Path
		}
		return stats[i].Method < stats[j].Method
	})
	return APIInsights{Endpoints: stats}
}

// ---- Sentinel ----

// SentinelTopicStat holds cycle/publish counts for one topic.
type SentinelTopicStat struct {
	Topic     string
	Cycles    int
	Published int
	Skipped   int
}

// SentinelInsights is the parsed summary from sentinel pod logs.
type SentinelInsights struct {
	Topics []SentinelTopicStat
}

// ParseSentinelLogs parses JSON sentinel log lines and returns per-topic cycle/publish counts.
func ParseSentinelLogs(lines []string) SentinelInsights {
	counts := map[string]*SentinelTopicStat{}

	for _, line := range lines {
		if !strings.HasPrefix(line, "{") {
			continue
		}
		var m map[string]interface{}
		if err := json.Unmarshal([]byte(line), &m); err != nil {
			continue
		}
		msg, _ := m["message"].(string)
		if !strings.HasPrefix(msg, "Trigger cycle completed") {
			continue
		}
		topic, _ := m["topic"].(string)
		if topic == "" {
			continue
		}
		published := extractInt(msg, "published=")
		skipped := extractInt(msg, "skipped=")

		if counts[topic] == nil {
			counts[topic] = &SentinelTopicStat{Topic: topic}
		}
		counts[topic].Cycles++
		counts[topic].Published += published
		counts[topic].Skipped += skipped
	}

	stats := make([]SentinelTopicStat, 0, len(counts))
	for _, v := range counts {
		stats = append(stats, *v)
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Topic < stats[j].Topic
	})
	return SentinelInsights{Topics: stats}
}

// extractInt parses the integer value after prefix in s (e.g. "published=3").
func extractInt(s, prefix string) int {
	idx := strings.Index(s, prefix)
	if idx < 0 {
		return 0
	}
	rest := s[idx+len(prefix):]
	end := strings.IndexAny(rest, " \t\n")
	if end >= 0 {
		rest = rest[:end]
	}
	n, _ := strconv.Atoi(rest)
	return n
}

// ---- Adapters ----

// AdapterPhaseStat holds the execution count for one phase of one adapter.
type AdapterPhaseStat struct {
	Name  string
	Count int
}

// AdapterStat holds execution and phase counts for one adapter component.
type AdapterStat struct {
	Name       string
	Executions int
	Phases     []AdapterPhaseStat
}

// AdapterInsights is the parsed summary from adapter pod logs.
type AdapterInsights struct {
	Adapters []AdapterStat
}

// ParseAdapterLogs parses logfmt adapter log lines and returns per-adapter execution/phase counts.
func ParseAdapterLogs(lines []string) AdapterInsights {
	type adapterKey = string
	executions := map[adapterKey]int{}
	phases := map[adapterKey]map[string]int{}

	for _, line := range lines {
		if strings.HasPrefix(line, "{") {
			continue
		}
		fields := kube.ParseLogfmt(line)
		component := fields["component"]
		if component == "" {
			// also try msg field parsing for logfmt without explicit component
			component = fields["msg"]
			if component == "" {
				continue
			}
			component = ""
			continue
		}
		msg := fields["msg"]

		if msg == "Processing event" {
			executions[component]++
			if phases[component] == nil {
				phases[component] = map[string]int{}
			}
		}

		// "Phase <name>: RUNNING" marks a phase that executed
		if strings.HasPrefix(msg, "Phase ") && strings.Contains(msg, ": RUNNING") {
			phaseName := strings.TrimPrefix(msg, "Phase ")
			if colon := strings.Index(phaseName, ":"); colon >= 0 {
				phaseName = phaseName[:colon]
			}
			if phases[component] == nil {
				phases[component] = map[string]int{}
			}
			phases[component][phaseName]++
		}
	}

	// Merge all known components from both executions and phases maps.
	allComponents := map[string]struct{}{}
	for c := range executions {
		allComponents[c] = struct{}{}
	}
	for c := range phases {
		allComponents[c] = struct{}{}
	}

	stats := make([]AdapterStat, 0, len(allComponents))
	for component := range allComponents {
		phaseMap := phases[component]
		phaseStats := make([]AdapterPhaseStat, 0, len(phaseMap))
		// Canonical phase order.
		for _, pn := range []string{"param_extraction", "preconditions", "resources", "post_actions"} {
			if c, ok := phaseMap[pn]; ok {
				phaseStats = append(phaseStats, AdapterPhaseStat{Name: pn, Count: c})
			}
		}
		// Any extra phases not in canonical order.
		for pn, c := range phaseMap {
			if pn != "param_extraction" && pn != "preconditions" && pn != "resources" && pn != "post_actions" {
				phaseStats = append(phaseStats, AdapterPhaseStat{Name: pn, Count: c})
			}
		}
		stats = append(stats, AdapterStat{
			Name:       component,
			Executions: executions[component],
			Phases:     phaseStats,
		})
	}
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].Name < stats[j].Name
	})
	return AdapterInsights{Adapters: stats}
}
