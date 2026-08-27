-- +migrate Up

CREATE TABLE families (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE users (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  family_id INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  display_name TEXT NOT NULL,
  role TEXT NOT NULL CHECK (role IN ('parent', 'child')),
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_users_family_id ON users(family_id);

CREATE TABLE rewards (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  family_id INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  description TEXT NOT NULL DEFAULT '',
  cost_stars INTEGER NOT NULL CHECK (cost_stars > 0),
  active INTEGER NOT NULL DEFAULT 1,
  approval_required INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX idx_rewards_family_id ON rewards(family_id);

CREATE TABLE star_ledger_entries (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  family_id INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  child_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  amount INTEGER NOT NULL,
  entry_type TEXT NOT NULL CHECK (entry_type IN ('award', 'revoke', 'redeem')),
  note TEXT NOT NULL DEFAULT '',
  related_reward_id INTEGER REFERENCES rewards(id),
  created_by_user_id INTEGER REFERENCES users(id),
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_ledger_child ON star_ledger_entries(child_user_id);
CREATE INDEX idx_ledger_family ON star_ledger_entries(family_id);

CREATE TABLE redemptions (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  family_id INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  child_user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  reward_id INTEGER NOT NULL REFERENCES rewards(id),
  stars_spent INTEGER NOT NULL,
  status TEXT NOT NULL CHECK (status IN ('pending', 'approved', 'rejected')),
  ledger_entry_id INTEGER REFERENCES star_ledger_entries(id),
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  resolved_at TEXT,
  resolved_by_user_id INTEGER REFERENCES users(id)
);

CREATE INDEX idx_redemptions_family ON redemptions(family_id);
CREATE INDEX idx_redemptions_status ON redemptions(status);

-- +migrate Down

DROP TABLE IF EXISTS redemptions;
DROP TABLE IF EXISTS star_ledger_entries;
DROP TABLE IF EXISTS rewards;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS families;
