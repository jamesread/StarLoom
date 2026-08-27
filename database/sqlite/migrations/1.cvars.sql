-- +migrate Up

CREATE TABLE cvars (
  cvar_key TEXT NOT NULL PRIMARY KEY,
  cvar_value_int INTEGER,
  cvar_value_string TEXT,
  cvar_main_type TEXT NOT NULL,
  cvar_title TEXT NOT NULL DEFAULT '',
  cvar_description TEXT NOT NULL DEFAULT '',
  cvar_category TEXT NOT NULL DEFAULT '',
  cvar_ordinal INTEGER NOT NULL DEFAULT 0
);

-- +migrate Down

DROP TABLE IF EXISTS cvars;
