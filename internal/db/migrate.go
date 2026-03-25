package db

import (
	"context"
	"embed"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	if ctx == nil {
		ctx = context.Background()
	}
	entries, err := migrationFiles.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("read migrations: %w", err)
	}
	filenames := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		filenames = append(filenames, entry.Name())
	}
	sort.Strings(filenames)
	for _, name := range filenames {
		body, err := migrationFiles.ReadFile("migrations/" + name)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		for _, stmt := range splitSQLStatements(string(body)) {
			if _, err := pool.Exec(ctx, stmt); err != nil {
				return fmt.Errorf("apply migration %s: %w", name, err)
			}
		}
	}
	return nil
}

func splitSQLStatements(sql string) []string {
	parts := strings.Split(sql, ";")
	stmts := make([]string, 0, len(parts))
	for _, part := range parts {
		stmt := strings.TrimSpace(part)
		if stmt == "" {
			continue
		}
		stmts = append(stmts, stmt)
	}
	return stmts
}
