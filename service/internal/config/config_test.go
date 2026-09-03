package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequiredMigration(t *testing.T) {
	latest, err := latestSQLiteMigration()
	require.NoError(t, err)
	require.Equal(t, latest, RequiredMigration,
		"update RequiredMigration in config.go when adding sqlite migrations")
}

func latestSQLiteMigration() (string, error) {
	root, err := repoRoot()
	if err != nil {
		return "", err
	}
	migrationsDir := filepath.Join(root, "database", "sqlite", "migrations")
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return "", err
	}
	var latest string
	var latestID = -1
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".sql") {
			continue
		}
		id, ok := migrationNumericPrefix(name)
		if !ok || id <= latestID {
			continue
		}
		latestID = id
		latest = name
	}
	if latest == "" {
		return "", os.ErrNotExist
	}
	return latest, nil
}

func migrationNumericPrefix(name string) (int, bool) {
	dot := strings.IndexByte(name, '.')
	if dot <= 0 {
		return 0, false
	}
	id, err := strconv.Atoi(name[:dot])
	if err != nil {
		return 0, false
	}
	return id, true
}

func repoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "database", "sqlite", "migrations")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}

func TestLoadDefaults(t *testing.T) {
	dir := t.TempDir()
	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, dir, cfg.ConfigDir)
	require.Equal(t, "sqlite", cfg.DBDriver)
	require.Equal(t, filepath.Join(dir, "starapp.db"), cfg.DBPath)
	require.True(t, cfg.ShowFooter)
}

func TestLoadFromYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	err := os.WriteFile(path, []byte("http_addr: \":9090\"\nshow_footer: false\n"), 0o644)
	require.NoError(t, err)

	cfg, err := Load(dir)
	require.NoError(t, err)
	require.Equal(t, ":9090", cfg.HTTPAddr)
	require.False(t, cfg.ShowFooter)
}
