package migrate

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindSQLiteMigrationsDirFromRepo(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	// tests run from service/internal/migrate — repo root is ../../../
	repo := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	t.Chdir(repo)

	dir, err := FindSQLiteMigrationsDir()
	require.NoError(t, err)
	require.FileExists(t, filepath.Join(dir, "0.base.sql"))
	require.FileExists(t, filepath.Join(dir, "2.webhook-targets-events.sql"))
}
