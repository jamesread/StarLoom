package migrate

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jamesread/starapp/service/internal/config"
	"github.com/jamesread/starapp/service/internal/store"
)

func TestAssertRequiredMissing(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	st, err := store.OpenSQLite(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	err = AssertRequired(context.Background(), st, config.RequiredMigration, dbPath)
	require.Error(t, err)
	require.Contains(t, err.Error(), dbPath)
	require.Contains(t, err.Error(), "migrations table could not be queried")
}

func TestAssertRequiredPresent(t *testing.T) {
	st := store.OpenMemory()
	t.Cleanup(func() { _ = st.Close() })

	err := AssertRequired(context.Background(), st, config.RequiredMigration, ":memory:")
	require.NoError(t, err)
}
