package migrate

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindSQLiteMigrationsDir locates database/sqlite/migrations for in-process migrate.
func FindSQLiteMigrationsDir() (string, error) {
	candidates := []string{}
	if env := os.Getenv("STARAPP_MIGRATIONS_DIR"); env != "" {
		candidates = append(candidates, env)
	}
	candidates = append(candidates,
		"database/sqlite/migrations",
		filepath.Join("..", "database", "sqlite", "migrations"),
		"/var/app/database/sqlite/migrations",
	)

	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}
		info, err := os.Stat(abs)
		if err != nil || !info.IsDir() {
			continue
		}
		entries, err := os.ReadDir(abs)
		if err != nil {
			continue
		}
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
				return abs, nil
			}
		}
	}
	return "", fmt.Errorf("sqlite migrations directory not found (set STARAPP_MIGRATIONS_DIR or run from the repo root)")
}
