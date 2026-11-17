package migrate

import (
	"fmt"
	"os"
	"path/filepath"
)

// FindSQLiteMigrationsDir locates database/sqlite/migrations for in-process migrate.
func FindSQLiteMigrationsDir() (string, error) {
	for _, candidate := range migrationDirCandidates() {
		if abs, ok := migrationsDirWithSQL(candidate); ok {
			return abs, nil
		}
	}
	return "", fmt.Errorf("sqlite migrations directory not found (set STARAPP_MIGRATIONS_DIR or run from the repo root)")
}

func migrationDirCandidates() []string {
	candidates := []string{}
	if env := os.Getenv("STARAPP_MIGRATIONS_DIR"); env != "" {
		candidates = append(candidates, env)
	}
	return append(candidates,
		"database/sqlite/migrations",
		filepath.Join("..", "database", "sqlite", "migrations"),
		"/var/app/database/sqlite/migrations",
	)
}

func migrationsDirWithSQL(candidate string) (string, bool) {
	abs, err := filepath.Abs(candidate)
	if err != nil {
		return "", false
	}
	info, err := os.Stat(abs)
	if err != nil || !info.IsDir() {
		return "", false
	}
	if !dirContainsSQL(abs) {
		return "", false
	}
	return abs, true
}

func dirContainsSQL(abs string) bool {
	entries, err := os.ReadDir(abs)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if !e.IsDir() && filepath.Ext(e.Name()) == ".sql" {
			return true
		}
	}
	return false
}
