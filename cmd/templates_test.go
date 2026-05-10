package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTemplate_AutoCreatesDefault(t *testing.T) {
	dir := t.TempDir()

	body, created, err := loadTemplate(dir, "cluster", "")
	if err != nil {
		t.Fatalf("loadTemplate: %v", err)
	}
	if !created {
		t.Error("expected created=true when template file did not exist")
	}
	if body["kind"] != "Cluster" {
		t.Errorf("expected kind=Cluster, got %v", body["kind"])
	}
	if body["name"] != "my-cluster" {
		t.Errorf("expected name=my-cluster, got %v", body["name"])
	}

	// File should now exist on disk.
	if _, err := os.Stat(filepath.Join(dir, "cluster-template.json")); os.IsNotExist(err) {
		t.Error("expected cluster-template.json to be written to config dir")
	}
}

func TestLoadTemplate_ExistingFile(t *testing.T) {
	dir := t.TempDir()
	custom := `{"kind":"Cluster","name":"custom-cluster","labels":{},"spec":{"region":"eu-west-1","version":"5.0.0","counter":"1"}}`
	if err := os.WriteFile(filepath.Join(dir, "cluster-template.json"), []byte(custom), 0600); err != nil {
		t.Fatal(err)
	}

	body, created, err := loadTemplate(dir, "cluster", "")
	if err != nil {
		t.Fatalf("loadTemplate: %v", err)
	}
	if created {
		t.Error("expected created=false when template file already exists")
	}
	if body["name"] != "custom-cluster" {
		t.Errorf("expected name=custom-cluster, got %v", body["name"])
	}
	spec, _ := body["spec"].(map[string]any)
	if spec["region"] != "eu-west-1" {
		t.Errorf("expected region=eu-west-1, got %v", spec["region"])
	}
}

func TestLoadTemplate_FlagFileOverride(t *testing.T) {
	dir := t.TempDir()
	// Write a different template as the config-dir default.
	if err := os.WriteFile(filepath.Join(dir, "cluster-template.json"),
		[]byte(`{"kind":"Cluster","name":"config-dir-cluster","labels":{},"spec":{}}`), 0600); err != nil {
		t.Fatal(err)
	}

	// Write the override file.
	overridePath := filepath.Join(dir, "override.json")
	if err := os.WriteFile(overridePath,
		[]byte(`{"kind":"Cluster","name":"override-cluster","labels":{},"spec":{}}`), 0600); err != nil {
		t.Fatal(err)
	}

	body, created, err := loadTemplate(dir, "cluster", overridePath)
	if err != nil {
		t.Fatalf("loadTemplate with -f: %v", err)
	}
	if created {
		t.Error("expected created=false when -f flag is used")
	}
	if body["name"] != "override-cluster" {
		t.Errorf("expected override-cluster, got %v", body["name"])
	}
}

func TestLoadTemplate_MalformedJSON(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "cluster-template.json"),
		[]byte(`{not valid json`), 0600); err != nil {
		t.Fatal(err)
	}

	_, _, err := loadTemplate(dir, "cluster", "")
	if err == nil {
		t.Fatal("expected error for malformed JSON template")
	}
}

func TestLoadTemplate_FlagFileMissing(t *testing.T) {
	dir := t.TempDir()
	_, _, err := loadTemplate(dir, "cluster", "/nonexistent/path/template.json")
	if err == nil {
		t.Fatal("expected error when -f file does not exist")
	}
}

func TestLoadTemplate_NodepoolDefault(t *testing.T) {
	dir := t.TempDir()

	body, created, err := loadTemplate(dir, "nodepool", "")
	if err != nil {
		t.Fatalf("loadTemplate nodepool: %v", err)
	}
	if !created {
		t.Error("expected created=true for missing nodepool template")
	}
	if body["kind"] != "NodePool" {
		t.Errorf("expected kind=NodePool, got %v", body["kind"])
	}
	if body["name"] != "my-nodepool" {
		t.Errorf("expected name=my-nodepool, got %v", body["name"])
	}

	if _, err := os.Stat(filepath.Join(dir, "nodepool-template.json")); os.IsNotExist(err) {
		t.Error("expected nodepool-template.json to be written to config dir")
	}
}
