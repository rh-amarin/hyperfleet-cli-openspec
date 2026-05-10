package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	_ "embed"
)

//go:embed assets/cluster-template.json
var clusterTemplateDefault []byte

//go:embed assets/nodepool-template.json
var nodepoolTemplateDefault []byte

func embeddedDefault(resource string) []byte {
	if resource == "cluster" {
		return clusterTemplateDefault
	}
	return nodepoolTemplateDefault
}

// loadTemplate returns the parsed JSON template for a create command.
// If flagFile is non-empty it reads from that path and does not touch the config dir.
// Otherwise it reads <configDir>/<resource>-template.json, writing the built-in default
// if the file does not exist. created is true when the file was just written.
func loadTemplate(configDir, resource, flagFile string) (map[string]any, bool, error) {
	var raw []byte
	created := false

	if flagFile != "" {
		b, err := os.ReadFile(flagFile)
		if err != nil {
			return nil, false, fmt.Errorf("loading template: %w", err)
		}
		raw = b
	} else {
		templatePath := filepath.Join(configDir, resource+"-template.json")
		if _, err := os.Stat(templatePath); os.IsNotExist(err) {
			if err := os.MkdirAll(configDir, 0700); err != nil {
				return nil, false, fmt.Errorf("loading template: creating config dir: %w", err)
			}
			if err := os.WriteFile(templatePath, embeddedDefault(resource), 0600); err != nil {
				return nil, false, fmt.Errorf("loading template: writing default: %w", err)
			}
			created = true
			raw = embeddedDefault(resource)
		} else {
			b, err := os.ReadFile(templatePath)
			if err != nil {
				return nil, false, fmt.Errorf("loading template: %w", err)
			}
			raw = b
		}
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, false, fmt.Errorf("loading template: %w", err)
	}
	return body, created, nil
}

// ensureSpecMap returns body["spec"] as map[string]any, creating it if absent.
func ensureSpecMap(body map[string]any) map[string]any {
	if s, ok := body["spec"].(map[string]any); ok {
		return s
	}
	m := map[string]any{}
	body["spec"] = m
	return m
}
