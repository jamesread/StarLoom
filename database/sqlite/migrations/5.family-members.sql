-- +migrate Up

ALTER TABLE users RENAME TO family_members;

ALTER TABLE family_members ADD COLUMN user_account_id INTEGER REFERENCES user_accounts(id) ON DELETE SET NULL;
ALTER TABLE family_members ADD COLUMN avatar_path TEXT;

DROP INDEX IF EXISTS idx_users_family_id;
CREATE INDEX idx_family_members_family_id ON family_members(family_id);
CREATE INDEX idx_family_members_user_account_id ON family_members(user_account_id);

ALTER TABLE star_ledger_entries RENAME COLUMN child_user_id TO child_member_id;
ALTER TABLE star_ledger_entries RENAME COLUMN created_by_user_id TO created_by_member_id;

ALTER TABLE redemptions RENAME COLUMN child_user_id TO child_member_id;
ALTER TABLE redemptions RENAME COLUMN resolved_by_user_id TO resolved_by_member_id;
ALTER TABLE redemptions ADD COLUMN fulfilled_at TEXT;

-- +migrate Down

ALTER TABLE redemptions DROP COLUMN fulfilled_at;
ALTER TABLE redemptions RENAME COLUMN resolved_by_member_id TO resolved_by_user_id;
ALTER TABLE redemptions RENAME COLUMN child_member_id TO child_user_id;

ALTER TABLE star_ledger_entries RENAME COLUMN created_by_member_id TO created_by_user_id;
ALTER TABLE star_ledger_entries RENAME COLUMN child_member_id TO child_user_id;

DROP INDEX IF EXISTS idx_family_members_user_account_id;
DROP INDEX IF EXISTS idx_family_members_family_id;
CREATE INDEX idx_users_family_id ON family_members(family_id);

ALTER TABLE family_members DROP COLUMN avatar_path;
ALTER TABLE family_members DROP COLUMN user_account_id;

ALTER TABLE family_members RENAME TO users;
