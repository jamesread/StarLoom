package store

import (
	"context"
	"database/sql"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"

	iamsqlite "github.com/jamesread/armature-iam/store/sqlite"
)

// OpenSQLite opens (or creates) a SQLite database at dbPath.
func OpenSQLite(dbPath string) (*SQLite, error) {
	db, err := openConfiguredSQLite(dbPath)
	if err != nil {
		return nil, err
	}
	return &SQLite{db: db, SQLite: iamsqlite.New(db)}, nil
}

func openConfiguredSQLite(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	db.SetConnMaxLifetime(-1)
	if err := configureSQLite(db); err != nil {
		_ = db.Close()
		return nil, err
	}
	return db, nil
}

func configureSQLite(db *sql.DB) error {
	ctx := context.Background()
	if _, err := db.ExecContext(ctx, `PRAGMA journal_mode=WAL`); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `PRAGMA synchronous=NORMAL`); err != nil {
		return err
	}
	return db.PingContext(ctx)
}
