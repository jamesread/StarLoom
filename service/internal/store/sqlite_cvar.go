package store

import (
	"context"
	"database/sql"
	"errors"
)

func cvarSelectSQL() string {
	return `SELECT cvar_key, cvar_main_type, cvar_value_int, cvar_value_string,
		cvar_title, cvar_description, cvar_category, cvar_ordinal FROM cvars`
}

func scanCvar(row interface {
	Scan(dest ...any) error
}) (*CvarRow, error) {
	var c CvarRow
	var valueInt sql.NullInt64
	var valueString sql.NullString
	if err := row.Scan(
		&c.Key, &c.MainType, &valueInt, &valueString,
		&c.Title, &c.Description, &c.Category, &c.Ordinal,
	); err != nil {
		return nil, err
	}
	if valueInt.Valid {
		c.ValueInt = int(valueInt.Int64)
	}
	if valueString.Valid {
		c.ValueString = valueString.String
	}
	return &c, nil
}

func (s *SQLite) ListCvars(ctx context.Context) ([]CvarRow, error) {
	rows, err := s.db.QueryContext(ctx, cvarSelectSQL()+` ORDER BY cvar_ordinal, cvar_key`)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	out := make([]CvarRow, 0)
	for rows.Next() {
		row, err := scanCvar(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *row)
	}
	return out, rows.Err()
}

func (s *SQLite) FindCvar(ctx context.Context, key string) (*CvarRow, error) {
	row := s.db.QueryRowContext(ctx, cvarSelectSQL()+` WHERE cvar_key = ?`, key)
	c, err := scanCvar(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (s *SQLite) InsertCvarIfMissing(ctx context.Context, row CvarRow) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO cvars (
			cvar_key, cvar_value_int, cvar_value_string, cvar_main_type,
			cvar_title, cvar_description, cvar_category, cvar_ordinal
		) VALUES (?, ?, NULLIF(?, ''), ?, ?, ?, ?, ?)
		ON CONFLICT(cvar_key) DO UPDATE SET
			cvar_title = excluded.cvar_title,
			cvar_description = excluded.cvar_description,
			cvar_category = excluded.cvar_category,
			cvar_ordinal = excluded.cvar_ordinal`,
		row.Key, row.ValueInt, row.ValueString, row.MainType,
		row.Title, row.Description, row.Category, row.Ordinal,
	)
	return err
}

func (s *SQLite) UpdateCvar(ctx context.Context, key string, valueInt int, valueString string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE cvars SET cvar_value_int = ?, cvar_value_string = NULLIF(?, '')
		 WHERE cvar_key = ?`,
		valueInt, valueString, key,
	)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}
