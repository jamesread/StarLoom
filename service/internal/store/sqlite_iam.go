package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jamesread/starapp/service/internal/rbac"
)

var errNoSuperuser = errors.New("refusing to leave the system without a superuser")

func userAccountSelectSQL() string {
	return `SELECT id, username, password_hash, created_by, created_at, updated_at FROM user_accounts`
}

func scanUserAccount(row interface{ Scan(dest ...any) error }) (*UserAccountRow, error) {
	var u UserAccountRow
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedBy, &u.CreatedAt, &u.UpdatedAt); err != nil {
		return nil, err
	}
	return &u, nil
}

func sessionSelectSQL() string {
	return `SELECT id, sid, user_account_id, impersonator_user_id, created_at, updated_at FROM sessions`
}

func scanSession(row interface{ Scan(dest ...any) error }) (*SessionRow, error) {
	var s SessionRow
	var impersonator sql.NullInt64
	if err := row.Scan(&s.ID, &s.SID, &s.UserAccountID, &impersonator, &s.CreatedAt, &s.UpdatedAt); err != nil {
		return nil, err
	}
	if impersonator.Valid {
		v := int(impersonator.Int64)
		s.ImpersonatorUserID = &v
	}
	return &s, nil
}

func apiKeySelectSQL() string {
	return `SELECT id, user_account_id, name, key_value, read_only, COALESCE(last_used_at, ''), created_at, updated_at FROM api_keys`
}

func scanAPIKey(row interface{ Scan(dest ...any) error }) (*APIKeyRow, error) {
	var k APIKeyRow
	var readOnly int
	if err := row.Scan(&k.ID, &k.UserAccountID, &k.Name, &k.KeyValue, &readOnly, &k.LastUsedAt, &k.CreatedAt, &k.UpdatedAt); err != nil {
		return nil, err
	}
	k.ReadOnly = readOnly != 0
	return &k, nil
}

func (s *SQLite) CountUserAccounts(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM user_accounts`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count user accounts: %w", err)
	}
	return n, nil
}

func (s *SQLite) GetUserByUsername(ctx context.Context, username string) (*UserAccountRow, error) {
	row := s.db.QueryRowContext(ctx, userAccountSelectSQL()+` WHERE username = ?`, username)
	u, err := scanUserAccount(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by username: %w", err)
	}
	return u, nil
}

func (s *SQLite) GetUserByID(ctx context.Context, id int) (*UserAccountRow, error) {
	row := s.db.QueryRowContext(ctx, userAccountSelectSQL()+` WHERE id = ?`, id)
	u, err := scanUserAccount(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return u, nil
}

func (s *SQLite) ListUserAccounts(ctx context.Context) ([]UserAccountRow, error) {
	rows, err := s.db.QueryContext(ctx, userAccountSelectSQL()+` ORDER BY username COLLATE NOCASE`)
	if err != nil {
		return nil, fmt.Errorf("list user accounts: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]UserAccountRow, 0)
	for rows.Next() {
		u, err := scanUserAccount(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user account: %w", err)
		}
		out = append(out, *u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list user accounts: %w", err)
	}
	return out, nil
}

func (s *SQLite) CreateUserAccount(ctx context.Context, username, passwordHash, createdBy string) (int, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO user_accounts (username, password_hash, created_by) VALUES (?, ?, ?)`,
		username, passwordHash, createdBy,
	)
	if err != nil {
		return 0, fmt.Errorf("create user account: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("create user account id: %w", err)
	}
	return int(id), nil
}

func (s *SQLite) DeleteUserAccount(ctx context.Context, id int) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM user_accounts WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete user account: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user account rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLite) UpdateUserPassword(ctx context.Context, id int, passwordHash string) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE user_accounts SET password_hash = ?, updated_at = datetime('now') WHERE id = ?`,
		passwordHash, id,
	)
	if err != nil {
		return fmt.Errorf("update user password: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update user password rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLite) CreateSession(ctx context.Context, sid string, userID int, impersonatorID *int) error {
	var impersonator any
	if impersonatorID != nil {
		impersonator = *impersonatorID
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO sessions (sid, user_account_id, impersonator_user_id) VALUES (?, ?, ?)`,
		sid, userID, impersonator,
	)
	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}
	return nil
}

