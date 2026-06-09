CREATE TYPE account_action AS ENUM ('deactivate', 'delete', 'reactivate');

ALTER TABLE users
    ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMPTZ;

CREATE TABLE account_action_tokens (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action      account_action NOT NULL,
    token       TEXT NOT NULL UNIQUE,
    expires_at  TIMESTAMPTZ NOT NULL,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_account_action_tokens_user_id ON account_action_tokens(user_id);
CREATE INDEX idx_account_action_tokens_expires_at ON account_action_tokens(expires_at);
