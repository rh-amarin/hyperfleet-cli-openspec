// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// validConfigSections is the canonical set of sections accepted by `hf config set/get`.
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
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage hf configuration",
	Long: `Manage hf configuration.

Subcommands: show, get, set, env, doctor.`,
}

// configShowCmd prints the resolved configuration as YAML.
// With an optional env-name argument it displays that environment profile instead.
var configShowCmd = &cobra.Command{
	Use:   "show [env-name]",
	Short: "Show the resolved configuration, or a named environment profile",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}

		// If a specific environment name is provided, delegate to env-show logic.
		if len(args) == 1 {
			name := args[0]
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

		w := cmd.OutOrStdout()

		stateKeys := []string{"active-environment", "cluster-id", "cluster-name", "nodepool-id"}
		stateVals := make(map[string]string)
		for _, k := range stateKeys {
			if v := s.GetState(k); v != "" {
				stateVals[k] = v
			}
		}

		allSections := []string{"state", "hyperfleet", "kubernetes", "maestro", "port-forward", "database", "rabbitmq", "registry"}
		out := make(map[string]map[string]string, len(allSections))
		if len(stateVals) > 0 {
			out["state"] = stateVals
		}
		for _, sec := range []string{"hyperfleet", "kubernetes", "maestro", "port-forward", "database", "rabbitmq", "registry"} {
			vals := resolvedSection(s, sec)
			if len(vals) > 0 {
				out[sec] = vals
			}
		}

		b, err := marshalYAMLOrdered(out, allSections)
		if err != nil {
			return err
		}
		_, err = w.Write(b)
		return err
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

// configSetCmd writes a config value to config.yaml.
var configSetCmd = &cobra.Command{
	Use:   "set <section> <key> <value>",
	Short: "Set a configuration value",
	Args:  helpOnNoArgs(3),
	RunE: func(cmd *cobra.Command, args []string) error {
		section, key, value := args[0], args[1], args[2]
		if !validConfigSections[section] {
			return fmt.Errorf("[ERROR] Unknown config section '%s'", section)
		}
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}
		return s.Set(section, key, value)
	},
}

// configDoctorCmd checks API connectivity with a 5-second timeout.
var configDoctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check connectivity to the HyperFleet API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}
		baseURL := s.Get("hyperfleet", "api-url")
		if baseURL == "" {
			baseURL = "http://localhost:8000"
		}

		client := &http.Client{Timeout: 5 * time.Second}
		target := strings.TrimRight(baseURL, "/") + "/healthz"
		resp, err := client.Get(target)
		if err != nil {
			fmt.Fprintf(cmd.OutOrStdout(), "[ERROR] Cannot reach API server at %s: %v\n", baseURL, err)
			return fmt.Errorf("[ERROR] Cannot reach API server at %s: %w", baseURL, err)
		}
		resp.Body.Close()
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			fmt.Fprintf(cmd.OutOrStdout(), "[OK] API server reachable at %s\n", baseURL)
			return nil
		}
		msg := fmt.Sprintf("[ERROR] Cannot reach API server at %s: HTTP %d", baseURL, resp.StatusCode)
		fmt.Fprintln(cmd.OutOrStdout(), msg)
		return fmt.Errorf("%s", msg)
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
		active := s.ActiveEnvironment()

		w := cmd.OutOrStdout()
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tAPI URL\tACTIVE")
		for _, name := range names {
			apiURLVal := envProfileAPIURL(s, name)
			activeMarker := ""
			if name == active {
				activeMarker = "✓"
			}
			fmt.Fprintf(tw, "%s\t%s\t%s\n", name, apiURLVal, activeMarker)
		}
		return tw.Flush()
	},
}

// Flags for env create — declared as package-level vars so tests can reset them.
var (
	envCreateAPIURL    string
	envCreateAPIToken  string
	envCreateClusterID string
	envCreateNPID      string
)

var configEnvCreateCmd = &cobra.Command{
	Use:     "create <name>",
	Aliases: []string{"new"},
	Short:   "Create a new environment profile",
	Args:    helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		s := config.NewFromEnv()
		if err := s.Load(); err != nil {
			return fmt.Errorf("[ERROR] loading config: %w", err)
		}

		profPath := filepath.Join(s.ConfigDir(), "environments", name+".yaml")
		if _, err := os.Stat(profPath); err == nil {
			return fmt.Errorf("[ERROR] Environment '%s' already exists", name)
		}

		prof := map[string]map[string]string{}
		if envCreateAPIURL != "" {
			configSetNested(prof, "hyperfleet", "api-url", envCreateAPIURL)
		}
		if envCreateAPIToken != "" {
			configSetNested(prof, "hyperfleet", "token", envCreateAPIToken)
		}
		if envCreateClusterID != "" {
			configSetNested(prof, "state", "cluster-id", envCreateClusterID)
		}
		if envCreateNPID != "" {
			configSetNested(prof, "state", "nodepool-id", envCreateNPID)
		}

		if err := os.MkdirAll(filepath.Dir(profPath), 0700); err != nil {
			return err
		}
		b, err := yaml.Marshal(prof)
		if err != nil {
			return err
		}
		if err := os.WriteFile(profPath, b, 0600); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Environment '%s' created. Run 'hf config env activate %s' to use it.\n", name, name)
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
			return fmt.Errorf("[ERROR] Environment '%s' not found", name)
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
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	configCmd.AddCommand(configShowCmd)
	configCmd.AddCommand(configGetCmd)
	configCmd.AddCommand(configSetCmd)
	configCmd.AddCommand(configDoctorCmd)
	configCmd.AddCommand(configEnvCmd)

	configEnvCmd.AddCommand(configEnvListCmd)
	configEnvCmd.AddCommand(configEnvCreateCmd)
	configEnvCmd.AddCommand(configEnvActivateCmd)
	configEnvCmd.AddCommand(configEnvDeleteCmd)
	configEnvCmd.AddCommand(configEnvShowCmd)

	configEnvCreateCmd.Flags().StringVar(&envCreateAPIURL, "api-url", "", "API server URL for this environment")
	configEnvCreateCmd.Flags().StringVar(&envCreateAPIToken, "api-token", "", "API bearer token for this environment")
	configEnvCreateCmd.Flags().StringVar(&envCreateClusterID, "cluster-id", "", "default cluster ID for this environment")
	configEnvCreateCmd.Flags().StringVar(&envCreateNPID, "nodepool-id", "", "default nodepool ID for this environment")
}

// ---- helpers ----

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
		return []string{"api-url", "api-version", "token", "gcp-project"}
	case "kubernetes":
		return []string{"context", "namespace"}
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

// envProfileAPIURL reads the api-url from a named environment profile file.
func envProfileAPIURL(s *config.Store, name string) string {
	profPath := filepath.Join(s.ConfigDir(), "environments", name+".yaml")
	raw, err := os.ReadFile(profPath)
	if err != nil {
		return ""
	}
	var prof map[string]map[string]string
	if err := yaml.Unmarshal(raw, &prof); err != nil {
		return ""
	}
	if hf, ok := prof["hyperfleet"]; ok {
		return hf["api-url"]
	}
	return ""
}

// configSetNested initialises inner maps as needed and sets a value.
func configSetNested(m map[string]map[string]string, section, key, value string) {
	if _, ok := m[section]; !ok {
		m[section] = map[string]string{}
	}
	m[section][key] = value
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
