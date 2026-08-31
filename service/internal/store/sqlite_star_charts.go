package store

import (
	"context"
	"database/sql"
	"fmt"
)

func (s *SQLite) ListStarCharts(ctx context.Context, familyID int, includeInactive bool) ([]StarChartRow, error) {
	q := `SELECT id, family_id, name, sort_order, active, created_at
		FROM star_charts WHERE family_id = ?`
	if !includeInactive {
		q += ` AND active = 1`
	}
	q += ` ORDER BY sort_order, name, id`
	rows, err := s.db.QueryContext(ctx, q, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []StarChartRow
	for rows.Next() {
		var sc StarChartRow
		var active int
		if err := rows.Scan(&sc.ID, &sc.FamilyID, &sc.Name, &sc.SortOrder, &active, &sc.CreatedAt); err != nil {
			return nil, err
		}
		sc.Active = active != 0
		out = append(out, sc)
	}
	return out, rows.Err()
}

func (s *SQLite) GetStarChartByID(ctx context.Context, id int) (*StarChartRow, error) {
	var sc StarChartRow
	var active int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, family_id, name, sort_order, active, created_at FROM star_charts WHERE id = ?`, id,
	).Scan(&sc.ID, &sc.FamilyID, &sc.Name, &sc.SortOrder, &active, &sc.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	sc.Active = active != 0
	return &sc, nil
}

func (s *SQLite) GetDefaultStarChartID(ctx context.Context, familyID int) (int, error) {
	var id int
	err := s.db.QueryRowContext(ctx,
		`SELECT id FROM star_charts WHERE family_id = ? AND active = 1 ORDER BY sort_order, id LIMIT 1`, familyID,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, fmt.Errorf("no star chart found")
	}
	return id, err
}

func (s *SQLite) CreateStarChart(ctx context.Context, familyID int, name string, sortOrder int) (int, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO star_charts (family_id, name, sort_order) VALUES (?, ?, ?)`,
		familyID, name, sortOrder)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id64), nil
}

func (s *SQLite) UpdateStarChart(ctx context.Context, id int, name string, sortOrder int, active bool) error {
	activeInt := 0
	if active {
		activeInt = 1
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE star_charts SET name = ?, sort_order = ?, active = ? WHERE id = ?`,
		name, sortOrder, activeInt, id)
	return err
}

func (s *SQLite) DeleteStarChart(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM star_charts WHERE id = ?`, id)
	return err
}

func (s *SQLite) CountChoresForStarChart(ctx context.Context, starChartID int) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM chores WHERE star_chart_id = ?`, starChartID,
	).Scan(&count)
	return count, err
}

func (s *SQLite) CountChoresForStarChartAndMember(ctx context.Context, starChartID, memberID int) (int, error) {
	var count int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM chores c
		 JOIN chore_assignments ca ON ca.chore_id = c.id
		 WHERE c.star_chart_id = ? AND c.active = 1 AND ca.child_member_id = ?`,
		starChartID, memberID,
	).Scan(&count)
	return count, err
}
