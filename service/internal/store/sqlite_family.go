package store

import (
	"context"
	"database/sql"
	"fmt"
)

func scanFamily(row *sql.Row) (*FamilyRow, error) {
	var f FamilyRow
	if err := row.Scan(&f.ID, &f.Name, &f.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &f, nil
}

func (s *SQLite) CountFamilies(ctx context.Context) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM families`).Scan(&n)
	return n, err
}

func (s *SQLite) GetFamilyByID(ctx context.Context, id int) (*FamilyRow, error) {
	return scanFamily(s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at FROM families WHERE id = ?`, id))
}

func (s *SQLite) GetFirstFamily(ctx context.Context) (*FamilyRow, error) {
	return scanFamily(s.db.QueryRowContext(ctx,
		`SELECT id, name, created_at FROM families ORDER BY id LIMIT 1`))
}

func (s *SQLite) CreateFamily(ctx context.Context, name string) (int, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO families (name) VALUES (?)`, name)
	if err != nil {
		return 0, fmt.Errorf("create family: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *SQLite) UpdateFamilyName(ctx context.Context, id int, name string) error {
	_, err := s.db.ExecContext(ctx, `UPDATE families SET name = ? WHERE id = ?`, name, id)
	return err
}

func scanMember(scanner interface {
	Scan(dest ...any) error
}) (*FamilyMemberRow, error) {
	var m FamilyMemberRow
	var accountID sql.NullInt64
	var avatarPath sql.NullString
	var starColor sql.NullString
	var username sql.NullString
	if err := scanner.Scan(
		&m.ID, &m.FamilyID, &accountID, &m.DisplayName, &m.Role, &avatarPath, &starColor, &m.CreatedAt, &username,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if accountID.Valid {
		v := int(accountID.Int64)
		m.UserAccountID = &v
	}
	if avatarPath.Valid {
		m.AvatarPath = avatarPath.String
	}
	if starColor.Valid {
		m.StarColor = starColor.String
	}
	if username.Valid {
		m.Username = username.String
	}
	return &m, nil
}

const memberSelect = `SELECT fm.id, fm.family_id, fm.user_account_id, fm.display_name, fm.role,
  fm.avatar_path, fm.star_color, fm.created_at, ua.username
  FROM family_members fm
  LEFT JOIN user_accounts ua ON ua.id = fm.user_account_id`

func (s *SQLite) GetMemberByID(ctx context.Context, id int) (*FamilyMemberRow, error) {
	return scanMember(s.db.QueryRowContext(ctx, memberSelect+` WHERE fm.id = ?`, id))
}

func (s *SQLite) GetMemberByAccountID(ctx context.Context, accountID int) (*FamilyMemberRow, error) {
	return scanMember(s.db.QueryRowContext(ctx, memberSelect+` WHERE fm.user_account_id = ?`, accountID))
}

func (s *SQLite) ListMembersByFamily(ctx context.Context, familyID int) ([]FamilyMemberRow, error) {
	rows, err := s.db.QueryContext(ctx, memberSelect+` WHERE fm.family_id = ? ORDER BY fm.role, fm.display_name`, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []FamilyMemberRow{}
	for rows.Next() {
		var m FamilyMemberRow
		var accountID sql.NullInt64
		var avatarPath sql.NullString
		var starColor sql.NullString
		var username sql.NullString
		if err := rows.Scan(
			&m.ID, &m.FamilyID, &accountID, &m.DisplayName, &m.Role, &avatarPath, &starColor, &m.CreatedAt, &username,
		); err != nil {
			return nil, err
		}
		if accountID.Valid {
			v := int(accountID.Int64)
			m.UserAccountID = &v
		}
		if avatarPath.Valid {
			m.AvatarPath = avatarPath.String
		}
		if starColor.Valid {
			m.StarColor = starColor.String
		}
		if username.Valid {
			m.Username = username.String
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (s *SQLite) CreateMember(ctx context.Context, familyID int, displayName, role string, accountID *int, starColor string) (int, error) {
	var aid any
	if accountID != nil {
		aid = *accountID
	}
	var color any
	if starColor != "" {
		color = starColor
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO family_members (family_id, display_name, role, user_account_id, star_color) VALUES (?, ?, ?, ?, ?)`,
		familyID, displayName, role, aid, color)
	if err != nil {
		return 0, fmt.Errorf("create member: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *SQLite) UpdateMember(ctx context.Context, id int, displayName, starColor string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE family_members SET display_name = ?, star_color = ? WHERE id = ?`,
		displayName, starColor, id)
	return err
}

func (s *SQLite) SetMemberUserAccount(ctx context.Context, id int, accountID int) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE family_members SET user_account_id = ? WHERE id = ? AND user_account_id IS NULL`,
		accountID, id)
	if err != nil {
		return fmt.Errorf("set member user account: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("member not found or already has login")
	}
	return nil
}

func (s *SQLite) DeleteMember(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM family_members WHERE id = ? AND role = ?`, id, MemberRoleChild)
	return err
}