func (s *SQLite) GetSessionBySID(ctx context.Context, sid string) (*SessionRow, error) {
	row := s.db.QueryRowContext(ctx, sessionSelectSQL()+` WHERE sid = ?`, sid)
	sess, err := scanSession(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session by sid: %w", err)
	}
	return sess, nil
}

func (s *SQLite) DeleteSession(ctx context.Context, sid string) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE sid = ?`, sid)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete session rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLite) DeleteSessionsForUser(ctx context.Context, userID int) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM sessions WHERE user_account_id = ?`, userID)
	if err != nil {
		return fmt.Errorf("delete sessions for user: %w", err)
	}
	return nil
}

func (s *SQLite) ListAPIKeysForUser(ctx context.Context, userID int) ([]APIKeyRow, error) {
	rows, err := s.db.QueryContext(ctx, apiKeySelectSQL()+` WHERE user_account_id = ? ORDER BY name, id`, userID)
	if err != nil {
		return nil, fmt.Errorf("list api keys for user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]APIKeyRow, 0)
	for rows.Next() {
		k, err := scanAPIKey(rows)
		if err != nil {
			return nil, fmt.Errorf("scan api key: %w", err)
		}
		out = append(out, *k)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list api keys for user: %w", err)
	}
	return out, nil
}

func (s *SQLite) CreateAPIKey(ctx context.Context, userID int, name, keyValue string, readOnly bool) (int, error) {
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO api_keys (user_account_id, name, key_value, read_only) VALUES (?, ?, ?, ?)`,
		userID, name, keyValue, boolToInt(readOnly),
	)
	if err != nil {
		return 0, fmt.Errorf("create api key: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("create api key id: %w", err)
	}
	return int(id), nil
}

func (s *SQLite) DeleteAPIKey(ctx context.Context, id, userID int) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM api_keys WHERE id = ? AND user_account_id = ?`, id, userID)
	if err != nil {
		return fmt.Errorf("delete api key: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete api key rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLite) GetUserByAPIKey(ctx context.Context, keyValue string) (*UserAccountRow, bool, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT u.id, u.username, u.password_hash, u.created_by, u.created_at, u.updated_at, k.read_only
		 FROM api_keys k
		 INNER JOIN user_accounts u ON u.id = k.user_account_id
		 WHERE k.key_value = ?`, keyValue,
	)
	var u UserAccountRow
	var readOnly int
	if err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.CreatedBy, &u.CreatedAt, &u.UpdatedAt, &readOnly); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf("get user by api key: %w", err)
	}
	return &u, readOnly != 0, nil
}

func (s *SQLite) TouchAPIKeyUsed(ctx context.Context, keyValue string) error {
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	res, err := s.db.ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = ?, updated_at = datetime('now') WHERE key_value = ?`,
		now, keyValue,
	)
	if err != nil {
		return fmt.Errorf("touch api key used: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("touch api key used rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLite) LoadEffectiveRBAC(ctx context.Context, userID int) (*rbac.EffectiveRBAC, error) {
	var superCount int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM user_group_memberships ugm
		 INNER JOIN rbac_group_roles gr ON gr.user_group_id = ugm.user_group_id
		 INNER JOIN rbac_roles r ON r.id = gr.role_id
		 WHERE ugm.user_account_id = ? AND r.name = ?`,
		userID, rbac.RoleSuperuser,
	).Scan(&superCount); err != nil {
		return nil, fmt.Errorf("load effective rbac superuser count: %w", err)
	}

	out := &rbac.EffectiveRBAC{
		IsSuperuser: superCount > 0,
		Permissions: map[string]bool{},
	}

	if out.IsSuperuser {
		rows, err := s.db.QueryContext(ctx, `SELECT name FROM rbac_permissions ORDER BY name`)
		if err != nil {
			return nil, fmt.Errorf("load effective rbac permissions: %w", err)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, fmt.Errorf("scan rbac permission: %w", err)
			}
			out.Permissions[name] = true
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("load effective rbac permissions: %w", err)
		}
	} else {
		rows, err := s.db.QueryContext(ctx,
			`SELECT DISTINCT p.name FROM rbac_permissions p
			 INNER JOIN rbac_role_permissions rp ON rp.permission_id = p.id
			 INNER JOIN rbac_group_roles gr ON gr.role_id = rp.role_id
			 INNER JOIN user_group_memberships ugm ON ugm.user_group_id = gr.user_group_id
			 WHERE ugm.user_account_id = ?
			 ORDER BY p.name`, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("load effective rbac permissions: %w", err)
		}
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			var name string
			if err := rows.Scan(&name); err != nil {
				return nil, fmt.Errorf("scan rbac permission: %w", err)
			}
			out.Permissions[name] = true
		}
		if err := rows.Err(); err != nil {
			return nil, fmt.Errorf("load effective rbac permissions: %w", err)
		}
	}

	roleRows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT r.name FROM rbac_roles r
		 INNER JOIN rbac_group_roles gr ON gr.role_id = r.id
		 INNER JOIN user_group_memberships ugm ON ugm.user_group_id = gr.user_group_id
		 WHERE ugm.user_account_id = ?
		 ORDER BY r.name`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("load effective rbac roles: %w", err)
	}
	defer func() { _ = roleRows.Close() }()
	for roleRows.Next() {
		var name string
		if err := roleRows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan rbac role: %w", err)
		}
		out.RoleNames = append(out.RoleNames, name)
	}
	if err := roleRows.Err(); err != nil {
		return nil, fmt.Errorf("load effective rbac roles: %w", err)
	}

	return out, nil
}

func (s *SQLite) EnsureRBACBootstrap(ctx context.Context) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("ensure rbac bootstrap begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO user_groups (name) VALUES (?), (?), (?), (?)`,
		rbac.GroupEveryone, rbac.GroupAdministrators, rbac.GroupParents, rbac.GroupChildren,
	); err != nil {
		return fmt.Errorf("ensure rbac bootstrap groups: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO rbac_group_roles (user_group_id, role_id)
		 SELECT g.id, r.id FROM user_groups g CROSS JOIN rbac_roles r
		 WHERE g.name = ? AND r.name = ?`,
		rbac.GroupEveryone, rbac.RoleMember,
	); err != nil {
		return fmt.Errorf("ensure rbac bootstrap everyone role: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO rbac_group_roles (user_group_id, role_id)
		 SELECT g.id, r.id FROM user_groups g CROSS JOIN rbac_roles r
		 WHERE g.name = ? AND r.name = ?`,
		rbac.GroupAdministrators, rbac.RoleSuperuser,
	); err != nil {
		return fmt.Errorf("ensure rbac bootstrap administrators role: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO rbac_group_roles (user_group_id, role_id)
		 SELECT g.id, r.id FROM user_groups g CROSS JOIN rbac_roles r
		 WHERE g.name = ? AND r.name = ?`,
		rbac.GroupParents, rbac.RoleParent,
	); err != nil {
		return fmt.Errorf("ensure rbac bootstrap parents role: %w", err)
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO rbac_group_roles (user_group_id, role_id)
		 SELECT g.id, r.id FROM user_groups g CROSS JOIN rbac_roles r
		 WHERE g.name = ? AND r.name = ?`,
		rbac.GroupChildren, rbac.RoleChild,
	); err != nil {
		return fmt.Errorf("ensure rbac bootstrap children role: %w", err)
	}

	var superCount int
	if err := tx.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT ugm.user_account_id)
		 FROM user_group_memberships ugm
		 INNER JOIN rbac_group_roles gr ON gr.user_group_id = ugm.user_group_id
		 INNER JOIN rbac_roles r ON r.id = gr.role_id
		 WHERE r.name = ?`, rbac.RoleSuperuser,
	).Scan(&superCount); err != nil {
		return fmt.Errorf("ensure rbac bootstrap superuser count: %w", err)
	}

	if superCount == 0 {
		if _, err := tx.ExecContext(ctx,
			`INSERT OR IGNORE INTO user_group_memberships (user_account_id, user_group_id)
			 SELECT u.id, g.id FROM user_accounts u
			 CROSS JOIN user_groups g
			 WHERE g.name = ?
			 AND u.id = (SELECT MIN(id) FROM user_accounts)`,
			rbac.GroupAdministrators,
		); err != nil {
			return fmt.Errorf("ensure rbac bootstrap first superuser: %w", err)
		}
	}

	if _, err := tx.ExecContext(ctx,
		`INSERT OR IGNORE INTO user_group_memberships (user_account_id, user_group_id)
		 SELECT u.id, g.id FROM user_accounts u
		 CROSS JOIN user_groups g
		 WHERE g.name = ?
		 AND NOT EXISTS (
		   SELECT 1 FROM user_group_memberships ugm WHERE ugm.user_account_id = u.id
		 )`, rbac.GroupEveryone,
	); err != nil {
		return fmt.Errorf("ensure rbac bootstrap everyone memberships: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("ensure rbac bootstrap commit: %w", err)
	}
	return nil
}

func (s *SQLite) EnsureUserInEveryoneGroup(ctx context.Context, userID int) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO user_group_memberships (user_account_id, user_group_id)
		 SELECT ?, g.id FROM user_groups g WHERE g.name = ?`,
		userID, rbac.GroupEveryone,
	)
	if err != nil {
		return fmt.Errorf("ensure user in everyone group: %w", err)
	}
	return nil
}

func (s *SQLite) CountUsersWithSuperuserViaGroups(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT ugm.user_account_id)
		 FROM user_group_memberships ugm
		 INNER JOIN rbac_group_roles gr ON gr.user_group_id = ugm.user_group_id
		 INNER JOIN rbac_roles r ON r.id = gr.role_id
		 WHERE r.name = ?`, rbac.RoleSuperuser,
	).Scan(&n); err != nil {
		return 0, fmt.Errorf("count users with superuser via groups: %w", err)
	}
	return n, nil
}

func (s *SQLite) ensureSuperuserCoverage(ctx context.Context) error {
	n, err := s.CountUsersWithSuperuserViaGroups(ctx)
	if err != nil {
		return err
	}
	if n == 0 {
		return errNoSuperuser
	}
	return nil
}

func (s *SQLite) ListRBACPermissions(ctx context.Context) ([]RBACPermissionRow, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description FROM rbac_permissions ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list rbac permissions: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]RBACPermissionRow, 0)
	for rows.Next() {
		var p RBACPermissionRow
		if err := rows.Scan(&p.ID, &p.Name, &p.Description); err != nil {
			return nil, fmt.Errorf("scan rbac permission: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list rbac permissions: %w", err)
	}
	return out, nil
}

func (s *SQLite) loadRBACRole(ctx context.Context, id int) (*RBACRoleRow, error) {
	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, description FROM rbac_roles WHERE id = ?`, id,
	)
	var role RBACRoleRow
	if err := row.Scan(&role.ID, &role.Name, &role.Description); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("load rbac role: %w", err)
	}

	permIDs, err := s.ListRolePermissionIDs(ctx, id)
	if err != nil {
		return nil, err
	}
	role.PermissionIDs = permIDs

	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM rbac_group_roles WHERE role_id = ?`, id,
	).Scan(&role.GroupCount); err != nil {
		return nil, fmt.Errorf("load rbac role group count: %w", err)
	}

	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(DISTINCT ugm.user_account_id)
		 FROM user_group_memberships ugm
		 INNER JOIN rbac_group_roles gr ON gr.user_group_id = ugm.user_group_id
		 WHERE gr.role_id = ?`, id,
	).Scan(&role.UserCount); err != nil {
		return nil, fmt.Errorf("load rbac role user count: %w", err)
	}

	return &role, nil
}

