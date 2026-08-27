package migrate

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/store"
)

func TestRunSQLiteUp(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	repo := filepath.Clean(filepath.Join(wd, "..", "..", ".."))
	t.Chdir(repo)

	migrationsDir, err := FindSQLiteMigrationsDir()
	require.NoError(t, err)

	dbPath := filepath.Join(t.TempDir(), "migrate-test.db")
	st, err := store.OpenSQLite(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	n, err := RunSQLiteUp(st, migrationsDir)
	require.NoError(t, err)
	require.Positive(t, n)

	require.NoError(t, AssertRequired(context.Background(), st, config.RequiredMigration, dbPath))

	n2, err := RunSQLiteUp(st, migrationsDir)
	require.NoError(t, err)
	require.Equal(t, 0, n2)
}
