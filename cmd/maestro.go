// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/rh-amarin/hyperfleet-cli/internal/maestro"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/spf13/cobra"
)

// maestroCmd is the top-level group for Maestro API operations.
var maestroCmd = &cobra.Command{
	Use:   "maestro",
	Short: "Interact with the Maestro API",
	Long: `Interact with the Maestro API.

Subcommands: list, get, delete, bundles, consumers.`,
}

// newMaestroClient builds a Maestro client from the active config store.
func newMaestroClient() (*maestro.Client, error) {
	s, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return maestro.NewFromConfig(s), nil
}

// ---- maestro list ----

var maestroListCmd = &cobra.Command{
	Use:   "list",
	Short: "List Maestro resource-bundles for the configured consumer",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		consumer := s.Get("maestro", "consumer")

		c := maestro.NewFromConfig(s)
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		var list *maestro.ResourceBundleList
		if consumer != "" {
			list, err = c.ListResourceBundlesByConsumer(context.Background(), consumer)
		} else {
			list, err = c.ListResourceBundles(context.Background())
		}
		if err != nil {
			return err
		}
		return p.Print(list)
	},
}

// ---- maestro bundles ----

var maestroBundlesCmd = &cobra.Command{
	Use:   "bundles",
	Short: "List all Maestro resource-bundles (no consumer filter)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newMaestroClient()
		if err != nil {
			return err
		}
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		list, err := c.ListResourceBundles(context.Background())
		if err != nil {
			return err
		}
		return p.Print(list)
	},
}

// ---- maestro consumers ----

var maestroConsumersCmd = &cobra.Command{
	Use:   "consumers",
	Short: "List all Maestro consumers",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newMaestroClient()
		if err != nil {
			return err
		}
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		list, err := c.ListConsumers(context.Background())
		if err != nil {
			return err
		}
		return p.Print(list)
	},
}

// ---- maestro get ----

var maestroGetCmd = &cobra.Command{
	Use:   "get [name]",
	Short: "Get a Maestro resource-bundle by name",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newMaestroClient()
		if err != nil {
			return err
		}
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		if name == "" {
			// Interactive selection.
			list, err := c.ListResourceBundles(context.Background())
			if err != nil {
				return err
			}
			if len(list.Items) == 0 {
				p.Warn("No resource bundles available")
				return nil
			}
			name, err = selectResourceBundle(cmd, list.Items)
			if err != nil {
				return err
			}
		}

		rb, err := c.GetResourceBundle(context.Background(), name)
		if err != nil {
			return err
		}
		if rb == nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] No resource bundle found matching '%s'\n", name)
			return nil
		}
		return p.Print(rb)
	},
}

// ---- maestro delete ----

var maestroDeleteCmd = &cobra.Command{
	Use:   "delete [name]",
	Short: "Delete a Maestro resource-bundle by name",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newMaestroClient()
		if err != nil {
			return err
		}
		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())

		name := ""
		if len(args) > 0 {
			name = args[0]
		}

		list, listErr := c.ListResourceBundles(context.Background())
		if listErr != nil {
			return listErr
		}

		if name == "" {
			if len(list.Items) == 0 {
				p.Warn("No resource bundles available")
				return nil
			}
			name, err = selectResourceBundle(cmd, list.Items)
			if err != nil {
				return err
			}
		}

		// Resolve name to ID.
		var target *maestro.ResourceBundle
		for i := range list.Items {
			if list.Items[i].Name == name {
				target = &list.Items[i]
				break
			}
		}
		if target == nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "[WARN] No resource bundle found matching '%s'\n", name)
			return nil
		}

		if err := c.DeleteResourceBundle(context.Background(), target.ID); err != nil {
			return err
		}
		p.Info(fmt.Sprintf("Deleted resource bundle '%s' (%s)", name, target.ID))
		return nil
	},
}

// selectResourceBundle prints a numbered menu and reads the user's choice from stdin.
func selectResourceBundle(cmd *cobra.Command, items []maestro.ResourceBundle) (string, error) {
	fmt.Fprintln(cmd.OutOrStdout(), "Available resource bundles:")
	for i, rb := range items {
		fmt.Fprintf(cmd.OutOrStdout(), "  %d) %s (%s)\n", i+1, rb.Name, rb.ID)
	}
	fmt.Fprint(cmd.OutOrStdout(), "Select [1-"+strconv.Itoa(len(items))+"]: ")

	scanner := bufio.NewScanner(cmd.InOrStdin())
	if !scanner.Scan() {
		return "", fmt.Errorf("no input provided")
	}
	line := strings.TrimSpace(scanner.Text())
	idx, err := strconv.Atoi(line)
	if err != nil || idx < 1 || idx > len(items) {
		return "", fmt.Errorf("invalid selection: %q", line)
	}
	return items[idx-1].Name, nil
}

func init() {
	rootCmd.AddCommand(maestroCmd)

	maestroCmd.AddCommand(maestroListCmd)
	maestroCmd.AddCommand(maestroBundlesCmd)
	maestroCmd.AddCommand(maestroConsumersCmd)
	maestroCmd.AddCommand(maestroGetCmd)
	maestroCmd.AddCommand(maestroDeleteCmd)
}
