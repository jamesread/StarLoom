-- +migrate Up

ALTER TABLE rewards ADD COLUMN multi_spend INTEGER NOT NULL DEFAULT 0;

-- +migrate Down

ALTER TABLE rewards DROP COLUMN multi_spend;
