package store

import (
	"context"
	"database/sql"
	"fmt"

	iamsqlite "github.com/jamesread/armature-iam/store/sqlite"
)

// SQLite persists StarApp data in SQLite.
type SQLite struct {
	*iamsqlite.SQLite
	db *sql.DB
}

// SQLDB exposes the underlying connection for migrations.
func (s *SQLite) SQLDB() *sql.DB {
	return s.db
}

func (s *SQLite) Close() error {
	return s.db.Close()
}

func (s *SQLite) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *SQLite) HasMigration(ctx context.Context, id string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM migrations WHERE id = ?`, id).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("query migrations: %w", err)
	}
	return n > 0, nil
}

func (s *SQLite) LatestMigration(ctx context.Context) (string, error) {
	var id sql.NullString
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM migrations ORDER BY applied_at DESC, id DESC LIMIT 1`).Scan(&id)
	if err == sql.ErrNoRows {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("latest migration: %w", err)
	}
	if !id.Valid {
		return "", nil
	}
	return id.String, nil
}
