package store

import (
	"context"
	"database/sql"
)

const domainRBACSQL = `
INSERT OR IGNORE INTO rbac_permissions (name, description) VALUES
('family.view', 'View family name and member list'),
('family.manage', 'Create and update family'),
('members.manage', 'Add, edit, and remove child members'),
('members.avatar', 'Upload and remove child avatars'),
('stars.view_family', 'View all children balances and history'),
('stars.view_own', 'View own balance and history'),
('stars.award', 'Award stars to children'),
('stars.revoke', 'Revoke or correct star entries'),
('rewards.manage', 'Manage reward catalog'),
('rewards.view', 'Browse active rewards'),
('redemptions.approve', 'Approve or reject redemptions'),
('redemptions.request', 'Request reward redemptions'),
('chores.manage', 'Create, edit, and pause chores'),
('chores.complete', 'Mark chore completions and award stars'),
('chores.view_family', 'View weekly star chart for all children');

INSERT OR IGNORE INTO rbac_roles (name, description) VALUES
('parent', 'Family administrator'),
('child', 'Child account with limited access');

INSERT OR IGNORE INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM rbac_roles r CROSS JOIN rbac_permissions p
WHERE r.name = 'parent' AND p.name IN (
  'app.access', 'family.view', 'family.manage', 'members.manage', 'members.avatar',
  'stars.view_family', 'stars.view_own', 'stars.award', 'stars.revoke',
  'rewards.manage', 'rewards.view', 'redemptions.approve', 'redemptions.request',
  'chores.manage', 'chores.complete', 'chores.view_family'
);

INSERT OR IGNORE INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM rbac_roles r CROSS JOIN rbac_permissions p
WHERE r.name = 'child' AND p.name IN (
  'app.access', 'stars.view_own', 'rewards.view', 'redemptions.request', 'chores.view_family'
);

INSERT OR IGNORE INTO user_groups (name) VALUES ('Parents'), ('Children');

INSERT OR IGNORE INTO rbac_group_roles (user_group_id, role_id)
SELECT g.id, r.id FROM user_groups g CROSS JOIN rbac_roles r
WHERE g.name = 'Parents' AND r.name = 'parent';

INSERT OR IGNORE INTO rbac_group_roles (user_group_id, role_id)
SELECT g.id, r.id FROM user_groups g CROSS JOIN rbac_roles r
WHERE g.name = 'Children' AND r.name = 'child';
`

func seedDomainRBAC(db *sql.DB) error {
	_, err := db.Exec(domainRBACSQL)
	return err
}

func (s *SQLite) SeedDomainRBAC(_ context.Context) error {
	return seedDomainRBAC(s.db)
}

func (m *Memory) SeedDomainRBAC(_ context.Context) error {
	if m.SQLite == nil {
		return nil
	}
	return seedDomainRBAC(m.SQLite.DB())
}
