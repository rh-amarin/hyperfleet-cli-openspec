// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/selector"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// validConfigSections is the canonical set of sections accepted by `hf config set`.
var validConfigSections = map[string]bool{
	"hyperfleet":   true,
	"kubernetes":   true,
	"maestro":      true,
	"port-forward": true,
	"database":     true,
	"rabbitmq":     true,
	"registry":     true,
}

// secretConfigKeys are redacted when displaying configuration.
var secretConfigKeys = map[string]bool{
	"token":    true,
	"password": true,
}

// configSetSel is the selector used by hf config set interactive mode; swapped in tests.
var configSetSel selector.PreviewSelector = selector.FuzzyPreviewSelector{}

const configSetHeader = "hf config set  —  select a key to edit\ntype to filter  ·  ↑↓ navigate  ·  Enter to set value  ·  Esc to cancel"

// configCmd is the top-level group for configuration management.
// With no subcommand it runs show.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage hf configuration",
	Long: `Manage hf configuration.

Subcommands: show, get, set, env.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		_ = cmd.Help()
		fmt.Fprintln(cmd.OutOrStdout())
		return configShowCmd.RunE(cmd, args)
	},
}

// configShowCmd prints the active environment file, or a named profile when given.
var configShowCmd = &cobra.Command{
	Use:   "show [env-name]",
	Short: "Show the environment config file and active state (colorized YAML)",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}

		envName := s.ActiveEnvironment()
		if len(args) == 1 {
			envName = args[0]
		}
		if envName == "" {
			return fmt.Errorf("[ERROR] no active environment\n  → run 'hf config env create <name>' to create one\n  → run 'hf config env activate <name>' to activate an existing one")
		}
		return showEnvProfile(cmd, s, envName)
	},
}

// configGetCmd prints a single config or state value.
// Key format: "section.key" for config values (e.g. hyperfleet.api-url),
// or a plain key for state values (e.g. cluster-id).
var configGetCmd = &cobra.Command{
	Use:   "get <key>",
	Short: "Get a configuration or state value (use section.key for config, plain key for state)",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		key := args[0]
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}
		if idx := strings.Index(key, "."); idx > 0 {
			section, field := key[:idx], key[idx+1:]
			val := s.Get(section, field)
			if val == "" {
				return fmt.Errorf("[ERROR] Config key '%s' not found", key)
			}
			fmt.Fprintln(cmd.OutOrStdout(), val)
			return nil
		}
		val := s.GetState(key)
		if val == "" {
			return fmt.Errorf("[ERROR] State key '%s' not found", key)
		}
		fmt.Fprintln(cmd.OutOrStdout(), val)
		return nil
	},
}

// configSetCmd writes a config value to the active environment file.
// Key format: "section.key" (e.g. hyperfleet.api-url).
// With no arguments, launches an interactive fuzzy-finder to pick the parameter.
var configSetCmd = &cobra.Command{
	Use:   "set [section.key value]",
	Short: "Set a configuration value (interactive when called with no arguments)",
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 || len(args) == 2 {
			return nil
		}
		_ = cmd.Help()
		return fmt.Errorf("")
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}

		if len(args) == 0 {
			return configSetInteractive(cmd, s)
		}

		dotKey, value := args[0], args[1]
		idx := strings.Index(dotKey, ".")
		if idx <= 0 {
			return fmt.Errorf("[ERROR] key must be in section.key format (e.g. hyperfleet.api-url)")
		}
		section, key := dotKey[:idx], dotKey[idx+1:]
		if !validConfigSections[section] {
			return fmt.Errorf("[ERROR] unknown config section '%s'", section)
		}
		if err := s.Set(section, key, value); err != nil {
			return err
		}
		return configShowCmd.RunE(cmd, nil)
	},
}

// configSetInteractive runs the fuzzy-find → prompt → set → show flow for hf config set (no args).
func configSetInteractive(cmd *cobra.Command, s *config.Store) error {
	configSections := []string{"hyperfleet", "kubernetes", "maestro", "port-forward", "database", "rabbitmq", "registry"}
	type param struct{ dotKey, value string }
	var params []param
	for _, sec := range configSections {
		for _, k := range knownKeysForSection(sec) {
			v := s.Get(sec, k)
			if secretConfigKeys[k] {
				if v != "" {
					v = "<set>"
				} else {
					v = "<not set>"
				}
			}
			params = append(params, param{sec + "." + k, v})
		}
	}

	items := make([]selector.Item, len(params))
	for i, p := range params {
		items[i] = selector.Item{Name: p.dotKey, ID: p.value}
	}

	preview := renderConfigPreview(s)
	previewFn := func(_ int) string { return preview }
	idx, err := configSetSel.SelectWithPreview(items, previewFn, configSetHeader)
	if err != nil {
		return err
	}
	if idx < 0 {
		return nil
	}

	dotKey := params[idx].dotKey
	fmt.Fprintf(cmd.OutOrStdout(), "Value for %s: ", dotKey)
	scanner := bufio.NewScanner(cmd.InOrStdin())
	scanner.Scan()
	value := scanner.Text()

	sep := strings.Index(dotKey, ".")
	section, key := dotKey[:sep], dotKey[sep+1:]
	if err := s.Set(section, key, value); err != nil {
		return err
	}
	return configShowCmd.RunE(cmd, nil)
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
}

// ---- helpers ----

// renderConfigPreview renders the active environment file as colorized YAML for picker previews.
func renderConfigPreview(s *config.Store) string {
	active := s.ActiveEnvironment()
	if active == "" {
		return ""
	}
	raw, err := os.ReadFile(s.EnvFilePath(active))
	if err != nil {
		return err.Error()
	}
	display, err := formatEnvFileForDisplay(raw, noColor)
	if err != nil {
		return fmt.Sprintf("[ERROR] %v", err)
	}
	return display
}

// resolvedSection returns all key/value pairs for a section with secrets masked.
func resolvedSection(s *config.Store, section string) map[string]string {
	keys := knownKeysForSection(section)
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		v := s.Get(section, k)
		if secretConfigKeys[k] {
			if v != "" {
				out[k] = "<set>"
			} else {
				out[k] = "<not set>"
			}
		} else {
			out[k] = v
		}
	}
	return out
}

// knownKeysForSection returns the canonical keys for each config section.
func knownKeysForSection(section string) []string {
	switch section {
	case "hyperfleet":
		return []string{"api-url", "api-version", "token", "gcp-project", "namespace"}
	case "kubernetes":
		return []string{"context", "kubeconfig"}
	case "maestro":
		return []string{"consumer", "http-endpoint", "grpc-endpoint", "namespace"}
	case "port-forward":
		return []string{"api-port", "pg-port", "maestro-http-port", "maestro-http-remote-port", "maestro-grpc-port", "maestro-grpc-remote-port"}
	case "database":
		return []string{"host", "port", "name", "user", "password"}
	case "rabbitmq":
		return []string{"host", "mgmt-port", "user", "password", "vhost"}
	case "registry":
		return []string{"name"}
	}
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

// writeTable writes a tabwriter-aligned table to w.
func writeTable(w io.Writer, headers []string, rows [][]string) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, strings.Join(headers, "\t"))
	for _, row := range rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}
