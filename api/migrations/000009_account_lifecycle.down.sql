DROP TABLE IF EXISTS account_action_tokens;
ALTER TABLE users DROP COLUMN IF EXISTS deactivated_at;
DROP TYPE IF EXISTS account_action;
