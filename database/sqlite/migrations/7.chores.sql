-- +migrate Up

CREATE TABLE chores (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  family_id INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  star_reward INTEGER NOT NULL DEFAULT 1 CHECK (star_reward > 0),
  weekday_mask INTEGER NOT NULL DEFAULT 127 CHECK (weekday_mask >= 0 AND weekday_mask <= 127),
  active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_chores_family_id ON chores(family_id);

CREATE TABLE chore_assignments (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  chore_id INTEGER NOT NULL REFERENCES chores(id) ON DELETE CASCADE,
  child_member_id INTEGER NOT NULL REFERENCES family_members(id) ON DELETE CASCADE,
  UNIQUE(chore_id, child_member_id)
);
CREATE INDEX idx_chore_assignments_chore_id ON chore_assignments(chore_id);
CREATE INDEX idx_chore_assignments_child_member_id ON chore_assignments(child_member_id);

CREATE TABLE chore_pauses (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  family_id INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  start_date TEXT NOT NULL,
  end_date TEXT NOT NULL,
  reason TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_chore_pauses_family_id ON chore_pauses(family_id);

CREATE TABLE chore_completions (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  assignment_id INTEGER NOT NULL REFERENCES chore_assignments(id) ON DELETE CASCADE,
  completion_date TEXT NOT NULL,
  ledger_entry_id INTEGER REFERENCES star_ledger_entries(id) ON DELETE SET NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(assignment_id, completion_date)
);
CREATE INDEX idx_chore_completions_assignment_id ON chore_completions(assignment_id);
CREATE INDEX idx_chore_completions_date ON chore_completions(completion_date);

INSERT OR IGNORE INTO rbac_permissions (name, description) VALUES
('chores.manage', 'Create, edit, and pause chores'),
('chores.complete', 'Mark chore completions and award stars'),
('chores.view_family', 'View weekly star chart for all children');

INSERT OR IGNORE INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM rbac_roles r CROSS JOIN rbac_permissions p
WHERE r.name = 'parent' AND p.name IN ('chores.manage', 'chores.complete', 'chores.view_family');

INSERT OR IGNORE INTO rbac_role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM rbac_roles r CROSS JOIN rbac_permissions p
WHERE r.name = 'child' AND p.name IN ('chores.view_family');

-- +migrate Down

DELETE FROM rbac_role_permissions WHERE permission_id IN (
  SELECT id FROM rbac_permissions WHERE name IN ('chores.manage', 'chores.complete', 'chores.view_family')
);
DELETE FROM rbac_permissions WHERE name IN ('chores.manage', 'chores.complete', 'chores.view_family');

DROP TABLE IF EXISTS chore_completions;
DROP TABLE IF EXISTS chore_pauses;
DROP TABLE IF EXISTS chore_assignments;
DROP TABLE IF EXISTS chores;
