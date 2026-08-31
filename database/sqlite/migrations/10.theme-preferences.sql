-- +migrate Up

ALTER TABLE user_preferences DROP COLUMN theme_toggle_enabled;

-- +migrate Down

ALTER TABLE user_preferences
  ADD COLUMN theme_toggle_enabled INTEGER NOT NULL DEFAULT 0;
