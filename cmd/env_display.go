// Package cmd contains display helpers for environment profile output.
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"sort"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"
)

// secretConfigKeys are redacted when displaying configuration.
var secretConfigKeys = map[string]bool{
	"token":    true,
	"password": true,
}

// showEnvProfile prints a named environment profile with optional active state.
func showEnvProfile(cmd *cobra.Command, s *config.Store, name string) error {
	profPath := s.EnvFilePath(name)
	raw, err := os.ReadFile(profPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("[ERROR] environment '%s' not found", name)
		}
		return err
	}

	w := cmd.OutOrStdout()
	nc := noColor
	if !nc {
		if f, ok := w.(*os.File); ok && !term.IsTerminal(int(f.Fd())) {
			nc = true
		}
	}

	display, err := formatEnvFileForDisplay(raw, nc)
	if err != nil {
		return err
	}
	if _, err = fmt.Fprint(w, display); err != nil {
		return err
	}

	isActive := s.ActiveEnvironment() == name
	if isActive {
		stateDisplay, err := formatStateForDisplay(s, nc)
		if err != nil {
			return err
		}
		if stateDisplay != "" {
			fmt.Fprintln(w, output.SectionSeparator(nc))
			if _, err = fmt.Fprint(w, stateDisplay); err != nil {
				return err
			}
		}
	}

	fmt.Fprintln(w, output.SectionSeparator(nc))
	if isActive {
		fmt.Fprintf(w, "Environment file: %s [active]\n", profPath)
	} else {
		fmt.Fprintf(w, "Environment file: %s\n", profPath)
	}
	fmt.Fprintf(w, "State file:       %s\n", s.StateFilePath())
	fmt.Fprintln(w, "\nEdit these files to change configuration and runtime state.")
	return nil
}

// formatEnvFileForDisplay returns environment file YAML with secrets redacted and syntax coloring.
func formatEnvFileForDisplay(raw []byte, nc bool) (string, error) {
	redacted, err := redactEnvFileYAML(raw)
	if err != nil {
		return "", err
	}
	return output.ColorizeYAMLSections(string(redacted), nc), nil
}

// formatStateForDisplay returns a colorized state: block for non-empty runtime state, or "" if none.
func formatStateForDisplay(s *config.Store, nc bool) (string, error) {
	vals := s.NonEmptyState()
	if len(vals) == 0 {
		return "", nil
	}
	raw, err := marshalStateBlock(vals)
	if err != nil {
		return "", err
	}
	return output.ColorizeYAMLSections(string(raw), nc), nil
}

func marshalStateBlock(vals map[string]string) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	stateNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	keys := make([]string, 0, len(vals))
	for k := range vals {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		stateNode.Content = append(stateNode.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: k},
			&yaml.Node{Kind: yaml.ScalarNode, Value: vals[k]},
		)
	}

	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	root.Content = append(root.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: "state"},
		stateNode,
	)
	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// redactEnvFileYAML returns env file bytes with secret scalar values masked.
func redactEnvFileYAML(raw []byte) ([]byte, error) {
	var doc yaml.Node
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	if len(doc.Content) > 0 {
		redactSecretsYAMLNode(doc.Content[0])
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(&doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func redactSecretsYAMLNode(n *yaml.Node) {
	if n == nil {
		return
	}
	switch n.Kind {
	case yaml.MappingNode:
		for i := 0; i+1 < len(n.Content); i += 2 {
			keyNode := n.Content[i]
			valNode := n.Content[i+1]
			if keyNode.Kind == yaml.ScalarNode && secretConfigKeys[keyNode.Value] && valNode.Kind == yaml.ScalarNode {
				if valNode.Value != "" {
					valNode.Value = "<set>"
				} else {
					valNode.Value = "<not set>"
				}
			}
			redactSecretsYAMLNode(valNode)
		}
	case yaml.SequenceNode:
		for _, child := range n.Content {
			redactSecretsYAMLNode(child)
		}
	case yaml.DocumentNode:
		for _, child := range n.Content {
			redactSecretsYAMLNode(child)
		}
	}
}
