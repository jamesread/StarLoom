package store

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func WeekdaysToMask(weekdays []int) int {
	mask := 0
	for _, d := range weekdays {
		if d >= 1 && d <= 7 {
			mask |= 1 << (d - 1)
		}
	}
	if mask == 0 {
		mask = 127
	}
	return mask
}

func MaskToWeekdays(mask int) []int {
	out := []int{}
	for d := 1; d <= 7; d++ {
		if mask&(1<<(d-1)) != 0 {
			out = append(out, d)
		}
	}
	return out
}

func weekdayFromDate(date string) int {
	t, err := parseDate(date)
	if err != nil {
		return 0
	}
	wd := int(t.Weekday())
	if wd == 0 {
		return 7
	}
	return wd
}

func IsScheduledOnDate(mask int, date string) bool {
	wd := weekdayFromDate(date)
	if wd == 0 {
		return false
	}
	return mask&(1<<(wd-1)) != 0
}

// ListChores returns a family's chores in display order.
func (s *SQLite) ListChores(ctx context.Context, familyID int, starChartID int, includeInactive bool) ([]ChoreWithAssignments, error) {
	q := `SELECT id, family_id, star_chart_id, title, star_reward, weekday_mask, active, sort_order, created_at
		FROM chores WHERE family_id = ?`
	args := []any{familyID}
	if starChartID > 0 {
		q += ` AND star_chart_id = ?`
		args = append(args, starChartID)
	}
	if !includeInactive {
		q += ` AND active = 1`
	}
	q += ` ORDER BY sort_order, title`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChoreWithAssignments
	for rows.Next() {
		var c ChoreRow
		var active int
		var chartID sql.NullInt64
		if err := rows.Scan(&c.ID, &c.FamilyID, &chartID, &c.Title, &c.StarReward, &c.WeekdayMask, &active, &c.SortOrder, &c.CreatedAt); err != nil {
			return nil, err
		}
		if chartID.Valid {
			c.StarChartID = int(chartID.Int64)
		}
		c.Active = active != 0
		assigns, err := s.listAssignmentsForChore(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, ChoreWithAssignments{Chore: c, Assignments: assigns})
	}
	return out, rows.Err()
}

func (s *SQLite) listAssignmentsForChore(ctx context.Context, choreID int) ([]ChoreAssignmentRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, chore_id, child_member_id FROM chore_assignments WHERE chore_id = ? ORDER BY id`, choreID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChoreAssignmentRow
	for rows.Next() {
		var a ChoreAssignmentRow
		if err := rows.Scan(&a.ID, &a.ChoreID, &a.ChildMemberID); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

// GetChoreByID returns one chore with its assignments.
func (s *SQLite) GetChoreByID(ctx context.Context, id int) (*ChoreWithAssignments, error) {
	var c ChoreRow
	var active int
	var chartID sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, family_id, star_chart_id, title, star_reward, weekday_mask, active, sort_order, created_at FROM chores WHERE id = ?`, id,
	).Scan(&c.ID, &c.FamilyID, &chartID, &c.Title, &c.StarReward, &c.WeekdayMask, &active, &c.SortOrder, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if chartID.Valid {
		c.StarChartID = int(chartID.Int64)
	}
	c.Active = active != 0
	assigns, err := s.listAssignmentsForChore(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	return &ChoreWithAssignments{Chore: c, Assignments: assigns}, nil
}

// CreateChore appends a chore to the end of the order.
func (s *SQLite) CreateChore(ctx context.Context, familyID, starChartID int, title string, starReward, weekdayMask int, childMemberIDs []int) (int, error) {
	if weekdayMask == 0 {
		weekdayMask = 127
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO chores (family_id, star_chart_id, title, star_reward, weekday_mask, sort_order)
			VALUES (?, ?, ?, ?, ?, (SELECT COALESCE(MAX(sort_order), 0) + 1 FROM chores WHERE family_id = ?))`,
		familyID, starChartID, title, starReward, weekdayMask, familyID)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	choreID := int(id64)
	for _, childID := range childMemberIDs {
		if _, err := s.db.ExecContext(ctx,
			`INSERT INTO chore_assignments (chore_id, child_member_id) VALUES (?, ?)`, choreID, childID); err != nil {
			return 0, err
		}
	}
	return choreID, nil
}

func (s *SQLite) UpdateChore(ctx context.Context, id int, starChartID int, title string, starReward, weekdayMask int, active bool, childMemberIDs []int) error {
	activeInt := 0
	if active {
		activeInt = 1
	}
	if weekdayMask == 0 {
		weekdayMask = 127
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE chores SET star_chart_id = ?, title = ?, star_reward = ?, weekday_mask = ?, active = ? WHERE id = ?`,
		starChartID, title, starReward, weekdayMask, activeInt, id)
	if err != nil {
		return err
	}
	if _, err := s.db.ExecContext(ctx, `DELETE FROM chore_assignments WHERE chore_id = ?`, id); err != nil {
		return err
	}
	for _, childID := range childMemberIDs {
		if _, err := s.db.ExecContext(ctx,
			`INSERT INTO chore_assignments (chore_id, child_member_id) VALUES (?, ?)`, id, childID); err != nil {
			return err
		}
	}
	return nil
}

// ReorderChores numbers the given chores from one.
func (s *SQLite) ReorderChores(ctx context.Context, familyID int, choreIDs []int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	for i, id := range choreIDs {
		if _, err := tx.ExecContext(ctx,
			`UPDATE chores SET sort_order = ? WHERE id = ? AND family_id = ?`, i+1, id, familyID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *SQLite) DeactivateChore(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `UPDATE chores SET active = 0 WHERE id = ?`, id)
	return err
}

func (s *SQLite) ListChorePauses(ctx context.Context, familyID int) ([]ChorePauseRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, family_id, start_date, end_date, reason, created_at FROM chore_pauses
		 WHERE family_id = ? ORDER BY start_date DESC`, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ChorePauseRow
	for rows.Next() {
		var p ChorePauseRow
		if err := rows.Scan(&p.ID, &p.FamilyID, &p.StartDate, &p.EndDate, &p.Reason, &p.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *SQLite) CreateChorePause(ctx context.Context, familyID int, startDate, endDate, reason string) (int, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO chore_pauses (family_id, start_date, end_date, reason) VALUES (?, ?, ?, ?)`,
		familyID, startDate, endDate, reason)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	return int(id64), err
}

func (s *SQLite) DeleteChorePause(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM chore_pauses WHERE id = ?`, id)
	return err
}

func (s *SQLite) IsDatePaused(ctx context.Context, familyID int, date string) (bool, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM chore_pauses WHERE family_id = ? AND ? >= start_date AND ? <= end_date`,
		familyID, date, date).Scan(&n)
	return n > 0, err
}

