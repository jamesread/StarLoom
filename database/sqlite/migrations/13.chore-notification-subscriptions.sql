-- +migrate Up

CREATE TABLE chore_notification_subscriptions (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  family_id INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  subscriber_member_id INTEGER NOT NULL REFERENCES family_members(id) ON DELETE CASCADE,
  child_member_id INTEGER REFERENCES family_members(id) ON DELETE CASCADE,
  chore_id INTEGER REFERENCES chores(id) ON DELETE CASCADE,
  created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_chore_notify_sub_subscriber ON chore_notification_subscriptions(subscriber_member_id);
CREATE INDEX idx_chore_notify_sub_family ON chore_notification_subscriptions(family_id);

-- +migrate Down

DROP TABLE IF EXISTS chore_notification_subscriptions;
