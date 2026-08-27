package migrate

import (
	"context"
	"fmt"

	"github.com/jamesread/starapp/service/internal/store"
)

// AssertRequired checks that the store has the required sql-migrate id applied.
func AssertRequired(ctx context.Context, st store.Store, required, dbPath string) error {
	ok, err := st.HasMigration(ctx, required)
	if err != nil {
		return fmt.Errorf("database %q: migrations table could not be queried after migrate: %w", dbPath, err)
	}
	if ok {
		return nil
	}
	latest, _ := st.LatestMigration(ctx)
	if latest == "" {
		latest = "null"
	}
	return fmt.Errorf("database %q: requires migration %s but schema is at %s", dbPath, required, latest)
}