func (s *SQLite) ListRBACRoles(ctx context.Context) ([]RBACRoleRow, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM rbac_roles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list rbac roles: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]RBACRoleRow, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan rbac role id: %w", err)
		}
		role, err := s.loadRBACRole(ctx, id)
		if err != nil {
			return nil, err
		}
		if role != nil {
			out = append(out, *role)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list rbac roles: %w", err)
	}
	return out, nil
}

func (s *SQLite) GetRBACRole(ctx context.Context, id int) (*RBACRoleRow, error) {
	return s.loadRBACRole(ctx, id)
}

func (s *SQLite) rbacRoleNameByID(ctx context.Context, id int) (string, error) {
	var name string
	if err := s.db.QueryRowContext(ctx, `SELECT name FROM rbac_roles WHERE id = ?`, id).Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("rbac role name by id: %w", err)
	}
	return name, nil
}

func isSystemRBACRole(name string) bool {
	return name == rbac.RoleSuperuser || name == rbac.RoleMember
}

func (s *SQLite) setRBACRolePermissionsTx(ctx context.Context, tx *sql.Tx, roleID int, permissionIDs []int) error {
	if _, err := tx.ExecContext(ctx, `DELETE FROM rbac_role_permissions WHERE role_id = ?`, roleID); err != nil {
		return fmt.Errorf("clear rbac role permissions: %w", err)
	}
	for _, pid := range permissionIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO rbac_role_permissions (role_id, permission_id) VALUES (?, ?)`,
			roleID, pid,
		); err != nil {
			return fmt.Errorf("insert rbac role permission: %w", err)
		}
	}
	return nil
}

func (s *SQLite) CreateRBACRole(ctx context.Context, name, description string, permissionIDs []int) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("create rbac role begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`INSERT INTO rbac_roles (name, description) VALUES (?, ?)`,
		name, description,
	)
	if err != nil {
		return 0, fmt.Errorf("create rbac role: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("create rbac role id: %w", err)
	}
	roleID := int(id)
	if err := s.setRBACRolePermissionsTx(ctx, tx, roleID, permissionIDs); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("create rbac role commit: %w", err)
	}
	return roleID, nil
}

func (s *SQLite) UpdateRBACRole(ctx context.Context, id int, name, description string, permissionIDs []int) error {
	curName, err := s.rbacRoleNameByID(ctx, id)
	if err != nil {
		return err
	}
	if isSystemRBACRole(curName) && name != curName {
		return fmt.Errorf("cannot rename system role %q", curName)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("update rbac role begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	res, err := tx.ExecContext(ctx,
		`UPDATE rbac_roles SET name = ?, description = ?, updated_at = datetime('now') WHERE id = ?`,
		name, description, id,
	)
	if err != nil {
		return fmt.Errorf("update rbac role: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update rbac role rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}

	if curName != rbac.RoleSuperuser {
		if err := s.setRBACRolePermissionsTx(ctx, tx, id, permissionIDs); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("update rbac role commit: %w", err)
	}
	return nil
}

func (s *SQLite) DeleteRBACRole(ctx context.Context, id int) error {
	name, err := s.rbacRoleNameByID(ctx, id)
	if err != nil {
		return err
	}
	if isSystemRBACRole(name) {
		return fmt.Errorf("cannot delete system role %q", name)
	}

	res, err := s.db.ExecContext(ctx, `DELETE FROM rbac_roles WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete rbac role: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete rbac role rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLite) SetRBACRolePermissions(ctx context.Context, roleID int, permissionIDs []int) error {
	name, err := s.rbacRoleNameByID(ctx, roleID)
	if err != nil {
		return err
	}
	if name == rbac.RoleSuperuser {
		return fmt.Errorf("cannot set permissions for system role %q", name)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("set rbac role permissions begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if err := s.setRBACRolePermissionsTx(ctx, tx, roleID, permissionIDs); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("set rbac role permissions commit: %w", err)
	}
	return nil
}

func (s *SQLite) ListRolePermissionIDs(ctx context.Context, roleID int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT permission_id FROM rbac_role_permissions WHERE role_id = ? ORDER BY permission_id`, roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("list role permission ids: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan role permission id: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list role permission ids: %w", err)
	}
	return out, nil
}

func (s *SQLite) ListPermissionRoleNames(ctx context.Context, permissionID int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT r.name FROM rbac_roles r
		 INNER JOIN rbac_role_permissions rp ON rp.role_id = r.id
		 WHERE rp.permission_id = ?
		 ORDER BY r.name`, permissionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list permission role names: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan permission role name: %w", err)
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list permission role names: %w", err)
	}
	return out, nil
}

