-- +migrate Up

CREATE TABLE webhook_targets (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  url TEXT NOT NULL,
  secret TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  created TEXT NOT NULL DEFAULT (datetime('now')),
  updated TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE webhook_events (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  webhook_target_id INTEGER NOT NULL REFERENCES webhook_targets(id) ON DELETE CASCADE,
  event TEXT NOT NULL,
  UNIQUE (webhook_target_id, event)
);

CREATE INDEX idx_webhook_events_event ON webhook_events(event);

-- +migrate Down

DROP TABLE IF EXISTS webhook_events;
DROP TABLE IF EXISTS webhook_targets;