func (s *SQLite) GetAssignment(ctx context.Context, choreID, childMemberID int) (*ChoreAssignmentRow, error) {
	var a ChoreAssignmentRow
	err := s.db.QueryRowContext(ctx,
		`SELECT id, chore_id, child_member_id FROM chore_assignments WHERE chore_id = ? AND child_member_id = ?`,
		choreID, childMemberID).Scan(&a.ID, &a.ChoreID, &a.ChildMemberID)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (s *SQLite) GetCompletion(ctx context.Context, assignmentID int, date string) (*ChoreCompletionRow, error) {
	var c ChoreCompletionRow
	var ledgerID sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT id, assignment_id, completion_date, ledger_entry_id, created_at
		 FROM chore_completions WHERE assignment_id = ? AND completion_date = ?`,
		assignmentID, date).Scan(&c.ID, &c.AssignmentID, &c.CompletionDate, &ledgerID, &c.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if ledgerID.Valid {
		v := int(ledgerID.Int64)
		c.LedgerEntryID = &v
	}
	return &c, nil
}

func (s *SQLite) InsertChoreCompletion(ctx context.Context, assignmentID int, date string, ledgerEntryID int) (int, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO chore_completions (assignment_id, completion_date, ledger_entry_id) VALUES (?, ?, ?)`,
		assignmentID, date, ledgerEntryID)
	if err != nil {
		return 0, err
	}
	id64, err := res.LastInsertId()
	return int(id64), err
}

func (s *SQLite) DeleteChoreCompletion(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM chore_completions WHERE id = ?`, id)
	return err
}

func (s *SQLite) ListCompletionsForWeek(ctx context.Context, familyID int, weekStart, weekEnd string) ([]WeeklyChartCompletion, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT cc.assignment_id, ca.chore_id, ca.child_member_id, cc.completion_date,
		       c.star_reward, cc.ledger_entry_id
		FROM chore_completions cc
		JOIN chore_assignments ca ON ca.id = cc.assignment_id
		JOIN chores c ON c.id = ca.chore_id
		WHERE c.family_id = ? AND cc.completion_date >= ? AND cc.completion_date <= ?
		ORDER BY cc.completion_date`, familyID, weekStart, weekEnd)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []WeeklyChartCompletion
	for rows.Next() {
		var w WeeklyChartCompletion
		var ledgerID sql.NullInt64
		if err := rows.Scan(&w.AssignmentID, &w.ChoreID, &w.ChildMemberID, &w.CompletionDate, &w.StarsEarned, &ledgerID); err != nil {
			return nil, err
		}
		if ledgerID.Valid {
			v := int(ledgerID.Int64)
			w.LedgerEntryID = &v
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (s *SQLite) ListChoreLedgerEntryIDs(ctx context.Context, familyID int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT cc.ledger_entry_id FROM chore_completions cc
		JOIN chore_assignments ca ON ca.id = cc.assignment_id
		JOIN chores c ON c.id = ca.chore_id
		WHERE c.family_id = ? AND cc.ledger_entry_id IS NOT NULL`, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func (s *SQLite) ListBonusStarsForWeek(ctx context.Context, familyID int, weekStart, weekEnd string, choreLedgerIDs []int) (map[int]map[string]int, error) {
	q := `
		SELECT child_member_id, date(created_at) AS d, SUM(amount) AS total
		FROM star_ledger_entries
		WHERE family_id = ? AND entry_type = 'award'
		  AND date(created_at) >= ? AND date(created_at) <= ?
		  AND note NOT LIKE 'Chore:%'`
	args := []any{familyID, weekStart, weekEnd}
	if len(choreLedgerIDs) > 0 {
		placeholders := strings.Repeat("?,", len(choreLedgerIDs))
		placeholders = placeholders[:len(placeholders)-1]
		q += fmt.Sprintf(` AND id NOT IN (%s)`, placeholders)
		for _, id := range choreLedgerIDs {
			args = append(args, id)
		}
	}
	q += ` GROUP BY child_member_id, date(created_at)`
	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[int]map[string]int{}
	for rows.Next() {
		var childID int
		var d string
		var total int
		if err := rows.Scan(&childID, &d, &total); err != nil {
			return nil, err
		}
		if out[childID] == nil {
			out[childID] = map[string]int{}
		}
		out[childID][d] = total
	}
	return out, rows.Err()
}