func (s *SQLite) GetUserRbacRoleNames(ctx context.Context, userID int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT r.name FROM rbac_roles r
		 INNER JOIN rbac_group_roles gr ON gr.role_id = r.id
		 INNER JOIN user_group_memberships ugm ON ugm.user_group_id = gr.user_group_id
		 WHERE ugm.user_account_id = ?
		 ORDER BY r.name`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("get user rbac role names: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan user rbac role name: %w", err)
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get user rbac role names: %w", err)
	}
	return out, nil
}

func (s *SQLite) GetUserGroupRbacRoleIDs(ctx context.Context, groupID int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT role_id FROM rbac_group_roles WHERE user_group_id = ? ORDER BY role_id`, groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("get user group rbac role ids: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user group rbac role id: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("get user group rbac role ids: %w", err)
	}
	return out, nil
}

func (s *SQLite) SetUserGroupRbacRoles(ctx context.Context, groupID int, roleIDs []int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("set user group rbac roles begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM rbac_group_roles WHERE user_group_id = ?`, groupID); err != nil {
		return fmt.Errorf("clear user group rbac roles: %w", err)
	}
	for _, roleID := range roleIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO rbac_group_roles (user_group_id, role_id) VALUES (?, ?)`,
			groupID, roleID,
		); err != nil {
			return fmt.Errorf("insert user group rbac role: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("set user group rbac roles commit: %w", err)
	}
	return s.ensureSuperuserCoverage(ctx)
}

func (s *SQLite) ListRbacRoleGroupNames(ctx context.Context, roleID int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT g.name FROM user_groups g
		 INNER JOIN rbac_group_roles gr ON gr.user_group_id = g.id
		 WHERE gr.role_id = ?
		 ORDER BY g.name`, roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("list rbac role group names: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan rbac role group name: %w", err)
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list rbac role group names: %w", err)
	}
	return out, nil
}

