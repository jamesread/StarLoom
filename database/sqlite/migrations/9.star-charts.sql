-- +migrate Up

CREATE TABLE star_charts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  family_id INTEGER NOT NULL REFERENCES families(id) ON DELETE CASCADE,
  name TEXT NOT NULL,
  sort_order INTEGER NOT NULL DEFAULT 0,
  active INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT (datetime('now')),
  UNIQUE(family_id, name)
);
CREATE INDEX idx_star_charts_family_id ON star_charts(family_id);

INSERT INTO star_charts (family_id, name, sort_order)
SELECT id, 'Star Chart', 0 FROM families;

ALTER TABLE chores ADD COLUMN star_chart_id INTEGER REFERENCES star_charts(id);

UPDATE chores
SET star_chart_id = (
  SELECT sc.id FROM star_charts sc
  WHERE sc.family_id = chores.family_id
  ORDER BY sc.sort_order, sc.id
  LIMIT 1
);

CREATE INDEX idx_chores_star_chart_id ON chores(star_chart_id);

-- +migrate Down

DROP INDEX IF EXISTS idx_chores_star_chart_id;
ALTER TABLE chores DROP COLUMN star_chart_id;
DROP TABLE IF EXISTS star_charts;
