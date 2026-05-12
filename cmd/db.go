// Package cmd contains the Cobra command definitions for the hf CLI.
package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rh-amarin/hyperfleet-cli/internal/db"
	"github.com/rh-amarin/hyperfleet-cli/internal/output"
	"github.com/spf13/cobra"
)

// dbQuerier is the Querier used by db commands; replaced by tests.
var dbQuerier db.Querier = db.DefaultQuerier

// dbCmd is the top-level group for all database operations.
var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Direct database operations",
	Long: `Direct PostgreSQL database operations.

Subcommands: query, exec, delete, config.`,
}

// ---- flag vars ----

var dbQueryFile string
var dbDeleteAll bool

// ---- helpers ----

// openDBPool loads config and opens a pgxpool.Pool.
func openDBPool(ctx context.Context) (*db.Config, *pgxpool.Pool, error) {
	s, err := loadConfig()
	if err != nil {
		return nil, nil, err
	}
	cfg := db.NewFromConfig(s)
	pool, err := cfg.Pool(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("[ERROR] %v", err)
	}
	return cfg, pool, nil
}

func init() {
	rootCmd.AddCommand(dbCmd)

	// hf db query
	dbQueryCmd.Flags().StringVarP(&dbQueryFile, "file", "f", "", "read SQL from file instead of argument")
	dbCmd.AddCommand(dbQueryCmd)

	// hf db exec
	dbCmd.AddCommand(dbExecCmd)

	// hf db delete
	dbDeleteCmd.Flags().BoolVar(&dbDeleteAll, "all", false, "delete records from all tables in dependency-safe order")
	dbDeleteCmd.ValidArgs = []string{"clusters", "nodepools", "adapter_statuses"}
	dbCmd.AddCommand(dbDeleteCmd)

	// hf db config
	dbCmd.AddCommand(dbConfigCmd)
}

// ---- hf db query ----

var dbQueryCmd = &cobra.Command{
	Use:   "query [sql]",
	Short: "Execute a SQL query and display results",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var sql string
		if dbQueryFile != "" {
			b, err := os.ReadFile(dbQueryFile)
			if err != nil {
				return fmt.Errorf("[ERROR] %v", err)
			}
			sql = string(b)
		} else if len(args) == 1 {
			sql = args[0]
		} else {
			return fmt.Errorf("[ERROR] provide SQL as argument or use -f <file>")
		}

		ctx := context.Background()
		_, pool, err := openDBPool(ctx)
		if err != nil {
			return err
		}
		defer pool.Close()

		headers, rows, err := dbQuerier.Query(ctx, pool, sql)
		if err != nil {
			return fmt.Errorf("[ERROR] %v", err)
		}

		if len(rows) == 0 {
			fmt.Fprintln(cmd.OutOrStdout(), "[INFO] No rows returned.")
			return nil
		}

		p := output.NewPrinter(outputFmt, noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		if outputFmt == "table" {
			return p.PrintTable(headers, rows)
		}
		// JSON or YAML: build []map[string]string
		data := make([]map[string]string, len(rows))
		for i, row := range rows {
			m := make(map[string]string, len(headers))
			for j, h := range headers {
				m[h] = row[j]
			}
			data[i] = m
		}
		return p.Print(data)
	},
}

// ---- hf db exec ----

var dbExecCmd = &cobra.Command{
	Use:   "exec <sql>",
	Short: "Execute a DML SQL statement",
	Args:  helpOnNoArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		_, pool, err := openDBPool(ctx)
		if err != nil {
			return err
		}
		defer pool.Close()

		n, err := dbQuerier.Exec(ctx, pool, args[0])
		if err != nil {
			return fmt.Errorf("[ERROR] %v", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Rows affected: %d\n", n)
		return nil
	},
}

// ---- hf db delete ----

// deleteOrder is the dependency-safe deletion sequence for ALL.
var deleteOrder = []struct {
	target string
	table  string
}{
	{"adapter_statuses", "adapter_statuses"},
	{"nodepools", "node_pools"},
	{"clusters", "clusters"},
}

var dbDeleteCmd = &cobra.Command{
	Use:   "delete [target]",
	Short: "Delete records from a table (clusters, nodepools, adapter_statuses) or use --all",
	Args: func(cmd *cobra.Command, args []string) error {
		all, _ := cmd.Flags().GetBool("all")
		if all {
			if len(args) > 0 {
				return fmt.Errorf("cannot specify a target when --all is used")
			}
			return nil
		}
		if len(args) == 0 {
			_ = cmd.Help()
			return fmt.Errorf("")
		}
		return cobra.ExactArgs(1)(cmd, args)
	},
	ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"clusters", "nodepools", "adapter_statuses"}, cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		validTargets := map[string]bool{
			"clusters": true, "nodepools": true, "adapter_statuses": true,
		}

		var target string
		if !dbDeleteAll {
			target = args[0]
			if !validTargets[target] {
				return fmt.Errorf("[ERROR] Unknown target '%s'. Valid targets are: clusters, nodepools, adapter_statuses.", target)
			}
		}

		ctx := context.Background()
		_, pool, err := openDBPool(ctx)
		if err != nil {
			return err
		}
		defer pool.Close()

		// Build list of (target, table) pairs to process.
		var targets []struct{ target, table string }
		if dbDeleteAll {
			for _, d := range deleteOrder {
				targets = append(targets, struct{ target, table string }{d.target, d.table})
			}
		} else {
			table, _ := db.DeleteTargetTable(target)
			targets = []struct{ target, table string }{{target, table}}
		}

		// Show row counts.
		for _, t := range targets {
			_, countRows, err := dbQuerier.Query(ctx, pool, fmt.Sprintf("SELECT COUNT(*) FROM %s", t.table))
			if err != nil || len(countRows) == 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: (could not count)\n", t.target)
			} else {
				fmt.Fprintf(cmd.OutOrStdout(), "%s: %s rows\n", t.target, countRows[0][0])
			}
		}

		// Prompt for confirmation.
		fmt.Fprint(cmd.OutOrStdout(), "Type 'yes' to confirm deletion: ")
		scanner := bufio.NewScanner(cmd.InOrStdin())
		scanner.Scan()
		answer := strings.TrimSpace(scanner.Text())
		if answer != "yes" {
			fmt.Fprintln(cmd.OutOrStdout(), "Aborted")
			return nil
		}

		// Execute deletions.
		for _, t := range targets {
			n, err := dbQuerier.Exec(ctx, pool, fmt.Sprintf("DELETE FROM %s", t.table))
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "[ERROR] Failed to delete from %s: %v\n", t.table, err)
				continue
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %d rows from %s\n", n, t.table)
		}
		return nil
	},
}

// ---- hf db config ----

var dbConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Show resolved database connection parameters",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := loadConfig()
		if err != nil {
			return err
		}
		cfg := db.NewFromConfig(s)
		pw := "<not set>"
		if cfg.Password != "" {
			pw = "<set>"
		}
		fmt.Fprintf(cmd.OutOrStdout(), "host:     %s\n", cfg.Host)
		fmt.Fprintf(cmd.OutOrStdout(), "port:     %s\n", cfg.Port)
		fmt.Fprintf(cmd.OutOrStdout(), "name:     %s\n", cfg.Name)
		fmt.Fprintf(cmd.OutOrStdout(), "user:     %s\n", cfg.User)
		fmt.Fprintf(cmd.OutOrStdout(), "password: %s\n", pw)
		return nil
	},
}
