-- +migrate Up

ALTER TABLE rewards ADD COLUMN availability_expression TEXT NOT NULL DEFAULT '';
