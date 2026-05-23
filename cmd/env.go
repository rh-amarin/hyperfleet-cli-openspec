// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"text/tabwriter"

	"github.com/rh-amarin/hyperfleet-cli/internal/config"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/rh-amarin/hyperfleet-cli/internal/selector"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

// envSel is the injectable PreviewSelector used by the bare hf env picker.
var envSel selector.PreviewSelector = selector.FuzzyPreviewSelector{}

const envPickerHeader = "hf env  —  select an environment to activate\ntype to filter  ·  ↑↓ navigate  ·  Enter to activate  ·  Esc to cancel"

// envCmd is the top-level env management command.
// Bare invocation launches an interactive fuzzy-picker with a YAML preview panel.
var envCmd = &cobra.Command{
	Use:   "env",
	Short: "Manage and activate environment profiles",
	Long: `Manage environment profiles (create, activate, list, delete, show).

Running 'hf env' with no subcommand launches an interactive picker:
  - Left panel: filterable list of environments (active one marked with ✓)
  - Right panel: full YAML configuration of the highlighted environment
Selecting an environment activates it and shows the full config.`,
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
			_ = cmd.Help()
			fmt.Fprintln(cmd.OutOrStdout(), "\nNo environments found. Run 'hf env create <name>' to create one.")
			return nil
		}

		active := s.ActiveEnvironment()
		items := make([]selector.Item, len(names))
		for i, name := range names {
			label := name
			if name == active {
				label = name + " ✓" // ✓
			}
			items[i] = selector.Item{Name: label}
		}

		previewFn := func(i int) string {
			raw, err := os.ReadFile(s.EnvFilePath(names[i]))
			if err != nil {
				return err.Error()
			}
			return output.ColorizeYAMLSections(string(raw), noColor)
		}

		idx, err := envSel.SelectWithPreview(items, previewFn, envPickerHeader)
		if err != nil {
			return err
		}
		if idx == -1 {
			return nil
		}

		chosen := names[idx]
		if err := s.ActivateEnvironment(chosen); err != nil {
			return fmt.Errorf("[ERROR] activating environment: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "[INFO] Activated environment '%s'.\n\n", chosen)
		return configShowCmd.RunE(cmd, nil)
	},
}

var envListCmd = &cobra.Command{
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
			fmt.Fprintln(cmd.OutOrStdout(), "No environments configured. Run 'hf env create <name>' to create one.")
			return nil
		}
		active := s.ActiveEnvironment()

		w := cmd.OutOrStdout()
		tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
		fmt.Fprintln(tw, "NAME\tACTIVE")
		for _, name := range names {
			activeMarker := ""
			if name == active {
				activeMarker = "✓" // ✓
			}
			fmt.Fprintf(tw, "%s\t%s\n", name, activeMarker)
		}
		return tw.Flush()
	},
}

var envCreateCmd = &cobra.Command{
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

var envActivateCmd = &cobra.Command{
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

var envDeleteCmd = &cobra.Command{
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

var envShowCmd = &cobra.Command{
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
	rootCmd.AddCommand(envCmd)
	envCmd.AddCommand(envListCmd, envCreateCmd, envActivateCmd, envDeleteCmd, envShowCmd)
}

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