func (s *SQLite) ListRbacRoleUsernames(ctx context.Context, roleID int) ([]string, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT DISTINCT u.username FROM user_accounts u
		 INNER JOIN user_group_memberships ugm ON ugm.user_account_id = u.id
		 INNER JOIN rbac_group_roles gr ON gr.user_group_id = ugm.user_group_id
		 WHERE gr.role_id = ?
		 ORDER BY u.username COLLATE NOCASE`, roleID,
	)
	if err != nil {
		return nil, fmt.Errorf("list rbac role usernames: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]string, 0)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("scan rbac role username: %w", err)
		}
		out = append(out, name)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list rbac role usernames: %w", err)
	}
	return out, nil
}

func (s *SQLite) GetMyPermissionsAudit(ctx context.Context, userID int) ([]string, []string, bool, []MyPermissionAuditRow, error) {
	groupRows, err := s.db.QueryContext(ctx,
		`SELECT g.name FROM user_groups g
		 INNER JOIN user_group_memberships ugm ON ugm.user_group_id = g.id
		 WHERE ugm.user_account_id = ?
		 ORDER BY g.name`, userID,
	)
	if err != nil {
		return nil, nil, false, nil, fmt.Errorf("get my permissions audit groups: %w", err)
	}
	groupNames := make([]string, 0)
	for groupRows.Next() {
		var name string
		if err := groupRows.Scan(&name); err != nil {
			_ = groupRows.Close()
			return nil, nil, false, nil, fmt.Errorf("scan group name: %w", err)
		}
		groupNames = append(groupNames, name)
	}
	if err := groupRows.Err(); err != nil {
		_ = groupRows.Close()
		return nil, nil, false, nil, fmt.Errorf("get my permissions audit groups: %w", err)
	}
	_ = groupRows.Close()

	roleNames, err := s.GetUserRbacRoleNames(ctx, userID)
	if err != nil {
		return nil, nil, false, nil, err
	}

	effective, err := s.LoadEffectiveRBAC(ctx, userID)
	if err != nil {
		return nil, nil, false, nil, err
	}

	permCatalog, err := s.ListRBACPermissions(ctx)
	if err != nil {
		return nil, nil, false, nil, err
	}

	grantRows, err := s.db.QueryContext(ctx,
		`SELECT p.name, g.name FROM rbac_permissions p
		 INNER JOIN rbac_role_permissions rp ON rp.permission_id = p.id
		 INNER JOIN rbac_group_roles gr ON gr.role_id = rp.role_id
		 INNER JOIN user_group_memberships ugm ON ugm.user_group_id = gr.user_group_id
		 INNER JOIN user_groups g ON g.id = ugm.user_group_id
		 WHERE ugm.user_account_id = ?
		 ORDER BY p.name, g.name`, userID,
	)
	if err != nil {
		return nil, nil, false, nil, fmt.Errorf("get my permissions audit grants: %w", err)
	}
	grantingByPerm := map[string][]string{}
	for grantRows.Next() {
		var permName, groupName string
		if err := grantRows.Scan(&permName, &groupName); err != nil {
			_ = grantRows.Close()
			return nil, nil, false, nil, fmt.Errorf("scan permission grant: %w", err)
		}
		grantingByPerm[permName] = append(grantingByPerm[permName], groupName)
	}
	if err := grantRows.Err(); err != nil {
		_ = grantRows.Close()
		return nil, nil, false, nil, fmt.Errorf("get my permissions audit grants: %w", err)
	}
	_ = grantRows.Close()

	auditRows := make([]MyPermissionAuditRow, 0, len(permCatalog))
	for _, p := range permCatalog {
		row := MyPermissionAuditRow{
			Permission: p.Name,
			Granted:    effective.Has(p.Name),
		}
		if effective.IsSuperuser {
			row.GrantingGroups = nil
		} else {
			row.GrantingGroups = grantingByPerm[p.Name]
			if row.GrantingGroups == nil {
				row.GrantingGroups = []string{}
			}
		}
		auditRows = append(auditRows, row)
	}

	return groupNames, roleNames, effective.IsSuperuser, auditRows, nil
}

func userGroupSelectSQL() string {
	return `SELECT g.id, g.name,
		(SELECT COUNT(*) FROM user_group_memberships m WHERE m.user_group_id = g.id) AS member_count,
		g.created_at, g.updated_at
		FROM user_groups g`
}

func scanUserGroup(row interface{ Scan(dest ...any) error }) (*UserGroupRow, error) {
	var g UserGroupRow
	if err := row.Scan(&g.ID, &g.Name, &g.MemberCount, &g.CreatedAt, &g.UpdatedAt); err != nil {
		return nil, err
	}
	return &g, nil
}

func isSystemUserGroup(name string) bool {
	return name == rbac.GroupEveryone || name == rbac.GroupAdministrators
}

func (s *SQLite) ListUserGroups(ctx context.Context) ([]UserGroupRow, error) {
	rows, err := s.db.QueryContext(ctx, userGroupSelectSQL()+` ORDER BY g.name`)
	if err != nil {
		return nil, fmt.Errorf("list user groups: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]UserGroupRow, 0)
	for rows.Next() {
		g, err := scanUserGroup(rows)
		if err != nil {
			return nil, fmt.Errorf("scan user group: %w", err)
		}
		out = append(out, *g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list user groups: %w", err)
	}
	return out, nil
}

func (s *SQLite) GetUserGroupByName(ctx context.Context, name string) (*UserGroupRow, error) {
	row := s.db.QueryRowContext(ctx, userGroupSelectSQL()+` WHERE g.name = ?`, name)
	g, err := scanUserGroup(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user group by name: %w", err)
	}
	return g, nil
}

func (s *SQLite) GetUserGroupByID(ctx context.Context, id int) (*UserGroupRow, error) {
	row := s.db.QueryRowContext(ctx, userGroupSelectSQL()+` WHERE g.id = ?`, id)
	g, err := scanUserGroup(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user group by id: %w", err)
	}
	return g, nil
}

func (s *SQLite) CreateUserGroup(ctx context.Context, name string) (int, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO user_groups (name) VALUES (?)`, name)
	if err != nil {
		return 0, fmt.Errorf("create user group: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("create user group id: %w", err)
	}
	return int(id), nil
}