func (s *SQLite) SetMemberAvatarPath(ctx context.Context, id int, path string) error {
	var v any
	if path != "" {
		v = path
	}
	_, err := s.db.ExecContext(ctx, `UPDATE family_members SET avatar_path = ? WHERE id = ?`, v, id)
	return err
}

func (s *SQLite) GetMemberBalance(ctx context.Context, memberID int) (int, error) {
	var balance sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(amount), 0) FROM star_ledger_entries WHERE child_member_id = ?`, memberID,
	).Scan(&balance)
	if err != nil {
		return 0, err
	}
	if !balance.Valid {
		return 0, nil
	}
	return int(balance.Int64), nil
}

func scanLedger(scanner interface {
	Scan(dest ...any) error
}) (*StarLedgerRow, error) {
	var e StarLedgerRow
	var rewardID sql.NullInt64
	var createdBy sql.NullInt64
	if err := scanner.Scan(
		&e.ID, &e.FamilyID, &e.ChildMemberID, &e.Amount, &e.EntryType, &e.Note,
		&rewardID, &createdBy, &e.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if rewardID.Valid {
		v := int(rewardID.Int64)
		e.RelatedRewardID = &v
	}
	if createdBy.Valid {
		v := int(createdBy.Int64)
		e.CreatedByMemberID = &v
	}
	return &e, nil
}

func (s *SQLite) ListLedger(ctx context.Context, memberID, limit int) ([]StarLedgerRow, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, family_id, child_member_id, amount, entry_type, note, related_reward_id, created_by_member_id, created_at
		 FROM star_ledger_entries WHERE child_member_id = ? ORDER BY created_at DESC, id DESC LIMIT ?`,
		memberID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []StarLedgerRow{}
	for rows.Next() {
		e, err := scanLedger(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *e)
	}
	return out, rows.Err()
}

func (s *SQLite) GetLastAward(ctx context.Context, memberID int) (*StarLedgerRow, error) {
	return scanLedger(s.db.QueryRowContext(ctx,
		`SELECT id, family_id, child_member_id, amount, entry_type, note, related_reward_id, created_by_member_id, created_at
		 FROM star_ledger_entries WHERE child_member_id = ? AND entry_type = ? ORDER BY created_at DESC, id DESC LIMIT 1`,
		memberID, LedgerTypeAward))
}

