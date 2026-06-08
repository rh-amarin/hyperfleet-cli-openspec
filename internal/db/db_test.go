package db

import (
	"testing"
)

func TestDSN(t *testing.T) {
	c := &Config{Host: "localhost", Port: "5432", Name: "mydb", User: "admin", Password: "s3cr3t"}
	want := "postgres://admin:s3cr3t@localhost:5432/mydb"
	if got := c.DSN(); got != want {
		t.Errorf("DSN() = %q; want %q", got, want)
	}
}

func TestDSNDefaults(t *testing.T) {
	c := &Config{Host: "db.example.com", Port: "5433", Name: "hyperfleet", User: "hf", Password: ""}
	dsn := c.DSN()
	if dsn != "postgres://hf:@db.example.com:5433/hyperfleet" {
		t.Errorf("unexpected DSN: %q", dsn)
	}
}

func TestCellStringNull(t *testing.T) {
	if got := cellString(nil); got != "NULL" {
		t.Errorf("nil should render as NULL, got %q", got)
	}
}

func TestCellStringBytes(t *testing.T) {
	if got := cellString([]byte("hello")); got != "hello" {
		t.Errorf("[]byte should render as string, got %q", got)
	}
}

func TestCellStringTruncation(t *testing.T) {
	long := make([]rune, 85)
	for i := range long {
		long[i] = 'x'
	}
	result := cellString(string(long))
	runes := []rune(result)
	if len(runes) != 80 {
		t.Errorf("truncated string should be 80 runes, got %d", len(runes))
	}
	if runes[79] != '…' {
		t.Errorf("last rune should be ellipsis, got %q", runes[79])
	}
}

func TestCellStringNoTruncationAt80(t *testing.T) {
	exact := make([]rune, 80)
	for i := range exact {
		exact[i] = 'a'
	}
	result := cellString(string(exact))
	if len([]rune(result)) != 80 {
		t.Errorf("80-char string should not be truncated, got len %d", len([]rune(result)))
	}
}

func TestDeleteTargetTable(t *testing.T) {
	cases := []struct {
		target string
		table  string
		ok     bool
	}{
		{"clusters", "clusters", true},
		{"nodepools", "node_pools", true},
		{"adapter_statuses", "adapter_statuses", true},
		{"resources", "resources", true},
		{"ALL", "", false},
		{"unknown", "", false},
	}
	for _, tc := range cases {
		got, ok := DeleteTargetTable(tc.target)
		if ok != tc.ok {
			t.Errorf("DeleteTargetTable(%q) ok=%v, want %v", tc.target, ok, tc.ok)
		}
		if ok && got != tc.table {
			t.Errorf("DeleteTargetTable(%q) = %q, want %q", tc.target, got, tc.table)
		}
	}
}

func TestDeleteTargets_AllIncludesResources(t *testing.T) {
	var found bool
	for _, d := range DeleteTargets {
		if d.Name == "resources" && d.Table == "resources" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("DeleteTargets missing resources: %+v", DeleteTargets)
	}
	names := ValidDeleteTargetNames()
	if len(names) != 4 {
		t.Fatalf("expected 4 delete targets, got %v", names)
	}
}