func (s *SQLite) userGroupNameByID(ctx context.Context, id int) (string, error) {
	var name string
	if err := s.db.QueryRowContext(ctx, `SELECT name FROM user_groups WHERE id = ?`, id).Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", sql.ErrNoRows
		}
		return "", fmt.Errorf("user group name by id: %w", err)
	}
	return name, nil
}

func (s *SQLite) DeleteUserGroup(ctx context.Context, id int) error {
	name, err := s.userGroupNameByID(ctx, id)
	if err != nil {
		return err
	}
	if isSystemUserGroup(name) {
		return fmt.Errorf("cannot delete system group %q", name)
	}

	res, err := s.db.ExecContext(ctx, `DELETE FROM user_groups WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete user group: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete user group rows affected: %w", err)
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (s *SQLite) ListUserGroupMemberIDs(ctx context.Context, groupID int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT user_account_id FROM user_group_memberships WHERE user_group_id = ? ORDER BY user_account_id`, groupID,
	)
	if err != nil {
		return nil, fmt.Errorf("list user group member ids: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user group member id: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list user group member ids: %w", err)
	}
	return out, nil
}

func (s *SQLite) ListUserGroupIDsForUser(ctx context.Context, userID int) ([]int, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT user_group_id FROM user_group_memberships WHERE user_account_id = ? ORDER BY user_group_id`, userID,
	)
	if err != nil {
		return nil, fmt.Errorf("list user group ids for user: %w", err)
	}
	defer func() { _ = rows.Close() }()

	out := make([]int, 0)
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan user group id: %w", err)
		}
		out = append(out, id)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list user group ids for user: %w", err)
	}
	return out, nil
}

func (s *SQLite) SetUserGroupMembers(ctx context.Context, groupID int, userIDs []int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("set user group members begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM user_group_memberships WHERE user_group_id = ?`, groupID); err != nil {
		return fmt.Errorf("clear user group members: %w", err)
	}
	for _, userID := range userIDs {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO user_group_memberships (user_account_id, user_group_id) VALUES (?, ?)`,
			userID, groupID,
		); err != nil {
			return fmt.Errorf("insert user group member: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("set user group members commit: %w", err)
	}
	return s.ensureSuperuserCoverage(ctx)
}
