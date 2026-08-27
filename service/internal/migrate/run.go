package migrate

import (
	"fmt"

	migrate "github.com/rubenv/sql-migrate"

	"github.com/jamesread/starapp/service/internal/store"
)

// RunSQLiteUp applies pending sql-migrate scripts for SQLite.
func RunSQLiteUp(st store.Store, migrationsDir string) (int, error) {
	sqlite, ok := st.(*store.SQLite)
	if !ok {
		return 0, fmt.Errorf("in-process migrations only supported for sqlite")
	}

	source := &migrate.FileMigrationSource{Dir: migrationsDir}
	ms := migrate.MigrationSet{TableName: "migrations"}
	return ms.Exec(sqlite.SQLDB(), "sqlite3", source, migrate.Up)
}
