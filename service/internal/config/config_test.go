package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequiredMigration(t *testing.T) {
	if RequiredMigration != "8.member-star-color.sql" {
		t.Fatalf("migration=%s", RequiredMigration)
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
