package insights

import (
	"testing"
)

// ---- API parser tests ----

func TestParseAPILogs_Normal(t *testing.T) {
	lines := []string{
		`{"message":"HTTP request completed","method":"GET","path":"/api/hyperfleet/v1/clusters","status_code":200}`,
		`{"message":"HTTP request completed","method":"GET","path":"/api/hyperfleet/v1/clusters","status_code":200}`,
		`{"message":"HTTP request completed","method":"GET","path":"/api/hyperfleet/v1/clusters/019e1240-dc6b-7706-b17c-f8d7646e6f65","status_code":200}`,
		`{"message":"HTTP request completed","method":"GET","path":"/api/hyperfleet/v1/clusters/019e1240-dc6b-7706-b17c-f8d7646e6f65","status_code":404}`,
		`{"message":"HTTP request completed","method":"POST","path":"/api/hyperfleet/v1/clusters/019e1240-dc6b-7706-b17c-f8d7646e6f65/statuses","status_code":201}`,
		`{"message":"HTTP request received","method":"GET","path":"/api/hyperfleet/v1/clusters","status_code":0}`,
	}
	got := ParseAPILogs(lines)

	if len(got.Endpoints) != 3 {
		t.Fatalf("expected 3 endpoints, got %d: %+v", len(got.Endpoints), got.Endpoints)
	}

	byKey := map[string]APIEndpointStat{}
	for _, e := range got.Endpoints {
		byKey[e.Method+" "+e.Path] = e
	}

	e := byKey["GET /api/hyperfleet/v1/clusters"]
	if e.OK != 2 || e.Err != 0 {
		t.Errorf("GET /clusters: want OK=2 Err=0, got OK=%d Err=%d", e.OK, e.Err)
	}

	e = byKey["GET /api/hyperfleet/v1/clusters/:id"]
	if e.OK != 1 || e.Err != 1 {
		t.Errorf("GET /clusters/:id: want OK=1 Err=1, got OK=%d Err=%d", e.OK, e.Err)
	}

	e = byKey["POST /api/hyperfleet/v1/clusters/:id/statuses"]
	if e.OK != 1 || e.Err != 0 {
		t.Errorf("POST /clusters/:id/statuses: want OK=1 Err=0, got OK=%d Err=%d", e.OK, e.Err)
	}
}

func TestParseAPILogs_Empty(t *testing.T) {
	got := ParseAPILogs(nil)
	if len(got.Endpoints) != 0 {
		t.Errorf("expected empty endpoints for nil input, got %d", len(got.Endpoints))
	}
	got = ParseAPILogs([]string{})
	if len(got.Endpoints) != 0 {
		t.Errorf("expected empty endpoints for empty input, got %d", len(got.Endpoints))
	}
}

func TestParseAPILogs_SkipsNonJSON(t *testing.T) {
	lines := []string{
		`not json at all`,
		`{"message":"HTTP request received","method":"GET","path":"/api/v1/clusters","status_code":0}`,
	}
	got := ParseAPILogs(lines)
	if len(got.Endpoints) != 0 {
		t.Errorf("expected 0 endpoints (no completed requests), got %d", len(got.Endpoints))
	}
}

// ---- Sentinel parser tests ----

func TestParseSentinelLogs_Normal(t *testing.T) {
	lines := []string{
		`{"message":"Trigger cycle completed total=1 published=1 skipped=0 duration=0.012s","topic":"ns-clusters","level":"info"}`,
		`{"message":"Trigger cycle completed total=1 published=1 skipped=0 duration=0.010s","topic":"ns-clusters","level":"info"}`,
		`{"message":"Trigger cycle completed total=2 published=0 skipped=2 duration=0.005s","topic":"ns-nodepools","level":"info"}`,
		`{"message":"Fetched resources count=1","topic":"ns-clusters","level":"info"}`,
	}
	got := ParseSentinelLogs(lines)

	if len(got.Topics) != 2 {
		t.Fatalf("expected 2 topics, got %d: %+v", len(got.Topics), got.Topics)
	}

	byTopic := map[string]SentinelTopicStat{}
	for _, s := range got.Topics {
		byTopic[s.Topic] = s
	}

	cl := byTopic["ns-clusters"]
	if cl.Cycles != 2 || cl.Published != 2 || cl.Skipped != 0 {
		t.Errorf("ns-clusters: want cycles=2 published=2 skipped=0, got %+v", cl)
	}

	np := byTopic["ns-nodepools"]
	if np.Cycles != 1 || np.Published != 0 || np.Skipped != 2 {
		t.Errorf("ns-nodepools: want cycles=1 published=0 skipped=2, got %+v", np)
	}
}

