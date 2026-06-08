// Package db provides a pgxpool wrapper for direct PostgreSQL access.
package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rh-amarin/hyperfleet-cli/internal/config"
)

// Config holds database connection parameters.
type Config struct {
	Host     string
	Port     string
	Name     string
	User     string
	Password string
}

// NewFromConfig reads database.* keys from the config store.
func NewFromConfig(s *config.Store) *Config {
	return &Config{
		Host:     s.Get("database", "host"),
		Port:     s.Get("database", "port"),
		Name:     s.Get("database", "name"),
		User:     s.Get("database", "user"),
		Password: s.Get("database", "password"),
	}
}

// DSN returns the PostgreSQL connection string.
func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		c.User, c.Password, c.Host, c.Port, c.Name)
}

// Pool opens a pgxpool connection pool using the config's DSN.
func (c *Config) Pool(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, c.DSN())
	if err != nil {
		return nil, fmt.Errorf("open pool: %w", err)
	}
	return pool, nil
}

// Querier is the interface satisfied by defaultQuerier; used for testing.
type Querier interface {
	Query(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) ([]string, [][]string, error)
	Exec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (int64, error)
}

// DefaultQuerier is the production implementation of Querier.
var DefaultQuerier Querier = defaultQuerier{}

type defaultQuerier struct{}

func (defaultQuerier) Query(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) ([]string, [][]string, error) {
	return Query(ctx, pool, sql, args...)
}

func (defaultQuerier) Exec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (int64, error) {
	return Exec(ctx, pool, sql, args...)
}

// Query executes a SELECT-style SQL statement and returns column headers and row data.
// NULL values are rendered as "NULL"; fields over 80 chars are truncated with "…".
func Query(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) ([]string, [][]string, error) {
	pgRows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, nil, err
	}
	defer pgRows.Close()

	descs := pgRows.FieldDescriptions()
	headers := make([]string, len(descs))
	for i, d := range descs {
		headers[i] = string(d.Name)
	}

	var rows [][]string
	for pgRows.Next() {
		vals, err := pgRows.Values()
		if err != nil {
			return nil, nil, err
		}
		row := make([]string, len(vals))
		for i, v := range vals {
			row[i] = cellString(v)
		}
		rows = append(rows, row)
	}
	if err := pgRows.Err(); err != nil {
		return nil, nil, err
	}
	return headers, rows, nil
}

// Exec executes a DML SQL statement and returns the number of rows affected.
func Exec(ctx context.Context, pool *pgxpool.Pool, sql string, args ...any) (int64, error) {
	tag, err := pool.Exec(ctx, sql, args...)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// cellString converts a query value to its string representation.
func cellString(v any) string {
	var s string
	switch val := v.(type) {
	case nil:
		s = "NULL"
	case []byte:
		s = string(val)
	case string:
		s = val
	default:
		s = fmt.Sprintf("%v", val)
	}
	if len([]rune(s)) > 80 {
		s = string([]rune(s)[:79]) + "…"
	}
	return s
}

// DeleteTarget maps a CLI target name to a PostgreSQL table.
type DeleteTarget struct {
	Name  string
	Table string
}

// DeleteTargets is the dependency-safe deletion order for hf db delete --all
// (dependents before owners).
var DeleteTargets = []DeleteTarget{
	{Name: "adapter_statuses", Table: "adapter_statuses"},
	{Name: "nodepools", Table: "node_pools"},
	{Name: "resources", Table: "resources"},
	{Name: "clusters", Table: "clusters"},
}

// DeleteTargetTable resolves a user-facing target to a table name.
func DeleteTargetTable(target string) (string, bool) {
	for _, d := range DeleteTargets {
		if strings.EqualFold(target, d.Name) {
			return d.Table, true
		}
	}
	return "", false
}

// ValidDeleteTargetNames returns CLI target names for completions and errors.
func ValidDeleteTargetNames() []string {
	names := make([]string, len(DeleteTargets))
	for i, d := range DeleteTargets {
		names[i] = d.Name
	}
	return names
}
