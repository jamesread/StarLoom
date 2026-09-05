-- +migrate Up

ALTER TABLE chores ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- +migrate Down

ALTER TABLE chores DROP COLUMN sort_order;
