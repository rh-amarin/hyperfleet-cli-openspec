// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/spf13/cobra"
	"golang.org/x/term"
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

// configCmd is the top-level group for configuration management.
// With no subcommand it runs show.
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage hf configuration",
	Long: `Manage hf configuration.

Subcommands: show, get, set, env.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return configShowCmd.RunE(cmd, args)
	},
}

// configShowCmd prints the resolved configuration of the active environment.
// With an optional env-name argument it displays that environment profile instead.
var configShowCmd = &cobra.Command{
	Use:   "show [env-name]",
	Short: "Show the active configuration, or a named environment profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}

		if len(args) == 1 {
			return showEnvProfile(cmd, s, args[0])
		}

		active := s.ActiveEnvironment()
		if active == "" {
			return fmt.Errorf("[ERROR] no active environment\n  → run 'hf config env create <name>' to create one\n  → run 'hf config env activate <name>' to activate an existing one")
		}

		w := cmd.OutOrStdout()

		// Determine whether to suppress ANSI color codes.
		nc := noColor
		if !nc {
			if f, ok := w.(*os.File); ok {
				if !term.IsTerminal(int(f.Fd())) {
					nc = true
				}
			} else {
				nc = true
			}
			if os.Getenv("NO_COLOR") != "" {
				nc = true
			}
		}

		stateKeys := []string{"active-environment", "cluster-id", "cluster-name", "nodepool-id"}
		stateVals := make(map[string]string)
		for _, k := range stateKeys {
			if v := s.GetState(k); v != "" {
				stateVals[k] = v
			}
		}

		fmt.Fprintln(w, s.EnvFilePath(active))

		// Marshal config sections and state block separately.
		configSections := []string{"hyperfleet", "kubernetes", "maestro", "port-forward", "database", "rabbitmq", "registry"}
		cfgMap := make(map[string]map[string]string, len(configSections))
		for _, sec := range configSections {
			vals := resolvedSection(s, sec)
			if len(vals) > 0 {
				cfgMap[sec] = vals
			}
		}

		cfgBytes, err := marshalYAMLOrdered(cfgMap, configSections)
		if err != nil {
			return err
		}
		_, err = fmt.Fprint(w, output.ColorizeYAMLSections(string(cfgBytes), nc))
		if err != nil {
			return err
		}

		if len(stateVals) > 0 {
			fmt.Fprintln(w, output.SectionSeparator(nc))
			stateMap := map[string]map[string]string{"state": stateVals}
			stateBytes, err := marshalYAMLOrdered(stateMap, []string{"state"})
			if err != nil {
				return err
			}
			_, err = fmt.Fprint(w, output.ColorizeYAMLSections(string(stateBytes), nc))
			if err != nil {
				return err
			}
		}
		return nil
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
var configSetCmd = &cobra.Command{
	Use:   "set <section.key> <value>",
	Short: "Set a configuration value in the active environment file",
	Args:  helpOnNoArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		dotKey, value := args[0], args[1]
		idx := strings.Index(dotKey, ".")
		if idx <= 0 {
			return fmt.Errorf("[ERROR] key must be in section.key format (e.g. hyperfleet.api-url)")
		}
		section, key := dotKey[:idx], dotKey[idx+1:]
		if !validConfigSections[section] {
			return fmt.Errorf("[ERROR] unknown config section '%s'", section)
		}
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}
		return s.Set(section, key, value)
	},
}

// ---- env subcommands ----

var configEnvCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage environment profiles",
}

var configEnvListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all environment profiles",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}
		names, err := s.ListEnvironments()
		if err != nil {
			return fmt.Errorf("[ERROR] listing environments: %w", err)
		}
		if len(names) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "No environments configured. Run 'hf config env create <name>' to create one.")
			return nil
		}
		active := s.ActiveEnvironment()

		w := cmd.OutOrStdout()
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tACTIVE")
		for _, name := range names {
			activeMarker := ""
			if name == active {
				activeMarker = "✓"
			}
			fmt.Fprintf(tw, "%s\t%s\n", name, activeMarker)
		}
		return tw.Flush()
	},
}

var configEnvCreateCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new environment profile from the default template and activate it",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}

		envDir := filepath.Join(s.ConfigDir(), "environments")
		if err := os.MkdirAll(envDir, 0700); err != nil {
			return err
		}
		profPath := filepath.Join(envDir, name+".yaml")
		if _, err := os.Stat(profPath); err == nil {
			return fmt.Errorf("[ERROR] environment '%s' already exists", name)
		}

		if err := os.WriteFile(profPath, config.ConfigTemplateYAML, 0600); err != nil {
			return err
		}
		if err := s.ActivateEnvironment(name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Environment '%s' created and activated.\nEdit your configuration: %s\n", name, profPath)
		return nil
	},
}

var configEnvActivateCmd = &cobra.Command{
	Use:   "activate <name>",
	Short: "Activate an environment profile",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}
		if err := s.ActivateEnvironment(name); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Active environment set to '%s'.\n", name)
		return nil
	},
}

var configEnvDeleteCmd = &cobra.Command{
	Use:     "delete <name>",
	Aliases: []string{"rm"},
	Short:   "Delete an environment profile",
	Args:    helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}

		profPath := filepath.Join(s.ConfigDir(), "environments", name+".yaml")
		if _, err := os.Stat(profPath); os.IsNotExist(err) {
			return fmt.Errorf("[ERROR] environment '%s' not found", name)
		}
		active := s.ActiveEnvironment()
		if err := os.Remove(profPath); err != nil {
			return err
		}
		if active == name {
			_ = s.SetState("active-environment", "")
		}
		return nil
	},
}

var configEnvShowCmd = &cobra.Command{
	Use:   "show <name>",
	Short: "Show an environment profile",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}
		return showEnvProfile(cmd, s, name)
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configEnvCmd)

	configEnvCmd.AddCommand(configEnvListCmd)
	configEnvCmd.AddCommand(configEnvCreateCmd)
	configEnvCmd.AddCommand(configEnvActivateCmd)
	configEnvCmd.AddCommand(configEnvDeleteCmd)
	configEnvCmd.AddCommand(configEnvShowCmd)
}

// ---- helpers ----

// showEnvProfile prints the content of a named environment profile file.
func showEnvProfile(cmd *cobra.Command, s *config.Store, name string) error {
	profPath := filepath.Join(s.ConfigDir(), "environments", name+".yaml")
	raw, err := os.ReadFile(profPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("[ERROR] environment '%s' not found", name)
		}
		return err
	}

	w := cmd.OutOrStdout()
	if s.ActiveEnvironment() == name {
		fmt.Fprintf(w, "[active] %s\n", profPath)
	} else {
		fmt.Fprintln(w, profPath)
	}

	var prof map[string]map[string]string
	if err := yaml.Unmarshal(raw, &prof); err != nil {
		return err
	}
	redactSecrets(prof)
	b, err := yaml.Marshal(prof)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	return err
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
		return []string{"context"}
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

// marshalYAMLOrdered marshals a section map to YAML with a stable section order.
func marshalYAMLOrdered(m map[string]map[string]string, order []string) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)

	root := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	for _, sec := range order {
		vals, ok := m[sec]
		if !ok {
			continue
		}
		root.Content = append(root.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: sec})
		secNode := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		keys := make([]string, 0, len(vals))
		for k := range vals {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			secNode.Content = append(secNode.Content,
				&yaml.Node{Kind: yaml.ScalarNode, Value: k},
				&yaml.Node{Kind: yaml.ScalarNode, Value: vals[k]},
			)
		}
		root.Content = append(root.Content, secNode)
	}
	doc := &yaml.Node{Kind: yaml.DocumentNode, Content: []*yaml.Node{root}}
	if err := enc.Encode(doc); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// redactSecrets masks secret values in a profile map in-place.
func redactSecrets(prof map[string]map[string]string) {
	for _, sec := range prof {
		for k, v := range sec {
			if secretConfigKeys[k] {
				if v != "" {
					sec[k] = "<set>"
				} else {
					sec[k] = "<not set>"
				}
			}
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
