-- +migrate Up

ALTER TABLE family_members ADD COLUMN star_color TEXT;

-- +migrate Down

ALTER TABLE family_members DROP COLUMN star_color;