func (s *SQLite) InsertLedgerEntry(ctx context.Context, row StarLedgerRow) (int, error) {
	var rewardID any
	if row.RelatedRewardID != nil {
		rewardID = *row.RelatedRewardID
	}
	var createdBy any
	if row.CreatedByMemberID != nil {
		createdBy = *row.CreatedByMemberID
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO star_ledger_entries (family_id, child_member_id, amount, entry_type, note, related_reward_id, created_by_member_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		row.FamilyID, row.ChildMemberID, row.Amount, row.EntryType, row.Note, rewardID, createdBy)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *SQLite) ListRewards(ctx context.Context, familyID int, includeInactive bool) ([]RewardRow, error) {
	q := `SELECT id, family_id, title, description, cost_stars, active, approval_required, availability_expression FROM rewards WHERE family_id = ?`
	if !includeInactive {
		q += ` AND active = 1`
	}
	q += ` ORDER BY cost_stars, title`
	rows, err := s.db.QueryContext(ctx, q, familyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []RewardRow{}
	for rows.Next() {
		var r RewardRow
		var active, approval int
		if err := rows.Scan(&r.ID, &r.FamilyID, &r.Title, &r.Description, &r.CostStars, &active, &approval, &r.AvailabilityExpression); err != nil {
			return nil, err
		}
		r.Active = active != 0
		r.ApprovalRequired = approval != 0
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLite) GetRewardByID(ctx context.Context, id int) (*RewardRow, error) {
	var r RewardRow
	var active, approval int
	err := s.db.QueryRowContext(ctx,
		`SELECT id, family_id, title, description, cost_stars, active, approval_required, availability_expression FROM rewards WHERE id = ?`, id,
	).Scan(&r.ID, &r.FamilyID, &r.Title, &r.Description, &r.CostStars, &active, &approval, &r.AvailabilityExpression)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.Active = active != 0
	r.ApprovalRequired = approval != 0
	return &r, nil
}

func (s *SQLite) CreateReward(ctx context.Context, familyID int, title, description string, costStars int, approvalRequired bool, availabilityExpression string) (int, error) {
	active := 1
	approval := 0
	if approvalRequired {
		approval = 1
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO rewards (family_id, title, description, cost_stars, active, approval_required, availability_expression) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		familyID, title, description, costStars, active, approval, availabilityExpression)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *SQLite) UpdateReward(ctx context.Context, id int, title, description string, costStars int, active, approvalRequired bool, availabilityExpression string) error {
	activeInt := 0
	if active {
		activeInt = 1
	}
	approvalInt := 0
	if approvalRequired {
		approvalInt = 1
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE rewards SET title = ?, description = ?, cost_stars = ?, active = ?, approval_required = ?, availability_expression = ? WHERE id = ?`,
		title, description, costStars, activeInt, approvalInt, availabilityExpression, id)
	return err
}

func (s *SQLite) DeactivateReward(ctx context.Context, id int) error {
	_, err := s.db.ExecContext(ctx, `UPDATE rewards SET active = 0 WHERE id = ?`, id)
	return err
}

func (s *SQLite) ListRedemptions(ctx context.Context, familyID int, status string) ([]RedemptionRow, error) {
	q := `SELECT r.id, r.family_id, r.child_member_id, r.reward_id, r.stars_spent, r.status,
		r.ledger_entry_id, r.created_at, COALESCE(r.resolved_at, ''), r.resolved_by_member_id,
		COALESCE(r.fulfilled_at, ''), rw.title, fm.display_name
		FROM redemptions r
		INNER JOIN rewards rw ON rw.id = r.reward_id
		INNER JOIN family_members fm ON fm.id = r.child_member_id
		WHERE r.family_id = ?`
	args := []any{familyID}
	if status != "" {
		q += ` AND r.status = ?`
		args = append(args, status)
	}
	q += ` ORDER BY r.created_at DESC, r.id DESC`

	rows, err := s.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []RedemptionRow{}
	for rows.Next() {
		var r RedemptionRow
		var ledgerID sql.NullInt64
		var resolvedBy sql.NullInt64
		if err := rows.Scan(
			&r.ID, &r.FamilyID, &r.ChildMemberID, &r.RewardID, &r.StarsSpent, &r.Status,
			&ledgerID, &r.CreatedAt, &r.ResolvedAt, &resolvedBy, &r.FulfilledAt, &r.RewardTitle, &r.ChildDisplayName,
		); err != nil {
			return nil, err
		}
		if ledgerID.Valid {
			v := int(ledgerID.Int64)
			r.LedgerEntryID = &v
		}
		if resolvedBy.Valid {
			v := int(resolvedBy.Int64)
			r.ResolvedByMemberID = &v
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *SQLite) GetRedemptionByID(ctx context.Context, id int) (*RedemptionRow, error) {
	var r RedemptionRow
	var ledgerID sql.NullInt64
	var resolvedBy sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT r.id, r.family_id, r.child_member_id, r.reward_id, r.stars_spent, r.status,
		r.ledger_entry_id, r.created_at, COALESCE(r.resolved_at, ''), r.resolved_by_member_id,
		COALESCE(r.fulfilled_at, ''), rw.title, fm.display_name
		FROM redemptions r
		INNER JOIN rewards rw ON rw.id = r.reward_id
		INNER JOIN family_members fm ON fm.id = r.child_member_id
		WHERE r.id = ?`, id,
	).Scan(
		&r.ID, &r.FamilyID, &r.ChildMemberID, &r.RewardID, &r.StarsSpent, &r.Status,
		&ledgerID, &r.CreatedAt, &r.ResolvedAt, &resolvedBy, &r.FulfilledAt, &r.RewardTitle, &r.ChildDisplayName,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if ledgerID.Valid {
		v := int(ledgerID.Int64)
		r.LedgerEntryID = &v
	}
	if resolvedBy.Valid {
		v := int(resolvedBy.Int64)
		r.ResolvedByMemberID = &v
	}
	return &r, nil
}

func (s *SQLite) CreateRedemption(ctx context.Context, familyID, childMemberID, rewardID, starsSpent int, status string, ledgerEntryID *int) (int, error) {
	var ledger any
	if ledgerEntryID != nil {
		ledger = *ledgerEntryID
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO redemptions (family_id, child_member_id, reward_id, stars_spent, status, ledger_entry_id)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		familyID, childMemberID, rewardID, starsSpent, status, ledger)
	if err != nil {
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (s *SQLite) ResolveRedemption(ctx context.Context, id int, status string, resolvedByMemberID int, ledgerEntryID *int) error {
	var ledger any
	if ledgerEntryID != nil {
		ledger = *ledgerEntryID
	}
	_, err := s.db.ExecContext(ctx,
		`UPDATE redemptions SET status = ?, resolved_at = datetime('now'), resolved_by_member_id = ?, ledger_entry_id = ?
		 WHERE id = ?`,
		status, resolvedByMemberID, ledger, id)
	return err
}

func (s *SQLite) CountPendingRedemptions(ctx context.Context, familyID int) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM redemptions WHERE family_id = ? AND status = ?`, familyID, RedemptionPending,
	).Scan(&n)
	return n, err
}

func (s *SQLite) CountApprovedRedemptionsForMemberRewardBetween(ctx context.Context, childMemberID, rewardID int, startDate, endDate string) (int, error) {
	var n int
	err := s.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM redemptions
		WHERE child_member_id = ? AND reward_id = ? AND status = ?
		  AND date(COALESCE(NULLIF(resolved_at, ''), created_at)) >= date(?)
		  AND date(COALESCE(NULLIF(resolved_at, ''), created_at)) <= date(?)`,
		childMemberID, rewardID, RedemptionApproved, startDate, endDate,
	).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("count approved redemptions: %w", err)
	}
	return n, nil
}
