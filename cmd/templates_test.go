package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplate_EmbeddedDefault_Cluster(t *testing.T) {
	dir := t.TempDir()

	body, err := loadTemplate("cluster", "")
	if err != nil {
		t.Fatalf("loadTemplate: %v", err)
	}
	if body["kind"] != "Cluster" {
		t.Errorf("expected kind=Cluster, got %v", body["kind"])
	}
	if body["name"] != "my-cluster" {
		t.Errorf("expected name=my-cluster, got %v", body["name"])
	}

	// File must NOT be written to the config dir.
	if _, err := os.Stat(filepath.Join(dir, "cluster-template.json")); !os.IsNotExist(err) {
		t.Error("cluster-template.json must not be written to config dir")
	}
}

func TestLoadTemplate_EmbeddedDefault_Nodepool(t *testing.T) {
	dir := t.TempDir()

	body, err := loadTemplate("nodepool", "")
	if err != nil {
		t.Fatalf("loadTemplate nodepool: %v", err)
	}
	if body["kind"] != "NodePool" {
		t.Errorf("expected kind=NodePool, got %v", body["kind"])
	}
	if body["name"] != "my-nodepool" {
		t.Errorf("expected name=my-nodepool, got %v", body["name"])
	}

	// File must NOT be written to the config dir.
	if _, err := os.Stat(filepath.Join(dir, "nodepool-template.json")); !os.IsNotExist(err) {
		t.Error("nodepool-template.json must not be written to config dir")
	}
}

func TestLoadTemplate_FlagFileOverride(t *testing.T) {
	dir := t.TempDir()
	overridePath := filepath.Join(dir, "override.json")
	if err := os.WriteFile(overridePath,
		[]byte(`{"kind":"Cluster","name":"override-cluster","labels":{},"spec":{}}`), 0600); err != nil {
		t.Fatal(err)
	}

	body, err := loadTemplate("cluster", overridePath)
	if err != nil {
		t.Fatalf("loadTemplate with -f: %v", err)
	}
	if body["name"] != "override-cluster" {
		t.Errorf("expected override-cluster, got %v", body["name"])
	}
}

func TestLoadTemplate_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	badPath := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badPath, []byte(`{not valid json`), 0600); err != nil {
		t.Fatal(err)
	}

	_, err := loadTemplate("cluster", badPath)
	if err == nil {
		t.Fatal("expected error for malformed JSON template")
	}
}

func TestLoadTemplate_FlagFileMissing(t *testing.T) {
	_, err := loadTemplate("cluster", "/nonexistent/path/template.json")
	if err == nil {
		t.Fatal("expected error when -f file does not exist")
	}
}
