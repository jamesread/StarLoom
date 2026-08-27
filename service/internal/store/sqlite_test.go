package store

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/jamesread/starapp/service/internal/config"
)

func TestSQLiteHasMigrationWithoutMigrations(t *testing.T) {
	st, err := OpenSQLite(t.TempDir() + "/test.db")
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	ok, err := st.HasMigration(context.Background(), config.RequiredMigration)
	require.Error(t, err)
	require.False(t, ok)
}

func TestSQLiteHasMigrationWithRecord(t *testing.T) {
	st, err := OpenSQLite(t.TempDir() + "/test.db")
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	_, err = st.db.Exec(`
CREATE TABLE migrations (
  id VARCHAR(255) NOT NULL PRIMARY KEY,
  applied_at DATETIME
)`)
	require.NoError(t, err)
	_, err = st.db.Exec(
		`INSERT INTO migrations (id, applied_at) VALUES (?, datetime('now'))`,
		config.RequiredMigration,
	)
	require.NoError(t, err)

	ok, err := st.HasMigration(context.Background(), config.RequiredMigration)
	require.NoError(t, err)
	require.True(t, ok)

	latest, err := st.LatestMigration(context.Background())
	require.NoError(t, err)
	require.Equal(t, config.RequiredMigration, latest)
}

func TestSQLiteLatestMigrationEmpty(t *testing.T) {
	st, err := OpenSQLite(t.TempDir() + "/test.db")
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })

	_, err = st.db.Exec(`
CREATE TABLE migrations (
  id VARCHAR(255) NOT NULL PRIMARY KEY,
  applied_at DATETIME
)`)
	require.NoError(t, err)

	latest, err := st.LatestMigration(context.Background())
	require.NoError(t, err)
	require.Equal(t, "", latest)
}
