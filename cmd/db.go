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

Subcommands: query, exec, delete.`,
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

func validDeleteTargetsList() string {
	return strings.Join(db.ValidDeleteTargetNames(), ", ")
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
	dbDeleteCmd.ValidArgs = db.ValidDeleteTargetNames()
	dbCmd.AddCommand(dbDeleteCmd)
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

var dbDeleteCmd = &cobra.Command{
	Use:   "delete [target]",
	Short: "Delete records from a table (clusters, nodepools, adapter_statuses, resources) or use --all",
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
		return db.ValidDeleteTargetNames(), cobra.ShellCompDirectiveNoFileComp
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		var target string
		if !dbDeleteAll {
			target = args[0]
			if _, ok := db.DeleteTargetTable(target); !ok {
				return fmt.Errorf("[ERROR] Unknown target '%s'. Valid targets are: %s.", target, validDeleteTargetsList())
			}
		}

		ctx := context.Background()
		_, pool, err := openDBPool(ctx)
		if err != nil {
			return err
		}
		defer pool.Close()

		targets := db.DeleteTargets
		if !dbDeleteAll {
			table, _ := db.DeleteTargetTable(target)
			targets = []db.DeleteTarget{{Name: target, Table: table}}
		}

		counts, err := fetchDeleteTargetCounts(ctx, pool, targets)
		if err != nil {
			return err
		}
		if err := printDeleteTargetCounts(cmd, counts, dbDeleteAll); err != nil {
			return err
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
			n, err := dbQuerier.Exec(ctx, pool, fmt.Sprintf("DELETE FROM %s", t.Table))
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "[ERROR] Failed to delete from %s: %v\n", t.Table, err)
				continue
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Deleted %d rows from %s\n", n, t.Table)
		}
		return nil
	},
}

type deleteTargetCount struct {
	target string
	table  string
	rows   string
}

func fetchDeleteTargetCounts(ctx context.Context, pool *pgxpool.Pool, targets []db.DeleteTarget) ([]deleteTargetCount, error) {
	out := make([]deleteTargetCount, 0, len(targets))
	for _, t := range targets {
		_, countRows, err := dbQuerier.Query(ctx, pool, fmt.Sprintf("SELECT COUNT(*) FROM %s", t.Table))
		if err != nil || len(countRows) == 0 {
			out = append(out, deleteTargetCount{target: t.Name, table: t.Table, rows: "(could not count)"})
			continue
		}
		out = append(out, deleteTargetCount{target: t.Name, table: t.Table, rows: countRows[0][0]})
	}
	return out, nil
}

func printDeleteTargetCounts(cmd *cobra.Command, counts []deleteTargetCount, all bool) error {
	if all {
		headers := []string{"TARGET", "TABLE", "ROWS"}
		rows := make([][]string, len(counts))
		for i, c := range counts {
			rows[i] = []string{c.target, c.table, c.rows}
		}
		p := output.NewPrinter("table", noColor, cmd.OutOrStdout(), cmd.ErrOrStderr())
		if err := p.PrintTable(headers, rows); err != nil {
			return err
		}
		fmt.Fprintln(cmd.OutOrStdout())
		return nil
	}
	for _, c := range counts {
		fmt.Fprintf(cmd.OutOrStdout(), "%s: %s rows\n", c.target, c.rows)
	}
	return nil
}

