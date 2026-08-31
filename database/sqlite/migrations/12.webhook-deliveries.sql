-- +migrate Up

CREATE TABLE webhook_deliveries (
  id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
  webhook_target_id INTEGER REFERENCES webhook_targets(id) ON DELETE SET NULL,
  event TEXT NOT NULL,
  url TEXT NOT NULL,
  success INTEGER NOT NULL,
  http_status INTEGER NOT NULL DEFAULT 0,
  error_message TEXT NOT NULL DEFAULT '',
  fired_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX idx_webhook_deliveries_fired_at ON webhook_deliveries(fired_at DESC);

-- +migrate Down

DROP TABLE IF EXISTS webhook_deliveries;