func TestParseSentinelLogs_Empty(t *testing.T) {
	got := ParseSentinelLogs(nil)
	if len(got.Topics) != 0 {
		t.Errorf("expected empty topics for nil input")
	}
}

func TestExtractInt(t *testing.T) {
	cases := []struct {
		s      string
		prefix string
		want   int
	}{
		{"published=3 skipped=0", "published=", 3},
		{"published=0 skipped=7", "skipped=", 7},
		{"total=12 published=11 skipped=1 duration=0.01s", "skipped=", 1},
		{"no match here", "published=", 0},
	}
	for _, c := range cases {
		if got := extractInt(c.s, c.prefix); got != c.want {
			t.Errorf("extractInt(%q, %q) = %d, want %d", c.s, c.prefix, got, c.want)
		}
	}
}

// ---- Adapter parser tests ----

func TestParseAdapterLogs_Normal(t *testing.T) {
	lines := []string{
		`time=2026-05-11T08:41:18Z level=INFO msg="Processing event" component=cl-deployment cluster_id=abc event_id=001`,
		`time=2026-05-11T08:41:18Z level=INFO msg="Phase param_extraction: RUNNING" component=cl-deployment event_id=001`,
		`time=2026-05-11T08:41:18Z level=INFO msg="Phase preconditions: RUNNING - 3 configured" component=cl-deployment event_id=001`,
		`time=2026-05-11T08:41:18Z level=INFO msg="Phase resources: RUNNING - 1 configured" component=cl-deployment event_id=001`,
		`time=2026-05-11T08:41:18Z level=INFO msg="Phase post_actions: RUNNING - 1 configured" component=cl-deployment event_id=001`,
		`time=2026-05-11T08:41:18Z level=INFO msg="Processing event" component=cl-job cluster_id=abc event_id=002`,
		`time=2026-05-11T08:41:18Z level=INFO msg="Phase param_extraction: RUNNING" component=cl-job event_id=002`,
		`time=2026-05-11T08:41:18Z level=DEBUG msg="Parameter extraction completed: extracted 3 params" component=cl-deployment event_id=001`,
		`{"Name":"Execute","SpanContext":{"TraceID":"abc"}}`,
	}
	got := ParseAdapterLogs(lines)

	if len(got.Adapters) != 2 {
		t.Fatalf("expected 2 adapters, got %d: %+v", len(got.Adapters), got.Adapters)
	}

	byName := map[string]AdapterStat{}
	for _, a := range got.Adapters {
		byName[a.Name] = a
	}

	dep := byName["cl-deployment"]
	if dep.Executions != 1 {
		t.Errorf("cl-deployment: want 1 execution, got %d", dep.Executions)
	}
	if len(dep.Phases) != 4 {
		t.Errorf("cl-deployment: want 4 phases, got %d: %+v", len(dep.Phases), dep.Phases)
	}

	job := byName["cl-job"]
	if job.Executions != 1 {
		t.Errorf("cl-job: want 1 execution, got %d", job.Executions)
	}
	if len(job.Phases) != 1 || job.Phases[0].Name != "param_extraction" {
		t.Errorf("cl-job: want [param_extraction], got %+v", job.Phases)
	}
}

func TestParseAdapterLogs_Empty(t *testing.T) {
	got := ParseAdapterLogs(nil)
	if len(got.Adapters) != 0 {
		t.Errorf("expected empty adapters for nil input")
	}
}

func TestParseAdapterLogs_SkipsJSONSpans(t *testing.T) {
	lines := []string{
		`{"Name":"HTTP GET","SpanContext":{"TraceID":"abc123"}}`,
		`{"Name":"Execute","SpanContext":{"TraceID":"abc123"}}`,
	}
	got := ParseAdapterLogs(lines)
	if len(got.Adapters) != 0 {
		t.Errorf("expected 0 adapters (only JSON span lines), got %d", len(got.Adapters))
	}
}
