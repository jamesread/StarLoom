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

func TestSQLiteRewardMultiSpend(t *testing.T) {
	wd, err := os.Getwd()
	require.NoError(t, err)
	t.Chdir(filepath.Clean(filepath.Join(wd, "..", "..", "..")))
	migrationsDir, err := FindSQLiteMigrationsDir()
	require.NoError(t, err)

	st, err := store.OpenSQLite(filepath.Join(t.TempDir(), "reward.db"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	_, err = RunSQLiteUp(st, migrationsDir)
	require.NoError(t, err)

	ctx := context.Background()
	familyID, err := st.CreateFamily(ctx, "Household")
	require.NoError(t, err)
	id, err := st.CreateReward(ctx, familyID, "Pocket money", "", 1, false, "", true)
	require.NoError(t, err)

	reward, err := st.GetRewardByID(ctx, id)
	require.NoError(t, err)
	require.True(t, reward.MultiSpend)

	require.NoError(t, st.UpdateReward(ctx, id, "Pocket money", "", 1, true, false, "", false))
	rewards, err := st.ListRewards(ctx, familyID, true)
	require.NoError(t, err)
	require.Len(t, rewards, 1)
	require.False(t, rewards[0].MultiSpend)
}
