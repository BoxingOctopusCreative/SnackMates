DROP INDEX IF EXISTS idx_friendships_requester_unnotified;

ALTER TABLE friendships
    DROP COLUMN IF EXISTS requester_notified_at;
