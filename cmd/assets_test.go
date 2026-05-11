package cmd

import (
	"testing"

	"gopkg.in/yaml.v3"
)

// TestConfigTemplateMatchesDefaults asserts that every key in the embedded
// config-template.yaml matches the corresponding built-in default in internal/config.
// This prevents the template and defaults from drifting apart.
func TestConfigTemplateMatchesDefaults(t *testing.T) {
	var tmpl map[string]map[string]string
	if err := yaml.Unmarshal(configTemplateYAML, &tmpl); err != nil {
		t.Fatalf("parse config-template.yaml: %v", err)
	}

	// Built-in defaults duplicated here for comparison — must stay in sync with internal/config.
	wantDefaults := map[string]map[string]string{
		"hyperfleet": {
			"api-url":     "http://localhost:8000",
			"api-version": "v1",
			"token":       "",
			"gcp-project": "hcm-hyperfleet",
		},
		"kubernetes": {
			"context":   "",
			"namespace": "",
		},
		"maestro": {
			"consumer":      "cluster1",
			"http-endpoint": "http://localhost:8100",
			"grpc-endpoint": "localhost:8090",
			"namespace":     "maestro",
		},
		"port-forward": {
			"api-port":                 "8000",
			"pg-port":                  "5432",
			"maestro-http-port":        "8100",
			"maestro-http-remote-port": "8000",
			"maestro-grpc-port":        "8090",
			"maestro-grpc-remote-port": "8090",
		},
		"database": {
			"host":     "localhost",
			"port":     "5432",
			"name":     "hyperfleet",
			"user":     "hyperfleet",
			"password": "foobar-bizz-buzz",
		},
		"rabbitmq": {
			"host":      "localhost",
			"mgmt-port": "15672",
			"user":      "guest",
			"password":  "guest",
			"vhost":     "/",
		},
		"registry": {
			"name": "",
		},
	}

	for section, wantKeys := range wantDefaults {
		tmplSection, ok := tmpl[section]
		if !ok {
			t.Errorf("template missing section %q", section)
			continue
		}
		for key, wantVal := range wantKeys {
			gotVal, exists := tmplSection[key]
			if !exists {
				t.Errorf("template missing key %s.%s", section, key)
				continue
			}
			if gotVal != wantVal {
				t.Errorf("template %s.%s = %q, want %q", section, key, gotVal, wantVal)
			}
		}
	}
}
