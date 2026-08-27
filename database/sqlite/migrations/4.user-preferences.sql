-- +migrate Up

CREATE TABLE user_preferences (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  updated_at TEXT NOT NULL DEFAULT (datetime('now')),
  user_account_id INTEGER NOT NULL UNIQUE REFERENCES user_accounts(id) ON DELETE CASCADE,
  language TEXT NOT NULL DEFAULT '',
  sidebar_enabled INTEGER NOT NULL DEFAULT 1,
  theme_toggle_enabled INTEGER NOT NULL DEFAULT 0
);

CREATE UNIQUE INDEX idx_user_preferences_user_account_id ON user_preferences(user_account_id);

-- +migrate Down

DROP TABLE IF EXISTS user_preferences;
