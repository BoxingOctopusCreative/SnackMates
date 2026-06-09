ALTER TABLE friendships
    ADD COLUMN requester_notified_at TIMESTAMPTZ;

UPDATE friendships
SET requester_notified_at = updated_at
WHERE status = 'accepted';

CREATE INDEX idx_friendships_requester_unnotified
    ON friendships (requester_id)
    WHERE status = 'accepted' AND requester_notified_at IS NULL;
