package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	_ "embed"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
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

func embeddedTemplateByName(name string) []byte {
	switch name {
	case "clusters.json":
		return clusterTemplateDefault
	case "nodepools.json":
		return nodepoolTemplateDefault
	default:
		return nil
	}
}

// loadTemplate returns the parsed JSON template for a create command.
// If flagFile is non-empty it reads from that path.
// Otherwise it uses the embedded default bytes directly — no disk read or write.
func loadTemplate(resource, flagFile string) (map[string]any, error) {
	var raw []byte

	if flagFile != "" {
		b, err := os.ReadFile(flagFile)
		if err != nil {
			return nil, fmt.Errorf("loading template: %w", err)
		}
		raw = b
	} else {
		raw = embeddedDefault(resource)
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("loading template: %w", err)
	}
	return body, nil
}

// loadResourceTemplate returns the parsed JSON template for a generic resource create command.
// Resolution order: --file flag > {config-dir}/templates/{templateName}.
func loadResourceTemplate(s *config.Store, templateName, flagFile string) (map[string]any, error) {
	var raw []byte

	if flagFile != "" {
		b, err := os.ReadFile(flagFile)
		if err != nil {
			return nil, fmt.Errorf("loading template: %w", err)
		}
		raw = b
	} else if templateName != "" {
		path := filepath.Join(s.ConfigDir(), "templates", templateName)
		b, err := os.ReadFile(path)
		if err != nil {
			raw = embeddedTemplateByName(templateName)
			if raw == nil {
				return nil, fmt.Errorf("loading template: %w", err)
			}
		} else {
			raw = b
		}
	} else {
		return nil, fmt.Errorf("no create template configured and no --file provided")
	}

	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return nil, fmt.Errorf("loading template: %w", err)
	}
	return body, nil
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
