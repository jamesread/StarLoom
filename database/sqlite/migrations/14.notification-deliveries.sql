-- +migrate Up

CREATE TABLE notification_deliveries (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  family_id INTEGER REFERENCES families(id) ON DELETE SET NULL,
  recipient_member_id INTEGER REFERENCES family_members(id) ON DELETE SET NULL,
  notification_type TEXT NOT NULL,
  title TEXT NOT NULL DEFAULT '',
  success INTEGER NOT NULL,
  error_message TEXT NOT NULL DEFAULT '',
  sent_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_notification_deliveries_sent_at ON notification_deliveries(sent_at DESC);

-- +migrate Down

DROP TABLE IF EXISTS notification_deliveries;
